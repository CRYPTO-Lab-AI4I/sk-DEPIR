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

func storagePattern(index uint64) uint64 { return (index+1)*0x9e3779b97f4a7c15 ^ 0xa5a5a5a5a5a5a5a5 }
func storageTransfer[Word transposeWord](gpu *cuda.RankStorage, slot int, offset uint64, tile []Word, upload bool) error {
	switch v := any(tile).(type) {
	case []uint32:
		return gpu.Transfer32(slot, offset, v, upload)
	case []uint64:
		return gpu.Transfer64(slot, offset, v, upload)
	}
	return fmt.Errorf("storage test slot width")
}
func storageIO[Word transposeWord](t *testing.T, gpu *cuda.RankStorage, slot int, data []Word, upload bool) {
	t.Helper()
	for offset := 0; offset < len(data); {
		take := min(127, len(data)-offset)
		if err := storageTransfer(gpu, slot, uint64(offset), data[offset:offset+take], upload); err != nil {
			t.Fatal(err)
		}
		offset += take
	}
}
func storageGPUCases[Word transposeWord](t *testing.T, w *mpi.World, s int, shapes [][2]uint64, output string, operation *uint64, emit func(map[string]any)) {
	for _, shape := range shapes {
		rows, cols := shape[0], shape[1]
		label := fmt.Sprintf("s%d-r%d-c%d", s, rows, cols)
		emit(map[string]any{"kind": "phase", "phase": label, "event": "started", "completed": 0, "total": 8})
		h, _, err := canonicalScalingCodeHeader(s, 3, 1, 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		base, err := NewBinaryField(s, h.Polynomial)
		if err != nil {
			t.Fatal(err)
		}
		tables := newPreparationProducts(base)
		n := uint64(4)
		for n < rows*cols {
			n *= 2
		}
		gpu, err := cuda.NewRankStorage(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 128 << 20}, rows, cols, w.Rank, w.Size, 127)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer gpu.Close()
			retained := gpu.OwnedBytes()
			if !gpu.AdmitsTranspose(rows, cols, w.Rank, w.Size, 127) || gpu.AdmitsTranspose(rows, cols, 1-w.Rank, w.Size, 127) {
				t.Fatal("ownership admission")
			}
			if err = gpu.Local(0, 0); err == nil {
				t.Fatal("storage alias accepted")
			}
			if err = gpu.Reset(3, rows, cols, false); err == nil {
				t.Fatal("slot accepted")
			}
			if err = gpu.Reset(0, rows+1, cols, false); err == nil {
				t.Fatal("shape accepted")
			}
			if err = storageTransfer(gpu, 0, 0, []Word{0}, true); rows == 1 && w.Rank == 1 && err == nil {
				t.Fatal("empty owner extent accepted")
			}
			first, count := rankBand(rows, w.Rank, w.Size)
			input := make([]Word, count*cols)
			sentinel := make([]Word, len(input))
			for j := range input {
				input[j] = Word(storagePattern(first*cols + uint64(j)))
				sentinel[j] = ^input[j]
			}
			// Inputs intentionally include all slot bits, rather than only legal field
			// values: storage must preserve bits and must not hide arithmetic validation.
			for layout := 0; layout < 4; layout++ {
				rowInterleaved, colInterleaved := layout&1 != 0, layout&2 != 0
				if err = gpu.Reset(0, rows, cols, rowInterleaved); err != nil {
					t.Fatal(err)
				}
				if err = gpu.Reset(2, rows, cols, false); err != nil {
					t.Fatal(err)
				}
				storageIO(t, gpu, 0, input, true)
				storageIO(t, gpu, 2, sentinel, true)
				if err = storageTransfer(gpu, 0, count*cols, []Word{0}, true); err == nil {
					t.Fatal("out of bounds accepted")
				}
				for pass := 0; pass < 2; pass++ {
					a, b, src, dst, interleaved := rows, cols, 0, 1, colInterleaved
					if pass == 1 {
						a, b, src, dst, interleaved = cols, rows, 1, 0, rowInterleaved
					}
					op := *operation
					*operation++
					p, e := newRankTransposePlan(op, a, b, w.Rank, w.Size, n, 127)
					if e != nil {
						t.Fatal(e)
					}
					proofs := 0
					link := transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, tile []Word, own, remote [64]byte) {
						proofs++
						if mpi.ProofMismatch(own, remote) {
							saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-offset-%d-%s.bin", w.Rank, p.Operation, p.Offset, direction)), tile)
						}
						emit(map[string]any{"kind": "tile-proof", "case": label, "pass": 2*layout + pass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
					}}
					w.Barrier()
					began := time.Now()
					e = transposeResident[Word](p, gpu, src, dst, interleaved, link, func(phase string, done, total uint64) {
						emit(map[string]any{"kind": "progress", "case": label, "pass": 2*layout + pass, "operation": op, "phase": phase, "completed": done, "total": total})
					})
					if e != nil {
						t.Fatal(e)
					}
					seconds := w.Max(time.Since(began).Seconds())
					out := make([]Word, p.OutputWords)
					storageIO(t, gpu, dst, out, false)
					expected := make([]Word, len(out))
					// Independent global coordinate formula, not a shared packing helper.
					for j := range expected {
						r, c := p.OutputFirst+uint64(j)/a, uint64(j)%a
						index := c*cols + r
						if pass == 1 {
							index = r*cols + c
						}
						expected[j] = Word(storagePattern(index))
					}
					prefix := fmt.Sprintf("rank-%d-%s-pass-%d", w.Rank, label, 2*layout+pass)
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-output.bin"), out)
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-expected.bin"), expected)
					if !reflect.DeepEqual(out, expected) {
						t.Fatal("resident transpose mismatch", label, layout, pass)
					}
					source := make([]Word, p.InputWords)
					storageIO(t, gpu, src, source, false)
					for j, value := range source {
						r, c := p.InputFirst+uint64(j)/b, uint64(j)%b
						index := r*cols + c
						if pass == 1 {
							index = c*cols + r
						}
						if value != Word(storagePattern(index)) {
							t.Fatal("source changed", label, layout, pass, j)
						}
					}
					third := make([]Word, len(sentinel))
					storageIO(t, gpu, 2, third, false)
					if !reflect.DeepEqual(third, sentinel) {
						t.Fatal("third slot changed")
					}
					if gpu.OwnedBytes() != retained {
						t.Fatal("resident storage grew on reuse")
					}
					width := 4
					if _, ok := any(Word(0)).(uint64); ok {
						width = 8
					}
					emit(map[string]any{"kind": "measurement", "case": label, "s": s, "rows": a, "columns": b, "pass": 2*layout + pass, "operation": op, "word_bytes": width, "words": len(out), "matches": true, "source_preserved": true, "third_preserved": true, "reused": true, "row_interleaved": rowInterleaved, "column_interleaved": colInterleaved, "owned_device_bytes": retained, "seconds": seconds, "proof_events": proofs})
				}
			}
		}()
		emit(map[string]any{"kind": "phase", "phase": label, "event": "completed", "completed": 8, "total": 8})
	}
}

func TestDistributedGPUStorage(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded resident GPU storage gate")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		op := uint64(1)
		shapes := [][2]uint64{{1, 1}, {1, 3}, {3, 1}, {7, 11}, {3, 17}, {16, 16}}
		storageGPUCases[uint32](t, w, 4, shapes, output, &op, emit)
		storageGPUCases[uint64](t, w, 17, append(shapes, [2]uint64{64, 64}), output, &op, emit)
		emit(map[string]any{"kind": "completed", "cases": 104})
	})
}
