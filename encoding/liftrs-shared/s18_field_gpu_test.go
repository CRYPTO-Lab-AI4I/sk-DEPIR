//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"testing"
)

// These bounded components use the canonical s18 field, not an s18 code grid.
// Reuse the existing high-extension FFT and exact schoolbook product drivers.
func s18GPUFieldCases(t *testing.T, product bool) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" || os.Getenv("PIR_GPU_MATRIX_STAGE_TAYLOR") != "1" {
		t.Skip("explicit bounded s18 field admission")
	}
	dims := []int{5, 12}
	if os.Getenv("PIR_S18_FIELD_SMALL") == "1" {
		dims = []int{5}
	}
	ranks := 2
	switch os.Getenv("PIR_S18_FIELD_RANKS") {
	case "", "2":
	case "4":
		ranks = 4
	default:
		t.Fatal("worker count admission")
	}
	withRankGPUStages(t, output, ranks, func(w *mpi.World, emit func(map[string]any)) {
		operation := uint64(1)
		if product {
			distributedProductGPUCases[uint64](t, w, 18, dims, output, &operation, emit, productTestOptions{Resident: true, LocalTaylor: true, TileWords: 127})
		} else {
			residentVectorCases[uint64](t, w, 18, dims, output, &operation, emit, true)
		}
		emit(map[string]any{"kind": "completed", "cases": 2 * len(dims)})
	})
}

func TestS18GPUFFT(t *testing.T)      { s18GPUFieldCases(t, false) }
func TestS18GPUProducts(t *testing.T) { s18GPUFieldCases(t, true) }

// Exercise the independent native/scalar oracle locally before scheduling; no
// GPU or MPI initialization occurs, and no full 2^36-element grid is allocated.
func TestS18NativeField(t *testing.T) {
	t.Setenv("PIR_GPU_MATRIX_STAGE_TAYLOR", "1")
	for _, ranks := range []int{2, 4} {
		for rank := 0; rank < ranks; rank++ {
			for _, dim := range []int{5, 12} {
				matrixStageOracle(t, 18, dim, rank, ranks)
			}
		}
	}
}
