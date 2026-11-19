//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"
)

var rootMPIBuffer = flag.String("root-mpi-buffer", "direct", "direct Go payloads or bounded MPI-owned staging")

type recordedRootLink struct {
	world   *mpi.World
	record  func(string, int, []uint32)
	capture func(string, int, []uint32, [64]byte, [64]byte)
}

// Fingerprints reuse the canonical bounded codec, avoiding a full-array clone.
func rootTransferFingerprint(p []uint32) [32]byte {
	h := sha256.New()
	if err := writeScalingCodeValues(h, 4, 1<<32, p, nil); err != nil {
		panic(err)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}
func rootProofMismatch(own, remote [64]byte) bool {
	return !bytes.Equal(own[:32], own[32:]) || !bytes.Equal(remote[:32], remote[32:]) || !bytes.Equal(own[:32], remote[:32])
}

// Index proofs must preserve even invalid high bits, before range validation.
func indexTransferFingerprint(p []uint64) [32]byte {
	h := sha256.New()
	var block [1024]byte
	for first := 0; first < len(p); {
		take := minInt(len(p)-first, len(block)/8)
		for j := 0; j < take; j++ {
			binary.LittleEndian.PutUint64(block[j*8:], p[first+j])
		}
		_, _ = h.Write(block[:take*8])
		first += take
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// Both payload types use the same symmetric check and durable-capture handshake.
func checkTransferProof(world *mpi.World, peer int, before, after [32]byte, capture func([64]byte, [64]byte)) error {
	var own [64]byte
	copy(own[:32], before[:])
	copy(own[32:], after[:])
	remote, err := world.ExchangeProof(peer, own)
	if err != nil {
		return err
	}
	capture(own, remote)
	if rootProofMismatch(own, remote) {
		// Both files are synced before the acknowledgement and possible MPI abort.
		if _, err = world.ExchangeProof(peer, own); err != nil {
			return err
		}
		return fmt.Errorf("transfer fingerprint mismatch peer%d", peer)
	}
	return nil
}
func (l recordedRootLink) check(direction string, peer int, p []uint32, before [32]byte) error {
	return checkTransferProof(l.world, peer, before, rootTransferFingerprint(p), func(own, remote [64]byte) { l.capture(direction, peer, p, own, remote) })
}
func (l recordedRootLink) SendRoot(peer int, p []uint32, max int) error {
	var before [32]byte
	if l.capture != nil {
		before = rootTransferFingerprint(p)
	}
	if err := l.world.SendRoot(peer, p, max); err != nil {
		return err
	}
	if l.capture != nil {
		if err := l.check("send", peer, p, before); err != nil {
			return err
		}
	}
	l.record("send", peer, p)
	return nil
}
func (l recordedRootLink) ReceiveRoot(peer, max int) ([]uint32, error) {
	p, err := l.world.ReceiveRoot(peer, max)
	if err != nil {
		return nil, err
	}
	if l.capture != nil {
		if err = l.check("receive", peer, p, rootTransferFingerprint(p)); err != nil {
			return nil, err
		}
	}
	l.record("receive", peer, p)
	return p, nil
}

// TestRankOwnedRootTree validates two reduction levels with four logical ranks.
// Only rank0 builds a full support/reference; every worker keeps its local shard.
// It does not implement distributed discovery or a distributed last-product FFT.
func TestRankOwnedRootTree(t *testing.T) {
	directory := os.Getenv("PIR_ROOT_TREE_OUTPUT")
	if directory == "" {
		t.Skip("explicit root tree experiment")
	}
	world, err := mpi.Open()
	if err != nil {
		t.Fatal(err)
	}
	success := false
	var failureLog func()
	defer func() {
		if !success {
			defer world.Abort()
			if failureLog != nil {
				failureLog()
			}
		} else if err := world.Close(); err != nil {
			t.Error(err)
		}
	}()
	rank, err := strconv.Atoi(os.Getenv("SLURM_PROCID"))
	if err != nil || rank != world.Rank || world.Size != 4 {
		t.Fatal("four Slurm/MPI ranks required")
	}
	chunk, err := strconv.Atoi(os.Getenv("PIR_ROOT_TREE_CHUNK_WORDS"))
	if err != nil || chunk < 1 || chunk > 1<<20 {
		t.Fatal("explicit MPI chunk admission")
	}
	world.ChunkWords = chunk
	switch *rootMPIBuffer {
	case "direct":
	case "staged":
		if err = world.EnableStaging(); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatal("invalid root-mpi-buffer")
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(directory, fmt.Sprintf("rank-%d", rank))
	if err = os.MkdirAll(dest, 0700); err != nil {
		t.Fatal(err)
	}
	began := time.Now()
	var mu sync.Mutex
	emit := func(row map[string]any) {
		mu.Lock()
		defer mu.Unlock()
		var usage syscall.Rusage
		if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
			panic(err)
		}
		if usage.Maxrss*1024 > 1<<30 {
			panic("rank1GiB RSS guard")
		}
		row["utc"], row["elapsed_seconds"] = time.Now().UTC().Format(time.RFC3339Nano), time.Since(began).Seconds()
		row["rank"], row["host"], row["peak_rss_bytes"] = rank, host, usage.Maxrss*1024
		if err := json.NewEncoder(os.Stdout).Encode(row); err != nil {
			panic(err)
		}
	}
	failureLog = func() { emit(map[string]any{"kind": "error", "phase": "root-tree", "event": "failed"}) }
	digest := func(values any) string {
		h := sha256.New()
		if err := binary.Write(h, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	save := func(name string, values any) {
		f, err := os.OpenFile(filepath.Join(dest, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		if err = binary.Write(f, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
		if err = f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	emit(map[string]any{"kind": "host", "ranks": world.Size, "visible_devices": os.Getenv("CUDA_VISIBLE_DEVICES"), "local_rank": os.Getenv("SLURM_LOCALID"), "chunk_words": world.ChunkWords, "buffer_mode": *rootMPIBuffer, "staging_bytes": world.StagingBytes()})
	cases := []struct {
		s, count int
		genuine  bool
	}{{8, 0, true}, {17, 1 << 16, false}, {17, 1 << 18, false}}
	small := os.Getenv("PIR_ROOT_TREE_SMALL") == "1"
	if small {
		cases = cases[:1]
	}
	for index, item := range cases {
		emit(map[string]any{"kind": "stage", "case": index, "phase": "setup-reference", "event": "started"})
		setupStart := time.Now()
		var fixture *rootFixture
		var reference []uint32
		referenceSeconds := 0.0
		if rank == 0 {
			fixture, err = makeRootFixture(item.s, item.count, item.genuine)
			if err != nil {
				t.Fatal(err)
			}
			native, err := cpp.NewPreparation(item.s, fixture.header.Polynomial, fixture.header.Omega, 1, uint64(fixture.maximum), 64<<20)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			lastReference := time.Now()
			reference, err = fixture.root(supportSorted, native, fixture.support, func(stage string) {
				emit(map[string]any{"kind": "stage", "case": index, "phase": "CPU-reference/" + stage, "event": "started"})
			}, func(stage string, done, total uint64, unit string) {
				if done == total || time.Since(lastReference) >= 30*time.Second {
					emit(map[string]any{"kind": "progress", "case": index, "phase": "CPU-reference/" + stage, "completed": done, "total": total, "unit": unit})
					lastReference = time.Now()
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			if err = native.Close(); err != nil {
				t.Fatal(err)
			}
			if len(reference) != len(fixture.support)+1 || reference[0] != 1 {
				t.Fatal("CPU reference degree/normalization")
			}
			for witness := 0; witness < 4; witness++ {
				exponent := fixture.support[witness*(len(fixture.support)-1)/3]
				point := fixture.extension.pow(fixture.header.Omega, uint64(fixture.n-1-exponent))
				var value uint64
				for j := len(reference) - 1; j >= 0; j-- {
					value = fixture.extension.mul(value, point) ^ uint64(reference[j])
				}
				if value != 0 {
					t.Fatal("independent scalar root witness")
				}
			}
			referenceSeconds = time.Since(start).Seconds()
			save(fmt.Sprintf("case-%d-reference.bin", index), reference)
		} else {
			fixture, err = makeRootArithmetic(item.s)
			if err != nil {
				t.Fatal(err)
			}
		}
		count := int(world.ShareCount(uint64(len(fixture.support))))
		if count < 1 || count > 1<<18 {
			t.Fatal("bounded global support count")
		}
		// Admit only this owner's largest subtree, not a full final product
		// on every rank. Greedy orbit loads differ by at most two exponents.
		steps, e := rootTreeSteps(rank, world.Size)
		if e != nil {
			t.Fatal(e)
		}
		subtreeRanks := 1
		for _, step := range steps {
			if !step.Send {
				subtreeRanks += minInt(step.Stride, world.Size-step.Peer)
			}
		}
		bound := minInt(count+1, (count/world.Size+2)*subtreeRanks+1)
		maximum := 4
		for maximum < bound {
			maximum *= 2
		}
		h := fixture.header
		tables := fixture.base.lookup
		gpu, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: item.s,
			Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0},
			MaxElements: uint64(maximum), DeviceBytes: 64 << 20})
		if err != nil {
			t.Fatal(err)
		}
		emit(map[string]any{"kind": "fixture", "case": index, "s": item.s, "support_count": count, "coordinator_support_count": len(fixture.support),
			"reference_count": len(reference), "reference_sha256": digest(reference), "reference_seconds": referenceSeconds, "setup_seconds": time.Since(setupStart).Seconds(), "max_elements": maximum, "subtree_ranks": subtreeRanks})
		samples := []int{-1, 0, 1, 2}
		diagnostic := os.Getenv("PIR_ROOT_TREE_DIAGNOSTICS") == "1"
		if diagnostic && index == 2 && !small {
			repeats, e := strconv.Atoi(os.Getenv("PIR_ROOT_TREE_REPEATS"))
			if e != nil || repeats < 2 || repeats > 12 {
				t.Fatal("diagnostic repeat admission")
			}
			samples = []int{-1, 0, 1}
			for extra := 3; extra <= repeats; extra++ {
				samples = append(samples, extra)
			}
			samples = append(samples, 2)
		}
		if small {
			samples = []int{-1, 2}
		}
		for _, sample := range samples {
			world.Barrier()
			start := time.Now()
			var shard []int
			packets := 0
			routed := 0
			lastRouting := time.Now()
			emit(map[string]any{"kind": "stage", "case": index, "sample": sample, "phase": "route", "event": "started"})
			// Per-peer packet sequence binds captures even when routing interleaves owners.
			packetSequence := make([]int, world.Size)
			capture := func(phase, direction string, peer, sequence, wordBytes int, values any, own, remote [64]byte) {
				mismatch := rootProofMismatch(own, remote)
				name := ""
				if mismatch {
					name = fmt.Sprintf("case-%d-sample-%d-%s-%d-failed-%s-%d.bin", index, sample, phase, sequence, direction, peer)
					save(name, values)
				}
				v := reflect.ValueOf(values)
				row := map[string]any{"kind": "wire-proof", "case": index, "sample": sample, "phase": phase, "sequence": sequence, "direction": direction, "peer": peer,
					"words": v.Len(), "word_bytes": wordBytes, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote), "sha256": digest(values), "file": name, "matches": !mismatch, "buffer": fmt.Sprintf("%x", v.Pointer())}
				emit(row)
				if mismatch {
					row["kind"] = "transport-failure"
					emit(row)
				}
			}
			checkIndices := func(direction string, peer int, packet []uint64, before [32]byte) error {
				sequence := packetSequence[peer]
				packetSequence[peer]++
				return checkTransferProof(world, peer, before, indexTransferFingerprint(packet), func(own, remote [64]byte) { capture("support", direction, peer, sequence, 8, packet, own, remote) })
			}
			if rank == 0 {
				err = routeRootSupport(supportSorted, fixture.support, fixture.n, 1<<uint(item.s), world.Size, 16384, func(owner int, packet []uint64) error {
					if owner == 0 {
						for _, x := range packet {
							shard = append(shard, int(x))
						}
					} else {
						var before [32]byte
						if diagnostic {
							before = indexTransferFingerprint(packet)
						}
						if err := world.SendIndices(owner, packet); err != nil {
							return err
						}
						if diagnostic {
							if err := checkIndices("send", owner, packet, before); err != nil {
								return err
							}
						}
					}
					packets++
					routed += len(packet)
					if time.Since(lastRouting) >= 30*time.Second {
						emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "route", "completed": routed, "total": count, "unit": "exponents"})
						lastRouting = time.Now()
					}
					return nil
				})
				if err != nil {
					t.Fatal(err)
				}
				for peer := 1; peer < world.Size; peer++ {
					if err = world.SendIndices(peer, nil); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				for {
					packet, e := world.ReceiveIndices()
					if e != nil {
						t.Fatal(e)
					}
					if len(packet) == 0 {
						break
					}
					if diagnostic {
						if e = checkIndices("receive", 0, packet, indexTransferFingerprint(packet)); e != nil {
							t.Fatal(e)
						}
					}
					packets++
					if len(shard) > count-len(packet) {
						t.Fatal("support shard admission")
					}
					if time.Since(lastRouting) >= 30*time.Second {
						emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "route-receive", "completed": len(shard), "total": count, "unit": "local exponents", "scope": "global support bound; local target unknown"})
						lastRouting = time.Now()
					}
					for _, x := range packet {
						if x >= uint64(fixture.n-1) {
							t.Fatal("support exponent range")
						}
						shard = append(shard, int(x))
					}
				}
			}
			sort.Ints(shard)
			routeSeconds := time.Since(start).Seconds()
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "route", "event": "completed", "completed": len(shard), "total": len(shard), "unit": "local exponents", "packets": packets})
			rootStart := time.Now()
			last := time.Now()
			local, err := fixture.root(supportSorted, gpu, shard, func(stage string) {
				emit(map[string]any{"kind": "stage", "case": index, "sample": sample, "phase": "local-root/" + stage, "event": "started"})
			},
				func(stage string, done, total uint64, unit string) {
					if done == total || time.Since(last) >= 30*time.Second {
						emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "local-root/" + stage, "completed": done, "total": total, "unit": unit})
						last = time.Now()
					}
				})
			if err != nil {
				t.Fatal(err)
			}
			rootSeconds := time.Since(rootStart).Seconds()
			if len(local) != len(shard)+1 {
				t.Fatal("local root extent")
			}
			if sample == 2 {
				indices := make([]uint64, len(shard))
				for j, x := range shard {
					indices[j] = uint64(x)
				}
				save(fmt.Sprintf("case-%d-support.bin", index), indices)
				save(fmt.Sprintf("case-%d-local.bin", index), local)
			}
			stride := 0
			link := recordedRootLink{world: world, record: func(direction string, peer int, p []uint32) {
				name := fmt.Sprintf("case-%d-stride-%d-%s-%d.bin", index, stride, direction, peer)
				if sample == 2 {
					save(name, p)
				}
				emit(map[string]any{"kind": "transfer", "case": index, "sample": sample, "stride": stride, "direction": direction, "peer": peer, "words": len(p), "sha256": digest(p), "file": name})
			}}
			if diagnostic {
				link.capture = func(direction string, peer int, p []uint32, own, remote [64]byte) {
					capture("root", direction, peer, stride, 4, p, own, remote)
				}
				if index == 2 && sample == -1 {
					indices := make([]uint64, len(shard))
					for j, x := range shard {
						indices[j] = uint64(x)
					}
					save("case-2-diagnostic-support.bin", indices)
					save("case-2-diagnostic-local.bin", local)
				}
			}
			treeStart := time.Now()
			result, err := reduceRootProduct(rank, world.Size, count+1, 1<<uint(item.s), local, link,
				func(a, b []uint32) (p []uint32, e error) {
					defer recoverPreparation(&e)
					return polyMul(fixture.withEngine(gpu), a, b), nil
				},
				func(step rootTreeStep, event string, p []uint32) {
					stride = step.Stride
					if sample == 2 && event == "completed" && !step.Send {
						save(fmt.Sprintf("case-%d-merged-%d.bin", index, step.Stride), p)
					}
					emit(map[string]any{"kind": "tree", "case": index, "sample": sample, "stride": step.Stride, "peer": step.Peer, "sending": step.Send, "event": event, "words": len(p), "sha256": digest(p)})
				})
			local = nil
			shardCount := len(shard)
			shard = nil
			if err != nil {
				t.Fatal(err)
			}
			treeSeconds := time.Since(treeStart).Seconds()
			verifyStart := time.Now()
			if rank == 0 {
				if !reflect.DeepEqual(result, reference) {
					if diagnostic {
						save(fmt.Sprintf("case-%d-sample-%d-final-mismatch.bin", index, sample), result)
					}
					t.Fatal("unique final owner differs from CPU")
				}
				if sample == 2 {
					save(fmt.Sprintf("case-%d-final.bin", index), result)
				}
			} else if result != nil {
				t.Fatal("sender retained final ownership")
			}
			ownedWords := len(result)
			result = nil
			verifySeconds := time.Since(verifyStart).Seconds()
			seconds := time.Since(start).Seconds()
			// Fixed collective order, outside pipeline clock; maxima not additive.
			maxSeconds, maxRoute, maxRoot, maxTree, maxVerify := world.Max(seconds), world.Max(routeSeconds), world.Max(rootSeconds), world.Max(treeSeconds), world.Max(verifySeconds)
			emit(map[string]any{"kind": "measurement", "case": index, "sample": sample, "s": item.s, "support_count": count, "shard_count": shardCount, "ranks": world.Size,
				"seconds": maxSeconds, "route_seconds": maxRoute, "root_seconds": maxRoot, "tree_seconds": maxTree, "verify_seconds": maxVerify,
				"owned_result_words": ownedWords, "owned_device_bytes": gpu.OwnedBytes(), "validation_io": sample == 2, "matches": true})
			if gpu.OwnedBytes() > 64<<20 {
				t.Fatal("rank device cap")
			}
		}
		if err = gpu.Close(); err != nil {
			t.Fatal(err)
		}
	}
	emit(map[string]any{"kind": "completed", "cases": len(cases)})
	success = true
}
