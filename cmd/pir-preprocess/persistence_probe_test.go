//go:build linux

package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Synthetic public base-field coefficients exercise the real persistence path,
// not a valid root/code/checkpoint. A fixed shard seed gives identical bytes for
// all buffer/concurrency modes without retaining a large input array.
func persistencePattern(w io.Writer, bytes, shard int) error {
	state := uint32(shard + 1)
	var block [65536]byte
	for first := 0; first < bytes; first += len(block) {
		take := len(block)
		if take > bytes-first {
			take = bytes - first
		}
		for j := 0; j < take; j += 4 {
			state ^= state << 13
			state ^= state >> 17
			state ^= state << 5
			binary.LittleEndian.PutUint32(block[j:], state&((1<<17)-1))
		}
		n, err := w.Write(block[:take])
		if err != nil {
			return err
		}
		if n != take {
			return io.ErrShortWrite
		}
	}
	return nil
}

type persistenceLog struct {
	emit         func(map[string]any)
	trial, shard int
}

func (l persistenceLog) Write(p []byte) (int, error) {
	var row map[string]any
	if err := json.Unmarshal(p, &row); err != nil {
		return 0, err
	}
	row["trial"], row["shard"] = l.trial, l.shard
	l.emit(row)
	return len(p), nil
}

type persistenceFile struct {
	Stage           string            `json:"stage"`
	Record          fileRecord        `json:"record"`
	Stats           payloadWriteStats `json:"stats"`
	PublishSeconds  float64           `json:"publish_seconds"`
	ReadbackSeconds float64           `json:"readback_seconds"`
	Writer          int               `json:"writer"`
	Shard           int               `json:"shard"`
	PID             int               `json:"pid"`
}

