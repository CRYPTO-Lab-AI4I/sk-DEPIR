// Measure allocation shapes used by preparation, not a code encoder or s=15 run.
package main

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"time"
	"unsafe"
)

var retained any

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	kind := flag.String("kind", "map", "map, factors or bitset")
	count := flag.Int("count", 1000000, "representative entries (at most ten million)")
	domain := flag.Int("domain", 1<<24, "bitset positions")
	out := flag.String("out", "", "fresh output directory")
	flag.Parse()
	if *count < 1 || *count > 10000000 || *domain < *count || *out == "" {
		panic("invalid probe arguments")
	}
	must(os.MkdirAll(*out, 0755))
	debug.SetMemoryLimit(1 << 30)
	runtime.GC()
	var before, allocated, live runtime.MemStats
	runtime.ReadMemStats(&before)
	profile, err := os.Create(filepath.Join(*out, "cpu.pprof"))
	must(err)
	must(pprof.StartCPUProfile(profile))
	// Establish the allocation baseline after profiler buffers exist. The pilot
	// measured those buffers as part of each object shape; keep that pilot separate.
	runtime.GC()
	runtime.ReadMemStats(&before)
	start := time.Now()
	capacity := *count
	switch *kind {
	case "map":
		values := make(map[int]bool, *count)
		for i := 0; i < *count; i++ {
			values[i] = true
		}
		retained = values
	case "factors":
		var values [][]uint32
		for i := 0; i < *count; i++ {
			values = append(values, []uint32{uint32(i), uint32(i >> 16), 1})
		}
		capacity = cap(values)
		retained = values
	case "bitset":
		values := make([]byte, (*domain+7)/8)
		for i := range values {
			values[i] = byte(i)
		}
		retained = values
	default:
		panic("unknown allocation kind")
	}
	allocateNS := time.Since(start).Nanoseconds()
	runtime.ReadMemStats(&allocated)
	runtime.GC()
	runtime.ReadMemStats(&live)
	heap, err := os.Create(filepath.Join(*out, "heap.pprof"))
	must(err)
	must(pprof.WriteHeapProfile(heap))
	must(heap.Close())
	pprof.StopCPUProfile()
	must(profile.Close())
	runtime.KeepAlive(retained)
	result := map[string]any{"kind": *kind, "count": *count, "domain": *domain, "capacity": capacity, "allocation_ns": allocateNS, "int_bytes": unsafe.Sizeof(int(0)), "slice_header_bytes": unsafe.Sizeof([]uint32{}), "heap_before_bytes": before.HeapAlloc, "heap_after_allocation_bytes": allocated.HeapAlloc, "heap_after_gc_bytes": live.HeapAlloc, "retained_delta_bytes": int64(live.HeapAlloc) - int64(before.HeapAlloc), "allocation_traffic_bytes": allocated.TotalAlloc - before.TotalAlloc, "live_objects_delta": int64(live.HeapObjects) - int64(before.HeapObjects), "go": runtime.Version(), "cpu_profile": "cpu.pprof", "heap_profile": "heap.pprof"}
	raw, err := json.MarshalIndent(result, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(*out, "measurement.json"), raw, 0644))
}
