//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"os"
	"testing"
)

// The production CUDA implementation remains CUDA-only buildable. This shared
// benchmark harness uses native clocks for comparable whole-session evidence.
func TestCUDACachedEncoding(t *testing.T) {
	out, cache := os.Getenv("PIR_CUDA_CACHED_OUT"), os.Getenv("PIR_CUDA_CACHED_CACHE")
	if out == "" {
		t.Skip("opt-in cached CUDA encoding")
	}
	if cache == "" {
		t.Fatal("fresh output cache required")
	}
	spec := Spec{16, 2, 5}
	streamed := os.Getenv("PIR_CUDA_STREAM") == "1"
	compact := os.Getenv("PIR_CUDA_COMPACT") == "1"
	reuse := os.Getenv("PIR_CUDA_REUSE") == "1" || compact || streamed
	capBytes := uint64(96 << 30)
	if reuse {
		capBytes = 56 << 30
	}
	if compact || streamed {
		capBytes = 40 << 30
	}
	memory := int64(64 << 30)
	if os.Getenv("PIR_CUDA_CACHED_S10") == "1" {
		spec, capBytes, memory = Spec{10, 2, 5}, 512<<20, 1<<30
	}
	if os.Getenv("PIR_CUDA_CACHED_S14") == "1" {
		spec, capBytes, memory = Spec{14, 2, 5}, 4<<30, 8<<30
	}
	if os.Getenv("PIR_CUDA_CACHED_SMOKE") == "1" {
		spec = Spec{6, 2, 4}
		capBytes = 512 << 20
		memory = 1 << 30
	}
	capBytes = scalingTableDeviceCap(capBytes)
	cfg := scalingFirstUseConfig{spec: spec, out: out, cache: cache, workers: 16, repeats: 2, memoryLimit: memory, onlineOnly: true}
	cfg.onlineSetup = func(plan *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
		gpu, err := cudaExperimentEncoder(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: capBytes}, reuse)
		if err != nil {
			return nil, nil, nil, err
		}
		payload, err := plan.CUDAArrayBytes()
		if err != nil {
			gpu.Close()
			return nil, nil, nil, err
		}
		if reuse {
			// Remove one full work array; add per-chunk parity and one flag.
			payload = payload - uint64(plan.Length())*4 + uint64((plan.Length()-1+4095)/4096+1)*4
		}
		if compact {
			payload = payload - uint64(plan.Length())*4 + 2*uint64(minInt(plan.Length(), 1<<20))
		}
		if streamed {
			payload -= 4 * uint64(plan.Length()+plan.k+1)
		}
		return func(message []uint32) ([]uint32, map[string]any, error) {
			word, stats, err := gpu.EncodeWithStats(message)
			metadata := map[string]any{"cuda_stages": stats}
			if compact && err == nil {
				extra, statsErr := gpu.backend.(interface{ CompactStats() ([3]float64, error) }).CompactStats()
				if statsErr != nil {
					return nil, nil, statsErr
				}
				metadata["compact_gpu_seconds"] = extra[0]
				metadata["expand_host_seconds"] = extra[1]
				metadata["bulk_download_bytes"] = uint64(extra[2])
			}
			if streamed && err == nil {
				extra, statsErr := gpu.backend.(interface{ StreamStats() ([4]float64, error) }).StreamStats()
				if statsErr != nil {
					return nil, nil, statsErr
				}
				metadata["code_upload_seconds"] = extra[0]
				metadata["code_upload_bytes"] = uint64(extra[1])
				metadata["spectrum_chunks"] = uint64(extra[2])
				metadata["stream_tile_bytes"] = uint64(extra[3])
			}
			return word, metadata, err
		}, func() { gpu.Close() }, map[string]any{"backend": "cuda", "reuse_buffers": reuse, "compact_output": compact, "streamed_code": streamed, "device_array_payload_bytes": payload, "owned_device_bytes": gpu.OwnedBytes(), "device_cap_bytes": capBytes}, nil
	}
	runScalingFirstUse(t, cfg)
}
