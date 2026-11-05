//go:build liftrs_cpp && cgo

// Profile complete native encoding against the retained hybrid path and hashes.
// This is a mathematical-encoder benchmark; no masking, packing or storage writes.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	native "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	liftrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	out := flag.String("out", "", "fresh result directory")
	cache := flag.String("cache", "", "retained minimal binary file at s14; small fixture cache directory otherwise")
	s := flag.Int("s", 14, "14 retained instance or 8 smoke fixture")
	backend := flag.String("backend", "native,hybrid,hybrid,native,native,hybrid", "ordered comma-separated cases")
	clocksOnly := flag.Bool("clocks-only", false, "run short busy/sleep controls only")
	workers := flag.String("workers", "16", "explicit native worker counts")
	repeats := flag.Int("repeats", 1, "profiled encoding repeats per count")
	flag.Parse()
	if *clocksOnly {
		must(os.MkdirAll(*out, 0755))
		var controls []map[string]any
		for trial := 0; trial < 6; trial++ {
			before, err := native.ReadClocks()
			must(err)
			start := time.Now()
			kind := "sleep"
			if trial%2 == 0 {
				kind = "busy"
				var x uint64 = 17
				for time.Since(start) < 250*time.Millisecond {
					for j := 0; j < 1000; j++ {
						x = x*6364136223846793005 + 1
					}
				}
				runtime.KeepAlive(x)
			} else {
				time.Sleep(250 * time.Millisecond)
			}
			elapsed := time.Since(start)
			after, err := native.ReadClocks()
			must(err)
			controls = append(controls, map[string]any{"kind": kind, "go_ns": elapsed.Nanoseconds(), "before": before, "after": after})
		}
		raw, err := json.MarshalIndent(controls, "", "  ")
		must(err)
		must(os.WriteFile(filepath.Join(*out, "clocks.json"), raw, 0644))
		return
	}
	if *out == "" || *cache == "" || (*s != 8 && *s != 14) || *repeats < 1 {
		panic("out/cache required; only s8/14")
	}
	must(os.MkdirAll(*out, 0755))
	debug.SetMemoryLimit(12 << 30)
	var phases []map[string]any
	save := func() {
		raw, err := json.MarshalIndent(phases, "", "  ")
		must(err)
		must(os.WriteFile(filepath.Join(*out, "phases.json"), raw, 0644))
	}
	phase := func(name string, fn func() map[string]any) {
		must(os.WriteFile(filepath.Join(*out, "state.json"), []byte(fmt.Sprintf(`{"phase":%q}`, name)), 0644))
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		f, err := os.Create(filepath.Join(*out, name+".pprof"))
		must(err)
		must(pprof.StartCPUProfile(f))
		nativeBefore, err := native.ReadClocks()
		must(err)
		var usageBefore, usageAfter syscall.Rusage
		must(syscall.Getrusage(syscall.RUSAGE_SELF, &usageBefore))
		start := time.Now()
		r := fn()
		elapsed := time.Since(start)
		nativeAfter, err := native.ReadClocks()
		must(err)
		must(syscall.Getrusage(syscall.RUSAGE_SELF, &usageAfter))
		if stats, ok := r["native_stages"].(liftrs.ScalingCPPStats); ok {
			sum := stats.TaylorNS + stats.ColumnsNS + stats.RowsNS + stats.BufferNS + stats.PointwiseNS + stats.AssemblyNS
			if sum > nativeAfter.SteadyNS-nativeBefore.SteadyNS+1_000_000 {
				panic("native stages exceed same-clock total")
			}
		}
		r["native_before"] = nativeBefore
		r["native_after"] = nativeAfter
		timevalNS := func(v syscall.Timeval) int64 { return v.Sec*1e9 + int64(v.Usec)*1000 }
		r["user_cpu_ns"] = timevalNS(usageAfter.Utime) - timevalNS(usageBefore.Utime)
		r["system_cpu_ns"] = timevalNS(usageAfter.Stime) - timevalNS(usageBefore.Stime)
		r["max_rss_native_units"] = usageAfter.Maxrss
		pprof.StopCPUProfile()
		must(f.Close())
		runtime.ReadMemStats(&after)
		r["phase"] = name
		r["elapsed_ns"] = elapsed.Nanoseconds()
		r["heap_before_bytes"] = before.HeapAlloc
		r["heap_after_bytes"] = after.HeapAlloc
		r["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
		r["go_sys_bytes"] = after.Sys
		phases = append(phases, r)
		save()
		fmt.Println(name, elapsed)
	}
	var plan *liftrs.ScalingOnlinePlan
	var oracle *liftrs.CyclicEncoder
	phase("load", func() map[string]any {
		if *s == 14 {
			f, err := os.Open(*cache)
			must(err)
			defer f.Close()
			plan, err = liftrs.LoadScalingOnline(f, "88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8", 1<<28)
			must(err)
		} else {
			must(os.MkdirAll(*cache, 0755))
			path := filepath.Join(*cache, "s8-cyclic.json")
			raw, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				basis, err := liftrs.DiscoverOrbit(liftrs.Spec{FieldDegree: 8, Dimension: 2, CurveDegree: 8}, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(1501))})
				must(err)
				oracle, err = basis.PrepareCyclicFFT()
				must(err)
				raw, err = oracle.MarshalBinary()
				must(err)
				must(os.WriteFile(path, raw, 0644))
			} else {
				must(err)
				oracle, err = liftrs.LoadCyclicEncoder(raw)
				must(err)
			}
			plan, err = oracle.ScalingPlan()
			must(err)
		}
		return map[string]any{"s": *s, "k": plan.Dimension(), "n": plan.Length(), "plan_array_bytes": plan.ResidentBytes(), "small_load_includes_generation_if_missing": *s == 8}
	})
	message := make([]uint32, plan.Dimension())
	random := rand.New(rand.NewSource(71))
	for i := range message {
		message[i] = uint32(random.Intn(1 << uint(*s)))
	}
	var want string
	digest := func(word []uint32) string {
		h := sha256.New()
		var block [65536]byte
		for start := 0; start < len(word); start += len(block) / 4 {
			n := len(block) / 4
			if n > len(word)-start {
				n = len(word) - start
			}
			for i := 0; i < n; i++ {
				binary.LittleEndian.PutUint32(block[4*i:], word[start+i])
			}
			h.Write(block[:4*n])
		}
		return hex.EncodeToString(h.Sum(nil))
	}
	if *s == 14 {
		want = "b828d4dda538e95b78d914dce52fbfa7fc29a139e4781b0dad83f061f3e26d6c"
	} else {
		word, err := oracle.Encode(message)
		must(err)
		want = digest(word)
		oracle = nil
	}
	for _, value := range strings.Split(*workers, ",") {
		count, err := strconv.Atoi(value)
		must(err)
		var encoder *liftrs.ScalingCPPEncoder
		var work *liftrs.ScalingCPPWorkspace
		phase(fmt.Sprintf("setup-w%d", count), func() map[string]any {
			var err error
			encoder, err = plan.WithCPPMatrix(count)
			must(err)
			return map[string]any{"workers": count}
		})
		phase(fmt.Sprintf("workspace-w%d", count), func() map[string]any {
			work = encoder.NewWorkspace()
			return map[string]any{"workspace_bytes": work.Bytes(), "input_bytes": len(message) * 4}
		})
		for repeat := 0; repeat < *repeats; repeat++ {
			for caseIndex, selected := range strings.Split(*backend, ",") {
				var word []uint32
				phase(fmt.Sprintf("encode-w%d-r%d-c%d-%s", count, repeat, caseIndex, selected), func() map[string]any {
					var err error
					var stats liftrs.ScalingCPPStats
					if selected == "native" {
						word, stats, err = encoder.EncodeWithStats(message, work)
					} else if selected == "hybrid" {
						word, stats, err = encoder.EncodeHybridWithStats(message, work)
					} else {
						panic("unknown backend")
					}
					must(err)
					return map[string]any{"backend": selected, "workers": count, "repeat": repeat, "output_bytes": len(word) * 4, "native_stages": stats}
				})
				phase(fmt.Sprintf("verify-w%d-r%d-c%d-%s", count, repeat, caseIndex, selected), func() map[string]any {
					got := digest(word)
					if got != want {
						panic("canonical hash differs: " + got)
					}
					return map[string]any{"canonical_sha256": got, "full_hash_match": true}
				})
				word = nil
			}
		}
		must(encoder.Close())
		work = nil
		encoder = nil
	}
	runtime.KeepAlive(plan)
	save()
}
