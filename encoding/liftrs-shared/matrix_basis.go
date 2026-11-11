package liftrs

import (
	"fmt"
	"math/bits"
)

// orderedTransform constructs sparse S_k for an arbitrary independent basis.
// values still evaluates canonical packed offsets; the native executor advances
// children using basis[k], rather than confusing basis indices with field points.
func orderedTransform(f orbitField, basis []uint32) (*additiveTransform, error) {
	width := bits.Len64(f.q) - 1
	levels := len(basis)
	if levels < 1 || levels > width {
		return nil, fmt.Errorf("invalid ordered basis length")
	}
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{1}
	eval := func(x uint32) uint32 {
		var y uint32
		for j, c := range p {
			y ^= f.mul(c, f.pow(x, 1<<uint(j)))
		}
		return y
	}
	for k, x := range basis {
		if uint64(x) >= f.q {
			return nil, fmt.Errorf("basis outside field")
		}
		for j, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{degree: 1 << uint(j), coefficient: c})
			}
		}
		a.values[k] = make([]uint32, width)
		for j := 0; j < width; j++ {
			a.values[k][j] = eval(1 << uint(j))
		}
		a.delta[k] = eval(x)
		if a.delta[k] == 0 {
			return nil, fmt.Errorf("dependent basis")
		}
		a.inverse[k] = f.inv(a.delta[k])
		next := make([]uint32, len(p)+1)
		for j, c := range p {
			next[j+1] ^= f.mul(c, c)
			next[j] ^= f.mul(a.delta[k], c)
		}
		p = next
	}
	return finishAdditiveTransform(a), nil
}

// matrixBasis is shared code-only construction, independent of CPU/GPU bindings.
type matrixBasis struct {
	split         int
	rows, columns *additiveTransform
	low, gamma    []uint32
}

func buildMatrixBasis(f orbitField, dimension int) (*matrixBasis, error) {
	if dimension < 2 || dimension > bits.Len64(f.q)-1 {
		return nil, fmt.Errorf("matrix dimension outside field")
	}
	split := dimension / 2
	low := make([]uint32, split+1)
	for j := range low {
		low[j] = 1 << uint(j)
	}
	rows, err := orderedTransform(f, low)
	if err != nil {
		return nil, err
	}
	gamma := make([]uint32, dimension-split)
	for j := range gamma {
		for _, term := range rows.terms[split] {
			gamma[j] ^= f.mul(term.coefficient, f.pow(1<<uint(split+j), term.degree))
		}
	}
	columns, err := orderedTransform(f, gamma)
	if err != nil {
		return nil, err
	}
	return &matrixBasis{split, rows, columns, low, gamma}, nil
}
