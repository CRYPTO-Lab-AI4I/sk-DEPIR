package liftrs

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

// productOracle uses the independent carryless/tower arithmetic, not either
// polynomial backend. It also fixes trailing-zero length semantics explicitly.
func productOracle(f orbitField, a, b []uint32, limit int) []uint32 {
	if len(a) == 0 || len(b) == 0 || limit <= 0 {
		return nil
	}
	out := make([]uint32, minInt(limit, len(a)+len(b)-1))
	mul := func(x, y uint32) uint32 { return binaryOracleMul(x, y, f.base.Degree(), f.base.Polynomial()) }
	if f.extension != nil {
		mul = func(x, y uint32) uint32 {
			s := uint(f.base.Degree())
			mask := uint32(1)<<s - 1
			z := quadraticOracleMul(*f.extension, QuadraticElement{A: x & mask, B: x >> s}, QuadraticElement{A: y & mask, B: y >> s})
			return z.A | z.B<<s
		}
	}
	for i, x := range a {
		for j, y := range b {
			if i+j < len(out) {
				out[i+j] ^= mul(x, y)
			}
		}
	}
	return out
}

func TestPolynomialProductsIndependent(t *testing.T) {
	base, ext, _, _ := orbitFields(4, 0x13)
	transform := newAdditiveTransform(ext)
	random := rand.New(rand.NewSource(204))
	for _, field := range []orbitField{base, ext} {
		for _, fft := range []bool{false, true} {
			f := field
			if fft {
				f.products = &polynomialProducts{transform: transform, cutoff: 0}
			}
			for _, n := range []int{0, 1, 2, 23, 24, 31, 32, 33, 64, 65, 127, 128, 129, 255, 256, 257} {
				a, b := make([]uint32, n), make([]uint32, n+3)
				for i := range a {
					a[i] = uint32(random.Intn(int(f.q)))
				}
				for i := range b {
					b[i] = uint32(random.Intn(int(f.q)))
				}
				// A zero highest coefficient must not shorten the returned convolution.
				b[len(b)-1] = 0
				beforeA, beforeB := append([]uint32{}, a...), append([]uint32{}, b...)
				for _, limit := range []int{0, 1, 2, n / 2, n, n + 1, 2*n + 8} {
					got := polyMulLow(f, a, b, limit)
					want := productOracle(f, a, b, limit)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("q=%d fft=%v n=%d limit=%d", f.q, fft, n, limit)
					}
				}
				if !reflect.DeepEqual(polyMul(f, a, b), productOracle(f, a, b, len(a)+len(b))) {
					t.Fatal("full product")
				}
				if !reflect.DeepEqual(a, beforeA) || !reflect.DeepEqual(b, beforeB) {
					t.Fatal("input modified")
				}
			}
		}
	}
}

func TestSubspaceTransforms(t *testing.T) {
	_, f, _, _ := orbitFields(4, 0x13)
	transform := newAdditiveTransform(f)
	random := rand.New(rand.NewSource(205))
	for n := 1; n <= int(f.q); n *= 2 {
		coefficients := make([]uint32, n)
		for i := range coefficients {
			coefficients[i] = uint32(random.Intn(int(f.q)))
		}
		values := append([]uint32(nil), coefficients...)
		transform.applySubspace(values, false)
		for point := range values {
			want := uint32(0)
			for i := len(coefficients) - 1; i >= 0; i-- {
				want = f.mul(want, uint32(point)) ^ coefficients[i]
			}
			if values[point] != want {
				t.Fatalf("forward n=%d point=%d", n, point)
			}
		}
		transform.applySubspace(values, true)
		if !reflect.DeepEqual(values, coefficients) {
			t.Fatal("inverse", n)
		}
	}
	for _, n := range []int{0, 3, 512} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("accepted length", n)
				}
			}()
			transform.applySubspace(make([]uint32, n), false)
		}()
	}
}

func TestFastPolynomialDivision(t *testing.T) {
	_, f, _, _ := orbitFields(4, 0x13)
	f.products = &polynomialProducts{transform: newAdditiveTransform(f), cutoff: 0}
	random := rand.New(rand.NewSource(206))
	for _, n := range []int{1, 23, 33, 65, 130, 257} {
		a, b := make([]uint32, n), make([]uint32, n+1)
		for i := range a {
			a[i] = uint32(random.Intn(int(f.q)))
		}
		for i := range b {
			b[i] = uint32(random.Intn(int(f.q)))
		}
		b[0] = 1
		b[len(b)-1] = 1
		product := productOracle(f, a, b, len(a)+len(b))
		q, r := polyDiv(f, product, b)
		if len(r) != 0 || !reflect.DeepEqual(q, polyTrim(a)) {
			t.Fatal("division", n)
		}
		inv := polyInverse(f, b, n)
		want := make([]uint32, n)
		want[0] = 1
		if !reflect.DeepEqual(productOracle(f, b, inv, n), want) {
			t.Fatal("series inverse", n)
		}
	}
}

