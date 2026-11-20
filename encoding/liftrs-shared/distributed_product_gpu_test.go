//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func productCoefficient(index, count uint64, s, operand int) uint32 {
	if index >= count {
		return 0
	}
	if index == 0 || index+1 == count {
		return 1
	}
	return uint32((index+1)*0x9e3779b9^uint64(operand+1)*0x1337) & ((uint32(1) << uint(s)) - 1)
}
func distributedProductGPUCases[Word transposeWord](t *testing.T, w *mpi.World, s int, dims []int, output string, operation *uint64, emit func(map[string]any)) {
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
		tables := newPreparationProducts(base)
		gpu, err := cuda.NewVectorPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 128 << 20}, capacity)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer gpu.Close()
			multiply := func(left, right []Word) error {
				switch a := any(left).(type) {
				case []uint32:
					return gpu.Pointwise32(a, any(right).([]uint32))
				case []uint64:
					return gpu.Pointwise64(a, any(right).([]uint64))
				}
				return fmt.Errorf("product slot width")
			}
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
			var retained uint64
			for shape := 0; shape < 2; shape++ {
				ac, bc := n/4+1, 3*n/4
				if shape == 1 {
					ac, bc = n/4, n/2
				}
				label := fmt.Sprintf("s%d-d%d-shape-%d", s, dim, shape)
				emit(map[string]any{"kind": "phase", "phase": label, "event": "started", "completed": 0, "total": 1})
				left, right := make([]Word, words), make([]Word, words)
				expected := make([]uint32, words)
				for j := range left {
					k := start + uint64(j)
					left[j] = Word(productCoefficient(k, ac, s, 0))
					right[j] = Word(productCoefficient(k, bc, s, 1))
					if k < ac+bc-1 {
						lo := uint64(0)
						if k >= bc {
							lo = k - bc + 1
						}
						hi := min(k, ac-1)
						for i := lo; i <= hi; i++ {
							expected[j] ^= base.Mul(productCoefficient(i, ac, s, 0), productCoefficient(k-i, bc, s, 1))
						}
					}
				}
				op := *operation
				span := uint64(2 + 2*(dim-dim/2))
				*operation += 3 * span
				p, e := newRankTransposePlan(op, rows, cols, w.Rank, w.Size, n, 127)
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
							return gpu.Apply32(a, columns, inverse, first, count)
						case []uint64:
							return gpu.Apply64(a, columns, inverse, first, count)
						}
						return fmt.Errorf("vector slot width")
					},
					Taylor: func(data []Word, level int, q bool, block, first, count uint64) error {
						before := mpi.PayloadFingerprint(data[count:])
						tile := cuda.TaylorTile{Level: level, Quotient: q, Block: block, First: first, Count: count}
						var e error
						switch a := any(data).(type) {
						case []uint32:
							e = gpu.Taylor32(a, tile)
						case []uint64:
							e = gpu.Taylor64(a, tile)
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
				last := time.Now()
				notice := func(phase string, done, total uint64) {
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
				out, e := rankPolynomialProduct(p, left, right, capacity, backend, notice)
				if e != nil {
					t.Fatal(e)
				}
				seconds := w.Max(time.Since(began).Seconds())
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
				emit(map[string]any{"kind": "measurement", "case": label, "s": s, "dimension": dim, "shape": shape, "a_count": ac, "b_count": bc, "first_coefficient": start, "words": words, "slot_bytes": h.ExtensionBytes, "output_word_bytes": 4, "seconds": seconds, "operation_start": op, "proof_events": proofs, "taylor_proof_events": taylorProofs, "owned_device_bytes": gpu.OwnedBytes(), "output_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(actual)), "matches": true})
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
		distributedProductGPUCases[uint32](t, w, 4, []int{4, 5, 8}, output, &operation, emit)
		distributedProductGPUCases[uint64](t, w, 17, []int{5, 9, 12}, output, &operation, emit)
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}