// TestPersistenceProbe is opt-in and keeps a fixed aggregate load in every case.
// Writers own disjoint publication stages. Completed synthetic payloads from
// earlier cases are pruned only after full verification; manifests/digests and
// the final complete case remain. Expensive preprocessing is never consumed or
// deleted. Ordinary builds/CLI defaults never select the experimental buffer.
func TestPersistenceProbe(t *testing.T) {
	directory := os.Getenv("PIR_PERSISTENCE_OUTPUT")
	if directory == "" {
		t.Skip("explicit persistence experiment")
	}
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	began := time.Now()
	var mu sync.Mutex
	emit := func(row map[string]any) {
		mu.Lock()
		defer mu.Unlock()
		var u syscall.Rusage
		if err := syscall.Getrusage(syscall.RUSAGE_SELF, &u); err != nil {
			panic(err)
		}
		if u.Maxrss*1024 > 2<<30 {
			panic("RSS guard")
		}
		row["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		row["elapsed_seconds"] = time.Since(began).Seconds()
		row["peak_rss_bytes"] = u.Maxrss * 1024
		if err := json.NewEncoder(os.Stdout).Encode(row); err != nil {
			panic(err)
		}
	}
	successful := false
	defer func() {
		if !successful {
			emit(map[string]any{"kind": "error", "phase": "persistence-probe", "event": "failed"})
		}
	}()
	fileBytes := 24 << 20
	if os.Getenv("PIR_PERSISTENCE_SMALL") == "1" {
		fileBytes = 128 << 10
	}
	var expected [4]string
	for shard := range expected {
		h := sha256.New()
		if err := persistencePattern(h, fileBytes, shard); err != nil {
			t.Fatal(err)
		}
		expected[shard] = hex.EncodeToString(h.Sum(nil))
	}
	type mode struct{ buffer, writers int }
	modes := []mode{{0, 1}, {4 << 20, 1}, {4 << 20, 2}, {0, 2}, {0, 4}, {4 << 20, 4}}
	order := append([]mode(nil), modes...)
	for j := len(modes) - 1; j >= 0; j-- {
		order = append(order, modes[j])
	}
	spec := lrs.Spec{FieldDegree: 17, Dimension: 2, CurveDegree: 5}
	for trial, mode := range order {
		emit(map[string]any{"kind": "progress", "trial": trial, "phase": "publication", "event": "started", "completed": 0, "total": 4, "unit": "files", "buffer_bytes": mode.buffer, "writers": mode.writers})
		results := make([]persistenceFile, 4)
		errors := make(chan error, mode.writers)
		var group sync.WaitGroup
		start := time.Now()
		completed := 0
		for writer := 0; writer < mode.writers; writer++ {
			group.Add(1)
			go func(writer int) {
				defer group.Done()
				for shard := writer; shard < 4; shard += mode.writers {
					stage := fmt.Sprintf("io-trial-%02d-file-%d", trial, shard)
					report := &reporter{persistenceLog{emit, trial, shard}, began}
					result := persistenceFile{Stage: stage, Writer: writer, Shard: shard, PID: os.Getpid()}
					m := manifest{Version: 1, Stage: stage, Spec: spec, Files: map[string]fileRecord{}}
					publishStart := time.Now()
					_, err := publish(directory, stage, m, func(dir string, m *manifest) error {
						record, err := writePayloadBuffered(dir, "payload.bin", uint64(fileBytes), report, mode.buffer, &result.Stats,
							hashed(func(w io.Writer) error { return persistencePattern(w, fileBytes, shard) }))
						if err != nil {
							return err
						}
						if record.SHA256 != expected[shard] {
							return fmt.Errorf("producer digest mismatch")
						}
						result.Record = record
						m.Files["payload.bin"] = record
						return nil
					})
					result.PublishSeconds = time.Since(publishStart).Seconds()
					if err != nil {
						errors <- err
						return
					}
					results[shard] = result
					mu.Lock()
					completed++
					done := completed
					mu.Unlock()
					emit(map[string]any{"kind": "progress", "trial": trial, "shard": shard, "phase": "publication", "event": "progress", "completed": done, "total": 4, "unit": "files"})
				}
			}(writer)
		}
		group.Wait()
		close(errors)
		publicationSeconds := time.Since(start).Seconds()
		for err := range errors {
			t.Fatal(err)
		}
		emit(map[string]any{"kind": "progress", "trial": trial, "phase": "publication", "event": "completed", "completed": 4, "total": 4, "unit": "files"})
		readStart := time.Now()
		for shard := range results {
			result := &results[shard]
			report := &reporter{persistenceLog{emit, trial, shard}, began}
			start := time.Now()
			loaded, err := readManifest(directory, result.Stage, spec)
			if err != nil {
				t.Fatal(err)
			}
			if loaded.Files["payload.bin"] != result.Record {
				t.Fatal("restart manifest differs")
			}
			if err = verifyFile(directory, result.Stage, "payload.bin", result.Record, report); err != nil {
				t.Fatal(err)
			}
			result.ReadbackSeconds = time.Since(start).Seconds()
			emit(map[string]any{"kind": "progress", "trial": trial, "shard": shard, "phase": "restart-readback", "event": "progress", "completed": shard + 1, "total": 4, "unit": "files"})
		}
		readbackSeconds := time.Since(readStart).Seconds()
		retained := trial == len(order)-1
		emit(map[string]any{"kind": "measurement", "trial": trial, "repeat": trial / 6, "writers": mode.writers, "buffer_bytes": mode.buffer,
			"codec_chunk_bytes": 65536, "file_bytes": fileBytes, "total_bytes": 4 * fileBytes, "publication_seconds": publicationSeconds,
			"readback_seconds": readbackSeconds, "publish_readback_seconds": time.Since(start).Seconds(), "files": results, "matches": true, "payloads_retained": retained})
		if !retained {
			// Delete only this probe's cheap, just-verified synthetic payloads;
			// leave every manifest, checksum/profile and failed partial stage.
			for _, result := range results {
				if err := os.Remove(filepath.Join(directory, result.Stage, "payload.bin")); err != nil {
					t.Fatal(err)
				}
			}
			emit(map[string]any{"kind": "progress", "trial": trial, "phase": "prune-verified-synthetic", "event": "completed", "completed": 4, "total": 4, "unit": "files"})
		}
	}
	emit(map[string]any{"kind": "completed", "trials": len(order), "file_bytes": fileBytes})
	successful = true
}
