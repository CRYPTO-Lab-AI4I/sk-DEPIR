//go:build liftrs_cuda && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"os"
	"strings"
	"testing"
)

// Keep the unvalidated reuse constructor private to opt-in tests; ordinary
// callers retain WithCUDA's verified layout and unchanged memory estimate.
func cudaExperimentEncoder(p *ScalingOnlinePlan, options CUDAOptions, reuse bool) (*ScalingCUDAEncoder, error) {
	if !reuse {
		return p.WithCUDA(options)
	}
	backend, err := newCUDAOnlineBackendSchedule(p, options, true, os.Getenv("PIR_CUDA_COMPACT") == "1", os.Getenv("PIR_CUDA_STREAM") == "1")
	if err != nil {
		return nil, err
	}
	return &ScalingCUDAEncoder{backend: backend, k: p.k, n: p.Length()}, nil
}

func checkCUDAReusePrimitives(t *testing.T, gpu *ScalingCUDAEncoder) {
	t.Helper()
	engine, ok := gpu.backend.(*cuda.Engine)
	if !ok {
		engine = gpu.backend.(*cudaStreamedBackend).Engine
	}
	if _, err := engine.Apply(make([]uint32, gpu.n), false); err == nil || !strings.Contains(err.Error(), "online-only") {
		t.Fatal("reuse primitive transform rejection", err)
	}
	if _, _, err := engine.Product([]uint32{1}, []uint32{1}, 1, gpu.n); err == nil || !strings.Contains(err.Error(), "online-only") {
		t.Fatal("reuse primitive product rejection", err)
	}
}
