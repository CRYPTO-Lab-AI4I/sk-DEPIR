//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestPreprocessingRecoveryBoundaries(t *testing.T) {
	spec := Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 3}
	limits := PreprocessingLimits{MaxElements: 256, MaxHostArrayBytes: 1 << 20, NativeBytes: 8 << 20}
	recovery := &PreprocessingRecovery{}
	session, err := NewCodePreprocessor(spec, PreprocessingOptions{Backend: "cpp", Workers: 2, Limits: limits, Recovery: recovery})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	stop := errors.New("test safe stop")
	var saved *PreprocessingCheckpoint
	var raw []byte
	var record ProgressRecord
	capture := func(c *PreprocessingCheckpoint) error {
		var payload bytes.Buffer
		var e error
		record, e = c.Write(&payload)
		if e != nil {
			return e
		}
		raw = append([]byte(nil), payload.Bytes()...)
		saved, e = ReadPreprocessingCheckpoint(bytes.NewReader(raw), record, limits)
		if e != nil {
			return e
		}
		return stop
	}
	recovery.Save = capture
	if _, err = session.Discover(OrbitOptions{}); !errors.Is(err, stop) {
		t.Fatal("discovery stop", err)
	}
	if saved == nil || saved.Record().Kind != "discovery" {
		t.Fatal("no discovery checkpoint")
	}
	damaged := append([]byte(nil), raw...)
	damaged[0] ^= 1
	if _, err = ReadPreprocessingCheckpoint(bytes.NewReader(damaged), record, limits); err == nil {
		t.Fatal("damaged recovery accepted")
	}
	if _, err = ReadPreprocessingCheckpoint(bytes.NewReader(append(raw, 0)), record, limits); err == nil {
		t.Fatal("trailing recovery bytes accepted")
	}
	wrong := record
	wrong.Bytes++
	if _, err = ReadPreprocessingCheckpoint(bytes.NewReader(raw), wrong, limits); err == nil {
		t.Fatal("bad size accepted")
	}
	recovery.Discovery = saved
	recovery.Save = nil
	support, err := session.Discover(OrbitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = support.Write(io.Discard); err != nil {
		t.Fatal(err)
	}
	recovery.Preparation = nil
	recovery.Save = capture
	if _, err = session.Prepare(support, PreparationReuse{}); !errors.Is(err, stop) {
		t.Fatal("root stop", err)
	}
	if saved.Record().Kind != "roots" {
		t.Fatal("no root frontier")
	}
	rootSnapshot := saved
	before := append([]byte(nil), raw...)
	var root *RootCheckpoint
	reuse := PreparationReuse{SaveRoot: func(c *RootCheckpoint) error { root = c; return nil }}
	recovery.Preparation = rootSnapshot
	recovery.Save = nil
	first, err := session.Prepare(support, reuse)
	if err != nil {
		t.Fatal("root resume", err)
	}
	var unchanged bytes.Buffer
	if _, err = rootSnapshot.Write(&unchanged); err != nil || !bytes.Equal(before, unchanged.Bytes()) {
		t.Fatal("loaded frontier mutated", err)
	}
	for _, target := range []uint64{8, support.Parameters().N - support.Parameters().K} {
		recovery.Preparation = nil
		recovery.Save = func(c *PreprocessingCheckpoint) error {
			if c.Record().Kind == "reciprocal" && c.Record().Next == target {
				return capture(c)
			}
			return nil
		}
		if _, err = session.Prepare(support, PreparationReuse{Root: root}); !errors.Is(err, stop) {
			t.Fatal("reciprocal stop", target, err)
		}
		recovery.Preparation = saved
		recovery.Save = nil
		resumed, e := session.Prepare(support, PreparationReuse{Root: root})
		if e != nil {
			t.Fatal("reciprocal resume", target, e)
		}
		var a, b bytes.Buffer
		if e = WriteScalingCode(&a, first.CodeData()); e != nil {
			t.Fatal(e)
		}
		if e = WriteScalingCode(&b, resumed.CodeData()); e != nil || !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Fatal("different resumed code", e)
		}
	}

	baseField, e := NewBinaryField(spec.FieldDegree, support.Parameters().Polynomial)
	if e != nil {
		t.Fatal(e)
	}
	base := orbitField{q: 16, base: baseField, lookup: newPreparationProducts(baseField)}
	builder := newFactorProductBuilder(base, 3)
	factors := make([][4]uint32, len(support.exponents))
	if e = session.backend.RootFactorsInto(support.exponents, factors); e != nil {
		t.Fatal(e)
	}
	for j, index := range support.exponents {
		partner := int(uint64(index) * 16 % 255)
		if partner < index {
			continue
		}
		factor := factors[j]
		builder.add(factor[0], factor[1], factor[2], int(factor[3]))
	}
	sections := [][]uint32{builder.storage[:3*len(builder.factors)]}
	sections = append(sections, builder.levels...)
	frontier := &PreprocessingCheckpoint{record: ProgressRecord{Kind: "roots", Parameters: support.Parameters(), SupportSHA256: support.Checksum(), Next: support.Parameters().K}, sections: sections}
	if e = capture(frontier); !errors.Is(e, stop) {
		t.Fatal(e)
	}
	recovery.Preparation = saved
	recovery.Save = nil
	rebuilt, e := session.Prepare(support, PreparationReuse{})
	if e != nil {
		t.Fatal("multi-level frontier", e)
	}
	var reference, rebuiltBytes bytes.Buffer
	if e = WriteScalingCode(&reference, first.CodeData()); e != nil {
		t.Fatal(e)
	}
	if e = WriteScalingCode(&rebuiltBytes, rebuilt.CodeData()); e != nil || !bytes.Equal(reference.Bytes(), rebuiltBytes.Bytes()) {
		t.Fatal("multi-level frontier mismatch", e)
	}

	oracle, e := DiscoverOrbit(spec, OrbitOptions{})
	if e != nil {
		t.Fatal(e)
	}
	encoder, e := oracle.PrepareCyclicFFT()
	if e != nil {
		t.Fatal(e)
	}
	plan, e := encoder.ScalingPlan()
	if e != nil {
		t.Fatal(e)
	}
	var expected, actual bytes.Buffer
	if e = WriteScalingOnline(&expected, plan); e != nil {
		t.Fatal(e)
	}
	if e = WriteScalingCode(&actual, first.CodeData()); e != nil || !bytes.Equal(expected.Bytes(), actual.Bytes()) {
		t.Fatal("independent Go canonical mismatch", e)
	}
	// A valid digest with the wrong parent must still fail before reuse.
	bad := *rootSnapshot
	bad.record.SupportSHA256 = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	recovery.Preparation = &bad
	if _, e = session.Prepare(support, PreparationReuse{}); e == nil {
		t.Fatal("wrong parent accepted")
	}
}
