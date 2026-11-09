//go:build liftrs_cpp && cgo

package liftrs

import (
	native "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"reflect"
	"testing"
)

func TestCPPCountBoundary(t *testing.T) {
	for _, extension := range []bool{false, true} {
		probe, err := SparseTransformProbe(16, extension)
		if err != nil || (extension && probe[0] != 3) || (!extension && probe[0] != 1) {
			t.Fatal("probe did not select the required count width", extension, err)
		}
	}
	if _, err := SparseTransformProbe(17, true); err == nil {
		t.Fatal("public probe accepted values wider than uint32")
	}
	polynomial, _ := BinaryPolynomial(16)
	_, f, _, _ := orbitFields(16, polynomial)
	a := newSparseTransform(f)
	words := cppDescription(a)
	if words[0] != 3 || words[4] != 0 || words[5] != 1 {
		t.Fatal("wide native header")
	}
	p, err := native.NewPlan(words)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	input := []uint32{0xffffffff, 1, 0x80000000, 17, 0xffff0000, 2, 3, 4}
	got := append([]uint32(nil), input...)
	const offset = uint32(0xffffffff)
	if _, err := p.ApplyAt(got, offset, false); err != nil {
		t.Fatal(err)
	}
	for i, value := range got {
		var want uint32
		for j := len(input) - 1; j >= 0; j-- {
			want = f.mul(want, offset^uint32(i)) ^ input[j]
		}
		if value != want {
			t.Fatal("native full-width Horner", i)
		}
	}
	if _, err := p.ApplyAt(got, offset, true); err != nil || !reflect.DeepEqual(got, input) {
		t.Fatal("native full-width inverse", err)
	}
	bad := append([]uint32(nil), words...)
	bad[5] = 1 << 16 // A 48-bit domain cannot use 32-bit coordinate buffers.
	if invalid, err := native.NewPlan(bad); err == nil {
		invalid.Close()
		t.Fatal("oversized native domain accepted")
	}
	bad = words[:6]
	if invalid, err := native.NewPlan(bad); err == nil {
		invalid.Close()
		t.Fatal("short wide header accepted")
	}
	// Preparing the 32-level matrix description must remain small; no N-sized arrays.
	matrix, err := newMatrixCPP(f, 32)
	if err != nil {
		t.Fatal(err)
	}
	matrix.close()
}
