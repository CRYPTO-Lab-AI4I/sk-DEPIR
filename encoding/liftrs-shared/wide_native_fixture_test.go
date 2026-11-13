package liftrs

import (
	"encoding/binary"
	"math/bits"
	"math/rand"
	"os"
	"sort"
	"testing"
)

// Trusted experiment fixture only, not a new cache format. Sparse S_k setup
// never allocates2^(2s) coefficients. Expected evaluations use independent
// carryless polynomial arithmetic rather than either recursive transform.
func TestWideNativeFixtures(t *testing.T) {
	path := os.Getenv("PIR_WIDE_NATIVE_FIXTURE")
	if path == "" {
		t.Skip("opt-in bounded native wide fixtures")
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	write := func(v uint64) {
		if err := binary.Write(file, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	write(1)
	write(60)
	random := rand.New(rand.NewSource(170051))
	for _, s := range []int{8, 16, 17, 18, 20} {
		polynomial, _ := BinaryPolynomial(s)
		f, err := newWideOrbitField(s, polynomial)
		if err != nil {
			t.Fatal(err)
		}
		oracle := func(a, b uint64) uint64 {
			mask := uint64(1)<<uint(s) - 1
			x := QuadraticElement{A: uint32(a & mask), B: uint32(a >> uint(s))}
			y := QuadraticElement{A: uint32(b & mask), B: uint32(b >> uint(s))}
			z := quadraticOracleMul(f.tower, x, y)
			return uint64(z.A) | uint64(z.B)<<uint(s)
		}
		for _, levels := range []int{3, 7, 10} {
			for _, ordered := range []bool{false, true} {
				basis := make([]uint64, levels)
				for j := range basis {
					bit := j
					if ordered {
						bit = j - 1
						if j == 0 {
							bit = 2*s - 1
						}
					}
					basis[j] = uint64(1) << uint(bit)
				}
				for _, nonzero := range []bool{false, true} {
					start := uint64(0)
					if nonzero {
						start = uint64(1) << uint(2*s-2)
					}
					write(uint64(s))
					write(uint64(polynomial))
					write(uint64(f.tower.Delta()))
					write(uint64(levels))
					write(start)
					for _, b := range basis {
						write(b)
					}
					polynomialTerms := map[uint64]uint64{1: 1}
					for k := 0; k < levels; k++ {
						degrees := make([]uint64, 0, len(polynomialTerms))
						for d := range polynomialTerms {
							degrees = append(degrees, d)
						}
						sort.Slice(degrees, func(i, j int) bool { return degrees[i] < degrees[j] })
						write(uint64(len(degrees)))
						eval := func(x uint64) uint64 {
							var result uint64
							for _, d := range degrees {
								result ^= f.mul(polynomialTerms[d], f.pow(x, d))
							}
							return result
						}
						for _, d := range degrees {
							write(d)
							write(polynomialTerms[d])
						}
						for j := 0; j < 2*s; j++ {
							write(eval(uint64(1) << uint(j)))
						}
						delta := eval(basis[k])
						inverse, err := f.inverse(delta)
						if err != nil {
							t.Fatal("dependent basis", s, k, err)
						}
						write(delta)
						write(inverse)
						next := map[uint64]uint64{}
						for d, c := range polynomialTerms {
							next[2*d] ^= f.mul(c, c)
							next[d] ^= f.mul(delta, c)
						}
						polynomialTerms = map[uint64]uint64{}
						for d, c := range next {
							if c != 0 {
								polynomialTerms[d] = c
							}
						}
					}
					n := 1 << uint(levels)
					input := make([]uint64, n)
					dense := n
					if n > 128 {
						dense = 16
					}
					for i := 0; i < dense; i++ {
						input[i] = random.Uint64() & (f.cardinality - 1)
					}
					input[n-1] = f.cardinality - 1
					// Every fixture has values above the32-bit boundary when s>=17.
					input[0] = uint64(1) << uint(2*s-1)
					write(uint64(n))
					for _, x := range input {
						write(x)
					}
					for i := 0; i < n; i++ {
						point := start
						for value := uint(i); value != 0; value &= value - 1 {
							point ^= basis[bits.TrailingZeros(value)]
						}
						var want uint64
						for j := dense - 1; j >= 0; j-- {
							want = oracle(want, point) ^ input[j]
						}
						if dense < n {
							// Independent sparse high-degree evaluation exercises
							// every native polynomial-division level without a
							// quadratic-size dense oracle at n=1024.
							power, base, exponent := uint64(1), point, uint(n-1)
							for exponent != 0 {
								if exponent&1 != 0 {
									power = oracle(power, base)
								}
								exponent >>= 1
								if exponent != 0 {
									base = oracle(base, base)
								}
							}
							want ^= oracle(input[n-1], power)
						}
						write(want)
					}
				}
			}
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
