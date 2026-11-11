//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

type gpuPreparationExecutor struct {
	gpu      *gpumatrix.Engine
	fallback transformExecutor
	calls    int
	stats    [8]float64
	resident bool
	minimum  int
	products int
}

func (e *gpuPreparationExecutor) threshold() int {
	if e.minimum > 0 {
		return e.minimum
	}
	return 4096
}
func (e *gpuPreparationExecutor) canConvolve(n int) bool { return e.resident && n >= e.threshold() }
func (e *gpuPreparationExecutor) convolve(a, b []uint32, limit, n int) []uint32 {
	out, stats, err := e.gpu.Product(a, b, limit, n)
	if err != nil {
		panic(err)
	}
	e.products++
	for i, v := range stats {
		e.stats[i] += v
	}
	return out
}

func (e *gpuPreparationExecutor) apply(data []uint32, inverse bool) {
	if len(data) < e.threshold() {
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
	if raw := os.Getenv("PIR_GPU_PRODUCT_CUTOFF"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 24 || value > 32768 {
			t.Fatal("product cutoff must be 24..32768")
		}
		cfg.preparationCutoff = value
	}
	backend := os.Getenv("PIR_GPU_BACKEND")
	if backend != "cpu" && backend != "gpu" && backend != "resident" {
		t.Fatal("unknown preprocessing backend")
	}
	var executor *gpuPreparationExecutor
	var setupSeconds float64
	if backend != "cpu" {
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
			executor = &gpuPreparationExecutor{gpu: engine, fallback: basis.extensionFFT.executor, resident: backend == "resident"}
			basis.extensionFFT.executor = executor
			setupSeconds = time.Since(start).Seconds()
			return engine.Close, nil
		}
	}
	runScalingFirstUse(t, cfg)
	result := map[string]any{"backend": backend, "product_cutoff_override": cfg.preparationCutoff, "setup_seconds": setupSeconds, "boundaries": "Extension transforms >=4096 points use GPU. Resident backend also keeps eligible complete polynomial convolutions on device, including pointwise products; outer overlap-add/prefix assembly and small products remain Go. Base discovery transforms and complete online encoding remain CPU. CPU native scratch is retained."}
	if executor != nil {
		result["calls"] = executor.calls
		result["product_calls"] = executor.products
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
