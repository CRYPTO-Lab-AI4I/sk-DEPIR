package liftrs

import (
	"encoding/binary"
	"os"
	"testing"
)

// The server communication probe consumes canonical field metadata only; no
// GPU or MPI dependency enters ordinary builds or this opt-in fixture builder.
func TestGPUTransportFieldFixture(t *testing.T) {
	path := os.Getenv("PIR_GPU_TRANSPORT_FIELD")
	if path == "" {
		t.Skip("explicit GPU transport metadata fixture")
	}
	h, _, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	f, err := NewBinaryField(17, h.Polynomial)
	if err != nil {
		t.Fatal(err)
	}
	tables := newPreparationProducts(f)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	// Local bounded fixture, not a public cache schema. Counts precede arrays.
	for _, value := range []any{uint32(0x534b4754), uint32(17), h.Polynomial, h.Delta,
		uint32(len(tables.log)), tables.log, uint32(len(tables.powers)), tables.powers} {
		if err = binary.Write(out, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
}
