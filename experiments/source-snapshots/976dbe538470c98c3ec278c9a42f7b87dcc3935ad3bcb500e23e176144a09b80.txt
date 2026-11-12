//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"
	"testing"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
)

// P=1+X and G=sum_{j=0}^{N-2}X^j satisfy P*G=1+X^(N-1).
// K=1 therefore has an independently known constant canonical word. This is an
// arithmetic/assembly fixture, not a discovered catalogue code or security claim.
func TestCUDACompactBoundaries(t *testing.T) {
	out := os.Getenv("PIR_CUDA_COMPACT_BOUNDARY_OUT")
	if out == "" {
		t.Skip("opt-in compact-output boundary fixtures")
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(signals)
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-signals:
			pprof.StopCPUProfile()
			os.Exit(4)
		case <-done:
		}
	}()
	var records []map[string]any
	for _, s := range []int{6, 10, 11} {
		poly, err := BinaryPolynomial(s)
		if err != nil {
			t.Fatal(err)
		}
		base, f, delta, omega := orbitFields(s, poly)
		n := 1 << uint(2*s)
		spectrum := make([]uint32, n)
		for i := 0; i < n-1; i++ {
			spectrum[i] = 1
		}
		native, err := newMatrixCPP(f, 2*s)
		if err != nil {
			t.Fatal(err)
		}
		if err = cpp.ApplyMatrix(native.rows, native.columns, spectrum, make([]uint32, n), native.split, 16, false); err != nil {
			t.Fatal(err)
		}
		native.close()
		plan := &ScalingOnlinePlan{basis: OrbitBasis{spec: Spec{s, 2, 5}, polynomial: poly, delta: delta, omega: omega, base: base, extension: f, extensionFFT: newSparseTransform(f)}, k: 1, p: []uint32{1, 1}, spectrum: spectrum}
		if cache := os.Getenv("PIR_CUDA_COMPACT_BOUNDARY_CACHE"); cache != "" {
			if err := saveCUDAOnlineFixture(plan, filepath.Join(cache, fmt.Sprintf("s%d-code.bin", s))); err != nil {
				t.Fatal(err)
			}
		}
		backend, err := newCUDAOnlineBackendKind(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, true, true)
		if err != nil {
			t.Fatal(err)
		}
		gpu := &ScalingCUDAEncoder{backend: backend, k: 1, n: n}
		for _, message := range []uint32{0, 1, (1 << uint(s)) - 1} {
			word, err := gpu.Encode([]uint32{message})
			if err != nil {
				gpu.Close()
				t.Fatal(err)
			}
			for _, value := range word {
				if value != message {
					gpu.Close()
					t.Fatal("constant compact word", s, message, value)
				}
			}
		}
		owned := gpu.OwnedBytes()
		gpu.Close()
		if gpu.OwnedBytes() != 0 {
			t.Fatal("compact boundary leak")
		}
		if s == 11 {
			// A legal extension value with bit16 set would truncate to zero in
			// uint16. Altering evaluation at0 adds it to coefficient0; the
			// compaction range flag must reject it before returning output.
			bad := *plan
			bad.spectrum = append([]uint32(nil), spectrum...)
			bad.spectrum[0] ^= 1 << 16
			b, err := newCUDAOnlineBackendKind(&bad, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, true, true)
			if err != nil {
				t.Fatal(err)
			}
			invalid := &ScalingCUDAEncoder{backend: b, k: 1, n: n}
			_, err = invalid.Encode([]uint32{1})
			invalid.Close()
			if err == nil || !strings.Contains(err.Error(), "escaped base field") {
				t.Fatal("compact narrowing hid an invalid extension value", err)
			}
		}
		records = append(records, map[string]any{"s": s, "n": n, "matches": 3, "owned_bytes": owned, "tile_relation": map[int]string{6: "smaller", 10: "equal", 11: "multiple"}[s], "invalid_truncation_rejected": s == 11})
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "boundaries.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
