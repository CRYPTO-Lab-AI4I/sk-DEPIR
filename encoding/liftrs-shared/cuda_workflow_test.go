//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Reuse the same all-systematic/curve/packing harness across explicit CPU/GPU
// sessions. The public CUDA implementation itself does not depend on C++ tags.
func TestCUDAEncodingWorkflow(t *testing.T) {
	out, cache := os.Getenv("PIR_CUDA_WORKFLOW_OUT"), os.Getenv("PIR_CUDA_WORKFLOW_CACHE")
	if out == "" {
		t.Skip("opt-in complete CUDA encoding workflow")
	}
	reference := os.Getenv("PIR_CUDA_WORKFLOW_REFERENCE")
	for _, backend := range []string{"cpu", "cuda"} {
		dest := filepath.Join(cache, backend)
		if err := os.MkdirAll(dest, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"online.bin", "online-metadata.json"} {
			if err := os.Symlink(filepath.Join(reference, name), filepath.Join(dest, name)); err != nil {
				t.Fatal(err)
			}
		}
		results := filepath.Join(out, backend)
		if err := os.MkdirAll(results, 0755); err != nil {
			t.Fatal(err)
		}
		cfg := scalingFirstUseConfig{spec: Spec{10, 2, 5}, out: results, cache: dest, workers: 16, repeats: 2, memoryLimit: 4 << 30, onlineOnly: true}
		if backend == "cuda" {
			cfg.onlineSetup = func(plan *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
				options := CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}
				var gpu *ScalingCUDAEncoder
				var err error
				mode := "cuda"
				if os.Getenv("PIR_CUDA_HOST_WORKFLOW") == "1" {
					gpu, err = plan.WithCUDAHostBacked(options, 65536)
					mode = "cuda-host-backed"
				} else {
					gpu, err = plan.WithCUDA(options)
				}
				if err != nil {
					return nil, nil, nil, err
				}
				return func(message []uint32) ([]uint32, map[string]any, error) {
					word, stats, err := gpu.EncodeWithStats(message)
					return word, map[string]any{"cuda_stages": stats}, err
				}, func() { gpu.Close() }, map[string]any{"backend": mode, "device_bytes": gpu.OwnedBytes()}, nil
			}
		}
		runScalingFirstUse(t, cfg)
	}
	var digests []string
	for _, backend := range []string{"cpu", "cuda"} {
		raw, err := os.ReadFile(filepath.Join(cache, backend, "word-metadata.json"))
		if err != nil {
			t.Fatal(err)
		}
		var meta struct{ SHA256 string }
		if err = json.Unmarshal(raw, &meta); err != nil {
			t.Fatal(err)
		}
		digests = append(digests, meta.SHA256)
	}
	if digests[0] != digests[1] {
		t.Fatal("CPU/CUDA packed words differ")
	}
}
