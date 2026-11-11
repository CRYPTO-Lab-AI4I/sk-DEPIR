//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"syscall"
	"testing"
	"time"
)

func gpuScalingDigest(data []uint32) string {
	h := sha256.New()
	var buf [4096]byte
	for offset := 0; offset < len(data); {
		count := minInt(len(buf)/4, len(data)-offset)
		for i, v := range data[offset : offset+count] {
			binary.LittleEndian.PutUint32(buf[4*i:], v)
		}
		h.Write(buf[:4*count])
		offset += count
	}
	return hex.EncodeToString(h.Sum(nil))
}

func TestGPUScaling(t *testing.T) {
	out := os.Getenv("PIR_GPU_SCALING_OUT")
	if out == "" {
		t.Skip("opt-in bounded GPU scaling")
	}
	debug.SetMemoryLimit(4 << 30)
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
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
	save := func(name string, value any) {
		raw, err := json.MarshalIndent(value, "", "  ")
		check(err)
		check(os.WriteFile(filepath.Join(out, name), raw, 0644))
	}
	base, f, _, _ := orbitFields(16, 0x1002d)
	table := newPreparationProducts(base.base)
	library := os.Getenv("PIR_GPU_LIBRARY")
	first, last, deviceCap := 22, 25, uint64(512<<20)
	if stage := os.Getenv("PIR_GPU_SCALING_STAGE"); stage == "large" {
		first, last, deviceCap = 26, 28, 4<<30
	} else if stage != "" {
		t.Fatal("unknown scaling stage")
	}
	start := time.Now()
	engine, err := gpumatrix.NewBounded(library, 16, 0x1002d, f.extension.Delta(), table.log, table.powers, 1<<uint(last), deviceCap)
	check(err)
	defer engine.Close()
	setup := time.Since(start).Seconds()
	before := engine.OwnedBytes()
	wide, err := gpuMatrixDescriptionV2(f, 32)
	check(err)
	check(engine.CheckPlan(wide))
	if engine.OwnedBytes() != before {
		t.Fatal("metadata check allocated device memory")
	}
	if engine.AddPlan(wide) == nil || engine.OwnedBytes() != before {
		t.Fatal("full domain admitted outside budget")
	}
	for _, bad := range [][]uint32{wide[:len(wide)-1], {0x324d474c, 0, 2, 16, 32}, append(append([]uint32(nil), wide...), 0)} {
		if engine.CheckPlan(bad) == nil {
			t.Fatal("invalid wide metadata accepted")
		}
	}
	if _, err := gpumatrix.NewBounded(library, 16, 0x1002d, f.extension.Delta(), table.log, table.powers, 1<<26, 512<<20); err == nil {
		t.Fatal("underfunded engine admitted")
	}
	save("metadata.json", map[string]any{"domain": uint64(1) << 32, "row_levels": 17, "device_before": before, "device_after": engine.OwnedBytes(), "full_plan_rejected_by_budget": true, "malformed_rejected": true, "underfunded_rejected": true, "setup_seconds": setup})
	if os.Getenv("PIR_GPU_SCALING_SMALL") == "1" {
		return
	}
	var rows []map[string]any
	for dimension := first; dimension <= last; dimension++ {
		func() {
			runtime.GC()
			profile, err := os.Create(filepath.Join(out, fmt.Sprintf("dimension-%d.pprof", dimension)))
			check(err)
			check(pprof.StartCPUProfile(profile))
			defer func() { pprof.StopCPUProfile(); check(profile.Close()) }()
			n := 1 << uint(dimension)
			description, err := gpuMatrixDescriptionV2(f, dimension)
			check(err)
			raw, err := json.Marshal(description)
			check(err)
			check(os.WriteFile(filepath.Join(os.Getenv("PIR_GPU_SCALING_CACHE"), fmt.Sprintf("plan-%d.json", dimension)), raw, 0644))
			start := time.Now()
			check(engine.AddPlan(description))
			gpuSetup := time.Since(start).Seconds()
			start = time.Now()
			plan, err := newMatrixCPP(f, dimension)
			check(err)
			cpu := &scalingPreparationCPP{plans: map[int]*matrixCPPPlan{n: plan}, scratch: make([]uint32, n), workers: 16}
			defer cpu.close()
			cpuSetup := time.Since(start).Seconds()
			rng := rand.New(rand.NewSource(int64(1900 + dimension)))
			input := make([]uint32, n)
			for i := range input {
				input[i] = rng.Uint32()
			}
			want := append([]uint32(nil), input...)
			start = time.Now()
			cpu.apply(want, false)
			cpuForward := time.Since(start).Seconds()
			forwardHash := gpuScalingDigest(want)
			data := make([]uint32, n)
			var forwards, inverses, products [][8]float64
			for repeat := 0; repeat < 2; repeat++ {
				copy(data, input)
				stats, err := engine.Apply(data, false)
				check(err)
				if !reflect.DeepEqual(data, want) {
					t.Fatal("GPU forward mismatch", dimension)
				}
				forwards = append(forwards, stats)
				stats, err = engine.Apply(data, true)
				check(err)
				if !reflect.DeepEqual(data, input) {
					t.Fatal("GPU inverse mismatch", dimension)
				}
				inverses = append(inverses, stats)
			}
			start = time.Now()
			cpu.apply(want, true)
			cpuInverse := time.Since(start).Seconds()
			if !reflect.DeepEqual(want, input) {
				t.Fatal("CPU inverse mismatch")
			}
			// Unequal operands fill a full padded domain. CPU uses the existing native
			// transforms plus canonical Go field multiplication, separately identified.
			a, b := input[:n/2+1], input[n/2+1:]
			clear(want)
			clear(data)
			copy(want, a)
			copy(data, b)
			start = time.Now()
			cpu.apply(want, false)
			cpu.apply(data, false)
			for i := range want {
				want[i] = f.mul(want[i], data[i])
			}
			cpu.apply(want, true)
			cpuProduct := time.Since(start).Seconds()
			productHash := gpuScalingDigest(want[:n-1])
			for repeat := 0; repeat < 2; repeat++ {
				got, stats, err := engine.Product(a, b, n-1, n)
				check(err)
				if !reflect.DeepEqual(got, want[:n-1]) {
					t.Fatal("GPU product mismatch", dimension)
				}
				products = append(products, stats)
			}
			row := map[string]any{"dimension": dimension, "n": n, "cpu_workers": 16, "gpu_setup_seconds": gpuSetup, "cpu_setup_seconds": cpuSetup, "cpu_forward_seconds": cpuForward, "cpu_inverse_seconds": cpuInverse, "cpu_product_seconds": cpuProduct, "gpu_forward": forwards, "gpu_inverse": inverses, "gpu_products": products, "forward_sha256": forwardHash, "product_sha256": productHash, "owned_device_bytes": engine.OwnedBytes(), "match": true}
			rows = append(rows, row)
			save("measurements.json", rows)
			t.Logf("dimension%d CPU forward %.3fs, GPU %.3fs, device %.1f MiB", dimension, cpuForward, forwards[0][7], float64(engine.OwnedBytes())/(1<<20))
		}()
	}
	engine.Close()
	if engine.OwnedBytes() != 0 {
		t.Fatal("device state retained")
	}
	save("closed.json", map[string]any{"owned_device_bytes": engine.OwnedBytes()})
}
