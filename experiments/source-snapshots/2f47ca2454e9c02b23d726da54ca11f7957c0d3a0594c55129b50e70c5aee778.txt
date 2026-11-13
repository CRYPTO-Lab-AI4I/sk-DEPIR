//go:build liftrs_cuda && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/bits"
)

func newCUDAHostBackend(p *ScalingOnlinePlan, options CUDAOptions, tile int) (cudaOnlineBackend, error) {
	s, _, n, l, err := cudaShape(p)
	if err != nil {
		return nil, err
	}
	table := newPreparationProducts(p.basis.base.base)
	engine, err := cuda.NewHostBacked(options.LibraryPath, s, p.basis.polynomial, p.basis.delta, table.log, table.powers, n, options.DeviceMemoryLimit, tile)
	if err != nil {
		return nil, err
	}
	for _, count := range []int{l, n} {
		words, err := cudaMatrixDescription(p.basis.extension, bits.Len(uint(count))-1)
		if err == nil {
			err = engine.AddPlan(words)
		}
		if err != nil {
			engine.Close()
			return nil, err
		}
	}
	return &cudaHostBackend{Engine: engine, p: p.p, spectrum: p.spectrum, omega: p.basis.omega}, nil
}

type cudaHostBackend struct {
	*cuda.Engine
	p, spectrum, arena []uint32
	omega              uint32
	last               [16]float64
}

func (e *cudaHostBackend) EncodeOnline(input, word []uint32) ([10]float64, error) {
	if len(e.arena) != len(word) {
		e.arena = make([]uint32, len(word))
	}
	var err error
	e.last, err = e.Engine.EncodeOnlineBacked(e.p, e.spectrum, input, e.arena, word, e.omega)
	var result [10]float64
	copy(result[:], e.last[:10])
	return result, err
}
func (e *cudaHostBackend) HostBackedDetails() [16]float64 { return e.last }
func (e *cudaHostBackend) Close()                         { e.Engine.Close(); e.p = nil; e.spectrum = nil; e.arena = nil }
