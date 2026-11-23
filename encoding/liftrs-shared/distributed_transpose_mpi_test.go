//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// Input coordinates are generated locally; validation computes the transposed
// coordinate directly, independently of the packing schedule. No full matrix or
// full reference is gathered on rank0. Wide patterns exercise every storage bit.
func transposePattern[Word transposeWord](index uint64) Word {
	return Word(index*0x9e3779b97f4a7c15 ^ 0xa5a5a5a5a5a5a5a5)
}
func transposeMPICases[Word transposeWord](t *testing.T, w *mpi.World, emit func(map[string]any), output string, operation *atomic.Uint64) {
	var zero Word
	width := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	for caseIndex, shape := range [][2]uint64{{1, 3}, {7, 11}, {64, 128}, {1024, 256}} {
		rows, columns := shape[0], shape[1]
		tileWords := uint64(7)
		if caseIndex == 3 {
			tileWords = 16384
		}
		op := operation.Add(1)
		plan, err := newRankTransposePlan(op, rows, columns, w.Rank, w.Size, 1<<18, tileWords)
		if err != nil {
			t.Fatal(err)
		}
		input := make([]Word, int(plan.InputWords))
		for j := range input {
			input[j] = transposePattern[Word](plan.InputFirst*columns + uint64(j))
		}
		before := mpi.PayloadFingerprint(input)
		emit(map[string]any{"kind": "stage", "event": "started", "operation": op, "word_bytes": width, "rows": rows, "columns": columns, "tile_words": tileWords, "input_words": plan.InputWords, "output_words": plan.OutputWords, "scratch_words": plan.ScratchWords})
		link := transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, tile []Word, own, remote [64]byte) {
			bad := mpi.ProofMismatch(own, remote)
			file := ""
			if bad {
				file = fmt.Sprintf("rank-%d-op-%d-round-%d-offset-%d-%s.bin", w.Rank, p.Operation, p.Round, p.Offset, direction)
				f, e := os.OpenFile(output+"/"+file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				if e != nil {
					t.Fatal(e)
				}
				if e = binary.Write(f, binary.LittleEndian, tile); e != nil {
					t.Fatal(e)
				}
				if e = f.Sync(); e != nil {
					t.Fatal(e)
				}
				if e = f.Close(); e != nil {
					t.Fatal(e)
				}
			}
			emit(map[string]any{"kind": "tile-proof", "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "word_bytes": width, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote), "matches": !bad, "file": file})
		}}
		w.Barrier()
		start := time.Now()
		last := start
		out, err := transposeRank(plan, input, link, func(phase string, done, total uint64) {
			if done == total || time.Since(last) >= 30*time.Second {
				emit(map[string]any{"kind": "progress", "operation": op, "phase": phase, "completed": done, "total": total, "unit": "words"})
				last = time.Now()
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		seconds := time.Since(start).Seconds()
		// Check every output element and retain the original input fingerprint.
		if uint64(len(out)) != plan.OutputWords || mpi.PayloadFingerprint(input) != before {
			t.Fatal("transpose extent/input mutation")
		}
		for j, value := range out {
			column := plan.OutputFirst + uint64(j)/rows
			row := uint64(j) % rows
			if value != transposePattern[Word](row*columns+column) {
				t.Fatal("independent transpose coordinate mismatch", op, j)
			}
		}
		if caseIndex == 3 {
			file := fmt.Sprintf("%s/rank-%d-op-%d-output.bin", output, w.Rank, op)
			f, e := os.OpenFile(file, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if e != nil {
				t.Fatal(e)
			}
			if e = binary.Write(f, binary.LittleEndian, out); e != nil {
				t.Fatal(e)
			}
			if e = f.Sync(); e != nil {
				t.Fatal(e)
			}
			if e = f.Close(); e != nil {
				t.Fatal(e)
			}
		}
		w.Barrier()
		maxSeconds := w.Max(seconds)
		emit(map[string]any{"kind": "measurement", "operation": op, "rows": rows, "columns": columns, "word_bytes": width, "input_words": plan.InputWords, "output_words": plan.OutputWords, "scratch_words": plan.ScratchWords, "seconds": maxSeconds, "matches": true, "output_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(out))})
	}
}

// Preserve the pre-existing root20/21 and support10/11 wire paths after adding
// coefficient lanes. These are transport fixtures, not polynomial preparation.
func transposeLegacyLanes(t *testing.T, w *mpi.World, emit func(map[string]any)) {
	w.Barrier()
	if w.Rank == 1 {
		p := make([]uint32, 65537)
		for j := range p {
			p[j] = uint32(j) & 131071
		}
		p[0] = 1
		if err := w.SendRoot(0, p, len(p)); err != nil {
			t.Fatal(err)
		}
	}
	if w.Rank == 0 {
		p, err := w.ReceiveRoot(1, 65537)
		if err != nil || len(p) != 65537 {
			t.Fatal("legacy root extent", err)
		}
		for j, value := range p {
			want := uint32(j) & 131071
			if j == 0 {
				want = 1
			}
			if value != want {
				t.Fatal("legacy root value")
			}
		}
	}
	w.Barrier()
	emit(map[string]any{"kind": "legacy-check", "lane": "root32", "matches": true})
	if w.Rank == 0 {
		p := make([]uint64, 16384)
		for j := range p {
			p[j] = (1 << 33) | uint64(j)
		}
		if err := w.SendIndices(2, p); err != nil {
			t.Fatal(err)
		}
	}
	if w.Rank == 2 {
		p, err := w.ReceiveIndices()
		if err != nil || len(p) != 16384 {
			t.Fatal("legacy support extent", err)
		}
		for j, value := range p {
			if value != (1<<33)|uint64(j) {
				t.Fatal("legacy support high bits")
			}
		}
	}
	w.Barrier()
	emit(map[string]any{"kind": "legacy-check", "lane": "support64", "matches": true})
}

func TestDistributedTransposeMPI(t *testing.T) {
	output := os.Getenv("PIR_TRANSPOSE_OUTPUT")
	if output == "" {
		t.Skip("explicit distributed transpose experiment")
	}
	world, err := mpi.Open()
	if err != nil {
		t.Fatal(err)
	}
	success := false
	defer func() {
		if !success {
			world.Abort()
		} else if err := world.Close(); err != nil {
			t.Error(err)
		}
	}()
	if world.Size != 4 {
		t.Fatal("four ranks required")
	}
	world.ChunkWords = 1024
	mode := os.Getenv("PIR_TRANSPOSE_BUFFER")
	if mode == "staged" {
		if err = world.EnableStaging(); err != nil {
			t.Fatal(err)
		}
	} else if mode != "direct" {
		t.Fatal("explicit transpose buffer mode")
	}
	host, err := os.Hostname()
	if err != nil {
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
			panic("transpose1 GiB RSS guard")
		}
		row["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		row["elapsed_seconds"] = time.Since(began).Seconds()
		row["rank"] = world.Rank
		row["host"] = host
		row["peak_rss_bytes"] = usage.Maxrss * 1024
		if err := json.NewEncoder(os.Stdout).Encode(row); err != nil {
			panic(err)
		}
	}
	var operation atomic.Uint64
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		timer := time.NewTicker(30 * time.Second)
		defer timer.Stop()
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
				emit(map[string]any{"kind": "heartbeat", "operation": operation.Load(), "scope": "current operation, not a completion count; waiting is not progress"})
			}
		}
	}()
	emit(map[string]any{"kind": "host", "ranks": world.Size, "mode": mode, "chunk_words": world.ChunkWords, "staging_bytes": world.StagingBytes(), "mpi_transport": mpi.TransportSettings()})
	transposeMPICases[uint32](t, world, emit, output, &operation)
	transposeMPICases[uint64](t, world, emit, output, &operation)
	transposeLegacyLanes(t, world, emit)
	emit(map[string]any{"kind": "completed", "operations": operation.Load()})
	success = true
}

func TestTransposeDescriptorWidthAndOperation(t *testing.T) {
	packet := transposePacket{Operation: 1, Rows: 1 << 18, Columns: 1 << 18, Round: 3, Offset: 1 << 33, Words: 1 << 34}
	a := transposeDescriptor[uint32](packet, 2)
	b := transposeDescriptor[uint64](packet, 2)
	if a == b || binary.LittleEndian.Uint64(b[32:]) != 1<<33 || binary.LittleEndian.Uint64(b[40:]) != 1<<34 {
		t.Fatal("descriptor width/count truncation")
	}
	packet.Operation++
	if transposeDescriptor[uint32](packet, 2) == a {
		t.Fatal("operation alias")
	}
	packet.Operation--
	if transposeDescriptor[uint32](packet, 3) == a {
		t.Fatal("source alias")
	}
}
