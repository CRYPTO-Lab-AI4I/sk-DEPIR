//go:build liftrs_mpi && cgo

package hostbuffers

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

var mode = flag.String("buffer-mode", "direct", "direct or staged host payload")
var reclaim = flag.Bool("reclaim", false, "force Go reclamation before allocating each payload (diagnostic only)")

// Bound hashing scratch independently of payload and preserve all uint64 bits.
func fingerprint[T uint32 | uint64](values []T) [32]byte { return mpi.PayloadFingerprint(values) }

func loadPayload(path string) ([]uint32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) != 65537*4 {
		return nil, fmt.Errorf("replay extent")
	}
	p := make([]uint32, len(data)/4)
	for j := range p {
		p[j] = binary.LittleEndian.Uint32(data[j*4:])
	}
	return p, nil
}

// TestHostBuffers uses only MPI plus Go runtime work: no field or GPU code.
// Blocking transfers still borrow fresh Go slices, as in the failing tree path.
func TestHostBuffers(t *testing.T) {
	output := os.Getenv("PIR_HOST_BUFFER_OUTPUT")
	if output == "" {
		t.Skip("explicit host-buffer experiment")
	}
	original, err := loadPayload(os.Getenv("PIR_HOST_BUFFER_PAYLOAD"))
	if err != nil {
		t.Fatal(err)
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
	world.ChunkWords = 65536
	if *mode == "staged" {
		if err = world.EnableStaging(); err != nil {
			t.Fatal(err)
		}
	} else if *mode != "direct" {
		t.Fatal("buffer mode")
	}
	host, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(output, fmt.Sprintf("rank-%d", world.Rank))
	if err = os.MkdirAll(dest, 0700); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	var logMu sync.Mutex
	emit := func(row map[string]any) {
		logMu.Lock()
		defer logMu.Unlock()
		var usage syscall.Rusage
		if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
			t.Fatal(err)
		}
		if usage.Maxrss*1024 > 1<<30 {
			panic("rank 1 GiB RSS guard")
		}
		row["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		row["elapsed_seconds"] = time.Since(start).Seconds()
		row["rank"] = world.Rank
		row["host"] = host
		row["peak_rss_bytes"] = usage.Maxrss * 1024
		if err := json.NewEncoder(os.Stdout).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	emit(map[string]any{"kind": "host", "godebug": os.Getenv("GODEBUG"), "mpi_transport": mpi.TransportSettings(), "mode": *mode, "reclaim": *reclaim, "staging_bytes": world.StagingBytes(), "chunk_words": world.ChunkWords, "payload_sha256": fmt.Sprintf("%x", fingerprint(original))})
	var done, current atomic.Int64
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				emit(map[string]any{"kind": "heartbeat", "case": current.Load() % 4, "cycle": current.Load() / 4, "completed": done.Load(), "total": 256, "unit": "handoffs", "scope": "waiting heartbeat; unchanged count is not progress"})
			}
		}
	}()
	// Interleave a captured same-node root with a cross-node root and both
	// support routes. uint64 support values deliberately exercise high bits.
	cases := []struct{ sender, receiver, width int }{{1, 0, 4}, {2, 0, 4}, {0, 1, 8}, {0, 2, 8}}
	for cycle := 0; cycle < 64; cycle++ {
		for index, item := range cases {
			current.Store(int64(cycle*4 + index))
			world.Barrier()
			// Churn is identical in both modes; forced reclamation is a separate arm,
			// not normal preprocessing behavior. Release it before making live payloads.
			garbage := make([]byte, 16<<20)
			for j := 0; j < len(garbage); j += 4096 {
				garbage[j] = byte(cycle + 1)
			}
			runtime.KeepAlive(garbage)
			garbage = nil
			if *reclaim {
				debug.FreeOSMemory()
			}
			var values any
			var before, after [32]byte
			began := time.Now()
			if item.width == 4 {
				var p []uint32
				if world.Rank == item.sender {
					p = append([]uint32(nil), original...)
					before = fingerprint(p)
					err = world.SendRoot(item.receiver, p, len(original))
				}
				if world.Rank == item.receiver {
					p, err = world.ReceiveRoot(item.sender, len(original))
					if err == nil {
						before = fingerprint(p)
					}
				}
				if p != nil {
					after = fingerprint(p)
					values = p
				}
			} else {
				var p []uint64
				if world.Rank == item.sender {
					p = make([]uint64, 16384)
					for j := range p {
						p[j] = (1 << 34) | uint64(original[j])
					}
					before = fingerprint(p)
					err = world.SendIndices(item.receiver, p)
				}
				if world.Rank == item.receiver {
					p, err = world.ReceiveIndices()
					if err == nil {
						before = fingerprint(p)
					}
				}
				if p != nil {
					after = fingerprint(p)
					values = p
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if values != nil {
				peer := item.sender
				direction := "receive"
				if world.Rank == item.sender {
					peer = item.receiver
					direction = "send"
				}
				err = world.CheckPayloadProof(peer, before, after, func(own, remote [64]byte) {
					bad := mpi.ProofMismatch(own, remote)
					name := ""
					if bad {
						name = fmt.Sprintf("cycle-%d-case-%d-%s.bin", cycle, index, direction)
						f, e := os.OpenFile(filepath.Join(dest, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
						if e != nil {
							t.Fatal(e)
						}
						if e = binary.Write(f, binary.LittleEndian, values); e != nil {
							t.Fatal(e)
						}
						if e = f.Sync(); e != nil {
							t.Fatal(e)
						}
						if e = f.Close(); e != nil {
							t.Fatal(e)
						}
					}
					emit(map[string]any{"kind": "proof", "cycle": cycle, "case": index, "direction": direction, "peer": peer, "word_bytes": item.width, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote), "matches": !bad, "file": name, "transfer_seconds": time.Since(began).Seconds()})
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			world.Barrier()
			done.Store(int64(cycle*4 + index + 1))
			emit(map[string]any{"kind": "progress", "cycle": cycle, "case": index, "completed": cycle*len(cases) + index + 1, "total": 64 * len(cases), "unit": "handoffs", "event": "completed"})
		}
	}
	emit(map[string]any{"kind": "completed", "handoffs": 256})
	success = true
}

func TestFingerprintHighBitsAndTail(t *testing.T) {
	a := []uint32{1, 2, 0x1abcd}
	b := []uint64{1, 1 << 34, ^uint64(0)}
	for _, values := range []any{a, b} {
		h := sha256.New()
		if err := binary.Write(h, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		var actual [32]byte
		switch v := values.(type) {
		case []uint32:
			actual = fingerprint(v)
		case []uint64:
			actual = fingerprint(v)
		}
		if string(actual[:]) != string(h.Sum(nil)) {
			t.Fatal("fingerprint bits/extent")
		}
	}
	var same, other [64]byte
	other[0] = 1
	if mpi.ProofMismatch(same, same) || !mpi.ProofMismatch(same, other) || !mpi.ProofMismatch(other, same) {
		t.Fatal("proof symmetry")
	}
}
