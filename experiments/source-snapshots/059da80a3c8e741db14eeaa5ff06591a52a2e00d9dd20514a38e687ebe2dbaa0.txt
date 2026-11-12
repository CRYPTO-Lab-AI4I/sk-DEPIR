//go:build liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// An independent K=1 identity gives a constant word without code discovery.
// Tiny configurable tiles cover every scheduling boundary used by large tiles.
func TestCUDAStreamBoundaries(t *testing.T) {
	out := os.Getenv("PIR_CUDA_STREAM_BOUNDARY_OUT")
	if out == "" {
		t.Skip("opt-in streamed-code boundaries")
	}
	const s = 4
	const n = 1 << (2 * s)
	poly, _ := BinaryPolynomial(s)
	base, f, delta, omega := orbitFields(s, poly)
	spectrum := make([]uint32, n)
	for i := 0; i < n-1; i++ {
		spectrum[i] = 1
	}
	fft := newSparseTransform(f)
	fft.apply(spectrum, false)
	plan := &ScalingOnlinePlan{basis: OrbitBasis{spec: Spec{s, 2, 5}, polynomial: poly, delta: delta, omega: omega, base: base, extension: f, extensionFFT: fft}, k: 1, p: []uint32{1, 1}, spectrum: spectrum}
	if cache := os.Getenv("PIR_CUDA_STREAM_BOUNDARY_CACHE"); cache != "" {
		if err := saveCUDAOnlineFixture(plan, filepath.Join(cache, "s4-code.bin")); err != nil {
			t.Fatal(err)
		}
	}
	backend, err := newCUDAOnlineBackendSchedule(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, true, false, true)
	if err != nil {
		t.Fatal(err)
	}
	b := backend.(*cudaStreamedBackend)
	gpu := &ScalingCUDAEncoder{backend: backend, k: 1, n: n}
	defer gpu.Close()
	original := append([]uint32(nil), spectrum...)
	owned := gpu.OwnedBytes()
	var records []map[string]any
	var kept, saved []uint32
	for _, tile := range []int{1, 64, n - 1, n} {
		b.tile = tile
		for _, value := range []uint32{0, 1, 15} {
			word, err := gpu.Encode([]uint32{value})
			if err != nil {
				t.Fatal(err)
			}
			for _, x := range word {
				if x != value {
					t.Fatal("constant streamed word", tile, value, x)
				}
			}
			if kept == nil {
				kept = word
				saved = append([]uint32(nil), word...)
			}
			if !reflect.DeepEqual(kept, saved) || !reflect.DeepEqual(spectrum, original) || gpu.OwnedBytes() != owned {
				t.Fatal("stream ownership/source/allocation")
			}
			st, err := b.StreamStats()
			if err != nil || uint64(st[1]) != 4*(n+1) || int(st[2]) != (n+tile-1)/tile || int(st[3]) != 4*tile || st[0] < 0 {
				t.Fatal("stream counters", st, err)
			}
		}
		records = append(records, map[string]any{"tile_elements": tile, "matches": 3})
	}
	word := make([]uint32, n)
	for _, tile := range []int{0, n + 1} {
		if _, err = b.Engine.EncodeOnlineStreamed([]uint32{1}, word, plan.p, spectrum, tile); err == nil {
			t.Fatal("accepted invalid tile")
		}
	}
	if _, err = b.Engine.EncodeOnline([]uint32{1}, word); err == nil {
		t.Fatal("stream engine accepted missing code")
	}
	if _, err = b.Engine.EncodeOnlineStreamed([]uint32{1}, spectrum, plan.p, spectrum, n); err == nil {
		t.Fatal("code/output alias")
	}
	if _, err = b.Engine.EncodeOnlineStreamed([]uint32{1}, word, plan.p[:1], spectrum, n); err == nil {
		t.Fatal("short P")
	}
	bad := append([]uint32(nil), plan.p...)
	bad[1] = 1 << s
	if _, err = b.Engine.EncodeOnlineStreamed([]uint32{1}, word, bad, spectrum, n); err == nil {
		t.Fatal("bad P range")
	}
	spectrum[0] = 1 << (2 * s)
	if _, err = b.Engine.EncodeOnlineStreamed([]uint32{1}, word, plan.p, spectrum, n); err == nil {
		t.Fatal("bad spectrum range")
	}
	spectrum[0] = original[0]
	spectrum[0] ^= 1 << s // Legal extension value, invalid base-field product.
	if _, err = gpu.Encode([]uint32{1}); err == nil {
		t.Fatal("accepted non-base product")
	}
	spectrum[0] = original[0]
	if _, err = gpu.Encode([]uint32{1}); err != nil {
		t.Fatal("stream error recovery", err)
	}
	gpu.Close()
	if gpu.OwnedBytes() != 0 || !reflect.DeepEqual(kept, saved) {
		t.Fatal("stream close/owning output")
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "boundaries.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
