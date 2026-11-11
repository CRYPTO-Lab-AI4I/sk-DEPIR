//go:build liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Requires only the CUDA build tag; the independent Go reference stays usable
// even when liftrs_cpp is omitted. Complete CPU/CUDA builds also run this test.
func TestCUDAOnlineEncoding(t *testing.T) {
	out := os.Getenv("PIR_CUDA_ONLINE_OUT")
	if out == "" {
		t.Skip("opt-in CUDA online validation")
	}
	cache := os.Getenv("PIR_CUDA_ONLINE_CACHE")
	if cache == "" {
		t.Fatal("persistent code fixture cache required")
	}
	reuse := os.Getenv("PIR_CUDA_REUSE") == "1"
	var results []map[string]any
	for _, spec := range []Spec{{2, 2, 1}, {3, 2, 2}, {4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		basis, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(2501))})
		if err != nil {
			t.Fatal(err)
		}
		code, err := basis.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		plan, err := code.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		if err = saveCUDAOnlineFixture(plan, filepath.Join(cache, fmt.Sprintf("s%d-code.bin", spec.FieldDegree))); err != nil {
			t.Fatal(err)
		}
		if reuse {
			_, _, n, l, _ := cudaShape(plan)
			if l > n/2 {
				candidate, err := cudaExperimentEncoder(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, true)
				if err == nil {
					candidate.Close()
					t.Fatal("reuse accepted oversized prefix")
				}
				results = append(results, map[string]any{"s": spec.FieldDegree, "unsupported_prefix": true})
				continue
			}
		}
		start := time.Now()
		gpu, err := cudaExperimentEncoder(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, reuse)
		if err != nil {
			t.Fatal(err)
		}
		defer gpu.Close()
		if reuse {
			checkCUDAReusePrimitives(t, gpu)
		}
		setup := time.Since(start).Seconds()
		if old := os.Getenv("PIR_OLD_ONLINE_LIBRARY"); old != "" && ((reuse && spec.FieldDegree == 6) || (!reuse && spec.FieldDegree == 4)) {
			if previous, err := cudaExperimentEncoder(plan, CUDAOptions{LibraryPath: old, DeviceMemoryLimit: 512 << 20}, reuse); err == nil {
				previous.Close()
				t.Fatal("old library exposes online encoding unexpectedly")
			} else if !strings.Contains(err.Error(), map[bool]string{false: "online CUDA ABI unavailable", true: "reuse GPU ABI unavailable"}[reuse]) {
				t.Fatal("unexpected old-library error", err)
			}
		}
		rng := rand.New(rand.NewSource(2502))
		input := make([]uint32, plan.Dimension())
		for i := range input {
			input[i] = uint32(rng.Intn(1 << uint(spec.FieldDegree)))
		}
		original := append([]uint32(nil), input...)
		cpuWork := plan.NewWorkspace()
		var stages []CUDAEncodeStats
		for repeat := 0; repeat < 3; repeat++ {
			if repeat == 1 {
				clear(input)
			} else if repeat == 2 {
				copy(input, original)
			}
			want, err := plan.Encode(input, cpuWork)
			if err != nil {
				t.Fatal(err)
			}
			got, stats, err := gpu.EncodeWithStats(input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("CPU/CUDA word mismatch", spec, repeat)
			}
			if repeat != 1 && !reflect.DeepEqual(input, original) {
				t.Fatal("CUDA changed message")
			}
			stages = append(stages, stats)
		}
		if _, err = gpu.Encode(input[:len(input)-1]); err == nil {
			t.Fatal("invalid length")
		}
		bad := append([]uint32(nil), input...)
		bad[0] = 1 << uint(spec.FieldDegree)
		if _, err = gpu.Encode(bad); err == nil {
			t.Fatal("out-of-field message")
		}
		got, err := gpu.Encode(input)
		if err != nil {
			t.Fatal(err)
		}
		want, err := plan.Encode(input, cpuWork)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("valid call after error", err)
		}
		owned := gpu.OwnedBytes()
		gpu.Close()
		gpu.Close()
		if gpu.OwnedBytes() != 0 {
			t.Fatal("device leak after close")
		}
		if _, err = gpu.Encode(input); !errors.Is(err, ErrCUDAClosed) {
			t.Fatal("closed encode")
		}
		results = append(results, map[string]any{"s": spec.FieldDegree, "k": plan.Dimension(), "n": plan.Length(), "setup_seconds": setup, "owned_bytes": owned, "stages": stages, "matches": 4, "closed": true, "invalid_messages": true})
	}
	raw, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "online-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}

func saveCUDAOnlineFixture(plan *ScalingOnlinePlan, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = WriteScalingOnline(file, plan); err != nil {
		return err
	}
	return file.Sync()
}

func TestCUDAMatrixDescriptions(t *testing.T) {
	for _, s := range []int{4, 8, 16} {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		for _, dimension := range []int{2, 2*s - 1, 2 * s} {
			words, err := cudaMatrixDescription(f, dimension)
			if err != nil {
				t.Fatal(err)
			}
			if words[0] != 0x324d474c || (uint64(words[1])|uint64(words[2])<<32) != uint64(1)<<uint(dimension) || words[5] != uint32(dimension/2+1) {
				t.Fatal("CUDA plan count/row levels")
			}
		}
	}
}
