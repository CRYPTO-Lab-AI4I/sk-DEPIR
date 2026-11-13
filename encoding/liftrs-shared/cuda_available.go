//go:build liftrs_cuda && cgo

package liftrs

import (
	"fmt"
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
	return newCUDAOnlineBackendMode(p, options, false)
}

// Memory modes remain explicit experiments after their s16 validation.
// Public WithCUDA continues to select the previously verified resident layout.
func newCUDAOnlineBackendMode(p *ScalingOnlinePlan, options CUDAOptions, reuse bool) (cudaOnlineBackend, error) {
	return newCUDAOnlineBackendKind(p, options, reuse, false)
}

// Compact mode remains an explicit experiment after its s16 validation.
func newCUDAOnlineBackendKind(p *ScalingOnlinePlan, options CUDAOptions, reuse, compact bool) (cudaOnlineBackend, error) {
	return newCUDAOnlineBackendSchedule(p, options, reuse, compact, false)
}

// Streamed mode retains host constants in Go only; CUDA never retains Go pointers.
func newCUDAOnlineBackendSchedule(p *ScalingOnlinePlan, options CUDAOptions, reuse, compact, streamed bool) (cudaOnlineBackend, error) {
	if streamed && (!reuse || compact) {
		return nil, fmt.Errorf("streamed mode requires fast reuse layout")
	}
	s, _, n, l, err := cudaShape(p)
	if err != nil {
		return nil, err
	}
	if reuse && l > n/2 {
		return nil, fmt.Errorf("reuse prefix operands exceed shared arena")
	}
	// Build public field constants only; device arithmetic is variable-time,
	// as documented for this prototype, not a constant-time secret primitive.
	table := newPreparationProducts(p.basis.base.base)
	create := cuda.NewBounded
	if reuse {
		create = cuda.NewBoundedReuse
	}
	if compact {
		create = cuda.NewBoundedCompact
	}
	if streamed {
		create = cuda.NewBoundedStreamed
	}
	engine, err := create(options.LibraryPath, s, p.basis.polynomial, p.basis.delta, table.log, table.powers, n, options.DeviceMemoryLimit)
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
	if streamed {
		return &cudaStreamedBackend{Engine: engine, p: p.p, spectrum: p.spectrum, tile: minInt(n, 1<<26)}, nil
	}
	return engine, nil
}

// Only this explicit prototype retains host code; the ordinary backend does not.
// The public verified plan's private arrays are immutable for this lifetime.
type cudaStreamedBackend struct {
	*cuda.Engine
	p, spectrum []uint32
	tile        int
}

func (e *cudaStreamedBackend) EncodeOnline(input, word []uint32) ([10]float64, error) {
	return e.Engine.EncodeOnlineStreamed(input, word, e.p, e.spectrum, e.tile)
}
func (e *cudaStreamedBackend) Close() { e.Engine.Close(); e.p = nil; e.spectrum = nil }
