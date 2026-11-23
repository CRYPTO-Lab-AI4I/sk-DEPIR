//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

// Explicit test-binary flags: ordinary preprocessing still selects bitmap.
var rootValidationFlag = flag.String("root-validation", "bitmap", "root/partition membership: bitmap or sorted")
var rootReferenceValidationFlag = flag.String("root-reference-validation", "bitmap", "CPU fixture membership, held equal in matched comparisons")

// Opt-in bounded scheduling experiment. Workers own separate executors; lookup
// tables and caller support are immutable. No CLI fallback or public mode added.
func TestIndependentRootWorkers(t *testing.T) {
	directory := os.Getenv("PIR_ROOT_WORKER_OUTPUT")
	if directory == "" {
		t.Skip("explicit independent root-worker experiment")
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	modeOf := func(name string) supportMembership {
		switch name {
		case "bitmap":
			return supportBitmap
		case "sorted":
			return supportSorted
		default:
			t.Fatal("unknown root membership", name)
			return supportBitmap
		}
	}
	mode, referenceMode := modeOf(*rootValidationFlag), modeOf(*rootReferenceValidationFlag)
	var logMu sync.Mutex
	began := time.Now()
	emit := func(fields map[string]any) {
		logMu.Lock()
		defer logMu.Unlock()
		fields["validation"] = *rootValidationFlag
		fields["reference_validation"] = *rootReferenceValidationFlag
		fields["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		fields["elapsed_seconds"] = time.Since(began).Seconds()
		if err := json.NewEncoder(os.Stdout).Encode(fields); err != nil {
			panic(err)
		}
	}
	save := func(name string, values []uint32) string {
		f, err := os.OpenFile(filepath.Join(directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		hash := sha256.New()
		if err = binary.Write(hash, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		if err = binary.Write(f, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%x", hash.Sum(nil))
	}
	type fixture struct {
		s, count int
		genuine  bool
	}
	cases := []fixture{{8, 0, true}, {12, 1 << 16, false}, {17, 1 << 16, false}, {17, 1 << 18, false}}
	if os.Getenv("PIR_ROOT_WORKER_SMALL") == "1" {
		cases = cases[:1]
	}
	for index, item := range cases {
		fixture, err := makeRootFixture(item.s, item.count, item.genuine)
		if err != nil {
			t.Fatal(err)
		}
		n, h, extension, support, maximum := fixture.n, fixture.header, fixture.extension, fixture.support, fixture.maximum
		lookup := fixture.base.lookup
		makeBase := fixture.withEngine

		progress := func(worker int) preparationNotice {
			last := time.Now()
			return func(stage string, done, total uint64, unit string) {
				if done == total || time.Since(last) >= 30*time.Second {
					emit(map[string]any{"kind": "progress", "case": index, "worker": worker, "phase": stage, "completed": done, "total": total, "unit": unit})
					last = time.Now()
				}
			}
		}
		root := func(engine rootWorkerEngine, subset []int, worker int) ([]uint32, error) {
			selected := mode
			if worker == -1 {
				selected = referenceMode
			}
			return fixture.root(selected, engine, subset, func(stage string) {
				emit(map[string]any{"kind": "stage", "case": index, "worker": worker, "phase": stage, "event": "started"})
			}, progress(worker))
		}

		emit(map[string]any{"kind": "fixture", "case": index, "event": "started", "s": item.s, "support_count": len(support), "genuine_support": item.genuine})
		cpu, err := cpp.NewPreparation(item.s, h.Polynomial, h.Omega, 4, uint64(maximum), 64<<20)
		if err != nil {
			t.Fatal(err)
		}
		expected, err := root(cpu, support, -1)
		if err != nil {
			t.Fatal(err)
		}
		if err = cpu.Close(); err != nil {
			t.Fatal(err)
		}
		if len(expected) != len(support)+1 || expected[0] != 1 {
			t.Fatal("CPU root degree/normalization")
		}
		// Independent extension-field Horner witnesses complement shared host
		// leaf arithmetic and complete native/GPU coefficient comparisons.
		for check := 0; check < 4; check++ {
			exponent := support[check*(len(support)-1)/3]
			point := extension.pow(h.Omega, uint64(n-1-exponent))
			var value uint64
			for j := len(expected) - 1; j >= 0; j-- {
				value = extension.mul(value, point) ^ uint64(expected[j])
			}
			if value != 0 {
				t.Fatal("independent root evaluation", index, exponent)
			}
			emit(map[string]any{"kind": "progress", "case": index, "phase": "scalar-root-oracle", "completed": check + 1, "total": 4, "unit": "roots"})
		}
		digest := save(fmt.Sprintf("case-%d-reference.bin", index), expected)
		emit(map[string]any{"kind": "fixture", "case": index, "event": "completed", "s": item.s, "support_count": len(support), "genuine_support": item.genuine, "sha256": digest})
		for _, workers := range []int{0, 1, 2, 4} {
			setupStart := time.Now()
			create := func(devices []int) *cuda.Preparation {
				e, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: item.s,
					Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: lookup.log, Powers: lookup.powers,
					Devices: devices, MaxElements: uint64(maximum), DeviceBytes: 128 << 20})
				if err != nil {
					t.Fatal(err)
				}
				return e
			}
			merge := create([]int{0, 1, 2, 3})
			engines := make([]*cuda.Preparation, workers)
			for worker := range engines {
				engines[worker] = create([]int{worker})
			}
			setupSeconds := time.Since(setupStart).Seconds()
			for sample := -1; sample < 2; sample++ {
				start := time.Now()
				partitionSeconds, workerSeconds, mergeSeconds := 0.0, 0.0, 0.0
				var actual []uint32
				var sizes []int
				if workers == 0 {
					actual, err = root(merge, support, 0)
					workerSeconds = time.Since(start).Seconds()
				} else {
					partitions, e := partitionRootSupportWith(mode, support, n, 1<<uint(item.s), workers)
					if e != nil {
						t.Fatal(e)
					}
					for _, part := range partitions {
						sizes = append(sizes, len(part))
					}
					partitionSeconds = time.Since(start).Seconds()
					products := make([][]uint32, workers)
					errors := make([]error, workers)
					var group sync.WaitGroup
					workerStart := time.Now()
					for worker := range engines {
						group.Add(1)
						go func(worker int) {
							defer group.Done()
							products[worker], errors[worker] = root(engines[worker], partitions[worker], worker)
						}(worker)
					}
					group.Wait()
					workerSeconds = time.Since(workerStart).Seconds()
					for _, e := range errors {
						if e != nil {
							t.Fatal(e)
						}
					}
					mergeStart := time.Now()
					func() {
						defer recoverPreparation(&err)
						for len(products) > 1 {
							var next [][]uint32
							for first := 0; first < len(products); first += 2 {
								if first+1 == len(products) {
									next = append(next, products[first])
								} else {
									next = append(next, polyMul(makeBase(merge), products[first], products[first+1]))
								}
							}
							products = next
						}
						actual = products[0]
					}()
					mergeSeconds = time.Since(mergeStart).Seconds()
				}
				seconds := time.Since(start).Seconds()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatal("independent workers/merge differ from CPU", index, workers, sample)
				}
				owned := merge.OwnedBytes()
				for _, e := range engines {
					owned += e.OwnedBytes()
				}
				file := ""
				if sample == 1 {
					file = fmt.Sprintf("case-%d-workers-%d.bin", index, workers)
					if save(file, actual) != digest {
						t.Fatal("persisted hash differs")
					}
				}
				emit(map[string]any{"kind": "measurement", "case": index, "s": item.s, "support_count": len(support), "genuine_support": item.genuine,
					"independent_workers": workers, "sample": sample, "setup_seconds": setupSeconds, "seconds": seconds,
					"partition_seconds": partitionSeconds, "worker_seconds": workerSeconds, "merge_seconds": mergeSeconds, "partition_sizes": sizes,
					"owned_device_bytes": owned, "sha256": digest, "file": file, "matches": true})
			}
			for _, e := range engines {
				if err = e.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err = merge.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	emit(map[string]any{"kind": "completed", "cases": len(cases)})
}
