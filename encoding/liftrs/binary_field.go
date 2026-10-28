package liftrs

import "fmt"

// BinaryField is immutable table-free polynomial-basis GF(2^s) arithmetic for
// 2<=s<=20. Bit i is the coefficient of X^i. Arithmetic methods require canonical
// elements in [0,2^s); validate external input before calling them. This portable
// backend is not constant-time: it must not be used as a constant-time primitive.
// Its storage is fixed-size and does not grow with the field cardinality.
type BinaryField struct {
	degree     uint
	polynomial uint32
}

// BinaryPolynomial returns this implementation's fixed irreducible polynomial
// identifier, including its leading bit. Existing prototype identifiers are
// preserved. Choosing another internal basis requires an explicit isomorphism.
func BinaryPolynomial(degree int) (uint32, error) {
	polynomials := [21]uint32{2: 0x7, 3: 0xb, 4: 0x13, 5: 0x25, 6: 0x43, 7: 0x83, 8: 0x11d,
		9: 0x211, 10: 0x409, 11: 0x805, 12: 0x1053, 13: 0x201b, 14: 0x402b, 15: 0x8003,
		16: 0x1002d, 17: 0x20009, 18: 0x40081, 19: 0x80027, 20: 0x100009}
	if degree < 2 || degree > 20 {
		return 0, fmt.Errorf("binary field requires degree 2..20")
	}
	return polynomials[degree], nil
}

// NewBinaryField validates an irreducible degree-s polynomial, using Rabin's
// criterion. Polynomial may differ from BinaryPolynomial's conventional choice;
// the identifier must be retained wherever field elements are persisted.
func NewBinaryField(degree int, polynomial uint32) (BinaryField, error) {
	if degree < 2 || degree > 20 || polynomial>>uint(degree) != 1 || polynomial&1 == 0 {
		return BinaryField{}, fmt.Errorf("invalid binary-field polynomial")
	}
	f := BinaryField{uint(degree), polynomial}
	x := uint32(2)
	for i := 1; i <= degree; i++ {
		x = f.Mul(x, x)
		if i <= degree/2 && binaryGCD(x^2, polynomial) != 1 {
			return BinaryField{}, fmt.Errorf("reducible binary-field polynomial")
		}
	}
	if x != 2 {
		return BinaryField{}, fmt.Errorf("reducible binary-field polynomial")
	}
	return f, nil
}
func binaryGCD(a, b uint32) uint32 {
	for b != 0 {
		for a != 0 && binaryDegree(a) >= binaryDegree(b) {
			a ^= b << uint(binaryDegree(a)-binaryDegree(b))
		}
		a, b = b, a
	}
	return a
}
func binaryDegree(x uint32) int {
	d := -1
	for x != 0 {
		d++
		x >>= 1
	}
	return d
}

// Degree returns s, the field-element width in bits.
func (f BinaryField) Degree() int { return int(f.degree) }

// Polynomial returns the irreducible polynomial bitmask including X^s.
func (f BinaryField) Polynomial() uint32 { return f.polynomial }

// Add adds two field elements. It is XOR in polynomial basis.
func (f BinaryField) Add(a, b uint32) uint32 { return a ^ b }

// Mul multiplies modulo the defining polynomial without lookup tables.
func (f BinaryField) Mul(a, b uint32) uint32 {
	var result uint32
	high := uint32(1) << f.degree
	for b != 0 {
		if b&1 != 0 {
			result ^= a
		}
		b >>= 1
		a <<= 1
		if a&high != 0 {
			a ^= f.polynomial
		}
	}
	return result
}

// Pow uses binary exponentiation, including 0^0=1 for polynomial evaluation.
func (f BinaryField) Pow(a uint32, n uint64) uint32 {
	r := uint32(1)
	for n > 0 {
		if n&1 != 0 {
			r = f.Mul(r, a)
		}
		n >>= 1
		if n != 0 {
			a = f.Mul(a, a)
		}
	}
	return r
}

