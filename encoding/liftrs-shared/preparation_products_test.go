package liftrs

import (
	"math/rand"
	"testing"
)

func TestPreparationProducts(t *testing.T) {
	for s := 2; s <= 16; s++ {
		polynomial, _ := BinaryPolynomial(s)
		f, err := NewBinaryField(s, polynomial)
		if err != nil {
			t.Fatal(err)
		}
		lookup := newPreparationProducts(f)
		if (len(lookup.log)+len(lookup.powers))*4 >= 768<<10 {
			t.Fatal("table budget")
		}
		rng := rand.New(rand.NewSource(int64(s)))
		for a := uint32(0); a < 1<<uint(s); a++ {
			bs := []uint32{0, 1, 2, (1 << uint(s)) - 1, rng.Uint32() & ((1 << uint(s)) - 1)}
			if s <= 8 {
				bs = make([]uint32, 1<<uint(s))
				for i := range bs {
					bs[i] = uint32(i)
				}
			}
			for _, b := range bs {
				if got, want := lookup.mul(a, b), binaryOracleMul(a, b, s, polynomial); got != want {
					t.Fatalf("s=%d a=%d b=%d got=%d want=%d", s, a, b, got, want)
				}
			}
		}
	}
	// A valid polynomial need not make element 2 primitive.
	f, _ := NewBinaryField(8, 0x11b)
	lookup := newPreparationProducts(f)
	for a := uint32(0); a < 256; a++ {
		for b := uint32(0); b < 256; b++ {
			if lookup.mul(a, b) != binaryOracleMul(a, b, 8, 0x11b) {
				t.Fatal("alternate polynomial")
			}
		}
	}
	f, _ = NewBinaryField(17, 0x20009)
	if newPreparationProducts(f) != nil {
		t.Fatal("unbounded table")
	}
}
