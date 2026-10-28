package liftrs

import "math/bits"

type linearTerm struct {
	degree      int
	coefficient uint32
}

// additiveTransform evaluates/interpolates monomial coefficients on the full
// binary coordinate grid. S_k=product_(v<2^k)(X+v) is linearized. Division by
// these sparse monic polynomials gives O(N log² N) arithmetic without a Cantor
// basis restriction. All transforms include monomial-basis conversion implicitly.
type additiveTransform struct {
	f              orbitField
	terms          [][]linearTerm
	values         [][]uint32 // values[k][j]=S_k(2^j); offsets evaluated by XOR
	delta, inverse []uint32
}

func newAdditiveTransform(f orbitField) *additiveTransform {
	levels := bits.Len32(f.q) - 1
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{0, 1}
	for k := 0; k < levels; k++ {
		for d, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{d, c})
			}
		}
		a.values[k] = make([]uint32, levels)
		for j := 0; j < levels; j++ {
			for _, term := range a.terms[k] {
				a.values[k][j] ^= f.mul(term.coefficient, f.pow(1<<uint(j), term.degree))
			}
		}
		a.delta[k] = a.values[k][k]
		a.inverse[k] = f.inv(a.delta[k])
		next := make([]uint32, 2*(len(p)-1)+1)
		for d, c := range p {
			next[2*d] ^= f.mul(c, c)
			next[d] ^= f.mul(a.delta[k], c)
		}
		p = next
	}
	return a
}
func (a *additiveTransform) offset(k int, offset uint32) uint32 {
	var out uint32
	for offset != 0 {
		j := bits.TrailingZeros32(offset)
		out ^= a.values[k][j]
		offset &= offset - 1
	}
	return out
}

// apply mutates exactly q coefficients/values; callers pad before transforming.
func (a *additiveTransform) apply(buf []uint32, inverse bool) {
	a.recurse(buf, len(a.terms), 0, inverse)
}
func (a *additiveTransform) recurse(buf []uint32, level int, offset uint32, inverse bool) {
	if level == 0 {
		return
	}
	k, half := level-1, len(buf)/2
	c := a.offset(k, offset)
	terms := a.terms[k][:len(a.terms[k])-1] // omit the monic leading term
	if !inverse {
		for i := len(buf) - 1; i >= half; i-- {
			v := buf[i]
			for _, t := range terms {
				buf[i-half+t.degree] ^= a.f.mul(v, t.coefficient)
			}
		}
		for j := 0; j < half; j++ {
			v := buf[half+j]
			left := buf[j] ^ a.f.mul(c, v)
			buf[j] = left
			buf[half+j] = left ^ a.f.mul(a.delta[k], v)
		}
	}
	a.recurse(buf[:half], k, offset, inverse)
	a.recurse(buf[half:], k, offset^uint32(half), inverse)
	if inverse {
		for j := 0; j < half; j++ {
			v := a.f.mul(buf[j]^buf[half+j], a.inverse[k])
			buf[j] ^= a.f.mul(c, v)
			buf[half+j] = v
		}
		// Ascending j preserves each quotient coefficient before lower-degree terms
		// from later coefficients add into its already-consumed slot.
		for j := 0; j < half; j++ {
			v := buf[half+j]
			for _, t := range terms {
				buf[j+t.degree] ^= a.f.mul(v, t.coefficient)
			}
		}
	}
}
