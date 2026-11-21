//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// rootProductInput connects local GPU factor construction to the shared
// distributed final product. This bounded gate replicates support/CPU reference
// and explicitly selects peer exchange or coefficient routing for comparison.
// Replicated support/native references do not establish a fullsize s18 layout.
// Only the final FFT/product is sharded, with no gathered GPU output.
func rootProductInput(t *testing.T, w *mpi.World, directory string, emit func(map[string]any), layout string) func(int, int, int, uint64, uint64) productCaseInput {
	return func(s, dim, shape int, first, words uint64) productCaseInput {
		t.Helper()
		n := 1 << uint(dim)
		if s != 17 || shape != 0 || (dim != 12 && dim != 14 && dim != 16) || w.Size != 2 {
			t.Fatal("root product fixture admission")
		}
		started := time.Now()
		fixture, err := makeRootFixture(s, n-4, false)
		if err != nil {
			t.Fatal(err)
		}
		parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<uint(s), 2)
		if err != nil || len(parts[0]) != n/2-2 || len(parts[1]) != n/2-2 {
			t.Fatal("root orbit partition", err)
		}
		phase := "root-reference"
		last := time.Now()
		checkpoint := func(stage string) {
			emit(map[string]any{"kind": "stage", "phase": phase + "/" + stage, "event": "started"})
		}
		notice := func(stage string, done, total uint64, unit string) {
			if done == total || time.Since(last) >= 30*time.Second {
				emit(map[string]any{"kind": "progress", "phase": phase + "/" + stage, "completed": done, "total": total, "unit": unit})
				last = time.Now()
			}
		}
		emit(map[string]any{"kind": "stage", "phase": phase, "event": "started"})
		native, err := cpp.NewPreparation(s, fixture.header.Polynomial, fixture.header.Omega, 1, uint64(fixture.maximum), 64<<20)
		if err != nil {
			t.Fatal(err)
		}
		reference, err := fixture.root(supportSorted, native, fixture.support, checkpoint, notice)
		closeErr := native.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
		if len(reference) != len(fixture.support)+1 || reference[0] != 1 {
			t.Fatal("root reference normalization/degree")
		}
		// Independent scalar witnesses check four support roots in the extension.
		for check := 0; check < 4; check++ {
			exponent := fixture.support[check*(len(fixture.support)-1)/3]
			point := fixture.extension.pow(fixture.header.Omega, uint64(fixture.n-1-exponent))
			var value uint64
			for j := len(reference) - 1; j >= 0; j-- {
				value = fixture.extension.mul(value, point) ^ uint64(reference[j])
			}
			if value != 0 {
				t.Fatal("root reference scalar witness")
			}
		}
		referenceSeconds := time.Since(started).Seconds()
		phase = "local-root"
		checkpoint("begin")
		tables := fixture.base.lookup
		gpu, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: fixture.header.Polynomial, Delta: fixture.header.Delta, Omega: fixture.header.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: uint64(fixture.maximum), DeviceBytes: 64 << 20})
		if err != nil {
			t.Fatal(err)
		}
		rootBegan := time.Now()
		local, err := fixture.root(supportSorted, gpu, parts[w.Rank], checkpoint, notice)
		owned := gpu.OwnedBytes()
		closeErr = gpu.Close() // Release factor workspace before distributed FFT work.
		if err != nil || closeErr != nil || owned > 64<<20 {
			t.Fatal("local root", err, closeErr)
		}
		rootSeconds := time.Since(rootBegan).Seconds()
		if len(local) != n/2-1 || local[0] != 1 {
			t.Fatal("local root extent/normalization")
		}
		phase = "root-operand-exchange"
		checkpoint("begin")
		before := mpi.PayloadFingerprint(local)
		exchangeBegan := time.Now()
		var peer []uint32
		result := productCaseInput{Expected: make([]uint32, words), ACount: uint64(len(local)), BCount: uint64(len(local)), Leading: reference[len(reference)-1]}
		var sentWords, receivedWords uint64
		if layout == "peer" {
			result.Left, result.Right = make([]uint32, words), make([]uint32, words)
			peer, err = w.Peer(local, n/2-1)
			if err != nil || len(peer) != len(local) || peer[0] != 1 {
				t.Fatal("root peer extent", err)
			}
			for _, v := range peer {
				if v >= 1<<uint(s) {
					t.Fatal("noncanonical peer root")
				}
			}
			sentWords, receivedWords = uint64(len(local)), uint64(len(peer))
			a, b := local, peer
			if w.Rank == 1 {
				a, b = peer, local
			}
			for j := range result.Left {
				k := first + uint64(j)
				if k < uint64(len(a)) {
					result.Left[j] = a[k]
				}
				if k < uint64(len(b)) {
					result.Right[j] = b[k]
				}
			}
		} else {
			link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
				if direction == "send" {
					sentWords += p.Words
				} else {
					receivedWords += p.Words
				}
				if mpi.ProofMismatch(own, remote) {
					saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("capture-rank-%d-operand-%d-offset-%d-%s.bin", w.Rank, p.Operand, p.Offset, direction)), data)
				}
				emit(map[string]any{"kind": "operand-proof", "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
			}}
			var inputs [2][]uint32
			inputs[w.Rank] = local
			shards, e := routePolynomialOperands(rankOperandPlan{Operation: 1, Domain: uint64(n), TileWords: 4096, Limit: 1 << uint(s), MaxLocalWords: words, Counts: [2]uint64{uint64(len(local)), uint64(len(local))}, Owners: [2]int{0, 1}, Rank: w.Rank, Ranks: w.Size}, inputs, link, func(operand int, done, total uint64) {
				notice(fmt.Sprintf("operand-%d", operand), done, total, "coefficients")
			})
			if e != nil {
				t.Fatal(e)
			}
			result.Left, result.Right = shards[0], shards[1]
		}
		if mpi.PayloadFingerprint(local) != before {
			t.Fatal("root source mutated during input routing")
		}
		exchangeSeconds := time.Since(exchangeBegan).Seconds()
		for j := range result.Expected {
			if k := first + uint64(j); k < uint64(len(reference)) {
				result.Expected[j] = reference[k]
			}
		}
		label := fmt.Sprintf("rank-%d-s%d-d%d-shape-%d", w.Rank, s, dim, shape)
		support := make([]uint64, len(parts[w.Rank]))
		for j, v := range parts[w.Rank] {
			support[j] = uint64(v)
		}
		saveMatrixStageWords(t, filepath.Join(directory, label+"-support.bin"), support)
		saveMatrixStageWords(t, filepath.Join(directory, label+"-local-root.bin"), local)
		if layout == "peer" {
			saveMatrixStageWords(t, filepath.Join(directory, label+"-peer-root.bin"), peer)
		}
		saveMatrixStageWords(t, filepath.Join(directory, label+"-left-input.bin"), result.Left)
		saveMatrixStageWords(t, filepath.Join(directory, label+"-right-input.bin"), result.Right)
		saveMatrixStageWords(t, filepath.Join(directory, label+"-reference.bin"), result.Expected)
		peerDigest := ""
		if layout == "peer" {
			peerDigest = fmt.Sprintf("%x", mpi.PayloadFingerprint(peer))
		}
		// Equal scalar references on both ranks are diagnostic scaffolding, never
		// retained as a distributed production memory model.
		emit(map[string]any{"kind": "root-input", "case": fmt.Sprintf("s%d-d%d-shape-%d", s, dim, shape), "s": s, "dimension": dim, "support_count": len(fixture.support), "local_count": len(parts[w.Rank]), "a_count": result.ACount, "b_count": result.BCount, "leading_coefficient": result.Leading, "reference_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(reference)), "local_sha256": fmt.Sprintf("%x", before), "peer_sha256": peerDigest, "peer_materialized": layout == "peer", "support_sha256": fmt.Sprintf("%x", mpi.PayloadFingerprint(support)), "reference_seconds": referenceSeconds, "local_root_seconds": rootSeconds, "exchange_seconds": exchangeSeconds, "input_seconds": time.Since(started).Seconds(), "root_owned_device_bytes": owned, "scalar_witnesses": 4, "local_preserved": true, "input_layout": layout, "sent_words": sentWords, "received_words": receivedWords, "route_tile_words": 4096})
		return result
	}
}

