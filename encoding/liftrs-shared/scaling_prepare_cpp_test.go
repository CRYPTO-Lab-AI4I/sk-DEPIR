//go:build liftrs_cpp && cgo

package liftrs

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestScalingPreparationCPP(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}} {
		b, c, err := newScalingPreparationCPP(spec, 4)
		if err != nil {
			t.Fatal(err)
		}
		b, err = discoverOrbitBasis(b, OrbitOptions{Random: rand.New(rand.NewSource(1601)), MaxSamples: 2 + (128+spec.FieldDegree-1)/spec.FieldDegree})
		if err != nil {
			t.Fatal(err)
		}
		ref, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(1601))})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(b.support, ref.support) {
			t.Fatal("native discovery differs")
		}
		e, err := b.prepareCyclicScaling(b.extensionFFT)
		if err != nil {
			t.Fatal(err)
		}
		original, err := ref.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(e.p, original.p) || !reflect.DeepEqual(e.g, original.g) || !reflect.DeepEqual(e.spectrum, original.spectrum) {
			t.Fatal("prepared constants differ")
		}
		c.close()
	}
}

// s=14 arithmetic is validated on small buffers before any large allocation.
func TestScalingCPPField14(t *testing.T) {
	polynomial, _ := BinaryPolynomial(14)
	_, f, _, _ := orbitFields(14, polynomial)
	basis := []uint32{3, 5, 1 << 14, 9}
	a, err := orderedTransform(f, basis)
	if err != nil {
		t.Fatal(err)
	}
	p, err := orderedCPP(a, basis)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	input := make([]uint32, 16)
	random := rand.New(rand.NewSource(1602))
	for i := range input {
		input[i] = uint32(random.Intn(int(f.q)))
	}
	got := append([]uint32(nil), input...)
	offset := uint32((1 << 27) + 1)
	if _, err := p.ApplyAt(got, offset, false); err != nil {
		t.Fatal(err)
	}
	for i, x := range got {
		point := offset
		for j, b := range basis {
			if i&(1<<j) != 0 {
				point ^= b
			}
		}
		var want uint32
		for j := len(input) - 1; j >= 0; j-- {
			want = f.mul(want, point) ^ input[j]
		}
		if x != want {
			t.Fatal("field14 Horner", i)
		}
	}
	if _, err := p.ApplyAt(got, offset, true); err != nil || !reflect.DeepEqual(got, input) {
		t.Fatal("field14 inverse", err)
	}
}
