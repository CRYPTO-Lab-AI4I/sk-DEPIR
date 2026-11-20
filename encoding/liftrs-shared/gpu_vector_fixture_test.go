package liftrs

import (
	"encoding/binary"
	"os"
	"testing"
)

// The stage probe needs only canonical field metadata, not a domain-sized cache.
// CPU polynomial arithmetic supplies independent transform outputs in the probe.
func TestGPUVectorStageFixtures(t *testing.T) {
	path := os.Getenv("PIR_GPU_VECTOR_FIELDS")
	if path == "" {
		t.Skip("explicit GPU vector-stage metadata fixture")
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	put := func(value any) {
		t.Helper()
		if err := binary.Write(out, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
	}
	put(uint32(2))
	for _, s := range []int{4, 17} {
		h, _, err := canonicalScalingCodeHeader(s, 3, 1, 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		field, err := NewBinaryField(s, h.Polynomial)
		if err != nil {
			t.Fatal(err)
		}
		tables := newPreparationProducts(field)
		for _, value := range []any{uint32(s), h.Polynomial, h.Delta, h.Omega, uint32(len(tables.log)), tables.log, uint32(len(tables.powers)), tables.powers} {
			put(value)
		}
	}
	if err := out.Sync(); err != nil {
		t.Fatal(err)
	}
}
