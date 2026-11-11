//go:build liftrs_cpp && cgo

package liftrs

// Opt-in s=15 admission experiment: bounded reusable preparation and complete
// native online encoding. Default public discovery guards are unchanged.
// Preparation heap checkpoints force GC; online timings do not.
import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestS15FirstUse(t *testing.T) {
	out, cache := os.Getenv("PIR_S15_OUT"), os.Getenv("PIR_S15_CACHE")
	if out == "" {
		t.Skip("opt-in retained first-use experiment")
	}
	if cache == "" {
		t.Fatal("cache required")
	}
	s, curve := 15, 5
	if os.Getenv("PIR_S15_SMOKE") == "1" {
		s, curve = 6, 4
	}
	repeats := 3
	if s == 6 || os.Getenv("PIR_S15_ONLINE_ONLY") == "1" {
		repeats = 1
	}
	runScalingFirstUse(t, scalingFirstUseConfig{spec: Spec{s, 2, curve}, out: out, cache: cache,
		workers: 16, repeats: repeats, memoryLimit: 24 << 30, onlineOnly: os.Getenv("PIR_S15_ONLINE_ONLY") == "1"})
}

// Private experiment configuration; public encoder defaults remain unchanged.
type scalingFirstUseConfig struct {
	spec              Spec
	out, cache        string
	workers, repeats  int
	memoryLimit       int64
	onlineOnly        bool
	prepareOnly       bool
	rootCheckpoint    bool
	preparationCutoff int                               // Experiment-only override; zero preserves the default.
	prepareExecutor   func(*OrbitBasis) (func(), error) // Optional experimental backend; cleanup follows the run.
	// Optional test harness backend; default remains the native CPU session.
	onlineSetup func(*ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error)
}

