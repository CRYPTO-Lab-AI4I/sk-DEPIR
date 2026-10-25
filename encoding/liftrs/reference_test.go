package liftrs

import (
	"reflect"
	"sync"
	"testing"
)

// oracleMul uses carryless convolution followed by long division, independently
// of the encoder's shift/reduce multiplication and its residue-reachability DP.
func oracleMul(a, b uint32, degree int, polynomial uint32) uint32 {
	var product uint32
	for i := 0; i < degree; i++ {
		for j := 0; j < degree; j++ {
			product ^= ((a >> uint(i)) & (b >> uint(j)) & 1) << uint(i+j)
		}
	}
	for i := 2*degree - 2; i >= degree; i-- {
		if product&(1<<uint(i)) != 0 {
			product ^= polynomial << uint(i-degree)
		}
	}
	return product
}

func oraclePow(a uint32, exponent, degree int, polynomial uint32) uint32 {
	out := uint32(1)
	for i := 0; i < exponent; i++ {
		out = oracleMul(out, a, degree, polynomial)
	}
	return out
}

func TestReferenceFieldArithmetic(t *testing.T) {
	for _, degree := range []int{2, 3, 4} {
		c, err := Preprocess(Spec{degree, 2, 1})
		if err != nil {
			t.Fatal(err)
		}
		f := c.field
		for a := uint32(0); a < f.q; a++ {
			for b := uint32(0); b < f.q; b++ {
				if got, want := f.mul(a, b), oracleMul(a, b, degree, f.polynomial); got != want {
					t.Fatalf("degree %d: %d*%d=%d, want %d", degree, a, b, got, want)
				}
			}
			if f.pow(a, 0) != 1 {
				t.Fatal("zeroth power must equal one, including at zero")
			}
			if a != 0 && f.mul(a, f.pow(a, int(f.q)-2)) != 1 {
				t.Fatalf("degree %d: no inverse for %d", degree, a)
			}
		}
	}
}

// Exhaust all curves, exponent tuples, and evaluation parameters over GF(4)
// for degrees 1 and 2, and GF(8) for degree 1. A restriction has degree <=q-2
// exactly when its q evaluations sum to zero. No monomial-DP logic is reused.
func TestMonomialMembershipAgainstAllTinyCurves(t *testing.T) {
	for _, spec := range []Spec{{2, 2, 1}, {2, 2, 2}, {3, 2, 1}} {
		c, err := Preprocess(spec)
		if err != nil {
			t.Fatal(err)
		}
		q := int(c.field.q)
		count := 1
		for i := 0; i <= spec.CurveDegree; i++ {
			count *= q
		}
		curves := make([][]uint32, count)
		for index := range curves {
			coeffs := make([]uint32, spec.CurveDegree+1)
			v := index
			for j := range coeffs {
				coeffs[j] = uint32(v % q)
				v /= q
			}
			curves[index] = make([]uint32, q)
			for z := 0; z < q; z++ {
				for j, coefficient := range coeffs {
					curves[index][z] ^= oracleMul(coefficient, oraclePow(uint32(z), j, spec.FieldDegree, c.Polynomial()), spec.FieldDegree, c.Polynomial())
				}
			}
		}
		want := make([][]uint32, 0)
		for ey := 0; ey < q; ey++ {
			for ex := 0; ex < q; ex++ {
				good := true
			curvePairs:
				for _, x := range curves {
					for _, y := range curves {
						var sum uint32
						for z := 0; z < q; z++ {
							sum ^= oracleMul(oraclePow(x[z], ex, spec.FieldDegree, c.Polynomial()), oraclePow(y[z], ey, spec.FieldDegree, c.Polynomial()), spec.FieldDegree, c.Polynomial())
						}
						if sum != 0 {
							good = false
							break curvePairs
						}
					}
				}
				if good {
					want = append(want, []uint32{uint32(ex), uint32(ey)})
				}
			}
		}
		if !reflect.DeepEqual(c.Monomials(), want) {
			t.Fatalf("spec %+v: DP basis %v, exhaustive curve basis %v", spec, c.Monomials(), want)
		}
		if c.Dimension() != len(want) {
			t.Fatal("incorrect verified message dimension")
		}
	}
}

