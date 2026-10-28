package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"sync"
	"testing"
	"time"
)

// Opt-in profiling harness: ordinary go test runs only the correctness tests.
// JSON contains observed heap/allocations, never a claimed OS memory bound.
func TestScalingProbe(t *testing.T) {
	dir := os.Getenv("PIR_SCALING_PROBE")
	if dir == "" {
		t.Skip("opt-in resource experiment")
	}
	cache := os.Getenv("PIR_SCALING_CACHE")
	if cache == "" {
		t.Fatal("cache directory required")
	}
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	var results []map[string]any
	save := func() {
		raw, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "phases.json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		peak := before.HeapAlloc
		var mu sync.Mutex
		stop, done := make(chan struct{}), make(chan struct{})
		go func() {
			defer close(done)
			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					mu.Lock()
					if m.HeapAlloc > peak {
						peak = m.HeapAlloc
					}
					mu.Unlock()
				}
			}
		}()
		file, err := os.Create(filepath.Join(dir, name+".pprof"))
		if err != nil {
			t.Fatal(err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		data := fn()
		elapsed := time.Since(start)
		pprof.StopCPUProfile()
		file.Close()
		close(stop)
		<-done
		runtime.ReadMemStats(&after)
		if after.HeapAlloc > peak {
			peak = after.HeapAlloc
		}
		data["phase"] = name
		data["elapsed_ns"] = elapsed.Nanoseconds()
		data["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
		data["baseline_heap_bytes"] = before.HeapAlloc
		data["sampled_peak_heap_bytes"] = peak
		data["gc_cycles"] = after.NumGC - before.NumGC
		data["heap_sampling_ms"] = 5
		results = append(results, data)
		save()
		t.Log(name, elapsed)
	}
	for s := 8; s <= 15; s++ {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		var expected []uint32
		if s <= 13 {
			phase(fmt.Sprintf("dense-setup-s%d", s), func() map[string]any {
				a := newAdditiveTransform(f)
				expected = cppDescription(a)
				return map[string]any{"s": s, "description_bytes": len(expected) * 4}
			})
		}
		phase(fmt.Sprintf("sparse-setup-s%d", s), func() map[string]any {
			a := newSparseTransform(f)
			words := cppDescription(a)
			if s <= 13 && !reflect.DeepEqual(words, expected) {
				t.Fatal("sparse description differs", s)
			}
			raw := make([]byte, 4*len(words))
			for i, x := range words {
				binary.LittleEndian.PutUint32(raw[4*i:], x)
			}
			path := filepath.Join(cache, fmt.Sprintf("s%d-transform.bin", s))
			if err := os.WriteFile(path, raw, 0644); err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(raw)
			return map[string]any{"s": s, "description_bytes": len(raw), "description_sha256": hex.EncodeToString(digest[:]), "cache": path, "dense_comparison": s <= 13}
		})
	}
	for _, s := range []int{8, 9} {
		curve := 8
		if s == 9 {
			curve = 7
		}
		var basis *OrbitBasis
		phase(fmt.Sprintf("discovery-s%d", s), func() map[string]any {
			var err error
			basis, err = DiscoverOrbit(Spec{s, 2, curve}, OrbitOptions{Random: rand.New(rand.NewSource(900 + int64(s)))})
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{"k": basis.Dimension(), "n": basis.Length()}
		})
		var baseline, candidate *CyclicEncoder
		phase(fmt.Sprintf("baseline-preparation-s%d", s), func() map[string]any {
			var err error
			baseline, err = basis.PrepareCyclicFFT()
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{}
		})
		phase(fmt.Sprintf("prototype-preparation-s%d", s), func() map[string]any {
			var err error
			candidate, err = basis.PrepareCyclicScalingPrototype()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(baseline.p, candidate.p) || !reflect.DeepEqual(baseline.g, candidate.g) || !reflect.DeepEqual(baseline.spectrum, candidate.spectrum) {
				t.Fatal("prepared constants mismatch")
			}
			return map[string]any{"pg_spectrum_equal": true}
		})
		phase(fmt.Sprintf("cache-save-s%d", s), func() map[string]any {
			raw, err := candidate.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(cache, fmt.Sprintf("s%d-cyclic.json", s))
			if err := os.WriteFile(path, raw, 0644); err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(raw)
			return map[string]any{"cache": path, "bytes": len(raw), "sha256": hex.EncodeToString(digest[:])}
		})
		plan, err := candidate.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		// Retain a minimal independently checksummed binary online plan and load
		// it without whole-file JSON or G/support arrays.
		path := filepath.Join(cache, fmt.Sprintf("s%d-online.bin", s))
		var checksum string
		phase(fmt.Sprintf("online-save-s%d", s), func() map[string]any {
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			hash := sha256.New()
			if err := writeScalingOnline(io.MultiWriter(file, hash), plan); err != nil {
				t.Fatal(err)
			}
			if err := file.Sync(); err != nil {
				t.Fatal(err)
			}
			file.Close()
			checksum = hex.EncodeToString(hash.Sum(nil))
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{"cache": path, "sha256": checksum, "bytes": info.Size()}
		})
		phase(fmt.Sprintf("online-load-s%d", s), func() map[string]any {
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			plan, err = loadScalingOnline(file, checksum, basis.Length())
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{"resident_array_bytes": plan.ResidentBytes(), "support_elements": len(plan.basis.support)}
		})
		// Native setup and execution measured separately in the tagged helper.
		scalingNativeProbe(t, s, phase, baseline, plan)
	}
	phase("matrix-disk-s4", func() map[string]any {
		_, f, _, _ := orbitFields(4, 0x13)
		n := int(f.q)
		input := make([]uint32, n)
		for i := range input {
			input[i] = uint32((i*11 + 5) % n)
		}
		want := append([]uint32(nil), input...)
		newSparseTransform(f).apply(want, false)
		rows := []map[string]any{}
		for _, split := range []int{2, 4, 6} {
			file, err := os.CreateTemp(cache, "matrix-s4-")
			if err != nil {
				t.Fatal(err)
			}
			d := &scalingDisk{file: file}
			for i, x := range input {
				d.put(i, x)
			}
			d.read = 0
			d.written = 0
			d.calls = 0
			start := time.Now()
			tile := matrixDisk(d, f, 8, split, false)
			elapsed := time.Since(start)
			forwardRead, forwardWrite, forwardCalls := d.read, d.written, d.calls
			for i, x := range want {
				if d.get(i) != x {
					t.Fatal("matrix forward")
				}
			}
			d.read = 0
			d.written = 0
			d.calls = 0
			matrixDisk(d, f, 8, split, true)
			inverseRead, inverseWrite := d.read, d.written
			for i, x := range input {
				if d.get(i) != x {
					t.Fatal("matrix inverse")
				}
			}
			file.Sync()
			file.Close()
			rows = append(rows, map[string]any{"split": split, "tile_payload_bytes": tile, "file_bytes": 4 * n, "forward_read_bytes": forwardRead, "forward_write_bytes": forwardWrite, "forward_syscalls": forwardCalls, "inverse_read_bytes": inverseRead, "inverse_write_bytes": inverseWrite, "forward_ns": elapsed.Nanoseconds(), "cache": file.Name(), "roundtrip": true})
		}
		return map[string]any{"n": n, "cases": rows, "note": "logical scalar file I/O; no storage-throughput claim"}
	})
	phase("routing-s4", func() map[string]any {
		b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(901))})
		if err != nil {
			t.Fatal(err)
		}
		e, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		message := make([]uint32, b.Dimension())
		for i := range message {
			message[i] = uint32(i % 16)
		}
		h := polyMulLow(b.base, e.p, message, b.Dimension())
		y := polyMul(b.base, e.g, h)
		y = append(y, 0)
		want, err := e.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		cases := []map[string]any{}
		for _, tile := range []int{4, 16, 64} {
			path := filepath.Join(cache, fmt.Sprintf("route-tile%d", tile))
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
			got, traffic := routeDisk(t, b, y, tile, path)
			if !reflect.DeepEqual(got, want) {
				t.Fatal("routing")
			}
			cases = append(cases, map[string]any{"tile_elements": tile, "logical_bucket_and_output_bytes": traffic, "bucket_bytes": 8 * (b.Length() - 1), "buckets": (b.Length() + tile - 1) / tile, "directory": path, "canonical_equal": true})
		}
		return map[string]any{"cases": cases}
	})
}
