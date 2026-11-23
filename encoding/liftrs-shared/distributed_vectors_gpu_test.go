//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"
)

func saveMatrixStageWords[Word transposeWord](t *testing.T, path string, data []Word) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err = binary.Write(f, binary.LittleEndian, data); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatal(err, closeErr)
	}
}

func matrixStagePattern(index uint64, s int) uint64 {
	return ((index+1)*0x9e3779b9 ^ (uint64(1) << uint(2*s-1))) & ((uint64(1) << uint(2*s)) - 1)
}

// Build an independent native oracle using one temporary CPU column plus this
// rank's expected rows. Gamma is the explicit subspace vanishing product, not
// metadata copied from the GPU builder. No full-domain array is gathered.
func matrixStageOracle(t *testing.T, s, dim int, rank, ranks int) (ScalingCodeHeader, []uint64) {
	t.Helper()
	h, f, err := canonicalScalingCodeHeader(s, 3, 1, 1<<uint(2*s))
	if err != nil {
		t.Fatal(err)
	}
	split := dim / 2
	rows, cols := uint64(1)<<uint(dim-split), uint64(1)<<uint(split)
	if os.Getenv("PIR_GPU_MATRIX_STAGE_TAYLOR") == "1" {
		// Small diagnostic oracle only: independent native full FFT on a local
		// synthetic grid (<=4096 words), never a coordinator/network gather.
		matrix, e := cpp.NewWideMatrix(s, h.Polynomial, dim, split)
		if e != nil {
			t.Fatal(e)
		}
		defer matrix.Close()
		full := make([]uint64, rows*cols)
		scratch := make([]uint64, len(full))
		for j := range full {
			full[j] = matrixStagePattern(uint64(j), s)
		}
		if _, e = matrix.Apply(full, scratch, 1, false); e != nil {
			t.Fatal(e)
		}
		first, count := rankBand(rows, rank, ranks)
		local := full[first*cols : (first+count)*cols]
		if len(local) > 0 {
			for _, j := range []int{0, len(local) - 1} {
				point := first*cols + uint64(j)
				want := uint64(0)
				for k := len(full) - 1; k >= 0; k-- {
					want = f.mul(want, point) ^ matrixStagePattern(uint64(k), s)
				}
				if local[j] != want {
					t.Fatal("native full FFT/scalar Horner", s, dim, point)
				}
			}
		}
		return h, append([]uint64(nil), local...)
	}
	low := make([]uint64, split)
	for k := range low {
		low[k] = uint64(1) << uint(k)
	}
	gamma := make([]uint64, dim-split)
	for k := range gamma {
		x := uint64(1) << uint(split+k)
		product := uint64(1)
		for a := uint64(0); a < cols; a++ {
			product = f.mul(product, x^a)
		}
		gamma[k] = product
	}
	rowPlan, err := cpp.NewWidePlan(s, h.Polynomial, low)
	if err != nil {
		t.Fatal(err)
	}
	defer rowPlan.Close()
	colPlan, err := cpp.NewWidePlan(s, h.Polynomial, gamma)
	if err != nil {
		t.Fatal(err)
	}
	defer colPlan.Close()
	first, count := rankBand(rows, rank, ranks)
	expected := make([]uint64, count*cols)
	column := make([]uint64, rows)
	for c := uint64(0); c < cols; c++ {
		for r := uint64(0); r < rows; r++ {
			column[r] = matrixStagePattern(r*cols+c, s)
		}
		if _, err = colPlan.ApplyAt(column, 0, false); err != nil {
			t.Fatal(err)
		}
		for r := uint64(0); r < count; r++ {
			expected[r*cols+c] = column[first+r]
		}
	}
	for r := uint64(0); r < count; r++ {
		if _, err = rowPlan.ApplyAt(expected[r*cols:(r+1)*cols], (first+r)*cols, false); err != nil {
			t.Fatal(err)
		}
	}
	return h, expected
}
func matrixStageGPUCases[Word transposeWord](t *testing.T, w *mpi.World, s int, dims []int, output string, operation *uint64, emit func(map[string]any)) {
	for _, dim := range dims {
		label := fmt.Sprintf("s%d-d%d", s, dim)
		emit(map[string]any{"kind": "phase", "phase": label, "event": "started", "completed": 0, "total": 2})
		h, reference := matrixStageOracle(t, s, dim, w.Rank, w.Size)
		n := uint64(1) << uint(dim)
		rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
		localFirst, localRows := rankBand(rows, w.Rank, w.Size)
		base, err := NewBinaryField(s, h.Polynomial)
		if err != nil {
			t.Fatal(err)
		}
		tables := newPreparationProducts(base)
		capacity := min(n, 3*max(rows, cols))
		gpu, err := cuda.NewVectorPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 128 << 20}, capacity)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer gpu.Close()
			input := make([]Word, localRows*cols)
			expected := make([]Word, len(input))
			for j := range input {
				input[j] = Word(matrixStagePattern(localFirst*cols+uint64(j), s))
				expected[j] = Word(reference[j])
			}
			apply := func(tile []Word, columns, inverse bool, first, count uint64) error {
				switch values := any(tile).(type) {
				case []uint32:
					return gpu.Apply32(values, columns, inverse, first, count)
				case []uint64:
					return gpu.Apply64(values, columns, inverse, first, count)
				}
				return fmt.Errorf("stage slot width")
			}
			if e := apply(input, false, false, n, 1); e == nil {
				t.Fatal("out-of-range vector accepted")
			}
			current := append([]Word(nil), input...)
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
				windows := taylorMPI[Word]{world: w, record: func(direction string, peer int, p taylorPacket, data []Word, own, remote [64]byte) {
					taylorProofs++
					if mpi.ProofMismatch(own, remote) {
						saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("capture-rank-%d-op-%d-block-%d-first-%d-slot-%d-source-%d-%s.bin", w.Rank, p.Operation, p.Target.Block, p.Target.First, p.Slot, p.Source, direction)), data)
					}
					emit(map[string]any{"kind": "taylor-proof", "case": label, "pass": pass, "peer": peer, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				degrees := func(level int) ([]uint64, error) {
					terms, e := gpu.TaylorDegrees(level)
					if e == nil {
						emit(map[string]any{"kind": "taylor-plan", "case": label, "pass": pass, "operation_start": op, "level": level, "degrees": terms, "capacity": capacity})
					}
					return terms, e
				}
				update := func(snapshot []Word, level int, q bool, block, first, count uint64) error {
					before := mpi.PayloadFingerprint(snapshot[count:])
					tile := cuda.TaylorTile{Level: level, Quotient: q, Block: block, First: first, Count: count}
					var e error
					switch values := any(snapshot).(type) {
					case []uint32:
						e = gpu.Taylor32(values, tile)
					case []uint64:
						e = gpu.Taylor64(values, tile)
					}
					if e == nil && before != mpi.PayloadFingerprint(snapshot[count:]) {
						return fmt.Errorf("Taylor mutated source snapshot")
					}
					return e
				}
				w.Barrier()
				began := time.Now()
				last := began
				notice := func(phase string, done, total uint64) {
					if done == 0 || done == total || time.Since(last) >= 30*time.Second {
						event := "progress"
						if done == 0 {
							event = "started"
						}
						if done == total {
							event = "completed"
						}
						emit(map[string]any{"kind": "progress", "event": event, "case": label, "pass": pass, "phase": phase, "completed": done, "total": total})
						last = time.Now()
					}
				}
				var got []Word
				if fullFFT {
					got, e = rankAdditiveTransform(p, current, capacity, pass == 1, link, windows, degrees, apply, update, notice)
				} else {
					got, e = rankMatrixVectors(p, current, capacity, pass == 1, link, apply, notice)
				}
				if e != nil {
					t.Fatal(e)
				}
				seconds := w.Max(time.Since(began).Seconds())
				if pass == 0 {
					retained = gpu.OwnedBytes()
				} else if gpu.OwnedBytes() != retained {
					t.Fatal("stage allocation grew on inverse reuse")
				}
				want := expected
				if pass == 1 {
					want = input
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatal("distributed GPU stage/native oracle", label, pass)
				}
				prefix := fmt.Sprintf("rank-%d-%s-pass-%d", w.Rank, label, pass)
				for _, array := range []struct {
					name string
					data []Word
				}{{"output", got}, {"expected", want}} {
					saveMatrixStageWords(t, filepath.Join(output, prefix+"-"+array.name+".bin"), array.data)
				}
				emit(map[string]any{"kind": "measurement", "case": label, "pass": pass, "s": s, "dimension": dim, "rank_first": localFirst, "rank_rows": localRows, "word_bytes": h.ExtensionBytes, "words": len(got), "seconds": seconds, "proof_events": proofs, "taylor_proof_events": taylorProofs, "full_fft": fullFFT, "operation_start": op, "owned_device_bytes": gpu.OwnedBytes(), "output_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(got)), "matches": true})
				current = got
			}
		}()
		emit(map[string]any{"kind": "phase", "phase": label, "event": "completed", "completed": 2, "total": 2})
	}
}

