package liftrs

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestBatchedFactorProducts(t *testing.T) {
	base, _, _, _ := orbitFields(4, 0x13)
	for _, batch := range []int{1, 2, 3, 5, 16, 32} {
		for _, count := range []int{0, 1, 2, 3, 5, 31, 32, 33, 65} {
			builder := newFactorProductBuilder(base, batch)
			var factors [][]uint32
			for i := 0; i < count; i++ {
				if i%3 == 0 {
					factor := []uint32{uint32(i % 16), 1}
					factors = append(factors, factor)
					builder.add(factor[0], 1, 0, 1)
				} else {
					factor := []uint32{uint32((i * 7) % 16), uint32((i * 3) % 16), 1}
					factors = append(factors, factor)
					builder.add(factor[0], factor[1], 1, 2)
				}
			}
			got, want := builder.finish(), rootProduct(base, factors)
			if !reflect.DeepEqual(got, want) {
				t.Fatal("batch/singleton/mixed-degree product", batch, count)
			}
		}
	}
}
func TestCompactSupportValidation(t *testing.T) {
	basis, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(1801))})
	if err != nil {
		t.Fatal(err)
	}
	if err := basis.validateSupport(); err != nil {
		t.Fatal(err)
	}
	cases := [][]int{nil, {-1}, {basis.Length() - 1}, {0, 0}, {1, 0}, {1}}
	for _, support := range cases {
		copyBasis := *basis
		copyBasis.support = support
		if copyBasis.validateSupport() == nil {
			t.Fatal("accepted invalid support", support)
		}
	}
	set := newExponentSet(130)
	for _, i := range []int{0, 63, 64, 65, 129} {
		set.add(i)
	}
	for i := 0; i < 130; i++ {
		want := i == 0 || i == 63 || i == 64 || i == 65 || i == 129
		if set.has(i) != want {
			t.Fatal("bit boundary", i)
		}
	}
}
func TestPreparationProductWorkspace(t *testing.T) {
	base, ext, _, _ := orbitFields(4, 0x13)
	transform := newSparseTransform(ext)
	for _, field := range []orbitField{base, ext} {
		reference := field
		pool := &productWorkspace{}
		field.products = &polynomialProducts{transform: transform, cutoff: 0, workspace: pool}
		for _, lengths := range [][2]int{{7, 5}, {32, 32}, {41, 13}, {97, 73}, {8, 11}, {145, 151}} {
			a, b := make([]uint32, lengths[0]), make([]uint32, lengths[1])
			for i := range a {
				a[i] = uint32((i*7 + 3) % int(field.q))
			}
			for i := range b {
				b[i] = uint32((i*5 + 1) % int(field.q))
			}
			for _, limit := range []int{3, 31, len(a) + len(b) - 1} {
				got, want := polyMulLow(field, a, b, limit), polyMulLow(reference, a, b, limit)
				if !reflect.DeepEqual(got, want) {
					t.Fatal("pooled convolution", field.q, lengths, limit)
				}
				if pool.busy {
					t.Fatal("workspace lease retained")
				}
			}
		}
		pool.release()
		if pool.bytes() != 0 {
			t.Fatal("released scratch retained")
		}
	}
	// Reciprocal reuse must keep zero padding when multiplication returns fewer
	// coefficients than the requested precision, including constant inputs.
	for _, a := range [][]uint32{{1}, {2}, {1, 1}, {3, 5, 7}} {
		g := polyInverse(base, a, 97)
		product := polyMulLow(base, a, g, 97)
		if len(g) != 97 || product[0] != 1 {
			t.Fatal("reciprocal shape")
		}
		for i := 1; i < len(product); i++ {
			if product[i] != 0 {
				t.Fatal("reciprocal identity", a, i)
			}
		}
	}
}
