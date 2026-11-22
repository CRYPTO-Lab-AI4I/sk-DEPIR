//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func distributedProductGPUCases[Word transposeWord](t *testing.T, w *mpi.World, s int, dims []int, output string, operation *uint64, emit func(map[string]any), opt productTestOptions) {
	residentMode, tileWords, localTaylor := opt.Resident, opt.TileWords, opt.LocalTaylor
	pattern := opt.Pattern
	if pattern == "" {
		pattern = "random"
	}
	shapes := opt.Shapes
	if shapes == 0 {
		shapes = 2
	}
	if (pattern != "random" && pattern != "ones" && !(pattern == "roots" && opt.Input != nil)) || (shapes != 1 && shapes != 2) {
		t.Fatal("product test pattern/shape admission")
	}
	for _, dim := range dims {
		n := uint64(1) << uint(dim)
		rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
		first, localRows := rankBand(rows, w.Rank, w.Size)
		start, words := first*cols, localRows*cols
		capacity := min(n, 3*max(rows, cols))
		h, field, err := canonicalScalingCodeHeader(s, 3, 1, 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		base, err := NewBinaryField(s, h.Polynomial)
		if err != nil {
			t.Fatal(err)
		}
		tables := newBoundedPreparationProducts(base, 18)
		options := cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: (128 << 20) / uint64(max(1, w.Size/2))}
		setupBegan := time.Now()
		var stage *cuda.VectorPreparation
		var resident *cuda.RankStorage
		var gpu interface {
			Close() error
			OwnedBytes() uint64
			TaylorDegrees(int) ([]uint64, error)
		}
		if residentMode {
			resident, err = cuda.NewRankStorage(options, rows, cols, w.Rank, w.Size, tileWords)
			gpu = resident
		} else {
			stage, err = cuda.NewVectorPreparation(options, capacity)
			gpu = stage
		}
		if err != nil {
			t.Fatal(err)
		}
		setupSeconds := w.Max(time.Since(setupBegan).Seconds())
		io := func(slot int, data []Word, upload bool) {
			for offset := 0; offset < len(data); {
				take := min(int(tileWords), len(data)-offset)
				if e := storageTransfer(resident, slot, uint64(offset), data[offset:offset+take], upload); e != nil {
					t.Fatal(e)
				}
				offset += take
			}
		}
		func() {
			defer gpu.Close()
			multiply := func(left, right []Word) error {
				switch a := any(left).(type) {
				case []uint32:
					return stage.Pointwise32(a, any(right).([]uint32))
				case []uint64:
					return stage.Pointwise64(a, any(right).([]uint64))
				}
				return fmt.Errorf("product slot width")
			}
			if residentMode {
				if localTaylor && os.Getenv("PIR_TAYLOR_LOCAL_GUARD") == "1" {
					localTaylorGuard[Word](t, resident, s, dim, w.Rank, w.Size, rows, cols, emit)
				}
				residentPointwiseCheck[Word](t, resident, field, s, dim, rows, cols, words, capacity, emit)
			} else {
				// Direct full-extension products exercise high components even when a small
				// polynomial's evaluation subspace lies entirely in the base field.
				a, b := make([]Word, min(capacity, 17)), make([]Word, min(capacity, 17))
				want := make([]Word, len(a))
				mask := (uint64(1) << uint(2*s)) - 1
				for j := range a {
					a[j] = Word((uint64(j+1)*0x9e3779b9 ^ (uint64(1) << uint(2*s-1))) & mask)
					b[j] = Word((uint64(j+3)*0x16379 ^ (uint64(1) << uint(s))) & mask)
					want[j] = Word(field.mul(uint64(a[j]), uint64(b[j])))
				}
				before := append([]Word(nil), b...)
				if e := multiply(a, b); e != nil || !reflect.DeepEqual(a, want) || !reflect.DeepEqual(b, before) {
					prefix := fmt.Sprintf("capture-rank-%d-s%d-d%d-pointwise", w.Rank, s, dim)
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-output.bin"), a)
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-expected.bin"), want)
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-right.bin"), b)
					t.Fatal("full extension pointwise", e)
				}
				saved := append([]Word(nil), a...)
				if e := multiply(a, a); e == nil || !reflect.DeepEqual(a, saved) {
					t.Fatal("pointwise overlap rejection", e)
				}
				emit(map[string]any{"kind": "pointwise-check", "s": s, "dimension": dim, "words": len(a), "matches": true})
			}
			var retained uint64
			for shape := 0; shape < shapes; shape++ {
				ac, bc := n/4+1, 3*n/4
				if shape == 1 {
					ac, bc = n/4, n/2
				}
				label := fmt.Sprintf("s%d-d%d-shape-%d", s, dim, shape)
				emit(map[string]any{"kind": "phase", "phase": label, "event": "started", "completed": 0, "total": 1})
				left, right := make([]Word, words), make([]Word, words)
				var expected []uint32
				leading := uint32(1)
				if opt.Input != nil {
					if pattern == "roots" && gpu.OwnedBytes() > 64<<20 {
						t.Fatal("root/product overlapping workspace cap")
					}
					input := opt.Input(s, dim, shape, start, words)
					ac, bc, leading = input.ACount, input.BCount, input.Leading
					if leading == 0 || uint64(leading) >= uint64(1)<<uint(s) || ac == 0 || bc == 0 || ac > n || bc > n-ac+1 || uint64(len(input.Left)) != words || len(input.Left) != len(input.Right) || len(input.Left) != len(input.Expected) {
						t.Fatal("injected product geometry/shards")
					}
					expected = input.Expected
					for j := range left {
						if input.Left[j] >= 1<<uint(s) || input.Right[j] >= 1<<uint(s) || expected[j] >= 1<<uint(s) {
							t.Fatal("injected product canonical coefficients")
						}
						left[j], right[j] = Word(input.Left[j]), Word(input.Right[j])
					}
				} else {
					expected = productExpected(start, words, ac, bc, s, base, pattern)
					for j := range left {
						k := start + uint64(j)
						left[j] = Word(productInputCoefficient(k, ac, s, 0, pattern))
						right[j] = Word(productInputCoefficient(k, bc, s, 1, pattern))
					}
				}
				op := *operation
				span := uint64(2 + 2*(dim-dim/2))
				*operation += 3 * span
				p, e := newRankTransposePlan(op, rows, cols, w.Rank, w.Size, n, tileWords)
				if e != nil {
					t.Fatal(e)
				}
				fftPass := 0
				proofs, taylorProofs := 0, 0
				transposes := transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, data []Word, own, remote [64]byte) {
					proofs++
					if mpi.ProofMismatch(own, remote) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-offset-%d-%s.bin", w.Rank, p.Operation, p.Offset, direction)), data)
					}
					emit(map[string]any{"kind": "tile-proof", "case": label, "pass": fftPass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				windows := taylorMPI[Word]{world: w, record: func(direction string, peer int, p taylorPacket, data []Word, own, remote [64]byte) {
					taylorProofs++
					if mpi.ProofMismatch(own, remote) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-block-%d-first-%d-slot-%d-source-%d-%s.bin", w.Rank, p.Operation, p.Target.Block, p.Target.First, p.Slot, p.Source, direction)), data)
					}
					emit(map[string]any{"kind": "taylor-proof", "case": label, "pass": fftPass, "peer": peer, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				backend := rankProductBackend[Word]{Transpose: transposes, Windows: windows,
					Degrees: func(level int) ([]uint64, error) {
						degrees, e := gpu.TaylorDegrees(level)
						if e == nil {
							emit(map[string]any{"kind": "taylor-plan", "case": label, "pass": fftPass, "operation_start": op + uint64(fftPass)*span, "level": level, "degrees": degrees, "capacity": capacity})
						}
						return degrees, e
					},
					Vectors: func(data []Word, columns, inverse bool, first, count uint64) error {
						switch a := any(data).(type) {
						case []uint32:
							return stage.Apply32(a, columns, inverse, first, count)
						case []uint64:
							return stage.Apply64(a, columns, inverse, first, count)
						}
						return fmt.Errorf("vector slot width")
					},
					Taylor: func(data []Word, level int, q bool, block, first, count uint64) error {
						before := mpi.PayloadFingerprint(data[count:])
						tile := cuda.TaylorTile{Level: level, Quotient: q, Block: block, First: first, Count: count}
						var e error
						switch a := any(data).(type) {
						case []uint32:
							e = stage.Taylor32(a, tile)
						case []uint64:
							e = stage.Taylor64(a, tile)
						}
						if e == nil && before != mpi.PayloadFingerprint(data[count:]) {
							return fmt.Errorf("Taylor source mutation")
						}
						return e
					},
					Multiply: func(left, right []Word) error {
						before := mpi.PayloadFingerprint(right)
						e := multiply(left, right)
						if e == nil && before != mpi.PayloadFingerprint(right) {
							return fmt.Errorf("pointwise right mutation")
						}
						return e
					},
				}
				var rightSpectrum []Word
				var inverseOwned uint64
				var captureDuration time.Duration
				last := time.Now()
				notice := func(phase string, done, total uint64) {
					if phase == "rank-product-pointwise" && total > 0 && done == total {
						inverseOwned = gpu.OwnedBytes()
					}
					if residentMode && phase == "rank-product-pointwise" && done == 0 {
						captureBegan := time.Now()
						rightSpectrum = make([]Word, words)
						io(1, rightSpectrum, false)
						captureDuration += time.Since(captureBegan)
					}
					if strings.HasPrefix(phase, "rank-product-") && done == 0 {
						switch phase {
						case "rank-product-left-forward":
							fftPass = 0
						case "rank-product-right-forward":
							fftPass = 1
						case "rank-product-inverse":
							fftPass = 2
						}
					}
					if done == 0 || done == total || time.Since(last) >= 30*time.Second {
						event := "progress"
						if done == 0 {
							event = "started"
						}
						if done == total {
							event = "completed"
						}
						emit(map[string]any{"kind": "progress", "event": event, "case": label, "pass": fftPass, "phase": phase, "completed": done, "total": total})
						last = time.Now()
					}
				}
				w.Barrier()
				began := time.Now()
				var out []Word
				if residentMode {
					for slot := 0; slot < 2; slot++ {
						if e = resident.Reset(slot, rows, cols, false); e != nil {
							t.Fatal(e)
						}
					}
					io(0, left, true)
					io(1, right, true)
					e = rankResidentPolynomialProduct[Word](p, resident, 0, 1, 2, capacity, transposes, windows, backend.Degrees, notice, localTaylor)
					if e == nil {
						out = make([]Word, words)
						io(0, out, false)
					}
				} else {
					out, e = rankPolynomialProduct(p, left, right, capacity, backend, notice)
				}
				if e != nil {
					t.Fatal(e)
				}
				localSeconds := time.Since(began).Seconds()
				localCapture := captureDuration.Seconds()
				localWork := localSeconds - localCapture
				if inverseOwned == 0 || gpu.OwnedBytes() != inverseOwned {
					t.Fatal("inverse workspace allocation grew")
				}
				seconds := w.Max(localSeconds)
				workSeconds := w.Max(localWork)
				captureSeconds := w.Max(localCapture)
				if residentMode {
					gotRight := make([]Word, words)
					io(1, gotRight, false)
					if !reflect.DeepEqual(gotRight, rightSpectrum) {
						t.Fatal("right spectrum changed by pointwise/inverse")
					}
				}
				actual := make([]uint32, len(out))
				for j, v := range out {
					if uint64(v) >= uint64(1)<<uint(s) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-%s-extension-output.bin", w.Rank, label)), out)
						t.Fatal("product outside base field", label, j)
					}
					actual[j] = uint32(v)
				}
				if shape == 0 {
					retained = gpu.OwnedBytes()
				} else if gpu.OwnedBytes() != retained {
					t.Fatal("product allocation grew on reuse")
				}
				prefix := fmt.Sprintf("rank-%d-%s", w.Rank, label)
				saveMatrixStageWords(t, filepath.Join(output, prefix+"-output.bin"), actual)
				saveMatrixStageWords(t, filepath.Join(output, prefix+"-expected.bin"), expected)
				if !reflect.DeepEqual(actual, expected) {
					t.Fatal("distributed product/schoolbook", label)
				}
				emit(map[string]any{"kind": "measurement", "case": label, "s": s, "dimension": dim, "shape": shape, "a_count": ac, "b_count": bc, "first_coefficient": start, "words": words, "slot_bytes": h.ExtensionBytes, "output_word_bytes": 4, "seconds": seconds, "work_seconds": workSeconds, "validation_capture_seconds": captureSeconds, "local_seconds": localSeconds, "local_work_seconds": localWork, "local_capture_seconds": localCapture, "operation_start": op, "proof_events": proofs, "taylor_proof_events": taylorProofs, "owned_device_bytes": gpu.OwnedBytes(), "output_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(actual)), "matches": true, "resident": residentMode, "oracle_pattern": pattern, "leading_coefficient": leading, "local_taylor": localTaylor, "tile_words": tileWords, "setup_seconds": setupSeconds, "right_spectrum_preserved": residentMode, "reused": true, "inverse_reused": true, "geometry_reuse_checked": shapes > 1})
				emit(map[string]any{"kind": "phase", "phase": label, "event": "completed", "completed": 1, "total": 1})
			}
		}()
	}
}
func TestDistributedGPUProducts(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded distributed polynomial products")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		operation := uint64(1)
		distributedProductGPUCases[uint32](t, w, 4, []int{4, 5, 8}, output, &operation, emit, productTestOptions{TileWords: 127})
		distributedProductGPUCases[uint64](t, w, 17, []int{5, 9, 12}, output, &operation, emit, productTestOptions{TileWords: 127})
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}

// TestDistributedResidentProducts shares operands, schoolbook oracle, routes and
// resource guards with the host-shard product reference; only storage changes.
func TestDistributedResidentProducts(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded resident products")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		operation := uint64(1)
		distributedProductGPUCases[uint32](t, w, 4, []int{4, 5, 8}, output, &operation, emit, productTestOptions{Resident: true, TileWords: 127})
		distributedProductGPUCases[uint64](t, w, 17, []int{5, 9, 12}, output, &operation, emit, productTestOptions{Resident: true, TileWords: 127})
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}

// TestMatchedGPUProducts changes only declared synthetic sizes/transport tiling.
// Exact schoolbook and route checks stay shared with the validated small gate.
func TestMatchedGPUProducts(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit matched product experiment")
	}
	tile, err := strconv.ParseUint(os.Getenv("PIR_PRODUCT_TILE_WORDS"), 10, 64)
	if err != nil || (tile != 127 && tile != 4096) {
		t.Fatal("matched tile admission")
	}
	dims := []int{12, 14}
	if os.Getenv("PIR_PRODUCT_DIMS") == "12" {
		dims = []int{12}
	} else if os.Getenv("PIR_PRODUCT_DIMS") != "12,14" {
		t.Fatal("matched dimension admission")
	}
	backend := os.Getenv("PIR_PRODUCT_BACKEND")
	if backend != "host" && backend != "resident" && backend != "resident-local" {
		t.Fatal("matched backend")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		operation := uint64(1)
		distributedProductGPUCases[uint64](t, w, 17, dims, output, &operation, emit, productTestOptions{Resident: backend != "host", TileWords: tile, LocalTaylor: backend == "resident-local"})
		emit(map[string]any{"kind": "completed", "cases": 2 * len(dims)})
	})
}

// TestScaledGPUProducts keeps resource/log-size admission explicit. The larger
// evaluation grid crosses the s17 base field while remaining in its extension.
func TestScaledGPUProducts(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit structured product scaling")
	}
	dim, err := strconv.Atoi(os.Getenv("PIR_SCALED_DIM"))
	if err != nil || (dim != 12 && dim != 14 && dim != 16 && dim != 18) {
		t.Fatal("scaled dimension admission")
	}
	pattern := os.Getenv("PIR_SCALED_PATTERN")
	if pattern != "random" && pattern != "ones" {
		t.Fatal("scaled pattern")
	}
	if pattern == "random" && dim > 14 {
		t.Fatal("quadratic oracle not admitted at larger sizes")
	}
	shapes, err := strconv.Atoi(os.Getenv("PIR_SCALED_SHAPES"))
	if err != nil || (shapes != 1 && shapes != 2) || (dim > 14 && shapes != 1) {
		t.Fatal("scaled proof-log admission")
	}
	backend := os.Getenv("PIR_PRODUCT_BACKEND")
	if backend != "host" && backend != "resident-local" {
		t.Fatal("scaled backend")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		op := uint64(1)
		distributedProductGPUCases[uint64](t, w, 17, []int{dim}, output, &op, emit, productTestOptions{Resident: backend != "host", LocalTaylor: backend == "resident-local", TileWords: 4096, Pattern: pattern, Shapes: shapes})
		emit(map[string]any{"kind": "completed", "cases": shapes})
	})
}
