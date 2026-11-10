//go:build liftrs_cpp && cgo

// local-calibration measures independent native tiles and bounded file traffic.
// It is not a parallel full-domain encoder or an external-memory implementation.
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
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func emit(v any) { check(json.NewEncoder(os.Stdout).Encode(v)) }
func measure(dir, name string, fn func() map[string]any) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	profile, err := os.Create(filepath.Join(dir, name+".pprof"))
	check(err)
	check(pprof.StartCPUProfile(profile))
	emit(map[string]any{"started": name})
	start := time.Now()
	r := fn()
	elapsed := time.Since(start)
	pprof.StopCPUProfile()
	check(profile.Close())
	runtime.ReadMemStats(&after)
	r["phase"] = name
	r["elapsed_ns"] = elapsed.Nanoseconds()
	r["allocated_bytes"] = after.TotalAlloc - before.TotalAlloc
	r["heap_after_bytes"] = after.HeapAlloc
	r["go_sys_bytes"] = after.Sys
	emit(r)
}
func digest(values []uint32) string {
	h := sha256.New()
	var b [64 << 10]byte
	for start := 0; start < len(values); start += len(b) / 4 {
		count := len(b) / 4
		if count > len(values)-start {
			count = len(values) - start
		}
		for i := 0; i < count; i++ {
			binary.LittleEndian.PutUint32(b[4*i:], values[start+i])
		}
		h.Write(b[:4*count])
	}
	return hex.EncodeToString(h.Sum(nil))
}
func transform(plan *native.Plan, tiles [][]uint32, workers int) {
	jobs := make(chan []uint32)
	var group sync.WaitGroup
	errors := make(chan error, workers)
	for w := 0; w < workers; w++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for tile := range jobs {
				if _, err := plan.Apply(tile, false); err != nil {
					errors <- err
					return
				}
			}
		}()
	}
	for _, tile := range tiles {
		jobs <- tile
	}
	close(jobs)
	group.Wait()
	close(errors)
	for err := range errors {
		check(err)
	}
}
func decode(raw []byte, values []uint32) {
	for i := range values {
		values[i] = binary.LittleEndian.Uint32(raw[4*i:])
	}
}
func encode(raw []byte, values []uint32) {
	for i, x := range values {
		binary.LittleEndian.PutUint32(raw[4*i:], x)
	}
}
func main() {
	dir := flag.String("out", "", "fresh profile/result directory")
	cache := flag.String("cache", "", "retained file/plan directory")
	quick := flag.Bool("quick", false, "small smoke only")
	cpuOnly := flag.Bool("cpu-only", false, "measure tiles only, without storage/pipeline")
	elementsFlag := flag.Int("tile-elements", 1<<20, "power-of-two elements per independent native tile")
	countFlag := flag.Int("tiles", 16, "independent tile count, bounded total working set")
	flag.Parse()
	if *dir == "" || *cache == "" {
		panic("out/cache required")
	}
	check(os.MkdirAll(*dir, 0755))
	check(os.MkdirAll(*cache, 0755))
	debug.SetMemoryLimit(2 << 30)
	var plan *native.Plan
	measure(*dir, "native-setup", func() map[string]any {
		words, err := liftrs.SparseTransformProbe(13, true)
		check(err)
		plan, err = native.NewPlan(words)
		check(err)
		raw := make([]byte, 4*len(words))
		for i, x := range words {
			binary.LittleEndian.PutUint32(raw[4*i:], x)
		}
		check(os.WriteFile(filepath.Join(*cache, "transform.bin"), raw, 0644))
		return map[string]any{"native_owned_bytes": plan.OwnedBytes(), "s": 13}
	})
	defer plan.Close()
	tileElements, tileCount, repeats := *elementsFlag, *countFlag, 3
	if tileElements < 2 || tileElements&(tileElements-1) != 0 || tileCount < 1 || uint64(tileElements)*uint64(tileCount) > 1<<24 {
		panic("invalid/budget-exceeding tile dimensions")
	}
	if *quick {
		tileElements, tileCount, repeats = 256, 4, 1
	}
	inputs, work := make([][]uint32, tileCount), make([][]uint32, tileCount)
	r := rand.New(rand.NewSource(1001))
	for j := range inputs {
		inputs[j] = make([]uint32, tileElements)
		work[j] = make([]uint32, tileElements)
		for i := range inputs[j] {
			inputs[j][i] = uint32(r.Intn(1 << 26))
		}
	}
	expected := make([]string, tileCount)
	for _, workers := range []int{1, 2, 4, 8, 16} {
		if *quick && workers > 4 {
			continue
		}
		runtime.GOMAXPROCS(workers)
		for repeat := 0; repeat < repeats; repeat++ {
			for j := range inputs {
				copy(work[j], inputs[j])
			}
			measure(*dir, fmt.Sprintf("cpu-w%d-r%d", workers, repeat), func() map[string]any {
				transform(plan, work, workers)
				return map[string]any{"workers": workers, "repeat": repeat, "tiles": tileCount, "tile_elements": tileElements, "input_bytes": 4 * tileCount * tileElements}
			})
			for j := range work {
				got := digest(work[j])
				if workers == 1 && repeat == 0 {
					expected[j] = got
				} else if got != expected[j] {
					panic("parallel transform differs")
				}
			}
		}
	}
	runtime.GOMAXPROCS(8)
	var storageChecked int64
	if !*cpuOnly {
		// A deterministic initialized file; buffer allocation and hashing are outside
		// timed transfers. F_NOCACHE is requested explicitly by platform helper.
		fileBytes := int64(256 << 20)
		if *quick {
			fileBytes = 1 << 20
		}
		storageChecked = fileBytes
		pattern := make([]byte, 4<<20)
		for i := range pattern {
			pattern[i] = byte((i*17 + i/4096) % 251)
		}
		for _, block := range []int{4096, 64 << 10, 1 << 20, 4 << 20} {
			for repeat := 0; repeat < repeats; repeat++ {
				path := filepath.Join(*cache, fmt.Sprintf("storage-b%d.bin", block))
				f, err := os.Create(path)
				check(err)
				policy, err := noCache(f)
				check(err)
				buffer := pattern[:block]
				measure(*dir, fmt.Sprintf("write-b%d-r%d", block, repeat), func() map[string]any {
					var calls uint64
					for at := int64(0); at < fileBytes; at += int64(block) {
						n := block
						if int64(n) > fileBytes-at {
							n = int(fileBytes - at)
						}
						count, err := f.Write(buffer[:n])
						check(err)
						if count != n {
							panic("short write")
						}
						calls++
					}
					return map[string]any{"bytes": fileBytes, "calls": calls, "block_bytes": block, "cache_policy": policy, "sync_included": false}
				})
				measure(*dir, fmt.Sprintf("sync-b%d-r%d", block, repeat), func() map[string]any {
					check(f.Sync())
					return map[string]any{"durability": "file fsync; not drive-cache F_FULLFSYNC"}
				})
				check(f.Close())
				f, err = os.Open(path)
				check(err)
				_, err = noCache(f)
				check(err)
				measure(*dir, fmt.Sprintf("read-b%d-r%d", block, repeat), func() map[string]any {
					var calls uint64
					for at := int64(0); at < fileBytes; at += int64(block) {
						n := block
						if int64(n) > fileBytes-at {
							n = int(fileBytes - at)
						}
						_, err := io.ReadFull(f, buffer[:n])
						check(err)
						calls++
					}
					return map[string]any{"bytes": fileBytes, "calls": calls, "block_bytes": block, "cache_policy": policy}
				})
				check(f.Close())
				// Independent validation outside timing, all bytes checked against the
				// pattern actually repeated per block, not only a final read buffer.
				f, err = os.Open(path)
				check(err)
				verify := make([]byte, block)
				for at := int64(0); at < fileBytes; at += int64(block) {
					n := block
					if int64(n) > fileBytes-at {
						n = int(fileBytes - at)
					}
					_, err := io.ReadFull(f, verify[:n])
					check(err)
					for i, x := range verify[:n] {
						if x != pattern[i] {
							panic("storage content mismatch")
						}
					}
				}
				check(f.Close())
			}
		}
		// Real read -> native FFT -> write pipeline, bounded slots. It measures tile
		// work only, not the matrix conversion/transposes of the external encoder.
		inputPath := filepath.Join(*cache, "tile-input.bin")
		input, err := os.Create(inputPath)
		check(err)
		raw := make([]byte, 4*tileElements)
		for _, tile := range inputs {
			encode(raw, tile)
			_, err = input.Write(raw)
			check(err)
		}
		check(input.Sync())
		check(input.Close())
		for _, slots := range []int{1, 2, 4} {
			for repeat := 0; repeat < repeats; repeat++ {
				input, err = os.Open(inputPath)
				check(err)
				policy, err := noCache(input)
				check(err)
				outputPath := filepath.Join(*cache, fmt.Sprintf("pipeline-slots%d.bin", slots))
				output, err := os.Create(outputPath)
				check(err)
				_, err = noCache(output)
				check(err)
				type slot struct {
					raw    []byte
					values []uint32
					index  int
				}
				pool := make(chan *slot, slots)
				for j := 0; j < slots; j++ {
					pool <- &slot{raw: make([]byte, 4*tileElements), values: make([]uint32, tileElements)}
				}
				measure(*dir, fmt.Sprintf("pipeline-slots%d-r%d", slots, repeat), func() map[string]any {
					if slots == 1 {
						s := <-pool
						for j := 0; j < tileCount; j++ {
							_, err := io.ReadFull(input, s.raw)
							check(err)
							decode(s.raw, s.values)
							_, err = plan.Apply(s.values, false)
							check(err)
							encode(s.raw, s.values)
							_, err = output.Write(s.raw)
							check(err)
						}
						pool <- s
					} else {
						loaded, computed := make(chan *slot, slots), make(chan *slot, slots)
						var group sync.WaitGroup
						group.Add(2)
						go func() {
							defer group.Done()
							defer close(loaded)
							for j := 0; j < tileCount; j++ {
								s := <-pool
								s.index = j
								_, err := io.ReadFull(input, s.raw)
								check(err)
								decode(s.raw, s.values)
								loaded <- s
							}
						}()
						go func() {
							defer group.Done()
							defer close(computed)
							for s := range loaded {
								_, err := plan.Apply(s.values, false)
								check(err)
								encode(s.raw, s.values)
								computed <- s
							}
						}()
						for s := range computed {
							_, err := output.Write(s.raw)
							check(err)
							pool <- s
						}
						group.Wait()
					}
					return map[string]any{"slots": slots, "tiles": tileCount, "tile_bytes": 4 * tileElements, "cache_policy": policy, "read_bytes": 4 * tileCount * tileElements, "write_bytes": 4 * tileCount * tileElements, "buffer_payload_bytes": 8 * slots * tileElements, "sync_included": false}
				})
				measure(*dir, fmt.Sprintf("pipeline-sync-slots%d-r%d", slots, repeat), func() map[string]any { check(output.Sync()); return map[string]any{} })
				check(input.Close())
				check(output.Close())
				output, err = os.Open(outputPath)
				check(err)
				for j := 0; j < tileCount; j++ {
					_, err := io.ReadFull(output, raw)
					check(err)
					decode(raw, work[0])
					if digest(work[0]) != expected[j] {
						panic("pipeline output mismatch")
					}
				}
				check(output.Close())
			}
		}
	}
	// One inverse per tile after final output validation, outside measured work.
	for j := range work {
		copy(work[j], inputs[j])
		_, err := plan.Apply(work[j], false)
		check(err)
		_, err = plan.Apply(work[j], true)
		check(err)
		for i, x := range work[j] {
			if x != inputs[j][i] {
				panic("tile inverse mismatch")
			}
		}
	}
	emit(map[string]any{"verified": true, "tiles": tileCount, "expected_sha256": expected, "storage_bytes_checked": storageChecked, "native_inverse": true})
}
