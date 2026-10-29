//go:build liftrs_cpp && cgo

package liftrs

// Opt-in fresh-process resource probes. Kept beside private online-file helpers;
// they do not expose a new public artifact API or modify any normal encoder path.
import (
	"crypto/sha256"
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

func TestCalibrationResource(t *testing.T) {
	mode := os.Getenv("PIR_CALIBRATION_RESOURCE")
	if mode == "" {
		t.Skip("opt-in resource calibration")
	}
	out, cache := os.Getenv("PIR_CALIBRATION_OUT"), os.Getenv("PIR_CALIBRATION_CACHE")
	if out == "" || cache == "" {
		t.Fatal("out/cache required")
	}
	debug.SetMemoryLimit(9 << 30)
	if err := os.MkdirAll(cache, 0755); err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		f, err := os.Create(filepath.Join(out, name+".pprof"))
		if err != nil {
			t.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		data := fn()
		elapsed := time.Since(start)
		pprof.StopCPUProfile()
		f.Close()
		runtime.ReadMemStats(&after)
		data["phase"] = name
		data["elapsed_ns"] = elapsed.Nanoseconds()
		data["heap_after_bytes"] = after.HeapAlloc
		data["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
		data["go_sys_bytes"] = after.Sys
		events = append(events, data)
		raw, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "phases.json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
		t.Log(name, elapsed)
	}
	switch mode {
	case "basis-s10":
		var basis *OrbitBasis
		phase("s10-discovery", func() map[string]any {
			var err error
			basis, err = DiscoverOrbit(Spec{10, 2, 6}, OrbitOptions{Random: rand.New(rand.NewSource(1002))})
			if err != nil {
				t.Fatal(err)
			}
			raw, err := basis.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(cache, "s10-basis.json"), raw, 0644); err != nil {
				t.Fatal(err)
			}
			return map[string]any{"k": basis.Dimension(), "n": basis.Length(), "bytes": len(raw)}
		})
	case "baseline-s10", "prototype-s10":
		raw, err := os.ReadFile(filepath.Join(cache, "s10-basis.json"))
		if err != nil {
			t.Fatal(err)
		}
		basis, err := LoadOrbit(raw)
		if err != nil {
			t.Fatal(err)
		}
		var e *CyclicEncoder
		phase(mode, func() map[string]any {
			var err error
			if mode == "baseline-s10" {
				e, err = basis.PrepareCyclicFFT()
			} else {
				e, err = basis.PrepareCyclicScalingPrototype()
			}
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{"k": basis.Dimension(), "n": basis.Length()}
		})
		// Serialization/hash are outside preparation timing; keep reusable small setup.
		raw, err = e.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(raw)
		if err := os.WriteFile(filepath.Join(cache, mode+".json"), raw, 0644); err != nil {
			t.Fatal(err)
		}
		events = append(events, map[string]any{"prepared_sha256": hex.EncodeToString(digest[:]), "bytes": len(raw)})
		data, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "phases.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
		runtime.KeepAlive(e)
	case "export-s13":
		original := os.Getenv("PIR_CALIBRATION_S13")
		f, err := os.Open(original)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		phase(mode, func() map[string]any {
			raw, err := io.ReadAll(io.LimitReader(f, (2<<30)+1))
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(raw)
			if hex.EncodeToString(digest[:]) != "a2a3bcd5913d821c99d08ceca05d6c5d73ce2abe6e3987910c0dcf5dfe97eb22" {
				t.Fatal("original cache checksum")
			}
			e, err := LoadCyclicEncoderWithLimit(raw, 2<<30)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := e.ScalingPlan()
			if err != nil {
				t.Fatal(err)
			}
			output, err := os.Create(filepath.Join(cache, "s13-online.bin"))
			if err != nil {
				t.Fatal(err)
			}
			hash := sha256.New()
			if err := writeScalingOnline(io.MultiWriter(output, hash), plan); err != nil {
				t.Fatal(err)
			}
			if err := output.Sync(); err != nil {
				t.Fatal(err)
			}
			output.Close()
			metadata := map[string]any{"s": 13, "k": plan.k, "n": plan.basis.Length(), "sha256": hex.EncodeToString(hash.Sum(nil)), "source_sha256": hex.EncodeToString(digest[:]), "source": "retained verified s13 cyclic JSON"}
			metadataBytes, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(cache, "s13-online-metadata.json"), metadataBytes, 0644); err != nil {
				t.Fatal(err)
			}
			return metadata
		})
	case "minimal-load-s13":
		raw, err := os.ReadFile(filepath.Join(cache, "s13-online-metadata.json"))
		if err != nil {
			t.Fatal(err)
		}
		var meta struct{ SHA256 string }
		if err := json.Unmarshal(raw, &meta); err != nil {
			t.Fatal(err)
		}
		var plan *ScalingOnlinePlan
		phase(mode, func() map[string]any {
			f, err := os.Open(filepath.Join(cache, "s13-online.bin"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			plan, err = loadScalingOnline(f, meta.SHA256, 1<<26)
			if err != nil {
				t.Fatal(err)
			}
			return map[string]any{"k": plan.k, "n": plan.basis.Length(), "plan_array_bytes": plan.ResidentBytes(), "support_elements": len(plan.basis.support)}
		})
		runtime.KeepAlive(plan)
	default:
		t.Fatal("unknown calibration mode", mode)
	}
}
