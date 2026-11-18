//go:build liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// TestDistributedRootProducts is an explicit two-node component gate. Both ranks
// rebuild the bounded immutable fixture, compute disjoint orbit-closed roots,
// exchange entire host polynomials, then merge on their GPU. Replicating the final
// merge/oracle here detects ownership errors; it is not a scalable merge tree or
// sharded support discovery. Ordinary tests never initialize MPI.
func TestDistributedRootProducts(t *testing.T) {
	directory := os.Getenv("PIR_DISTRIBUTED_ROOT_OUTPUT")
	if directory == "" {
		t.Skip("explicit two-node GPU root experiment")
	}
	world, err := mpi.Open()
	if err != nil {
		t.Fatal(err)
	}
	successful := false
	var failureLog func()
	defer func() {
		if !successful {
			defer world.Abort()
			if failureLog != nil {
				failureLog()
			}
		} else if err := world.Close(); err != nil {
			t.Error(err)
		}
	}()
	rank, err := strconv.Atoi(os.Getenv("SLURM_PROCID"))
	if err != nil || rank != world.Rank {
		t.Fatal("Slurm/MPI ownership")
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
		if usage.Maxrss*1024 > 2<<30 {
			panic("rank RSS guard")
		}
		row["utc"] = time.Now().UTC().Format(time.RFC3339Nano)
		row["elapsed_seconds"] = time.Since(began).Seconds()
		row["rank"] = rank
		row["host"] = host
		row["peak_rss_bytes"] = uint64(usage.Maxrss) * 1024
		if err := json.NewEncoder(os.Stdout).Encode(row); err != nil {
			panic(err)
		}
	}
	save := func(name string, values []uint32) string {
		file, err := os.OpenFile(filepath.Join(dest, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.New()
		if err = binary.Write(h, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		if err = binary.Write(file, binary.LittleEndian, values); err != nil {
			t.Fatal(err)
		}
		if err = file.Sync(); err != nil {
			t.Fatal(err)
		}
		if err = file.Close(); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	failureLog = func() {
		emit(map[string]any{"kind": "error", "phase": "workflow", "event": "failed", "remaining_work": "unknown"})
	}
	emit(map[string]any{"kind": "host", "visible_devices": os.Getenv("CUDA_VISIBLE_DEVICES")})
	cases := []struct {
		s, count int
		genuine  bool
	}{{8, 0, true}, {17, 1 << 16, false}, {17, 1 << 18, false}}
	if os.Getenv("PIR_DISTRIBUTED_ROOT_SMALL") == "1" {
		cases = cases[:1]
	}
	for index, item := range cases {
		emit(map[string]any{"kind": "progress", "case": index, "phase": "fixture", "event": "started", "completed": 0, "total": 1, "unit": "fixtures"})
		setup := time.Now()
		fixture, err := makeRootFixture(item.s, item.count, item.genuine)
		if err != nil {
			t.Fatal(err)
		}
		h := fixture.header
		tables := fixture.base.lookup
		native, err := cpp.NewPreparation(item.s, h.Polynomial, h.Omega, 1, uint64(fixture.maximum), 64<<20)
		if err != nil {
			t.Fatal(err)
		}
		gpu, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: item.s,
			Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers,
			Devices: []int{0}, MaxElements: uint64(fixture.maximum), DeviceBytes: 128 << 20})
		if err != nil {
			t.Fatal(err)
		}
		setupSeconds := time.Since(setup).Seconds()
		currentSample := -2
		stage := "CPU-reference"
		last := time.Now()
		notice := func(phase string, done, total uint64, unit string) {
			if done == total || time.Since(last) >= 30*time.Second {
				emit(map[string]any{"kind": "progress", "case": index, "phase": stage + "/" + phase, "sample": currentSample, "completed": done, "total": total, "unit": unit})
				last = time.Now()
			}
		}
		checkpoint := func(phase string) {
			emit(map[string]any{"kind": "stage", "case": index, "phase": stage + "/" + phase, "sample": currentSample, "event": "started"})
		}
		referenceStart := time.Now()
		reference, err := fixture.root(supportSorted, native, fixture.support, checkpoint, notice)
		if err != nil {
			t.Fatal(err)
		}
		if len(reference) != len(fixture.support)+1 || reference[0] != 1 {
			t.Fatal("CPU degree/normalization")
		}
		// Independent extension-field witnesses complement the complete native
		// coefficient oracle. Include endpoints and two interior support values.
		for check := 0; check < 4; check++ {
			exponent := fixture.support[check*(len(fixture.support)-1)/3]
			point := fixture.extension.pow(h.Omega, uint64(fixture.n-1-exponent))
			var value uint64
			for j := len(reference) - 1; j >= 0; j-- {
				value = fixture.extension.mul(value, point) ^ uint64(reference[j])
			}
			if value != 0 {
				t.Fatal("scalar root witness")
			}
		}
		referenceSeconds := time.Since(referenceStart).Seconds()
		digest := save(fmt.Sprintf("case-%d-reference.bin", index), reference)
		emit(map[string]any{"kind": "fixture", "case": index, "event": "completed", "s": item.s, "genuine_support": item.genuine,
			"support_count": len(fixture.support), "sha256": digest, "setup_seconds": setupSeconds, "reference_seconds": referenceSeconds})
		for sample := -1; sample < 2; sample++ {
			currentSample = sample
			world.Barrier()
			start := time.Now()
			stage = "partition"
			parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<uint(item.s), 2)
			if err != nil {
				t.Fatal(err)
			}
			partitionSeconds := time.Since(start).Seconds()
			phase := time.Now()
			stage = "local-root"
			emit(map[string]any{"kind": "stage", "case": index, "sample": sample, "phase": stage, "event": "started"})
			local, err := fixture.root(supportSorted, gpu, parts[rank], checkpoint, notice)
			if err != nil {
				t.Fatal(err)
			}
			rootSeconds := time.Since(phase).Seconds()
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "local-root", "event": "completed", "completed": len(parts[rank]), "total": len(parts[rank]), "unit": "exponents"})
			phase = time.Now()
			stage = "polynomial-transfer"
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": stage, "event": "started", "completed": 0, "total": len(local), "unit": "coefficients"})
			peer, err := world.Peer(local, len(fixture.support)+1)
			if err != nil {
				t.Fatal(err)
			}
			if len(peer) != len(parts[1-rank])+1 || peer[0] != 1 {
				t.Fatal("peer polynomial degree/normalization")
			}
			for _, v := range peer {
				if v >= 1<<uint(item.s) {
					t.Fatal("noncanonical peer coefficient")
				}
			}
			exchangeSeconds := time.Since(phase).Seconds()
			phase = time.Now()
			stage = "GPU-merge"
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "polynomial-transfer", "event": "completed", "completed": len(peer), "total": len(peer), "unit": "coefficients"})
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "GPU-merge", "event": "started", "completed": 0, "total": 1, "unit": "products"})
			var merged []uint32
			func() { defer recoverPreparation(&err); merged = polyMul(fixture.withEngine(gpu), local, peer) }()
			if err != nil {
				t.Fatal(err)
			}
			mergeSeconds := time.Since(phase).Seconds()
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "GPU-merge", "event": "completed", "completed": 1, "total": 1, "unit": "products"})
			phase = time.Now()
			stage = "verify"
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "verify", "event": "started", "completed": 0, "total": len(reference), "unit": "coefficients"})
			if !reflect.DeepEqual(merged, reference) {
				t.Fatal("distributed merged coefficients differ from complete CPU reference")
			}
			verifySeconds := time.Since(phase).Seconds()
			emit(map[string]any{"kind": "progress", "case": index, "sample": sample, "phase": "verify", "event": "completed", "completed": len(reference), "total": len(reference), "unit": "coefficients"})
			seconds := time.Since(start).Seconds()
			// Reduction is outside the pipeline; maxima are not additive.
			maxima := map[string]any{"seconds": world.Max(seconds), "partition_seconds": world.Max(partitionSeconds), "root_seconds": world.Max(rootSeconds),
				"exchange_seconds": world.Max(exchangeSeconds), "merge_seconds": world.Max(mergeSeconds), "verify_seconds": world.Max(verifySeconds)}
			if sample == 1 {
				save(fmt.Sprintf("case-%d-local.bin", index), local)
				save(fmt.Sprintf("case-%d-peer.bin", index), peer)
				if save(fmt.Sprintf("case-%d-merged.bin", index), merged) != digest {
					t.Fatal("saved merged hash")
				}
			}
			maxima["kind"] = "measurement"
			maxima["case"] = index
			maxima["sample"] = sample
			maxima["s"] = item.s
			maxima["support_count"] = len(fixture.support)
			maxima["local_count"] = len(parts[rank])
			maxima["peer_count"] = len(parts[1-rank])
			maxima["remote_bytes"] = 4 * len(local)
			maxima["owned_device_bytes"] = gpu.OwnedBytes()
			maxima["sha256"] = digest
			maxima["matches"] = true
			if gpu.OwnedBytes() > 128<<20 {
				t.Fatal("owned device guard")
			}
			emit(maxima)
		}
		if err = gpu.Close(); err != nil {
			t.Fatal(err)
		}
		if err = native.Close(); err != nil {
			t.Fatal(err)
		}
	}
	emit(map[string]any{"kind": "completed", "cases": len(cases)})
	successful = true
}
