package liftrs

import (
	"math/rand"
	"testing"
)

func TestWideOrbitArithmetic(t *testing.T) {
	random := rand.New(rand.NewSource(170049))
	for _, s := range []int{2, 8, 13, 16, 17, 18, 19, 20} {
		polynomial, err := BinaryPolynomial(s)
		if err != nil {
			t.Fatal(err)
		}
		f, err := newWideOrbitField(s, polynomial)
		if err != nil {
			t.Fatal(s, err)
		}
		if f.cardinality != uint64(1)<<uint(2*s) {
			t.Fatal("cardinality", s)
		}
		// Exercise the high extension bit explicitly: random samples alone are not
		// sufficient evidence against accidental uint32 truncation at the boundary.
		cases := []uint64{0, 1, f.mask, uint64(1) << uint(s), uint64(1) << uint(2*s-1), f.cardinality - 1}
		for i := 0; i < 256; i++ {
			cases = append(cases, random.Uint64()&(f.cardinality-1))
		}
		for i, a := range cases {
			b := cases[(i*17+3)%len(cases)]
			want := quadraticOracleMul(f.tower, f.unpack(a), f.unpack(b))
			if f.mul(a, b) != f.pack(want) || f.pack(f.unpack(a)) != a || f.mul(a, 1) != a {
				t.Fatal("oracle/packing", s, a, b)
			}
			if f.pow(a, f.cardinality) != a {
				t.Fatal("Frobenius", s, a)
			}
			inv, err := f.inverse(a)
			if a == 0 {
				if err == nil {
					t.Fatal("zero inverse accepted")
				}
			} else if err != nil || f.mul(a, inv) != 1 {
				t.Fatal("inverse", s, a, err)
			}
		}
		generator, err := f.generator()
		if err != nil || generator < uint64(1)<<uint(s) || f.pow(generator, f.cardinality-1) != 1 {
			t.Fatal("generator", s, err)
		}
		// Check primitive-order witnesses through the component-field API,
		// independently of the packed exponentiation used by generator search.
		remaining := f.cardinality - 1
		for divisor := uint64(2); divisor <= remaining/divisor; divisor++ {
			if remaining%divisor == 0 {
				if f.tower.Pow(f.unpack(generator), (f.cardinality-1)/divisor) == (QuadraticElement{A: 1}) {
					t.Fatal("nonprimitive generator", s, divisor)
				}
				for remaining%divisor == 0 {
					remaining /= divisor
				}
			}
		}
		if remaining > 1 && f.tower.Pow(f.unpack(generator), (f.cardinality-1)/remaining) == (QuadraticElement{A: 1}) {
			t.Fatal("nonprimitive generator", s, remaining)
		}
		if s <= 16 {
			_, old, delta, omega := orbitFields(s, polynomial)
			if uint64(omega) != generator || delta != f.tower.Delta() {
				t.Fatal("legacy coordinate ordering", s)
			}
			for _, a := range cases {
				b := cases[len(cases)-1]
				if f.mul(a, b) != uint64(old.mul(uint32(a), uint32(b))) {
					t.Fatal("legacy arithmetic", s)
				}
			}
		}
	}
	for _, s := range []int{-1, 0, 1, 21, 32} {
		if _, err := newWideOrbitField(s, 0); err == nil {
			t.Fatal("invalid degree accepted", s)
		}
	}
	if _, err := newWideOrbitField(17, 0); err == nil {
		t.Fatal("invalid polynomial accepted")
	}
}
