//go:build liftrs_cpp && cgo

// Profile the opt-in parallel matrix encoder against retained canonical evidence.
// This is a mathematical-encoder benchmark; no masking, packing or storage writes.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	liftrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	out := flag.String("out", "", "fresh result directory")
	cache := flag.String("cache", "", "retained minimal binary file at s13; small fixture cache directory otherwise")
	s := flag.Int("s", 13, "13 retained instance or 8 smoke fixture")
	workers := flag.String("workers", "8,16", "explicit native worker counts")
	repeats := flag.Int("repeats", 3, "profiled encoding repeats per count")
	flag.Parse()
	if *out == "" || *cache == "" || (*s != 8 && *s != 13) || *repeats < 1 {
		panic("out/cache required; only s8/13")
	}
	must(os.MkdirAll(*out, 0755))
	debug.SetMemoryLimit(9 << 30)
	var phases []map[string]any
	save := func() {
		raw, err := json.MarshalIndent(phases, "", "  ")
		must(err)
		must(os.WriteFile(filepath.Join(*out, "phases.json"), raw, 0644))
	}
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		f, err := os.Create(filepath.Join(*out, name+".pprof"))
		must(err)
		must(pprof.StartCPUProfile(f))
		start := time.Now()
		r := fn()
		elapsed := time.Since(start)
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
		if *s == 13 {
			f, err := os.Open(*cache)
			must(err)
			defer f.Close()
			plan, err = liftrs.LoadScalingOnline(f, "f9992fa0a354d1154e73ccf110899abd96fb3f08b532fc2ae8a8aa1cfb0b7c47", 1<<26)
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
	if *s == 13 {
		want = "6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79"
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
			var word []uint32
			phase(fmt.Sprintf("encode-w%d-r%d", count, repeat), func() map[string]any {
				var err error
				var stats liftrs.ScalingCPPStats
				word, stats, err = encoder.EncodeWithStats(message, work)
				must(err)
				return map[string]any{"workers": count, "repeat": repeat, "output_bytes": len(word) * 4, "native_stages": stats}
			})
			phase(fmt.Sprintf("verify-w%d-r%d", count, repeat), func() map[string]any {
				got := digest(word)
				if got != want {
					panic("canonical hash differs: " + got)
				}
				return map[string]any{"canonical_sha256": got, "full_hash_match": true}
			})
			word = nil
		}
		must(encoder.Close())
		work = nil
		encoder = nil
	}
	runtime.KeepAlive(plan)
	save()
}