func TestSystematicEncodingAndReuse(t *testing.T) {
	for _, spec := range []Spec{{2, 2, 1}, {2, 2, 2}, {2, 3, 1}, {3, 2, 2}, {4, 2, 2}} {
		c, err := Preprocess(spec)
		if err != nil {
			t.Fatal(err)
		}
		d, err := Preprocess(spec)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c, d) {
			t.Fatal("preprocessing is not deterministic")
		}
		positions := c.SystematicPositions()
		for i := 0; i < c.Dimension(); i++ {
			for value := uint32(0); value < c.field.q; value++ {
				message := make([]uint32, c.Dimension())
				message[i] = value
				word, err := c.Encode(message)
				if err != nil {
					t.Fatal(err)
				}
				for j, pos := range positions {
					if word[pos] != message[j] {
						t.Fatalf("spec %+v, row %d: systematic mismatch", spec, j)
					}
				}
			}
		}
		if len(positions) > 0 {
			positions[0] = -1
		}
		monomials := c.Monomials()
		monomials[0][0] = 999
		if c.SystematicPositions()[0] < 0 || c.Monomials()[0][0] == 999 {
			t.Fatal("mutable artifact escaped")
		}
		message := make([]uint32, c.Dimension())
		for i := range message {
			message[i] = uint32(i) % c.field.q
		}
		before := append([]uint32(nil), message...)
		word, err := c.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(message, before) {
			t.Fatal("Encode changed message")
		}
		word[0] ^= 1
		again, err := c.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(word, again) {
			t.Fatal("output alias escaped")
		}
	}
}

// Every possible message of the GF(4), t=2 code is checked. At every target
// grid position, enumerate every degree<=2 curve through it (all nonconstant
// coefficients). Sum the other q-1 evaluations with multiplicity to recover it.
func TestExhaustiveTinyMessageRecovery(t *testing.T) {
	c, err := Preprocess(Spec{2, 2, 2})
	if err != nil {
		t.Fatal(err)
	}
	q := int(c.field.q)
	messages := 1
	for i := 0; i < c.Dimension(); i++ {
		messages *= q
	}
	if messages > 1024 {
		t.Fatalf("unexpected tiny dimension %d", c.Dimension())
	}
	for index := 0; index < messages; index++ {
		message := make([]uint32, c.Dimension())
		v := index
		for i := range message {
			message[i] = uint32(v % q)
			v /= q
		}
		word, err := c.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		for target := 0; target < c.Length(); target++ {
			for coefficients := 0; coefficients < q*q*q*q; coefficients++ {
				v := coefficients
				ax, bx, ay, by := uint32(v%q), uint32(v/q%q), uint32(v/q/q%q), uint32(v/q/q/q%q)
				var got uint32
				for z := 1; z < q; z++ {
					z2 := oracleMul(uint32(z), uint32(z), 2, 7)
					x := uint32(target%q) ^ oracleMul(ax, uint32(z), 2, 7) ^ oracleMul(bx, z2, 2, 7)
					y := uint32(target/q) ^ oracleMul(ay, uint32(z), 2, 7) ^ oracleMul(by, z2, 2, 7)
					got ^= word[int(x)+q*int(y)]
				}
				if got != word[target] {
					t.Fatalf("message %v target %d curve %d: got %d want %d", message, target, coefficients, got, word[target])
				}
			}
		}
	}
}

