//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func residentVectorCases[Word transposeWord](t *testing.T, w *mpi.World, s int, dims []int, output string, operation *uint64, emit func(map[string]any)) {
	for _, dim := range dims {
		label := fmt.Sprintf("s%d-d%d", s, dim)
		emit(map[string]any{"kind": "phase", "phase": label, "event": "started", "completed": 0, "total": 2})
		h, reference := matrixStageOracle(t, s, dim, w.Rank, w.Size)
		base, err := NewBinaryField(s, h.Polynomial)
		if err != nil {
			t.Fatal(err)
		}
		tables := newPreparationProducts(base)
		n := uint64(1) << uint(dim)
		rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
		first, count := rankBand(rows, w.Rank, w.Size)
		gpu, err := cuda.NewRankStorage(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 128 << 20}, rows, cols, w.Rank, w.Size, 127)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer gpu.Close()
			input := make([]Word, count*cols)
			expected := make([]Word, len(input))
			sentinel := make([]Word, len(input))
			for j := range input {
				input[j] = Word(matrixStagePattern(first*cols+uint64(j), s))
				expected[j] = Word(reference[j])
				sentinel[j] = ^input[j]
			}
			// Reject an invalid physical layout before arithmetic, then prove an out-of-
			// field resident value rejects without changing even the offending slot.
			if gpu.Vectors(0, false, false) == nil {
				t.Fatal("row-major vector view accepted")
			}
			if err = gpu.Reset(0, rows, cols, true); err != nil {
				t.Fatal(err)
			}
			invalid := append([]Word(nil), input...)
			invalid[0] = Word(uint64(1) << uint(2*s))
			storageIO(t, gpu, 0, invalid, true)
			if gpu.Vectors(0, false, false) == nil {
				t.Fatal("out-of-field resident arithmetic accepted")
			}
			unchanged := make([]Word, len(invalid))
			storageIO(t, gpu, 0, unchanged, false)
			if !reflect.DeepEqual(unchanged, invalid) {
				t.Fatal("rejected arithmetic mutated input")
			}
			if err = gpu.Reset(0, rows, cols, false); err != nil {
				t.Fatal(err)
			}
			storageIO(t, gpu, 0, input, true)
			storageIO(t, gpu, 2, sentinel, true)
			var retained uint64
			for pass := 0; pass < 2; pass++ {
				op := *operation
				fullFFT := os.Getenv("PIR_GPU_MATRIX_STAGE_TAYLOR") == "1"
				span := uint64(2)
				if fullFFT {
					span += uint64(2 * (dim - dim/2))
				}
				*operation += span
				p, e := newRankTransposePlan(op, rows, cols, w.Rank, w.Size, n, 127)
				if e != nil {
					t.Fatal(e)
				}
				proofs := 0
				link := transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, tile []Word, own, remote [64]byte) {
					proofs++
					if mpi.ProofMismatch(own, remote) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-offset-%d-%s.bin", w.Rank, p.Operation, p.Offset, direction)), tile)
					}
					emit(map[string]any{"kind": "tile-proof", "case": label, "pass": pass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				taylorProofs := 0
				windows := taylorMPI[Word]{world: w, record: func(direction string, peer int, p taylorPacket, tile []Word, own, remote [64]byte) {
					taylorProofs++
					if mpi.ProofMismatch(own, remote) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-block-%d-first-%d-slot-%d-source-%d-%s.bin", w.Rank, p.Operation, p.Target.Block, p.Target.First, p.Slot, p.Source, direction)), tile)
					}
					emit(map[string]any{"kind": "taylor-proof", "case": label, "pass": pass, "peer": peer, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				capacity := min(n, 3*max(rows, cols))
				degrees := func(level int) ([]uint64, error) {
					terms, e := gpu.TaylorDegrees(level)
					if e == nil {
						emit(map[string]any{"kind": "taylor-plan", "case": label, "pass": pass, "operation_start": op, "level": level, "degrees": terms, "capacity": capacity})
					}
					return terms, e
				}
				w.Barrier()
				began := time.Now()
				notice := func(phase string, done, total uint64) {
					emit(map[string]any{"kind": "progress", "case": label, "pass": pass, "operation": op, "phase": phase, "completed": done, "total": total})
				}
				if fullFFT {
					e = rankResidentAdditiveTransform[Word](p, gpu, 0, 1, capacity, pass == 1, link, windows, degrees, notice)
				} else {
					e = rankResidentVectors[Word](p, gpu, 0, 1, pass == 1, link, notice)
				}
				if e != nil {
					t.Fatal(e)
				}
				seconds := w.Max(time.Since(began).Seconds())
				got := make([]Word, len(input))
				storageIO(t, gpu, 0, got, false)
				want := expected
				if pass == 1 {
					want = input
				}
				prefix := fmt.Sprintf("rank-%d-%s-pass-%d", w.Rank, label, pass)
				saveMatrixStageWords(t, filepath.Join(output, prefix+"-output.bin"), got)
				saveMatrixStageWords(t, filepath.Join(output, prefix+"-expected.bin"), want)
				if !reflect.DeepEqual(got, want) {
					t.Fatal("resident vectors/native oracle", label, pass)
				}
				third := make([]Word, len(sentinel))
				storageIO(t, gpu, 2, third, false)
				if !reflect.DeepEqual(third, sentinel) {
					t.Fatal("resident vectors changed third slot")
				}
				if pass == 0 {
					retained = gpu.OwnedBytes()
				} else if gpu.OwnedBytes() != retained {
					t.Fatal("resident vectors allocation grew on inverse")
				}
				width := 4
				if s == 17 {
					width = 8
				}
				emit(map[string]any{"kind": "measurement", "case": label, "pass": pass, "s": s, "dimension": dim, "rank_first": first, "rank_rows": count, "word_bytes": width, "words": len(got), "matches": true, "full_fft": fullFFT, "operation_start": op, "taylor_proof_events": taylorProofs, "resident": true, "third_preserved": true, "reused": true, "range_rejection_preserved": true, "owned_device_bytes": retained, "seconds": seconds, "proof_events": proofs, "output_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(got))})
			}
		}()
		emit(map[string]any{"kind": "phase", "phase": label, "event": "completed", "completed": 2, "total": 2})
	}
}

func TestDistributedResidentVectors(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded resident vector gate")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		op := uint64(1)
		residentVectorCases[uint32](t, w, 4, []int{4, 5, 8}, output, &op, emit)
		residentVectorCases[uint64](t, w, 17, []int{5, 9, 12}, output, &op, emit)
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}

// Explicit complete FFT gate; the shared driver selects independent full native
// and scalar Horner oracles when the runner enables Taylor conversion.
func TestDistributedResidentFFT(t *testing.T) {
	if os.Getenv("PIR_GPU_MATRIX_STAGE_TAYLOR") != "1" {
		t.Skip("explicit complete resident FFT gate")
	}
	TestDistributedResidentVectors(t)
}