// TestDistributedRootMerge exercises real root operands through the existing
// product driver; no second FFT, Taylor schedule, transport or oracle is added.
func TestDistributedRootMerge(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit bounded distributed root merge")
	}
	dim, err := strconv.Atoi(os.Getenv("PIR_ROOT_PRODUCT_DIM"))
	if err != nil || (dim != 12 && dim != 14 && dim != 16) {
		t.Fatal("root merge dimension")
	}
	layout := os.Getenv("PIR_ROOT_INPUT_LAYOUT")
	if layout == "" {
		layout = "peer"
	} // Preserve the previously validated explicit test entry.
	if layout != "peer" && layout != "routed" {
		t.Fatal("root input layout")
	}
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		op := uint64(1)
		distributedProductGPUCases[uint64](t, w, 17, []int{dim}, output, &op, emit, productTestOptions{Resident: true, LocalTaylor: true, TileWords: 4096, Pattern: "roots", Shapes: 1, Input: rootProductInput(t, w, output, emit, layout)})
		emit(map[string]any{"kind": "completed", "cases": 1})
	})
}

// Validate the admitted shape before allocating GPUs: balanced two-element
// orbits must yield the exact operand degrees expected by the shared reviewer.
func TestRootMergeFixturePartition(t *testing.T) {
	for _, dim := range []int{12, 14, 16} {
		n := 1 << uint(dim)
		fixture, err := makeRootFixture(17, n-4, false)
		if err != nil {
			t.Fatal(err)
		}
		parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<17, 2)
		if err != nil || len(parts) != 2 {
			t.Fatal(err)
		}
		for _, part := range parts {
			if len(part) != n/2-2 || part[len(part)-1] <= 1<<32 {
				t.Fatal("root fixture geometry/wide indices", dim)
			}
		}
	}
}
