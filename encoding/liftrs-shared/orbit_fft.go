package liftrs

import (
	liftrsgo "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
	"math/bits"
)

type linearTerm struct {
	degree      int
	coefficient uint32
	product     constantProduct
}

// additiveTransform evaluates/interpolates monomial coefficients on the full
// binary coordinate grid. S_k=product_(v<2^k)(X+v) is linearized. Division by
// these sparse monic polynomials gives O(N log² N) arithmetic without a Cantor
// basis restriction. All transforms include monomial-basis conversion implicitly.
type additiveTransform struct {
	f                              orbitField
	terms                          [][]linearTerm
	values                         [][]uint32 // values[k][j]=S_k(2^j); offsets evaluated by XOR
	delta, inverse                 []uint32
	deltaProducts, inverseProducts []constantProduct
	constantTableBytes             int
	executor                       transformExecutor
	goPlan                         *liftrsgo.Transform
}

func newAdditiveTransform(f orbitField) *additiveTransform {
	levels := bits.Len32(f.q) - 1
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{0, 1}
	for k := 0; k < levels; k++ {
		for d, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{degree: d, coefficient: c})
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
	// Only level constants are cached, not the changing recursion offsets. The
	// aggregate budget includes the row data of both coefficient and butterfly
	// products; exhausted budgets fall back to scalar multiplication.
	budget := 256 << 10
	prepare := func(c uint32) constantProduct {
		product := newConstantProduct(f, c, budget)
		used := len(product.rows) * 4
		budget -= used
		a.constantTableBytes += used
		return product
	}
	a.deltaProducts = make([]constantProduct, levels)
	a.inverseProducts = make([]constantProduct, levels)
	for k := 0; k < levels; k++ {
		for i := range a.terms[k] {
			a.terms[k][i].product = prepare(a.terms[k][i].coefficient)
		}
		a.deltaProducts[k] = prepare(a.delta[k])
		a.inverseProducts[k] = prepare(a.inverse[k])
	}

	a.goPlan = &liftrsgo.Transform{Field: liftrsgo.Field{Base: f.base, Extension: f.extension}, Values: a.values, Terms: make([][]liftrsgo.Term, levels), DeltaProducts: make([]liftrsgo.Product, levels), InverseProducts: make([]liftrsgo.Product, levels)}
	view := func(p constantProduct) liftrsgo.Product {
		return liftrsgo.Product{Coefficient: p.coefficient, Rows: p.rows}
	}
	for k := 0; k < levels; k++ {
		for _, term := range a.terms[k] {
			a.goPlan.Terms[k] = append(a.goPlan.Terms[k], liftrsgo.Term{Degree: term.degree, Product: view(term.product)})
		}
		a.goPlan.DeltaProducts[k] = view(a.deltaProducts[k])
		a.goPlan.InverseProducts[k] = view(a.inverseProducts[k])
	}
	return a
}

// apply mutates exactly q coefficients/values; callers pad before transforming.
func (a *additiveTransform) apply(buf []uint32, inverse bool) {
	if a.executor != nil {
		a.executor.apply(buf, inverse)
		return
	}
	a.goPlan.Recurse(buf, len(a.terms), 0, inverse)
}

// applySubspace uses the first len(buf) coordinates, a binary subspace. The
// power-of-two buffer must fit the field. Restricting the recursion, rather than
// padding to the entire field, makes small polynomial products economical.
func (a *additiveTransform) applySubspace(buf []uint32, inverse bool) {
	n := len(buf)
	if n == 0 || n&(n-1) != 0 || uint64(n) > uint64(a.f.q) {
		panic("invalid polynomial transform length")
	}
	if a.executor != nil {
		a.executor.apply(buf, inverse)
		return
	}
	a.goPlan.Recurse(buf, bits.Len(uint(n))-1, 0, inverse)
}
