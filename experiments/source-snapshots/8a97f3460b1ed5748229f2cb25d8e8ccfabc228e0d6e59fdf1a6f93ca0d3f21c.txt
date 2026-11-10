package liftrs

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestPreparationProductPrefixes(t *testing.T) {
	base, _, _, _ := orbitFields(16, 0x1002d)
	fast := base
	fast.lookup = newPreparationProducts(base.base)
	rng := rand.New(rand.NewSource(1502))
	for _, shape := range [][2]int{{1, 1}, {3, 7}, {17, 23}, {32, 32}, {33, 65}, {128, 32}} {
		a, b := make([]uint32, shape[0]), make([]uint32, shape[1])
		for i := range a {
			if i%7 != 0 {
				a[i] = rng.Uint32() & 65535
			}
		}
		for i := range b {
			if i%5 != 0 {
				b[i] = rng.Uint32() & 65535
			}
		}
		for _, limit := range []int{0, 1, 17, 31, 32, 33, len(a) + len(b) - 1} {
			if !reflect.DeepEqual(polyMulLow(fast, a, b, limit), polyMulLow(base, a, b, limit)) {
				t.Fatal("lookup prefix", shape, limit)
			}
		}
	}
}

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