// withDistributedGPUStages shares MPI ownership, logging and admitted guards.
func withDistributedGPUStages(t *testing.T, output string, body func(*mpi.World, func(map[string]any))) {
	withRankGPUStages(t, output, 2, body)
}

// Four workers share two admitted GPUs;64MiB/worker bounds aggregate owned
// payload to the existing128MiB/card budget. Ordinary gates remain two workers.
func withRankGPUStages(t *testing.T, output string, ranks int, body func(*mpi.World, func(map[string]any))) {
	for _, path := range []string{os.Args[0], os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), output} {
		if _, err := os.Stat(path); err != nil {
			t.Fatal("rank path preflight", path, err)
		}
	}
	w, err := mpi.Open()
	if err != nil {
		t.Fatal(err)
	}
	success := false
	defer func() {
		if !success {
			w.Abort()
		} else if err := w.Close(); err != nil {
			t.Error(err)
		}
	}()
	if (ranks != 2 && ranks != 4) || w.Size != ranks {
		t.Fatal("rank count admission")
	}
	w.ChunkWords = 1024
	if err = w.EnableStaging(); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	recorder, recordErr := newProofLog(os.Getenv("PIR_PIPELINE_PROOF_LOG"))
	if recordErr != nil {
		t.Fatal(recordErr)
	}
	began := time.Now()
	var mu sync.Mutex
	lastProofProgress := time.Now()
	emit := func(row map[string]any) {
		mu.Lock()
		defer mu.Unlock()
		var usage syscall.Rusage
		if e := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); e != nil || usage.Maxrss >= 1<<20 {
			panic("rank RSS guard")
		}
		hide, err := recorder.Record(row)
		if err != nil {
			panic(err)
		}
		if hide {
			if time.Since(lastProofProgress) < 30*time.Second {
				return
			}
			lastProofProgress = time.Now()
			row = map[string]any{"kind": "proof-progress", "phase": "transport-proof-recording", "completed": recorder.total, "unit": "proof records", "total_known": false}
		}
		row["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		row["rank"] = w.Rank
		row["elapsed_seconds"] = time.Since(began).Seconds()
		row["peak_rss_bytes"] = uint64(usage.Maxrss) * 1024
		if e := json.NewEncoder(os.Stdout).Encode(row); e != nil {
			panic(e)
		}
	}
	defer func() {
		// A footer write/resource failure must still abort peers,not enter finalize.
		complete := success
		success = false
		mu.Lock()
		summaries := recorder.Summaries(complete)
		mu.Unlock()
		for _, row := range summaries {
			emit(row)
		}
		success = complete
	}()
	host, e := os.Hostname()
	if e != nil {
		t.Fatal(e)
	}
	emit(map[string]any{"kind": "host", "host": host, "visible_devices": os.Getenv("CUDA_VISIBLE_DEVICES")})
	body(w, emit)
	w.Barrier()
	success = true
}

// TestDistributedGPUVectors explicitly selects separable or complete FFT cases.
func TestDistributedGPUVectors(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded MPI GPU matrix-vector gate")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		operation := uint64(1)
		matrixStageGPUCases[uint32](t, w, 4, []int{4, 5, 8}, output, &operation, emit)
		matrixStageGPUCases[uint64](t, w, 17, []int{5, 9, 12}, output, &operation, emit)
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}
