// Package liftrs implements bounded, non-cryptographic systematic encoding for
// small curve-lifted Reed-Solomon codes over binary fields. It is a correctness
// reference for locality q-1, not a production-size or cryptographic backend.
//
// Preprocess takes only a code specification. Its immutable Code can encode
// multiple messages concurrently; returned slices belong to the caller. No
// application partitioning, padding, masking, permutation, or I/O occurs here.
package liftrs

import "fmt"

// MaxCodewordLength bounds both exponent enumeration and the evaluation grid.
// With uint32 entries, the worst-case dense matrix occupies at most 256 KiB,
// excluding slice headers and temporary preprocessing data.
const MaxCodewordLength = 256

// Spec describes Lift-RS_q(m,q-1,t): every degree-at-most-t curve restriction
// has reduced degree at most q-2. Field elements use polynomial-basis integers.
// The reference backend accepts FieldDegree 2, 3, or 4, Dimension 2 or 3,
// 1 <= CurveDegree < q-1, and q^Dimension <= MaxCodewordLength.
// These deliberately tiny fields are correctness fixtures, not security profiles.
type Spec struct {
	FieldDegree int // q = 2^FieldDegree
	Dimension   int // number of evaluation coordinates, m
	CurveDegree int // t; unrelated to the message dimension K
}

// Code holds a reusable systematic generator matrix and its information set.
// Obtain an initialized Code through Preprocess. The zero value cannot encode.
// Code has no secret state; it is safe for concurrent read/Encode operations.
type Code struct {
	spec      Spec
	field     field
	length    int
	monomials [][]uint32
	generator [][]uint32
	positions []int
}

// Preprocess enumerates good monomials, evaluates them, and deterministically
// row-reduces the generator matrix. Exponent and coordinate tuples are ordered
// by x_0 + q*x_1 + ...; pivot columns are scanned in increasing order.
// Invalid or oversized specifications fail before grid/matrix allocation.
func Preprocess(spec Spec) (*Code, error) {
	var polynomial uint32
	switch spec.FieldDegree {
	case 2:
		polynomial = 0x7 // X^2 + X + 1
	case 3:
		polynomial = 0xb // X^3 + X + 1
	case 4:
		polynomial = 0x13 // X^4 + X + 1; matches utils/gf2n.go.
	default:
		return nil, fmt.Errorf("reference field degree must be 2, 3, or 4, got %d", spec.FieldDegree)
	}
	q := 1 << uint(spec.FieldDegree)
	if spec.Dimension < 2 || spec.Dimension > 3 {
		return nil, fmt.Errorf("reference dimension must be 2 or 3, got %d", spec.Dimension)
	}
	if spec.CurveDegree < 1 || spec.CurveDegree >= q-1 {
		return nil, fmt.Errorf("curve degree must be in [1,%d), got %d", q-1, spec.CurveDegree)
	}
	n := 1
	for i := 0; i < spec.Dimension; i++ {
		if n > MaxCodewordLength/q {
			return nil, fmt.Errorf("reference codeword length exceeds %d", MaxCodewordLength)
		}
		n *= q
	}
	c := &Code{spec: spec, field: field{uint32(q), polynomial}, length: n}
	for index := 0; index < n; index++ {
		exponents := coordinates(index, q, spec.Dimension)
		if !goodMonomial(exponents, spec.FieldDegree, spec.CurveDegree) {
			continue
		}
		c.monomials = append(c.monomials, exponents)
		row := make([]uint32, n)
		for position := range row {
			point := coordinates(position, q, spec.Dimension)
			value := uint32(1)
			for i, exponent := range exponents {
				value = c.field.mul(value, c.field.pow(point[i], int(exponent)))
			}
			row[position] = value
		}
		c.generator = append(c.generator, row)
	}
	if err := c.systematic(); err != nil {
		return nil, err
	}
	return c, nil
}

// Spec returns the code specification by value.
func (c *Code) Spec() Spec { return c.spec }

// Polynomial returns the field's irreducible polynomial as a coefficient bitmask,
// including the leading term. FieldDegree identifies its degree.
func (c *Code) Polynomial() uint32 { return c.field.polynomial }

