//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

func TestRootTransferProofAgreement(t *testing.T) {
	values := make([]uint32, 131073)
	for j := range values {
		values[j] = uint32(j) & ((1 << 17) - 1)
	}
	values[len(values)-1] = 0x1abcd
	expected := sha256.New()
	if err := binary.Write(expected, binary.LittleEndian, values); err != nil {
		t.Fatal(err)
	}
	fingerprint := rootTransferFingerprint(values)
	var want [32]byte
	copy(want[:], expected.Sum(nil))
	if fingerprint != want {
		t.Fatal("canonical fingerprint lost a boundary/tail word")
	}

	a := rootTransferFingerprint([]uint32{1, 3, 7})
	b := rootTransferFingerprint([]uint32{1, 3, 8})
	proof := func(x, y [32]byte) (p [64]byte) { copy(p[:32], x[:]); copy(p[32:], y[:]); return }
	stable := proof(a, a)
	if rootProofMismatch(stable, stable) {
		t.Fatal("equal data rejected")
	}
	for _, other := range [][64]byte{proof(b, b), proof(a, b), proof(b, a)} {
		if !rootProofMismatch(stable, other) || !rootProofMismatch(other, stable) {
			t.Fatal("disagreement not symmetric")
		}
	}
}
