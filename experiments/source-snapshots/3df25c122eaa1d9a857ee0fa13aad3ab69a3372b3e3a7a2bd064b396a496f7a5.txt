//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"testing"
	"time"
)

// Reuse existing public preprocessing; no 30-minute preparation or Go baseline
// rerun. One profiled prototype encoding must match the complete retained hash.
func TestScalingS13(t *testing.T) {
	dir := os.Getenv("PIR_SCALING_PROBE")
	if dir == "" {
		t.Skip("opt-in s13 validation")
	}
	source := os.Getenv("PIR_SCALING_S13_CACHE")
	if source == "" {
		t.Fatal("s13 cache required")
	}
	debug.SetMemoryLimit(9 << 30)
	records := []map[string]any{}
	stage := func(name string, fn func() map[string]any) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		file, err := os.Create(filepath.Join(dir, name+".pprof"))
		if err != nil {
			t.Fatal(err)
		}
		if err := pprof.StartCPUProfile(file); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		record := fn()
		elapsed := time.Since(start)
		pprof.StopCPUProfile()
		file.Close()
		runtime.ReadMemStats(&after)
		record["phase"] = name
		record["elapsed_ns"] = elapsed.Nanoseconds()
		record["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
		record["heap_before_bytes"] = before.HeapAlloc
		record["heap_after_bytes"] = after.HeapAlloc
		record["go_sys_bytes"] = after.Sys
		records = append(records, record)
		raw, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "s13-phases.json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
		t.Log(name, elapsed)
	}
	var e *CyclicEncoder
	stage("s13-load-original", func() map[string]any {
		file, err := os.Open(source)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		raw, err := io.ReadAll(io.LimitReader(file, (2<<30)+1))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(raw)
		if hex.EncodeToString(digest[:]) != "a2a3bcd5913d821c99d08ceca05d6c5d73ce2abe6e3987910c0dcf5dfe97eb22" {
			t.Fatal("retained cache changed")
		}
		e, err = LoadCyclicEncoderWithLimit(raw, 2<<30)
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{"cache_bytes": len(raw), "sha256": hex.EncodeToString(digest[:])}
	})
	var plan *ScalingOnlinePlan
	var native cppPlan
	stage("s13-online-plan", func() map[string]any {
		var err error
		plan, err = e.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		native, _, err = makeCPPPlan(plan.basis.extensionFFT)
		if err != nil {
			t.Fatal(err)
		}
		e = nil // Subsequent phase GC can discard G/support/original spectrum.
		return map[string]any{"plan_payload_bytes": plan.ResidentBytes(), "native_bytes": native.OwnedBytes()}
	})
	defer native.Close()
	var stats CPPEncodeStats
	plan.basis.extensionFFT.executor = &meteredCPP{native, &stats}
	random := rand.New(rand.NewSource(71))
	message := make([]uint32, plan.k)
	for i := range message {
		message[i] = uint32(random.Intn(int(plan.basis.base.q)))
	}
	var workspace *ScalingWorkspace
	stage("s13-workspace", func() map[string]any {
		workspace = plan.NewWorkspace()
		return map[string]any{"workspace_payload_bytes": workspace.Bytes(), "input_payload_bytes": len(message) * 4}
	})
	var word []uint32
	stage("s13-native-workspace-encode", func() map[string]any {
		var err error
		word, err = plan.Encode(message, workspace)
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{"native": stats, "output_payload_bytes": len(word) * 4}
	})
	stage("s13-verify", func() map[string]any {
		hash := sha256.New()
		var buf [64 << 10]byte
		for start := 0; start < len(word); start += len(buf) / 4 {
			count := minInt(len(buf)/4, len(word)-start)
			for i := 0; i < count; i++ {
				binary.LittleEndian.PutUint32(buf[4*i:], word[start+i])
			}
			hash.Write(buf[:4*count])
		}
		digest := hex.EncodeToString(hash.Sum(nil))
		if digest != "6120f0388f6035dd480e391b26ef93c04ed368a77f2f93c02f6d4f56b99a7b79" {
			t.Fatal("prototype differs from retained complete codeword")
		}
		point := plan.basis.omega
		for _, x := range message {
			if word[point] != x {
				t.Fatal("systematic position differs")
			}
			point = plan.basis.extension.mul(point, plan.basis.omega)
		}
		return map[string]any{"canonical_sha256": digest, "systematic_positions": len(message), "matches_original_baseline": true}
	})
	runtime.KeepAlive(workspace)
	runtime.KeepAlive(plan)
	runtime.KeepAlive(message)
	runtime.KeepAlive(word)
}
