//go:build liftrs_cpp && cgo

package liftrs

import (
	"math/rand"
	"reflect"
	"strings"
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

// Large-field arithmetic is validated on small buffers before any large allocation.
func TestScalingCPPLargeFields(t *testing.T) {
	for _, s := range []int{14, 15} {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		basis := []uint32{3, 5, 1 << uint(s), 9}
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
		offset := uint32((1 << uint(2*s-1)) + 1)
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
				t.Fatal("large field Horner", i)
			}
		}
		if _, err := p.ApplyAt(got, offset, true); err != nil || !reflect.DeepEqual(got, input) {
			t.Fatal("large field inverse", err)
		}
	}

}

func TestObservedPreparationWorkspace(t *testing.T) {
	b, c, err := newScalingPreparationCPP(Spec{6, 2, 4}, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	b, err = discoverOrbitBasis(b, OrbitOptions{Random: rand.New(rand.NewSource(1802)), MaxSamples: 24})
	if err != nil {
		t.Fatal(err)
	}
	b.base.products.cutoff = 0
	active, complete := 0, false
	e, err := b.prepareCyclicScalingObserved(b.extensionFFT, func(stage string, bytes uint64) {
		if strings.HasPrefix(stage, "active-") {
			active++
			if bytes == 0 {
				t.Error("active scratch missing")
			}
		}
		if stage == "complete" {
			complete = true
			if bytes != 0 {
				t.Error("scratch not released")
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if active == 0 || !complete || e.basis.base.products.workspace != nil || e.basis.extension.products.workspace != nil {
		t.Fatal("preparation workspace lifetime", active, complete)
	}
}
