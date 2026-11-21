package liftrs

import "testing"

// Product test options keep workload changes explicit. Existing tests retain
// random operands/two shapes; larger runs select one shape to bound proof logs.
type productTestOptions struct {
	Resident, LocalTaylor bool
	TileWords             uint64
	Pattern               string
	Shapes                int
}

func productCoefficient(index, count uint64, s, operand int) uint32 {
	if index >= count {
		return 0
	}
	if index == 0 || index+1 == count {
		return 1
	}
	return uint32((index+1)*0x9e3779b9^uint64(operand+1)*0x1337) & ((uint32(1) << uint(s)) - 1)
}
func productInputCoefficient(index, count uint64, s, operand int, pattern string) uint32 {
	if pattern == "ones" && operand == 1 {
		if index < count {
			return 1
		}
		return 0
	}
	return productCoefficient(index, count, s, operand)
}

// The ones pattern keeps both operands dense, but B=1+X+...+X^(bc-1).
// Characteristic two makes each product coefficient an XOR window of A. This
// independent oracle uses constant workspace and O(ac+ownedWords) work; it does
// not run an FFT or materialize a global operand. Arbitrary inputs keep schoolbook.
func productExpected(first, words, ac, bc uint64, s int, base BinaryField, pattern string) []uint32 {
	out := make([]uint32, words)
	if ac == 0 || bc == 0 {
		return out
	}
	if pattern == "ones" {
		lo := uint64(0)
		if first >= bc {
			lo = first - bc + 1
		}
		hi := min(ac, first+1)
		var window uint32
		for i := lo; i < hi; i++ {
			window ^= productCoefficient(i, ac, s, 0)
		}
		for j := range out {
			out[j] = window
			k := first + uint64(j) + 1
			if k < ac {
				window ^= productCoefficient(k, ac, s, 0)
			}
			if k >= bc && k-bc < ac {
				window ^= productCoefficient(k-bc, ac, s, 0)
			}
		}
		return out
	}
	for j := range out {
		k := first + uint64(j)
		if k >= ac+bc-1 {
			continue
		}
		lo := uint64(0)
		if k >= bc {
			lo = k - bc + 1
		}
		hi := min(k, ac-1)
		for i := lo; i <= hi; i++ {
			out[j] ^= base.Mul(productCoefficient(i, ac, s, 0), productCoefficient(k-i, bc, s, 1))
		}
	}
	return out
}
func TestRankProductLinearOracle(t *testing.T) {
	for _, s := range []int{4, 8, 17} {
		poly, err := BinaryPolynomial(s)
		if err != nil {
			t.Fatal(err)
		}
		base, err := NewBinaryField(s, poly)
		if err != nil {
			t.Fatal(err)
		}
		for _, ac := range []uint64{1, 3, 8, 17} {
			for _, bc := range []uint64{1, 2, 5, 17} {
				n := ac + bc + 3
				want := make([]uint32, n)
				for i := uint64(0); i < ac; i++ {
					for j := uint64(0); j < bc; j++ {
						want[i+j] ^= base.Mul(productInputCoefficient(i, ac, s, 0, "ones"), productInputCoefficient(j, bc, s, 1, "ones"))
					}
				}
				for first := uint64(0); first < n; first++ {
					got := productExpected(first, n-first, ac, bc, s, base, "ones")
					for j, v := range got {
						if v != want[first+uint64(j)] {
							t.Fatal(s, ac, bc, first, j, v, want[first+uint64(j)])
						}
					}
				}
			}
		}
	}
}