// Dimension returns K, the exact number of message field elements.
func (c *Code) Dimension() int { return len(c.generator) }

// Length returns N, the number of encoded field elements.
func (c *Code) Length() int { return c.length }

// SystematicPositions returns a copy of the K codeword positions that preserve
// message elements in order. A PIR request for message j targets this position.
func (c *Code) SystematicPositions() []int {
	return append([]int(nil), c.positions...)
}

// Monomials returns a deep copy of the accepted exponent tuples before row
// reduction. They specify the code space, not the systematic message basis.
func (c *Code) Monomials() [][]uint32 {
	out := make([][]uint32, len(c.monomials))
	for i, exponents := range c.monomials {
		out[i] = append([]uint32(nil), exponents...)
	}
	return out
}

// Encode maps exactly K field elements to an N-element systematic codeword.
// Input values must be polynomial-basis integers in [0,q). Encode does not pad
// messages or modify them. Each call allocates a separate codeword.
func (c *Code) Encode(message []uint32) ([]uint32, error) {
	if c == nil || len(c.generator) == 0 {
		return nil, fmt.Errorf("uninitialized reference code; call Preprocess")
	}
	if len(message) != c.Dimension() {
		return nil, fmt.Errorf("message has %d elements, want %d", len(message), c.Dimension())
	}
	for i, value := range message {
		if value >= c.field.q {
			return nil, fmt.Errorf("message element %d is outside [0,%d)", i, c.field.q)
		}
	}
	out := make([]uint32, c.length)
	for i, value := range message {
		for j, coefficient := range c.generator[i] {
			out[j] ^= c.field.mul(value, coefficient)
		}
	}
	return out, nil
}

// goodMonomial implements Lemma 4.8 and Algorithm 1 (paper pp. 27–28).
// At bit k, reachable contributions are u*2^k for 0 <= u <= t*wt_k.
// Track only residues reached with a POSITIVE total exponent: the all-zero
// path must not classify the constant monomial as bad. A positive multiple
// of q-1 reduces to degree q-1, which is forbidden by RS(q,q-1).
func goodMonomial(exponents []uint32, degree, t int) bool {
	modulus := (1 << uint(degree)) - 1
	positive := make([]bool, modulus)
	for k := 0; k < degree; k++ {
		weight := 0
		for _, exponent := range exponents {
			weight += int((exponent >> uint(k)) & 1)
		}
		next := make([]bool, modulus)
		for u := 0; u <= t*weight; u++ {
			contribution := (u << uint(k)) % modulus
			if u > 0 { // Extend the otherwise implicit all-zero path.
				next[contribution] = true
			}
			for residue, reachable := range positive {
				if reachable {
					next[(residue+contribution)%modulus] = true
				}
			}
		}
		if next[0] {
			return false // Remaining bits can all contribute zero.
		}
		positive = next
	}
	return true
}

func coordinates(index, q, dimension int) []uint32 {
	out := make([]uint32, dimension)
	for i := range out {
		out[i] = uint32(index % q)
		index /= q
	}
	return out
}

// systematic uses deterministic Gauss-Jordan row operations, which preserve
// the code space. Full row rank follows from independence of reduced monomials;
// verify it anyway rather than returning a partial information set.
func (c *Code) systematic() error {
	pivot := 0
	for column := 0; column < c.length && pivot < len(c.generator); column++ {
		row := pivot
		for row < len(c.generator) && c.generator[row][column] == 0 {
			row++
		}
		if row == len(c.generator) {
			continue
		}
		c.generator[pivot], c.generator[row] = c.generator[row], c.generator[pivot]
		p := c.generator[pivot]
		inverse := c.field.pow(p[column], int(c.field.q)-2)
		for j := range p {
			p[j] = c.field.mul(p[j], inverse)
		}
		for i, r := range c.generator {
			if i == pivot || r[column] == 0 {
				continue
			}
			factor := r[column]
			for j := range r {
				r[j] ^= c.field.mul(factor, p[j])
			}
		}
		c.positions = append(c.positions, column)
		pivot++
	}
	if pivot != len(c.generator) {
		return fmt.Errorf("monomial evaluation matrix has rank %d, want %d", pivot, len(c.generator))
	}
	return nil
}