// Inverse returns a^-1, or an error for zero or an out-of-range element.
func (f BinaryField) Inverse(a uint32) (uint32, error) {
	if a == 0 || a >= 1<<f.degree {
		return 0, fmt.Errorf("inverse requires a nonzero field element")
	}
	return f.Pow(a, (1<<f.degree)-2), nil
}

// Trace is the absolute trace GF(2^s)->GF(2), returned as 0 or 1.
func (f BinaryField) Trace(a uint32) uint32 {
	t := a
	for i := uint(1); i < f.degree; i++ {
		a = f.Mul(a, a)
		t ^= a
	}
	return t
}

// QuadraticElement is a+b*u over BinaryField, where u²=u+delta. Components are
// polynomial-basis base-field elements, avoiding a cardinality-sized integer or
// lookup table. In particular s=20 needs two uint32 components, not a 2^40 table.
type QuadraticElement struct{ A, B uint32 }

// QuadraticField is immutable table-free GF((2^s)^2). It uses the first base-field
// polynomial-basis integer of absolute trace one as delta. This fixes the tower
// and preserves existing orbit coordinates. Storage is fixed-size.
type QuadraticField struct {
	base  BinaryField
	delta uint32
}

// NewQuadraticField chooses the deterministic irreducible tower u²+u+delta.
func NewQuadraticField(base BinaryField) (QuadraticField, error) {
	if base.degree < 2 || base.degree > 20 {
		return QuadraticField{}, fmt.Errorf("uninitialized base field")
	}
	// Trace is GF(2)-linear: the first trace-one integer must be a power of two.
	// Testing basis vectors avoids a cardinality-dependent delta search.
	for bit := uint(0); bit < base.degree; bit++ {
		delta := uint32(1) << bit
		if base.Trace(delta) == 1 {
			return QuadraticField{base, delta}, nil
		}
	}
	return QuadraticField{}, fmt.Errorf("trace-one tower element not found")
}

// Base returns the immutable base-field description.
func (f QuadraticField) Base() BinaryField { return f.base }

// Delta returns the base-field constant defining u²=u+delta.
func (f QuadraticField) Delta() uint32 { return f.delta }

// Add adds extension elements componentwise.
func (f QuadraticField) Add(x, y QuadraticElement) QuadraticElement {
	return QuadraticElement{x.A ^ y.A, x.B ^ y.B}
}

// Mul performs three component products plus multiplication by the fixed delta.
func (f QuadraticField) Mul(x, y QuadraticElement) QuadraticElement {
	ac, bd := f.base.Mul(x.A, y.A), f.base.Mul(x.B, y.B)
	return QuadraticElement{ac ^ f.base.Mul(f.delta, bd), f.base.Mul(x.A^x.B, y.A^y.B) ^ ac}
}

// Pow uses binary exponentiation; the exponent is independent of host int width.
func (f QuadraticField) Pow(x QuadraticElement, n uint64) QuadraticElement {
	r := QuadraticElement{A: 1}
	for n > 0 {
		if n&1 != 0 {
			r = f.Mul(r, x)
		}
		n >>= 1
		if n != 0 {
			x = f.Mul(x, x)
		}
	}
	return r
}

// Inverse uses the conjugate and a base-field norm, without extension-field tables.
func (f QuadraticField) Inverse(x QuadraticElement) (QuadraticElement, error) {
	if x.A >= 1<<f.base.degree || x.B >= 1<<f.base.degree {
		return QuadraticElement{}, fmt.Errorf("extension component outside base field")
	}
	norm := f.base.Mul(x.A, x.A^x.B) ^ f.base.Mul(f.delta, f.base.Mul(x.B, x.B))
	inverse, err := f.base.Inverse(norm)
	if err != nil {
		return QuadraticElement{}, err
	}
	return QuadraticElement{f.base.Mul(x.A^x.B, inverse), f.base.Mul(x.B, inverse)}, nil
}
