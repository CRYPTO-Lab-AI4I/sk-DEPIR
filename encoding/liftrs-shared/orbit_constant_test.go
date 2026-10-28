package liftrs

import (
	"math/rand"
	"testing"
)

// Compare table and fallback paths with independent polynomial arithmetic,
// including partial final nibbles and budget boundaries.
func TestConstantProducts(t *testing.T) {
	r := rand.New(rand.NewSource(8081))
	for s := 2; s <= 20; s++ {
		p, _ := BinaryPolynomial(s)
		base, err := NewBinaryField(s, p)
		if err != nil {
			t.Fatal(err)
		}
		q := uint32(1) << uint(s)
		fields := []orbitField{{q: q, base: base}}
		if s <= 13 {
			tower, err := NewQuadraticField(base)
			if err != nil {
				t.Fatal(err)
			}
			fields = append(fields, orbitField{q: q * q, base: base, extension: &tower})
		}
		for _, f := range fields {
			coefficients := []uint32{0, 1, f.q - 1, r.Uint32() & (f.q - 1)}
			for _, c := range coefficients {
				for _, budget := range []int{0, 63, 64, 4096} {
					product := newConstantProduct(f, c, budget)
					if len(product.rows)*4 > budget {
						t.Fatal("table budget exceeded")
					}
					for i := 0; i < 1000; i++ {
						x := r.Uint32() & (f.q - 1)
						var want uint32
						if f.extension == nil {
							want = binaryOracleMul(c, x, s, p)
						} else {
							y := quadraticOracleMul(*f.extension, QuadraticElement{A: c & (q - 1), B: c >> uint(s)}, QuadraticElement{A: x & (q - 1), B: x >> uint(s)})
							want = y.A | y.B<<uint(s)
						}
						if got := product.mul(f, x); got != want {
							t.Fatalf("s=%d c=%d x=%d budget=%d: %d != %d", s, c, x, budget, got, want)
						}
					}
				}
			}
		}
	}
}
func TestTransformConstantBudget(t *testing.T) {
	base, ext, _, _ := orbitFields(13, 0x201b)
	for _, f := range []orbitField{base, ext} {
		a := newAdditiveTransform(f)
		if a.constantTableBytes > 256<<10 {
			t.Fatal("transform exceeds constant-table budget")
		}
	}
}
