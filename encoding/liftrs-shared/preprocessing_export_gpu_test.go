//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

func TestGPUPreprocessingExport(t *testing.T) {
	spec := Spec{4, 2, 3}
	reference, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(171001))})
	if err != nil {
		t.Fatal(err)
	}
	encoder, err := reference.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := encoder.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	var expected bytes.Buffer
	if err = WriteScalingOnline(&expected, legacy); err != nil {
		t.Fatal(err)
	}
	options := preprocessingOptions(4, 2)
	options.Backend = "cuda"
	options.GPUDevices = gpuPreparationDevices(t)
	options.GPUPath = os.Getenv("PIR_GPU_PREPARATION_LIBRARY")
	options.GPUBytes = gpuPreparationCap(t)
	session, err := NewCodePreprocessor(spec, options)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	support, err := session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(171001))})
	if err != nil {
		t.Fatal(err)
	}
	var persisted bytes.Buffer
	if _, err = support.Write(&persisted); err != nil {
		t.Fatal(err)
	}
	indices := append([]int(nil), support.exponents...)
	var root *RootCheckpoint
	poly, err := session.PreparePolynomials(support, PreparationReuse{SaveRoot: func(r *RootCheckpoint) error { root = r; return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(poly.p, encoder.p) || !reflect.DeepEqual(poly.generator, encoder.g) {
		t.Fatal("independent P/G identity")
	}
	var before bytes.Buffer
	root.Write(&before)
	for _, tile := range []int{8, 24, 1 << 20} {
		var actual bytes.Buffer
		if err = session.WriteCode(&actual, poly, tile); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
			t.Fatal("canonical stream vs independent Go", tile, err)
		}
		if _, err = VerifyScalingCode(bytes.NewReader(actual.Bytes()), codeDigest(actual.Bytes()), ScalingCodeLimits{256, 1 << 20}); err != nil {
			t.Fatal(err)
		}
	}
	sentinel := errors.New("export sink failed")
	if err = session.WriteCode(codeFailWriter{sentinel}, poly, 24); !errors.Is(err, sentinel) {
		t.Fatal("writer failure", err)
	}
	if err = session.WriteCode(spectrumShortWriter{}, poly, 24); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal("short writer", err)
	}
	replay, err := session.PreparePolynomials(support, PreparationReuse{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	var actual, after bytes.Buffer
	root.Write(&after)
	if !bytes.Equal(before.Bytes(), after.Bytes()) || !reflect.DeepEqual(support.exponents, indices) {
		t.Fatal("caller support/root mutated")
	}
	if err = session.WriteCode(&actual, replay, 24); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
		t.Fatal("restart/reuse after write error", err)
	}
	if err = session.WriteCode(io.Discard, poly, 7); err == nil {
		t.Fatal("small tile cap")
	}
	if err = session.WriteCode(nil, poly, 24); err == nil {
		t.Fatal("nil writer")
	}
	other, err := NewCodePreprocessor(Spec{3, 2, 2}, preprocessingOptions(3, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	actual.Reset()
	if err = other.WriteCode(&actual, poly, 24); err == nil || actual.Len() != 0 {
		t.Fatal("wrong code session wrote output")
	}
	session.Close()
	if err = session.WriteCode(io.Discard, poly, 24); err == nil {
		t.Fatal("closed export")
	}
}
