//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/gpu-matrix"
	"io"
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

func TestS16ResidentIntervals(t *testing.T) {
	out := os.Getenv("PIR_S16_INTERVAL_OUT")
	if out == "" {
		t.Skip("opt-in bounded s16 intervals")
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
	go func() { <-signals; pprof.StopCPUProfile(); os.Exit(4) }()
	var stages []map[string]any
	save := func(name string, value any) {
		raw, err := json.MarshalIndent(value, "", "  ")
		check(err)
		check(os.WriteFile(filepath.Join(out, name), raw, 0644))
	}
	phase := func(name string, fn func() map[string]any) {
		runtime.GC()
		f, err := os.Create(filepath.Join(out, name+".pprof"))
		check(err)
		check(pprof.StartCPUProfile(f))
		start := time.Now()
		data := fn()
		seconds := time.Since(start).Seconds()
		pprof.StopCPUProfile()
		check(f.Close())
		data["phase"], data["seconds"] = name, seconds
		stages = append(stages, data)
		save("stages.json", stages)
		t.Log(name, seconds)
	}
	const prefix = 1 << 20
	var support []int
	phase("support-census", func() map[string]any {
		cache := os.Getenv("PIR_S16_SUPPORT")
		raw, err := os.ReadFile(filepath.Join(cache, "basis-metadata.json"))
		check(err)
		var meta struct {
			S, T, K, N, Bytes int
			SHA256            string
		}
		check(json.Unmarshal(raw, &meta))
		if meta.S != 16 || meta.T != 5 || meta.K != 918672421 || meta.N != 1<<32 || meta.Bytes != 4*meta.K || meta.SHA256 != os.Getenv("PIR_S16_SUPPORT_SHA256") {
			t.Fatal("support provenance")
		}
		file, err := os.Open(filepath.Join(cache, "support.bin"))
		check(err)
		defer file.Close()
		h := sha256.New()
		r := io.TeeReader(file, h)
		var block [65536]byte
		last := -1
		accepted, fixed, totalDegree, degree, count := 0, 0, 0, 0, 0
		var batches, fixedOrdinals []int
		for start := 0; start < meta.K; start += len(block) / 4 {
			n := minInt(len(block)/4, meta.K-start)
			_, err := io.ReadFull(r, block[:4*n])
			check(err)
			for j := 0; j < n; j++ {
				v := int(binary.LittleEndian.Uint32(block[4*j:]))
				if v <= last || v >= meta.N-1 {
					t.Fatal("support order/range")
				}
				last = v
				if len(support) < prefix {
					support = append(support, v)
				}
				partner := int(uint64(v) * 65536 % uint64(meta.N-1))
				if partner < v {
					continue
				}
				accepted++
				count++
				d := 2
				if partner == v {
					d = 1
					fixed++
					fixedOrdinals = append(fixedOrdinals, accepted-1)
				}
				degree += d
				totalDegree += d
				if count == 32768 {
					batches = append(batches, degree)
					degree = 0
					count = 0
				}
			}
		}
		var extra [1]byte
		n, err := r.Read(extra[:])
		if n != 0 || err != io.EOF {
			t.Fatal("support trailing bytes")
		}
		if hex.EncodeToString(h.Sum(nil)) != meta.SHA256 || totalDegree != meta.K {
			t.Fatal("support checksum/orbit degree")
		}
		if count > 0 {
			batches = append(batches, degree)
		}
		census := map[string]any{"s": 16, "t": 5, "k": meta.K, "n": meta.N, "support_sha256": meta.SHA256, "accepted_factors": accepted, "fixed_factors": fixed, "fixed_factor_ordinals": fixedOrdinals, "batch_factors": 32768, "batch_degrees": batches, "final_batch_factors": count, "prefix_positions": len(support), "limits": "Census only; later polynomial lengths are not allocated. Verified saved discovery, not a new independent discovery proof."}
		save("census.json", census)
		return map[string]any{"positions": meta.K, "accepted_factors": accepted, "fixed_factors": fixed, "batches": len(batches), "prefix_positions": len(support)}
	})
	base, ext, _, omega := orbitFields(16, 0x1002d)
	table := newPreparationProducts(base.base)
	var reference []uint32
	for _, backend := range []string{"cpu", "resident", "resident-4096"} {
		var executor *gpuPreparationExecutor
		native := &scalingPreparationCPP{plans: make(map[int]*matrixCPPPlan), workers: 4}
		fft := newSparseTransform(ext)
		var engine *gpumatrix.Engine
		phase(backend+"-setup", func() map[string]any {
			for dim := 2; dim <= 22; dim++ {
				p, err := newMatrixCPP(ext, dim)
				check(err)
				native.plans[1<<uint(dim)] = p
			}
			native.scratch = make([]uint32, 1<<22)
			fft.executor = native
			if backend != "cpu" {
				var err error
				engine, err = gpumatrix.New(os.Getenv("PIR_GPU_LIBRARY"), 16, 0x1002d, ext.extension.Delta(), table.log, table.powers)
				check(err)
				for dim := 12; dim <= 22; dim++ {
					words, _, _, err := gpuMatrixDescription(ext, dim)
					check(err)
					check(engine.AddPlan(words))
				}
				executor = &gpuPreparationExecutor{gpu: engine, fallback: native, resident: true}
				fft.executor = executor
			}
			return map[string]any{"workers": 4, "maximum_transform_points": 1 << 22, "cpu_transpose_bytes": len(native.scratch) * 4}
		})
		field := base
		field.lookup = table
		cutoff := 32768
		if backend == "resident" {
			cutoff = 1024
		}
		if backend == "resident-4096" {
			cutoff = 4096
		}
		workspace := &productWorkspace{}
		field.products = &polynomialProducts{transform: fft, cutoff: cutoff, workspace: workspace}
		builder := newFactorProductBuilder(field, 32768)
		totalFactors, totalDegree := 0, 0
		for interval := 0; interval < 2; interval++ {
			begin, end := interval*prefix/2, (interval+1)*prefix/2
			phase(fmt.Sprintf("%s-interval-%d", backend, interval), func() map[string]any {
				beforeFactors, beforeDegree := totalFactors, totalDegree
				for index := begin; index < end; index++ {
					i := support[index]
					partner := int(uint64(i) * base.q % uint64(ext.q-1))
					if partner < i {
						continue
					}
					root := ext.inv(ext.pow(omega, i))
					totalFactors++
					if partner == i {
						if uint64(root) >= base.q {
							t.Fatal("fixed root")
						}
						builder.add(root, 1, 0, 1)
						totalDegree++
					} else {
						conjugate := ext.pow(root, int(base.q))
						trace, norm := root^conjugate, ext.mul(root, conjugate)
						if uint64(trace) >= base.q || uint64(norm) >= base.q {
							t.Fatal("paired root")
						}
						builder.add(norm, trace, 1, 2)
						totalDegree += 2
					}
				}
				return map[string]any{"start_position": begin, "end_position": end, "factors_added": totalFactors - beforeFactors, "degree_added": totalDegree - beforeDegree, "product_cutoff": cutoff}
			})
		}
		var product []uint32
		phase(backend+"-finish", func() map[string]any {
			product = builder.finish()
			if len(product) != totalDegree+1 {
				t.Fatal("partial degree")
			}
			return map[string]any{"coefficients": len(product), "factors": totalFactors, "limits": "Finalizes this partial prefix only; not part of the two interval clocks or a complete code."}
		})
		if backend == "cpu" {
			reference = append([]uint32(nil), product...)
		} else if !reflect.DeepEqual(product, reference) {
			t.Fatal("partial CPU/resident product mismatch")
		}
		path := filepath.Join(os.Getenv("PIR_S16_INTERVAL_CACHE"), "partial-"+backend+".bin")
		file, err := os.Create(path)
		check(err)
		h := sha256.New()
		w := bufio.NewWriterSize(io.MultiWriter(file, h), 1<<20)
		check(binary.Write(w, binary.LittleEndian, product))
		check(w.Flush())
		check(file.Sync())
		check(file.Close())
		info := map[string]any{"coefficients": len(product), "sha256": hex.EncodeToString(h.Sum(nil)), "prefix_positions": prefix, "product_cutoff": cutoff, "scope": "monic partial root product; not normalized P or a complete code cache"}
		if engine != nil {
			info["product_calls"] = executor.products
			info["stage_seconds"] = executor.stats
			info["owned_device_bytes"] = engine.OwnedBytes()
			engine.Close()
			if engine.OwnedBytes() != 0 {
				t.Fatal("partial engine close")
			}
		}
		save(backend+"-partial.json", info)
		workspace.release()
		native.close()
	}
}
