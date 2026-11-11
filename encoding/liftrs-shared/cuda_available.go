//go:build liftrs_cuda && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/bits"
)

func cudaMatrixDescription(f orbitField, dimension int) ([]uint32, error) {
	m, err := buildMatrixBasis(f, dimension)
	if err != nil {
		return nil, err
	}
	n := uint64(1) << uint(dimension)
	words := []uint32{0x324d474c, uint32(n), uint32(n >> 32), uint32(m.split), uint32(2 * f.base.Degree())}
	tile := func(a *additiveTransform, basis []uint32) {
		words = append(words, uint32(len(basis)))
		words = append(words, basis...)
		for k := range basis {
			words = append(words, uint32(len(a.terms[k])))
			for _, term := range a.terms[k] {
				words = append(words, uint32(term.degree), term.coefficient)
			}
			words = append(words, a.values[k]...)
			words = append(words, a.delta[k], a.inverse[k])
		}
	}
	tile(m.rows, m.low)
	tile(m.columns, m.gamma)
	return words, nil
}

func newCUDAOnlineBackend(p *ScalingOnlinePlan, options CUDAOptions) (cudaOnlineBackend, error) {
	s, _, n, l, err := cudaShape(p)
	if err != nil {
		return nil, err
	}
	// Build public field constants only; device arithmetic is variable-time,
	// as documented for this prototype, not a constant-time secret primitive.
	table := newPreparationProducts(p.basis.base.base)
	engine, err := cuda.NewBounded(options.LibraryPath, s, p.basis.polynomial, p.basis.delta, table.log, table.powers, n, options.DeviceMemoryLimit)
	if err != nil {
		return nil, err
	}
	counts := []int{l}
	if n != l {
		counts = append(counts, n)
	}
	for _, count := range counts {
		words, err := cudaMatrixDescription(p.basis.extension, bits.Len(uint(count))-1)
		if err == nil {
			err = engine.AddPlan(words)
		}
		if err != nil {
			engine.Close()
			return nil, err
		}
	}
	if err = engine.SetOnline(p.p, p.spectrum, p.basis.omega); err != nil {
		engine.Close()
		return nil, err
	}
	return engine, nil
}
