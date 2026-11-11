//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Transform component only: avoiding scratch during transforms does not remove
// the complete encoder's need for safe canonical output assembly.
func TestGPUDirectRows(t *testing.T) {
	out, cache := os.Getenv("PIR_GPU_ROWS_OUT"), os.Getenv("PIR_GPU_ROWS_CACHE")
	if out == "" {
		t.Skip("opt-in transform memory/latency experiment")
	}
	if cache == "" {
		t.Fatal("retained fixture cache required")
	}
	layout := os.Getenv("PIR_GPU_ROWS_LAYOUT")
	layouts := []string{"fast"}
	switch layout {
	case "", "fast":
	case "direct":
		layouts = []string{"direct"}
	case "compare":
		layouts = []string{"fast", "direct"}
	default:
		t.Fatal("unknown layout")
	}
	dims := []int{8, 9, 16, 17, 20, 21, 24, 25}
	repeats := 3
	if os.Getenv("PIR_GPU_ROWS_SMOKE") == "1" {
		dims = []int{8, 9}
		repeats = 1
	}
	debug.SetMemoryLimit(4 << 30)
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	fprof, err := os.Create(filepath.Join(out, "cpu.pprof"))
	check(err)
	check(pprof.StartCPUProfile(fprof))
	defer fprof.Close()
	defer pprof.StopCPUProfile()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
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
	base, f, _, _ := orbitFields(16, 0x1002d)
	table := newPreparationProducts(base.base)
	library := os.Getenv("PIR_GPU_LIBRARY")
	var records []map[string]any
	save := func() {
		raw, err := json.MarshalIndent(records, "", "  ")
		check(err)
		check(os.WriteFile(filepath.Join(out, "transforms.json"), raw, 0644))
	}
	for _, dim := range dims {
		n := 1 << uint(dim)
		rng := rand.New(rand.NewSource(int64(3200 + dim)))
		input := make([]uint32, n)
		for i := range input {
			input[i] = rng.Uint32()
		}
		want := append([]uint32(nil), input...)
		scratch := make([]uint32, n)
		native, err := newMatrixCPP(f, dim)
		check(err)
		start := time.Now()
		check(cpp.ApplyMatrix(native.rows, native.columns, want, scratch, native.split, 16, false))
		referenceSeconds := time.Since(start).Seconds()
		if dim <= 9 {
			small := append([]uint32(nil), input...)
			newSparseTransform(f).applySubspace(small, false)
			if !reflect.DeepEqual(small, want) {
				t.Fatal("native/Go reference mismatch")
			}
		}
		native.close()
		scratch = nil
		words, err := cudaMatrixDescription(f, dim)
		check(err)
		desc, err := json.Marshal(words)
		check(err)
		check(os.WriteFile(filepath.Join(cache, fmt.Sprintf("plan-%d.json", dim)), desc, 0644))
		if dim == dims[len(dims)-1] {
			for label, values := range map[string][]uint32{"input": input, "reference": want} {
				file, err := os.Create(filepath.Join(cache, label+".bin"))
				check(err)
				var block [65536]byte
				for i := 0; i < n; i += len(block) / 4 {
					count := minInt(len(block)/4, n-i)
					for j := 0; j < count; j++ {
						binary.LittleEndian.PutUint32(block[4*j:], values[i+j])
					}
					_, err = file.Write(block[:4*count])
					check(err)
				}
				check(file.Close())
			}
		}
		owned := map[string]uint64{}
		for _, mode := range layouts {
			func() {
				setupStart := time.Now()
				engine, err := cuda.NewTransformBounded(library, 16, 0x1002d, f.extension.Delta(), table.log, table.powers, n, 512<<20, mode == "direct")
				check(err)
				defer engine.Close()
				before := engine.OwnedBytes()
				check(engine.AddPlan(words))
				setupSeconds := time.Since(setupStart).Seconds()
				owned[mode] = engine.OwnedBytes()
				if engine.OwnedBytes() > 512<<20 {
					t.Fatal("device cap exceeded")
				}
				if dim <= 9 {
					if _, _, err := engine.Product([]uint32{1}, []uint32{1}, 1, n); err == nil || !strings.Contains(err.Error(), "transform-only") {
						t.Fatal("product allowed", err)
					}
					if err := engine.SetOnline([]uint32{1, 1}, make([]uint32, n), f.extension.Delta()); err == nil || !strings.Contains(err.Error(), "transform-only") {
						t.Fatal("online allowed", err)
					}
					if old := os.Getenv("PIR_GPU_ROWS_OLD_LIBRARY"); old != "" {
						previous, err := cuda.NewTransformBounded(old, 16, 0x1002d, f.extension.Delta(), table.log, table.powers, n, 512<<20, true)
						if err == nil {
							previous.Close()
							t.Fatal("old library unexpectedly supports new layout")
						}
						if !strings.Contains(err.Error(), "transform layout ABI unavailable") {
							t.Fatal(err)
						}
					}
				}
				var forward, inverse []float64
				var forwardStages, inverseStages [][8]float64
				work := make([]uint32, n)
				for repeat := 0; repeat < repeats; repeat++ {
					copy(work, input)
					start := time.Now()
					st, err := engine.Apply(work, false)
					check(err)
					forward = append(forward, time.Since(start).Seconds())
					forwardStages = append(forwardStages, st)
					if !reflect.DeepEqual(work, want) {
						t.Fatal("forward mismatch", dim, mode, repeat)
					}
					if mode == "direct" && st[4] != 0 {
						t.Fatal("direct layout performed transpose")
					}
					start = time.Now()
					st, err = engine.Apply(work, true)
					check(err)
					inverse = append(inverse, time.Since(start).Seconds())
					inverseStages = append(inverseStages, st)
					if !reflect.DeepEqual(work, input) {
						t.Fatal("inverse/roundtrip mismatch", dim, mode, repeat)
					}
					if mode == "direct" && st[4] != 0 {
						t.Fatal("direct inverse performed transpose")
					}
				}
				engine.Close()
				if engine.OwnedBytes() != 0 {
					t.Fatal("close retained device memory")
				}
				if _, err := engine.Apply(work, false); err == nil {
					t.Fatal("closed engine accepted transform")
				}
				records = append(records, map[string]any{"seed": 3200 + dim, "field_s": 16, "polynomial": uint32(0x1002d), "delta": f.extension.Delta(), "dimension": dim, "n": n, "layout": mode, "owned_bytes": owned[mode], "table_bytes": before, "working_array_bytes": uint64(n) * 4 * map[bool]uint64{false: 2, true: 1}[mode == "direct"], "setup_seconds": setupSeconds, "reference_seconds": referenceSeconds, "forward_seconds": forward, "inverse_seconds": inverse, "forward_stages": forwardStages, "inverse_stages": inverseStages, "matches": true, "closed": true, "input_sha256": gpuScalingDigest(input), "reference_sha256": gpuScalingDigest(want)})
				save()
			}()
		}
		if layout == "compare" && owned["fast"]-owned["direct"] != uint64(n)*4 {
			t.Fatal("scratch saving differs from one array")
		}
		runtime.GC()
	}
}
