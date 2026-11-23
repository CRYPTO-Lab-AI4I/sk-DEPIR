//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"testing"
)

// Test reporting stays outside the shared session. The runtime returns errors;
// bounded tests keep their existing signatures and independent byte reviewers.
type pipelineGPU[Word transposeWord] struct {
	*rankGPUSession[Word]
	t *testing.T
}

func openPipelineGPU[Word transposeWord](t *testing.T, w *mpi.World, options cuda.PreparationOptions, dim int, label string, emit func(map[string]any)) *pipelineGPU[Word] {
	t.Helper()
	session, err := newRankGPUSession[Word](w, options, dim, label, emit)
	if err != nil {
		t.Fatal(err)
	}
	return &pipelineGPU[Word]{session, t}
}
func (d *pipelineGPU[Word]) Product(name string, a, b []uint32, ac, bc, keep uint64) []uint32 {
	d.t.Helper()
	out, err := d.rankGPUSession.Product(name, a, b, ac, bc, keep)
	if err != nil {
		d.t.Fatal(name, err)
	}
	return out
}
func (d *pipelineGPU[Word]) ProductPrefixInputs(name string, a, b []uint32, ac, bc, keep uint64) []uint32 {
	d.t.Helper()
	out, err := d.rankGPUSession.ProductPrefixInputs(name, a, b, ac, bc, keep)
	if err != nil {
		d.t.Fatal(name, err)
	}
	return out
}
func (d *pipelineGPU[Word]) Spectrum(name string, g []uint32, count uint64) []Word {
	d.t.Helper()
	out, err := d.rankGPUSession.Spectrum(name, g, count)
	if err != nil {
		d.t.Fatal(name, err)
	}
	return out
}
func (d *pipelineGPU[Word]) PrepareSpectrum(name string, g []uint32, count uint64) {
	d.t.Helper()
	if err := d.rankGPUSession.PrepareSpectrum(name, g, count); err != nil {
		d.t.Fatal(name, err)
	}
}
