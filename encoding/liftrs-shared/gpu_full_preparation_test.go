//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	"fmt"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/gpu-matrix"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGPUFullPreparation(t *testing.T) {
	out := os.Getenv("PIR_GPU_FULL_OUT")
	if out == "" {
		t.Skip("opt-in actual GPU-assisted preparation")
	}
	cache := os.Getenv("PIR_GPU_FULL_CACHE")
	if cache == "" {
		t.Fatal("cache required")
	}
	small := os.Getenv("PIR_GPU_FULL_SMALL") == "1"
	spec := Spec{16, 2, 5}
	workers := 32
	memory := int64(128 << 30)
	capBytes := uint64(64 << 30)
	if small {
		spec = Spec{10, 2, 5}
		workers = 4
		memory = 4 << 30
		capBytes = 512 << 20
	}
	run := func(out, cache string, online bool) {
		if !small {
			if _, err := os.Stat(filepath.Join(cache, "basis-metadata.json")); err != nil {
				t.Fatal("verified saved support required", err)
			}
		}
		if err := os.MkdirAll(out, 0755); err != nil {
			t.Fatal(err)
		}
		var executor *gpuPreparationExecutor
		var setup float64
		cfg := scalingFirstUseConfig{spec: spec, out: out, cache: cache, workers: workers, repeats: 1, memoryLimit: memory, prepareOnly: !online, rootCheckpoint: true, preparationCutoff: 4096}
		cfg.prepareExecutor = func(basis *OrbitBasis) (func(), error) {
			start := time.Now()
			table := newPreparationProducts(basis.base.base)
			engine, err := gpumatrix.NewBounded(os.Getenv("PIR_GPU_LIBRARY"), spec.FieldDegree, basis.polynomial, basis.delta, table.log, table.powers, 1<<uint(2*spec.FieldDegree), capBytes)
			if err != nil {
				return nil, err
			}
			for dim := 12; dim <= 2*spec.FieldDegree; dim++ {
				words, err := gpuMatrixDescriptionV2(basis.extension, dim)
				if err == nil {
					err = engine.AddPlan(words)
				}
				if err != nil {
					engine.Close()
					return nil, fmt.Errorf("prepare GPU dimension%d: %w", dim, err)
				}
			}
			executor = &gpuPreparationExecutor{gpu: engine, fallback: basis.extensionFFT.executor, resident: true, minimum: 4096}
			basis.extensionFFT.executor = executor
			setup = time.Since(start).Seconds()
			return engine.Close, nil
		}
		runScalingFirstUse(t, cfg)
		summary := map[string]any{"s": spec.FieldDegree, "setup_seconds": setup, "calls": executor.calls, "product_calls": executor.products, "stage_seconds": executor.stats, "device_cap_bytes": capBytes, "owned_after_close": executor.gpu.OwnedBytes(), "preparation_only": !online}
		if summary["owned_after_close"].(uint64) != 0 {
			t.Fatal("GPU state not closed")
		}
		raw, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(out, "gpu-summary.json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !small {
		run(out, cache, false)
		return
	}
	first := filepath.Join(cache, "first")
	run(filepath.Join(out, "first"), first, true)
	replay := filepath.Join(cache, "replay")
	if err := os.MkdirAll(replay, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"basis-metadata.json", "support.bin", "root-metadata.json", "root.bin"} {
		if err := os.Symlink(filepath.Join(first, name), filepath.Join(replay, name)); err != nil {
			t.Fatal(err)
		}
	}
	run(filepath.Join(out, "replay"), replay, false)
}