func TestRejectsInvalidInput(t *testing.T) {
	for _, spec := range []Spec{{0, 2, 1}, {8, 2, 5}, {4, 0, 2}, {4, 1000000, 2}, {4, 3, 2}, {4, 2, 0}, {4, 2, 15}, {4, 2, -1}} {
		if _, err := Preprocess(spec); err == nil {
			t.Errorf("accepted %+v", spec)
		}
	}
	c, err := Preprocess(Spec{2, 2, 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Encode(nil); err == nil {
		t.Fatal("accepted wrong message length")
	}
	message := make([]uint32, c.Dimension())
	message[0] = 4
	if _, err := c.Encode(message); err == nil {
		t.Fatal("accepted out-of-field value")
	}
	var zero Code
	if _, err := zero.Encode(nil); err == nil {
		t.Fatal("zero Code accepted message")
	}
	var absent *Code
	if _, err := absent.Encode(nil); err == nil {
		t.Fatal("nil Code accepted message")
	}
}

func TestConcurrentEncoding(t *testing.T) {
	c, err := Preprocess(Spec{3, 2, 2})
	if err != nil {
		t.Fatal(err)
	}
	message := make([]uint32, c.Dimension())
	message[0] = 3
	want, err := c.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := c.Encode(message)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Error("concurrent encoding mismatch", err)
			}
		}()
	}
	wg.Wait()
}

// In GF(4)=GF(2)[X]/(X^2+X+1), the squares of 0,1,X,X+1
// are 0,1,X+1,X. The code with t=2 has basis 1,x,x^2,y,y^2.
// This vector exercises a lifted monomial x^2 beyond total-degree-1 RM.
func TestGoldenGF4SquaredCoordinate(t *testing.T) {
	c, err := Preprocess(Spec{2, 2, 2})
	if err != nil {
		t.Fatal(err)
	}
	if c.Dimension() != 5 || c.Length() != 16 {
		t.Fatalf("unexpected code dimensions K=%d N=%d", c.Dimension(), c.Length())
	}
	if !reflect.DeepEqual(c.SystematicPositions(), []int{0, 1, 2, 4, 8}) {
		t.Fatalf("unexpected information set %v", c.SystematicPositions())
	}
	got, err := c.Encode([]uint32{0, 1, 3, 0, 0})
	if err != nil {
		t.Fatal(err)
	}
	want := []uint32{0, 1, 3, 2, 0, 1, 3, 2, 0, 1, 3, 2, 0, 1, 3, 2}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("golden codeword = %v, want %v", got, want)
	}
}

// Check arbitrary generated messages at every grid point for every accepted
// (s,m,t) combination. Curves are deterministic test fixtures through each
// target, with independent evaluation/arithmetic. This supplements, rather
// than replaces, the exhaustive GF(4) message/curve tests.
func TestEverySupportedSpecificationRecovery(t *testing.T) {
	for degree := 2; degree <= 4; degree++ {
		q := 1 << uint(degree)
		for dimension := 2; dimension <= 3; dimension++ {
			n := q * q
			if dimension == 3 {
				n *= q
			}
			if n > MaxCodewordLength {
				continue
			}
			for curveDegree := 1; curveDegree < q-1; curveDegree++ {
				spec := Spec{degree, dimension, curveDegree}
				c, err := Preprocess(spec)
				if err != nil {
					t.Fatal(err)
				}
				message := make([]uint32, c.Dimension())
				for i := range message {
					message[i] = uint32((i*i + 3*i + 1) % q)
				}
				word, err := c.Encode(message)
				if err != nil {
					t.Fatal(err)
				}
				for target := 0; target < n; target++ {
					var got uint32
					for z := 1; z < q; z++ {
						position, stride, remaining := 0, 1, target
						for coordinate := 0; coordinate < dimension; coordinate++ {
							value := uint32(remaining % q)
							remaining /= q
							for power := 1; power <= curveDegree; power++ {
								coefficient := uint32((target + 3*coordinate + power*power) % q)
								value ^= oracleMul(coefficient, oraclePow(uint32(z), power, degree, c.Polynomial()), degree, c.Polynomial())
							}
							position += int(value) * stride
							stride *= q
						}
						got ^= word[position]
					}
					if got != word[target] {
						t.Fatalf("spec %+v target %d: recovered %d, want %d", spec, target, got, word[target])
					}
				}
			}
		}
	}
}
