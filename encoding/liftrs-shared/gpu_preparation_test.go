//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/gpu-matrix"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type gpuPreparationExecutor struct {
	gpu      *gpumatrix.Engine
	fallback transformExecutor
	calls    int
	stats    [8]float64
}

func (e *gpuPreparationExecutor) apply(data []uint32, inverse bool) {
	if len(data) < 4096 {
		e.fallback.apply(data, inverse)
		return
	}
	stats, err := e.gpu.Apply(data, inverse)
	if err != nil {
		panic(err)
	}
	e.calls++
	for i, v := range stats {
		e.stats[i] += v
	}
}

func TestGPUPreparation(t *testing.T) {
	out := os.Getenv("PIR_GPU_PREP_OUT")
	if out == "" {
		t.Skip("opt-in GPU-assisted preparation")
	}
	cfg := scalingFirstUseConfig{spec: Spec{10, 2, 5}, out: out, cache: os.Getenv("PIR_GPU_PREP_CACHE"), workers: 4, repeats: 1, memoryLimit: 4 << 30, onlineOnly: os.Getenv("PIR_GPU_ONLINE_ONLY") == "1"}
	var executor *gpuPreparationExecutor
	var setupSeconds float64
	if os.Getenv("PIR_GPU_BACKEND") == "gpu" {
		cfg.prepareExecutor = func(basis *OrbitBasis) (func(), error) {
			start := time.Now()
			table := newPreparationProducts(basis.base.base)
			engine, err := gpumatrix.New(os.Getenv("PIR_GPU_LIBRARY"), 10, basis.polynomial, basis.delta, table.log, table.powers)
			if err != nil {
				return nil, err
			}
			for dimension := 12; dimension <= 20; dimension++ {
				words, _, _, err := gpuMatrixDescription(basis.extension, dimension)
				if err != nil {
					engine.Close()
					return nil, err
				}
				if err = engine.AddPlan(words); err != nil {
					engine.Close()
					return nil, err
				}
			}
			executor = &gpuPreparationExecutor{gpu: engine, fallback: basis.extensionFFT.executor}
			basis.extensionFFT.executor = executor
			setupSeconds = time.Since(start).Seconds()
			return engine.Close, nil
		}
	}
	runScalingFirstUse(t, cfg)
	result := map[string]any{"backend": os.Getenv("PIR_GPU_BACKEND"), "setup_seconds": setupSeconds, "boundaries": "Only extension transforms >=4096 points use GPU. Base discovery transforms, root products/orchestration, pointwise products and complete online encoding remain CPU; each GPU transform uploads/downloads. CPU native scratch is retained."}
	if executor != nil {
		result["calls"] = executor.calls
		result["stage_seconds"] = executor.stats
		result["owned_device_bytes_after_close"] = executor.gpu.OwnedBytes()
	}
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "gpu-summary.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
