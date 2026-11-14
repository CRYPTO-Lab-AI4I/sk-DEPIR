package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"reflect"
	"testing"
)

func preprocessingArtifactFixture(t *testing.T) (*CodeSupport, *CyclicEncoder) {
	t.Helper()
	basis, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(170066))})
	if err != nil {
		t.Fatal(err)
	}
	code, err := basis.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := canonicalScalingCodeHeader(4, 2, uint64(basis.Dimension()), uint64(basis.Length()))
	if err != nil {
		t.Fatal(err)
	}
	return &CodeSupport{parameters: SupportParameters{h.Spec, h.K, h.N, h.Polynomial, h.Delta, h.Omega, basis.Samples()}, exponents: append([]int(nil), basis.support...)}, code
}
func TestPreprocessingArtifacts(t *testing.T) {
	support, code := preprocessingArtifactFixture(t)
	var payload bytes.Buffer
	record, err := support.Write(&payload)
	if err != nil {
		t.Fatal(err)
	}
	limits := PreprocessingLimits{MaxElements: 256, MaxHostArrayBytes: 1 << 20}
	loaded, err := ReadCodeSupport(bytes.NewReader(payload.Bytes()), record, limits)
	if err != nil || !reflect.DeepEqual(loaded.exponents, support.exponents) || loaded.Checksum() != record.SHA256 || record.ElementBytes != 4 {
		t.Fatal("support persistence", err)
	}
	checkpoint := &RootCheckpoint{parameters: loaded.parameters, supportSHA: loaded.Checksum(), values: code.p}
	var root bytes.Buffer
	rootRecord, err := checkpoint.Write(&root)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := ReadPreparationRoot(bytes.NewReader(root.Bytes()), loaded, rootRecord, limits)
	if err != nil || !reflect.DeepEqual(saved.values, code.p) || saved.matches(loaded) != nil {
		t.Fatal("root persistence", err)
	}
	// Discovery sample count is provenance, not a change to the mathematical code.
	copySupport := &CodeSupport{parameters: loaded.parameters, exponents: loaded.exponents, checksum: loaded.Checksum()}
	copySupport.parameters.Samples++
	if err = saved.matches(copySupport); err != nil {
		t.Fatal("sampling provenance changed code identity", err)
	}
	for _, raw := range [][]byte{payload.Bytes()[:payload.Len()-1], append(append([]byte(nil), payload.Bytes()...), 0)} {
		if _, err = ReadCodeSupport(bytes.NewReader(raw), record, limits); err == nil {
			t.Fatal("support EOF")
		}
	}
	bad := record
	bad.ElementBytes = 8
	if _, err = ReadCodeSupport(bytes.NewReader(payload.Bytes()), bad, limits); err == nil {
		t.Fatal("support width")
	}
	bad = record
	bad.SHA256 = hex.EncodeToString(make([]byte, 32))
	if _, err = ReadCodeSupport(bytes.NewReader(payload.Bytes()), bad, limits); err == nil {
		t.Fatal("support digest")
	}
	bad = record
	bad.Parameters.Omega = 1
	if _, err = ReadCodeSupport(bytes.NewReader(payload.Bytes()), bad, limits); err == nil {
		t.Fatal("support generator")
	}
	bad = record
	bad.Parameters.K = 1 << 30
	if _, err = ReadCodeSupport(bytes.NewReader(nil), bad, limits); err == nil {
		t.Fatal("support cap")
	}
	raw := append([]byte(nil), payload.Bytes()...)
	copy(raw[4:8], raw[:4])
	bad = record
	digest := sha256.Sum256(raw)
	bad.SHA256 = hex.EncodeToString(digest[:])
	if _, err = ReadCodeSupport(bytes.NewReader(raw), bad, limits); err == nil {
		t.Fatal("duplicate support")
	}
	corrupt := rootRecord
	corrupt.SupportSHA256 = hex.EncodeToString(make([]byte, 32))
	if _, err = ReadPreparationRoot(bytes.NewReader(root.Bytes()), loaded, corrupt, limits); err == nil {
		t.Fatal("root binding")
	}
	raw = append([]byte(nil), root.Bytes()...)
	binary.LittleEndian.PutUint32(raw, 0)
	digest = sha256.Sum256(raw)
	corrupt = rootRecord
	corrupt.SHA256 = hex.EncodeToString(digest[:])
	if _, err = ReadPreparationRoot(bytes.NewReader(raw), loaded, corrupt, limits); err == nil {
		t.Fatal("root normalization")
	}
	if _, err = ReadPreparationRoot(bytes.NewReader(root.Bytes()[:root.Len()-1]), loaded, rootRecord, limits); err == nil {
		t.Fatal("root EOF")
	}
	if _, err = support.Write(codeFailWriter{err: io.ErrClosedPipe}); err == nil {
		t.Fatal("support writer failure")
	}
	if _, err = checkpoint.Write(codeFailWriter{err: io.ErrClosedPipe}); err == nil {
		t.Fatal("root writer failure")
	}
}

func TestPreprocessingWideSupportEncoding(t *testing.T) {
	n := uint64(1) << 34
	h, _, err := canonicalScalingCodeHeader(17, 5, 1, n)
	if err != nil {
		t.Fatal(err)
	}
	source := &CodeSupport{parameters: SupportParameters{h.Spec, 3, n, h.Polynomial, h.Delta, h.Omega, 0}, exponents: []int{0, 1 << 32, 1<<32 + 1}}
	var raw bytes.Buffer
	record, err := source.Write(&raw)
	if err != nil || record.ElementBytes != 8 || raw.Len() != 24 || binary.LittleEndian.Uint64(raw.Bytes()[8:]) != 1<<32 {
		t.Fatal("wide support slots", err)
	}
	// Metadata/slot witness only: no genuine s17 support or full domain bitset.
	if _, err = ReadCodeSupport(bytes.NewReader(raw.Bytes()), record, PreprocessingLimits{MaxElements: n, MaxHostArrayBytes: 1 << 20}); err == nil {
		t.Fatal("wide support bitset budget")
	}
	if _, err = source.Write(io.Discard); err != nil {
		t.Fatal(err)
	}
}