func runScalingFirstUse(t *testing.T, cfg scalingFirstUseConfig) {
	spec, out, cache := cfg.spec, cfg.out, cfg.cache
	s, curve := spec.FieldDegree, spec.CurveDegree
	workers, repeats := cfg.workers, cfg.repeats
	debug.SetMemoryLimit(cfg.memoryLimit)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)
	defer signal.Stop(signals)
	go func() { <-signals; pprof.StopCPUProfile(); os.Exit(4) }()
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	check(os.MkdirAll(cache, 0755))
	var phases []map[string]any
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		state, _ := json.Marshal(map[string]any{"phase": name, "event": "started", "timestamp_utc": time.Now().UTC().Format(time.RFC3339Nano)})
		check(os.WriteFile(filepath.Join(out, "state.json"), state, 0644))
		f, err := os.Create(filepath.Join(out, name+".pprof"))
		check(err)
		check(pprof.StartCPUProfile(f))
		stopped := false
		defer func() {
			if !stopped {
				pprof.StopCPUProfile()
				f.Close()
			}
		}()
		nativeBefore, clockErr := liftrscpp.ReadClocks()
		check(clockErr)
		start := time.Now()
		r := fn()
		elapsed := time.Since(start)
		nativeAfter, clockErr := liftrscpp.ReadClocks()
		check(clockErr)
		pprof.StopCPUProfile()
		stopped = true
		check(f.Close())
		runtime.ReadMemStats(&after)
		r["phase"] = name
		r["elapsed_ns"] = elapsed.Nanoseconds()
		r["native_before"] = nativeBefore
		r["native_after"] = nativeAfter
		r["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
		r["heap_before_bytes"] = before.HeapAlloc
		r["heap_after_bytes"] = after.HeapAlloc
		r["go_sys_bytes"] = after.Sys
		phases = append(phases, r)
		data, err := json.MarshalIndent(phases, "", "  ")
		check(err)
		check(os.WriteFile(filepath.Join(out, "phases.json"), data, 0644))
		t.Log(name, elapsed)
	}
	// Fixed-width public files are written atomically and hashed while streaming.
	saveWords := func(name string, values []uint32) map[string]any {
		path := filepath.Join(cache, name)
		file, err := os.Create(path + ".tmp")
		check(err)
		h := sha256.New()
		w := bufio.NewWriterSize(io.MultiWriter(file, h), 1<<20)
		var block [65536]byte
		for start := 0; start < len(values); start += len(block) / 4 {
			count := minInt(len(block)/4, len(values)-start)
			for j := 0; j < count; j++ {
				binary.LittleEndian.PutUint32(block[4*j:], values[start+j])
			}
			_, err := w.Write(block[:4*count])
			check(err)
		}
		check(w.Flush())
		check(file.Sync())
		check(file.Close())
		check(os.Rename(path+".tmp", path))
		return map[string]any{"path": name, "elements": len(values), "bytes": len(values) * 4, "sha256": hex.EncodeToString(h.Sum(nil))}
	}
	saveMeta := func(name string, m map[string]any) {
		raw, err := json.MarshalIndent(m, "", "  ")
		check(err)
		file, err := os.OpenFile(filepath.Join(cache, name+".tmp"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		check(err)
		_, err = file.Write(raw)
		check(err)
		check(file.Sync())
		check(file.Close())
		check(os.Rename(filepath.Join(cache, name+".tmp"), filepath.Join(cache, name)))
	}
	var basis *OrbitBasis
	var plan *ScalingOnlinePlan
	var setup *scalingPreparationCPP
	onlineMeta := filepath.Join(cache, "online-metadata.json")
	if !cfg.onlineOnly {
		var closeExecutor func()
		phase("executor-setup", func() map[string]any {
			var err error
			basis, setup, err = newScalingPreparationCPP(spec, workers)
			check(err)
			if cfg.prepareExecutor != nil {
				closeExecutor, err = cfg.prepareExecutor(basis)
				check(err)
			}
			return map[string]any{"s": s, "m": 2, "t": curve, "workers": workers, "scratch_bytes": len(setup.scratch) * 4}
		})
		defer setup.close()
		if closeExecutor != nil {
			defer closeExecutor()
		}
		basisMeta := filepath.Join(cache, "basis-metadata.json")
		phase("discovery", func() map[string]any {
			raw, err := os.ReadFile(basisMeta)
			if err == nil {
				var meta struct {
					S, T, K, N int
					SHA256     string
				}
				check(json.Unmarshal(raw, &meta))
				if meta.S != s || meta.T != curve || meta.N != basis.Length() || meta.K < 1 || meta.K > meta.N/2 {
					t.Fatal("basis cache spec")
				}
				file, err := os.Open(filepath.Join(cache, "support.bin"))
				check(err)
				defer file.Close()
				h := sha256.New()
				r := bufio.NewReaderSize(io.TeeReader(file, h), 1<<20)
				basis.support = make([]int, meta.K)
				var block [65536]byte
				for start := 0; start < meta.K; start += len(block) / 4 {
					n := minInt(len(block)/4, meta.K-start)
					_, err := io.ReadFull(r, block[:4*n])
					check(err)
					for j := 0; j < n; j++ {
						basis.support[start+j] = int(binary.LittleEndian.Uint32(block[4*j:]))
					}
				}
				if _, err := r.ReadByte(); err != io.EOF {
					t.Fatal("support trailing/read")
				}
				if hex.EncodeToString(h.Sum(nil)) != meta.SHA256 {
					t.Fatal("support checksum")
				}
				check(basis.validateSupport())
				return map[string]any{"reused": true, "k": basis.Dimension(), "n": basis.Length()}
			}
			if !os.IsNotExist(err) {
				check(err)
			}
			basis, err = discoverOrbitBasis(basis, OrbitOptions{Random: rand.New(rand.NewSource(1603)), MaxSamples: 2 + (128+s-1)/s})
			check(err)
			return map[string]any{"reused": false, "k": basis.Dimension(), "n": basis.Length(), "samples": basis.Samples()}
		})
		if _, err := os.Stat(basisMeta); os.IsNotExist(err) {
			phase("save-basis", func() map[string]any {
				file, err := os.Create(filepath.Join(cache, "support.bin.tmp"))
				check(err)
				h := sha256.New()
				w := bufio.NewWriterSize(io.MultiWriter(file, h), 1<<20)
				var block [65536]byte
				for start := 0; start < len(basis.support); start += len(block) / 4 {
					n := minInt(len(block)/4, len(basis.support)-start)
					for j := 0; j < n; j++ {
						binary.LittleEndian.PutUint32(block[4*j:], uint32(basis.support[start+j]))
					}
					_, err := w.Write(block[:4*n])
					check(err)
				}
				check(w.Flush())
				check(file.Sync())
				check(file.Close())
				check(os.Rename(filepath.Join(cache, "support.bin.tmp"), filepath.Join(cache, "support.bin")))
				m := map[string]any{"path": "support.bin", "elements": basis.Dimension(), "bytes": 4 * basis.Dimension(), "sha256": hex.EncodeToString(h.Sum(nil))}
				m["s"] = s
				m["t"] = curve
				m["k"] = basis.Dimension()
				m["n"] = basis.Length()
				m["seed"] = 1603
				saveMeta("basis-metadata.json", m)
				return m
			})
		}
		if basis.Dimension() > basis.Length()/2 {
			t.Fatal("discovered capacity exceeds admitted prefix/memory envelope")
		}
		var prepared *CyclicEncoder
		if _, err := os.Stat(onlineMeta); os.IsNotExist(err) {
			phase("preparation", func() map[string]any {
				var err error
				var checkpoints []map[string]any
				if cfg.preparationCutoff > 0 {
					basis.base.products.cutoff = cfg.preparationCutoff
				}
				var root []uint32
				var persistRoot func([]uint32) error
				if cfg.rootCheckpoint {
					var supportMeta struct{ SHA256 string }
					raw, err := os.ReadFile(basisMeta)
					check(err)
					check(json.Unmarshal(raw, &supportMeta))
					root, err = loadPreparationRoot(cache, spec, basis.Dimension(), basis.Length(), supportMeta.SHA256)
					check(err)
					persistRoot = func(p []uint32) error {
						m := saveWords("root.bin", p)
						m["s"] = s
						m["t"] = curve
						m["k"] = basis.Dimension()
						m["n"] = basis.Length()
						m["support_sha256"] = supportMeta.SHA256
						saveMeta("root-metadata.json", m)
						return nil
					}
				}
				prepared, err = basis.prepareCyclicScalingCheckpointed(basis.extensionFFT, func(stage string, workspace uint64) {
					label := stage
					if strings.HasPrefix(stage, "progress-") {
						label = "root-product"
					}
					state := map[string]any{"phase": label, "progress": stage, "timestamp_utc": time.Now().UTC().Format(time.RFC3339Nano)}
					raw, err := json.Marshal(state)
					check(err)
					check(os.WriteFile(filepath.Join(out, "state.json"), raw, 0644))
					if strings.HasPrefix(stage, "progress-") {
						return
					}
					var before, after runtime.MemStats
					runtime.ReadMemStats(&before)
					gcStart := time.Now()
					runtime.GC()
					gcNS := time.Since(gcStart).Nanoseconds()
					runtime.ReadMemStats(&after)
					name := fmt.Sprintf("preparation-%02d-%s.heap.pprof", len(checkpoints), stage)
					f, err := os.Create(filepath.Join(out, name))
					check(err)
					check(pprof.WriteHeapProfile(f))
					check(f.Close())
					checkpoints = append(checkpoints, map[string]any{"stage": stage, "workspace_bytes": workspace, "heap_before_gc": before.HeapAlloc, "live_heap_after_gc": after.HeapAlloc, "gc_ns": gcNS, "heap_profile": name})
					raw, err = json.MarshalIndent(checkpoints, "", "  ")
					check(err)
					check(os.WriteFile(filepath.Join(out, "checkpoints.json"), raw, 0644))
				}, root, persistRoot)
				check(err)
				return map[string]any{"k": basis.Dimension(), "n": basis.Length(), "identity_verified": true}
			})
			phase("save-prepared", func() map[string]any {
				// Borrow immutable arrays to stream, avoiding an extra P/spectrum copy.
				b := *prepared.basis
				b.support = nil
				b.extensionFFT = newSparseTransform(b.extension)
				b.baseFFT = nil
				b.base.products = nil
				b.extension.products = nil
				plan = &ScalingOnlinePlan{basis: b, k: basis.Dimension(), p: prepared.p, spectrum: prepared.spectrum}
				file, err := os.Create(filepath.Join(cache, "online.bin.tmp"))
				check(err)
				h := sha256.New()
				check(WriteScalingOnline(io.MultiWriter(file, h), plan))
				check(file.Sync())
				check(file.Close())
				check(os.Rename(filepath.Join(cache, "online.bin.tmp"), filepath.Join(cache, "online.bin")))
				info, err := os.Stat(filepath.Join(cache, "online.bin"))
				check(err)
				m := map[string]any{"s": s, "t": curve, "k": basis.Dimension(), "n": basis.Length(), "sha256": hex.EncodeToString(h.Sum(nil)), "bytes": info.Size(), "generator": saveWords("g.bin", prepared.g), "identity_verified": true}
				saveMeta("online-metadata.json", m)
				return m
			})
		}
		// Drop discovery/support/G and preparation scratch before a separately timed
		// online-file load. Reusing the public artifact does not retain old setup RAM.
		prepared = nil
		plan = nil
		basis = nil
		setup.close()
		if closeExecutor != nil {
			closeExecutor()
			closeExecutor = nil
		}
		phase("release-preparation", func() map[string]any { debug.FreeOSMemory(); return map[string]any{"released_to_os": true} })
	}
	if cfg.prepareOnly {
		return
	}
	phase("load-online", func() map[string]any {
		raw, err := os.ReadFile(onlineMeta)
		check(err)
		var m struct{ SHA256 string }
		check(json.Unmarshal(raw, &m))
		file, err := os.Open(filepath.Join(cache, "online.bin"))
		check(err)
		defer file.Close()
		plan, err = LoadScalingOnline(file, m.SHA256, 1<<uint(2*s))
		check(err)
		if plan.FieldDegree() != s || plan.basis.spec.CurveDegree != curve {
			t.Fatal("online cache spec mismatch")
		}
		return map[string]any{"plan_bytes": plan.ResidentBytes(), "k": plan.k, "n": plan.Length()}
	})
	var encode func([]uint32) ([]uint32, map[string]any, error)
	var closeOnline func()
	phase("online-setup", func() map[string]any {
		if cfg.onlineSetup != nil {
			var metadata map[string]any
			var err error
			encode, closeOnline, metadata, err = cfg.onlineSetup(plan)
			check(err)
			return metadata
		}
		encoder, err := plan.WithCPPMatrix(workers)
		check(err)
		workspace := encoder.NewWorkspace()
		encode = func(message []uint32) ([]uint32, map[string]any, error) {
			word, stats, err := encoder.EncodeWithStats(message, workspace)
			return word, map[string]any{"native_stages": stats}, err
		}
		closeOnline = func() { encoder.Close() }
		return map[string]any{"workers": workers, "workspace_bytes": workspace.Bytes()}
	})
	if closeOnline != nil {
		defer closeOnline()
	}
	random := rand.New(rand.NewSource(71))
	message := make([]uint32, plan.k)
	for i := range message {
		message[i] = uint32(random.Intn(1 << uint(s)))
	}
	var digest string
	var word []uint32
	for repeat := 0; repeat < repeats; repeat++ {
		word = nil
		phase(fmt.Sprintf("encode-%d", repeat), func() map[string]any {
			var err error
			var result map[string]any
			word, result, err = encode(message)
			check(err)
			result["input_bytes"] = len(message) * 4
			result["output_bytes"] = len(word) * 4
			return result
		})
		phase(fmt.Sprintf("verify-%d", repeat), func() map[string]any {
			point := plan.basis.omega
			for _, x := range message {
				if word[point] != x {
					t.Fatal("systematic mismatch")
				}
				point = plan.basis.extension.mul(point, plan.basis.omega)
			}
			// Independent carryless base arithmetic for degree-t curve restrictions.
			mul := func(a, b uint32) uint32 {
				var product uint64
				for j := 0; j < s; j++ {
					if b&(1<<uint(j)) != 0 {
						product ^= uint64(a) << uint(j)
					}
				}
				for j := 2*s - 2; j >= s; j-- {
					if product&(1<<uint(j)) != 0 {
						product ^= uint64(plan.basis.polynomial) << uint(j-s)
					}
				}
				return uint32(product)
			}
			r := rand.New(rand.NewSource(1604))
			for trial := 0; trial < 16; trial++ {
				var curves [2][]uint32
				for axis := range curves {
					curves[axis] = make([]uint32, curve+1)
					for j := range curves[axis] {
						curves[axis][j] = uint32(r.Intn(1 << uint(s)))
					}
				}
				var sum uint32
				for z := uint32(0); z < 1<<uint(s); z++ {
					var p [2]uint32
					for axis := range curves {
						for j := curve; j >= 0; j-- {
							p[axis] = mul(p[axis], z) ^ curves[axis][j]
						}
					}
					sum ^= word[p[0]+(1<<uint(s))*p[1]]
				}
				if sum != 0 {
					t.Fatal("curve restriction")
				}
			}
			h := sha256.New()
			var block [65536]byte
			for start := 0; start < len(word); start += len(block) / 4 {
				n := minInt(len(block)/4, len(word)-start)
				for j := 0; j < n; j++ {
					binary.LittleEndian.PutUint32(block[4*j:], word[start+j])
				}
				h.Write(block[:4*n])
			}
			got := hex.EncodeToString(h.Sum(nil))
			if digest != "" && digest != got {
				t.Fatal("repeat hash differs")
			}
			digest = got
			return map[string]any{"canonical_sha256": got, "systematic_checked": len(message), "curves_checked": 16, "repeat_match": true}
		})
	}
	phase("pack-output", func() map[string]any {
		file, err := os.Create(filepath.Join(cache, "word-packed.bin.tmp"))
		check(err)
		h := sha256.New()
		w := bufio.NewWriterSize(io.MultiWriter(file, h), 1<<20)
		var acc uint64
		var used uint
		var count int64
		for _, x := range word {
			acc |= uint64(x) << used
			used += uint(s)
			for used >= 8 {
				check(w.WriteByte(byte(acc)))
				acc >>= 8
				used -= 8
				count++
			}
		}
		if used != 0 {
			check(w.WriteByte(byte(acc)))
			count++
		}
		check(w.Flush())
		check(file.Sync())
		check(file.Close())
		check(os.Rename(filepath.Join(cache, "word-packed.bin.tmp"), filepath.Join(cache, "word-packed.bin")))
		if count != int64((len(word)*s+7)/8) {
			t.Fatal("packed length")
		}
		m := map[string]any{"packed_bytes": count, "sha256": hex.EncodeToString(h.Sum(nil)), "canonical_sha256": digest, "seed": 71}
		saveMeta("word-metadata.json", m)
		return m
	})
	phase("verify-packed", func() map[string]any {
		file, err := os.Open(filepath.Join(cache, "word-packed.bin"))
		check(err)
		defer file.Close()
		r := rand.New(rand.NewSource(1605))
		positions := []int{0, len(word) / 2, len(word) - 1}
		for j := 0; j < 1024; j++ {
			positions = append(positions, r.Intn(len(word)))
		}
		for _, index := range positions {
			bit := uint64(index) * uint64(s)
			shift := uint(bit % 8)
			count := (s + int(shift) + 7) / 8
			var data [4]byte
			_, err := file.ReadAt(data[:count], int64(bit/8))
			check(err)
			var x uint32
			for j := 0; j < count; j++ {
				x |= uint32(data[j]) << uint(8*j)
			}
			x = (x >> shift) & ((1 << uint(s)) - 1)
			if x != word[index] {
				t.Fatal("packed readback")
			}
		}
		return map[string]any{"positions_checked": len(positions), "readback_match": true}
	})
	runtime.KeepAlive(encode)
	runtime.KeepAlive(closeOnline)
	runtime.KeepAlive(word)
}
