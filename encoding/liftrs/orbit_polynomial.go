package liftrs

// Polynomial coefficients are in ascending order. Hybrid additive-FFT/Karatsuba
// products and Newton inversion support the Vandermonde solver.
// These bounded implementations allocate temporary slices, not external storage.
func polyTrim(a []uint32) []uint32 {
	for len(a) > 0 && a[len(a)-1] == 0 {
		a = a[:len(a)-1]
	}
	return a
}
func polyMul(f orbitField, a, b []uint32) []uint32 {
	return polyMulLow(f, a, b, len(a)+len(b)-1)
}
func karatsuba(f orbitField, a, b []uint32) []uint32 {
	n := len(a)
	if n <= 32 {
		c := make([]uint32, 2*n)
		for i, x := range a {
			if x != 0 {
				for j, y := range b {
					c[i+j] ^= f.mul(x, y)
				}
			}
		}
		return c
	}
	h := n / 2
	z0, z2 := karatsuba(f, a[:h], b[:h]), karatsuba(f, a[h:], b[h:])
	x, y := make([]uint32, h), make([]uint32, h)
	for i := range x {
		x[i] = a[i] ^ a[h+i]
		y[i] = b[i] ^ b[h+i]
	}
	z1 := karatsuba(f, x, y)
	c := make([]uint32, 2*n)
	for i := 0; i < n; i++ {
		c[i] ^= z0[i]
		c[n+i] ^= z2[i]
		c[h+i] ^= z1[i] ^ z0[i] ^ z2[i]
	}
	return c
}
func polyInverse(f orbitField, a []uint32, n int) []uint32 {
	g := []uint32{f.inv(a[0])}
	for len(g) < n {
		size := 2 * len(g)
		if size > n {
			size = n
		}
		aa := a
		if len(aa) > size {
			aa = aa[:size]
		}
		// In characteristic two Newton's g*(2-a*g) equals a*g².
		square := make([]uint32, 2*len(g)-1)
		for i, x := range g {
			square[2*i] = f.mul(x, x)
		}
		next := polyMulLow(f, aa, square, size)
		g = make([]uint32, size)
		copy(g, next)
	}
	return g
}
func polyReverse(a []uint32, n int) []uint32 {
	out := make([]uint32, n)
	for i := 0; i < n && i < len(a); i++ {
		out[n-1-i] = a[i]
	}
	return out
}
func polyDiv(f orbitField, a, b []uint32) (quotient, remainder []uint32) {
	a, b = polyTrim(a), polyTrim(b)
	if len(a) < len(b) {
		return nil, append([]uint32(nil), a...)
	}
	n := len(a) - len(b) + 1
	ar := polyReverse(a, len(a))
	br := polyReverse(b, len(b))
	quotient = polyMulLow(f, ar[:n], polyInverse(f, br, n), n)
	quotient = polyReverse(quotient[:n], n)
	product := polyMul(f, quotient, b)
	remainder = append([]uint32(nil), a...)
	for i, x := range product {
		remainder[i] ^= x
	}
	return polyTrim(quotient), polyTrim(remainder)
}

type productTree struct {
	polynomial  []uint32
	reciprocal  []uint32
	left, right *productTree
	root        uint32
}

func newProductTree(f orbitField, roots []uint32) *productTree {
	if len(roots) == 1 {
		return &productTree{polynomial: []uint32{roots[0], 1}, root: roots[0]}
	}
	h := len(roots) / 2
	l, r := newProductTree(f, roots[:h]), newProductTree(f, roots[h:])
	return &productTree{polynomial: polyMul(f, l.polynomial, r.polynomial), left: l, right: r}
}
func (t *productTree) evaluate(f orbitField, p []uint32, out []uint32) {
	if t.left == nil {
		v := uint32(0)
		for j := len(p) - 1; j >= 0; j-- {
			v = f.mul(v, t.root) ^ p[j]
		}
		out[0] = v
		return
	}
	l := t.left.remainder(f, p)
	r := t.right.remainder(f, p)
	h := len(t.left.polynomial) - 1
	t.left.evaluate(f, l, out[:h])
	t.right.evaluate(f, r, out[h:])
}

// prepare caches reversed-divisor reciprocals for the largest remainder each
// parent can request; this is code-only preprocessing, not encoding work.
func (t *productTree) prepare(f orbitField) {
	if t.left == nil {
		return
	}
	for _, child := range []*productTree{t.left, t.right} {
		n := len(t.polynomial) - len(child.polynomial)
		child.reciprocal = polyInverse(f, polyReverse(child.polynomial, len(child.polynomial)), n)
		child.prepare(f)
	}
}
func (t *productTree) remainder(f orbitField, p []uint32) []uint32 {
	p = polyTrim(p)
	if len(p) < len(t.polynomial) {
		return p
	}
	// During preparation no reciprocals exist yet; the generic division is used
	// only to precompute derivative weights, before any database is supplied.
	if len(t.reciprocal) == 0 {
		_, r := polyDiv(f, p, t.polynomial)
		return r
	}
	n := len(p) - len(t.polynomial) + 1
	rev := polyReverse(p, len(p))
	q := polyMulLow(f, rev[:n], t.reciprocal[:n], n)
	q = polyReverse(q[:n], n)
	product := polyMul(f, q, t.polynomial)
	r := append([]uint32(nil), p...)
	for i, x := range product {
		r[i] ^= x
	}
	return polyTrim(r)
}