var polynomialBenchmarkResult []uint32

// BenchmarkPolynomialKernels compares the old full-product-and-slice path,
// pruned Karatsuba, forced FFT, and automatic dispatch. Transform constants and
// seeded input allocation are outside the timed operation, as in code reuse.
func BenchmarkPolynomialKernels(b *testing.B) {
	base, ext, _, _ := orbitFields(9, 0x211)
	transform := newAdditiveTransform(ext)
	for _, field := range []struct {
		name   string
		f      orbitField
		cutoff int
	}{{"base", base, 32768}, {"extension", ext, 1024}} {
		for _, n := range []int{1024, 4096, 17200} {
			random := rand.New(rand.NewSource(207))
			a, c := make([]uint32, n), make([]uint32, n)
			for i := range a {
				a[i] = uint32(random.Intn(int(field.f.q)))
				c[i] = uint32(random.Intn(int(field.f.q)))
			}
			for _, low := range []bool{false, true} {
				limit := 2*n - 1
				if low {
					limit = n
				}
				for _, method := range []string{"old", "karatsuba", "fft", "auto"} {
					b.Run(fmt.Sprintf("%s/n%d/low%v/%s", field.name, n, low, method), func(b *testing.B) {
						f := field.f
						if method == "fft" {
							f.products = &polynomialProducts{transform: transform, cutoff: 0}
						}
						if method == "auto" {
							f.products = &polynomialProducts{transform: transform, cutoff: field.cutoff}
						}
						b.ReportAllocs()
						b.ResetTimer()
						for i := 0; i < b.N; i++ {
							if method == "old" {
								size := 1
								for size < n {
									size *= 2
								}
								x, y := make([]uint32, size), make([]uint32, size)
								copy(x, a)
								copy(y, c)
								polynomialBenchmarkResult = karatsuba(f, x, y)[:limit]
							} else {
								polynomialBenchmarkResult = polyMulLow(f, a, c, limit)
							}
						}
					})
				}
			}
		}
	}
}

func TestFFTProductsInDenseOraclePipeline(t *testing.T) {
	spec := Spec{4, 2, 1}
	basis, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(208))})
	if err != nil {
		t.Fatal(err)
	}
	// Force products below the production crossover to exercise the FFT in a code
	// small enough for an independently constructed dense generator oracle.
	basis.base.products.cutoff = 0
	basis.extension.products.cutoff = 0
	paper, err := basis.PreparePaper()
	if err != nil {
		t.Fatal(err)
	}
	cyclic, err := basis.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	dense, err := Preprocess(spec)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(209))
	for trial := 0; trial < 4; trial++ {
		message := make([]uint32, basis.Dimension())
		for i := range message {
			message[i] = uint32(random.Intn(16))
		}
		word, err := paper.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		other, err := cyclic.Encode(message)
		if err != nil || !reflect.DeepEqual(word, other) {
			t.Fatal("FFT encoder agreement", err)
		}
		input := make([]uint32, dense.Dimension())
		for j, pos := range dense.SystematicPositions() {
			input[j] = word[pos]
		}
		want, err := dense.Encode(input)
		if err != nil || !reflect.DeepEqual(word, want) {
			t.Fatal("dense oracle", err)
		}
	}
}

func TestUnbalancedPolynomialProducts(t *testing.T) {
	base, ext, _, _ := orbitFields(4, 0x13)
	random := rand.New(rand.NewSource(210))
	for _, field := range []orbitField{base, ext} {
		for _, cutoff := range []int{0, 1024} {
			f := field
			f.products = &polynomialProducts{transform: newAdditiveTransform(ext), cutoff: cutoff}
			a, b := make([]uint32, 401), make([]uint32, 25)
			for i := range a {
				a[i] = uint32(random.Intn(int(f.q)))
			}
			for i := range b {
				b[i] = uint32(random.Intn(int(f.q)))
			}
			for _, limit := range []int{1, 24, 100, 425, 500} {
				want := productOracle(f, a, b, limit)
				for _, operands := range [][2][]uint32{{a, b}, {b, a}} {
					got := polyMulLow(f, operands[0], operands[1], limit)
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("unbalanced q=%d cutoff=%d limit=%d", f.q, cutoff, limit)
					}
				}
			}
		}
	}
}
