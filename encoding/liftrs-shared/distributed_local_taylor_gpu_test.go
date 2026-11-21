//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"reflect"
	"testing"
)

// Prove the private canonical-value proof is invalidated after raw upload, even
// when local Taylor has already warmed it. Full products separately check all
// nonzero arithmetic and both physical layouts against schoolbook coefficients.
func localTaylorGuard[Word transposeWord](t *testing.T, gpu *cuda.RankStorage, s, dim, rank, ranks int, rows, cols uint64, emit func(map[string]any)) {
	t.Helper()
	first, count := rankBand(rows, rank, ranks)
	words := count * cols
	var plan rankTaylorPlan
	var ticket taylorTicket
	found := false
	for level := 0; level < dim-dim/2 && !found; level++ {
		terms, err := gpu.TaylorDegrees(level)
		if err != nil {
			t.Fatal(err)
		}
		if len(terms) == 0 {
			continue
		}
		for _, quotient := range []bool{true, false} {
			p := rankTaylorPlan{Operation: 1, Domain: rows * cols, Columns: cols, Capacity: uint64(len(terms) + 1), Rank: rank, Ranks: ranks, Level: level, Quotient: quotient, Degrees: terms}
			half, minimum, maximum, span, err := p.validate()
			if err != nil {
				t.Fatal(err)
			}
			cursor := newTaylorCursor(p, rank, half, minimum, maximum, span)
			if q, ok := cursor.take(); ok {
				plan, ticket, found = p, q, true
				break
			}
		}
	}
	if !found {
		t.Fatal("local Taylor guard has no owned target", first, words)
	}
	for _, interleaved := range []bool{false, true} {
		for _, mutation := range []string{"upload", "scatter", "localcopy"} {
			if err := gpu.Reset(0, rows, cols, interleaved); err != nil {
				t.Fatal(err)
			}
			snapshot := make([]Word, plan.Capacity)
			storage := residentTaylorStorage[Word]{gpu: gpu, slot: 0, tile: 127, local: true}
			if err := storage.Apply(plan, ticket, 0, snapshot); err != nil {
				t.Fatal("warm local Taylor", err)
			}
			expected := make([]Word, words)
			bad := []Word{Word(uint64(1) << uint(2*s))}
			switch mutation {
			case "upload":
				expected[0] = bad[0]
				if err := storageTransfer(gpu, 0, 0, bad, true); err != nil {
					t.Fatal(err)
				}
			case "scatter":
				peer := 1 - rank
				column, _ := rankBand(cols, peer, ranks)
				expected[column] = bad[0]
				var err error
				switch values := any(bad).(type) {
				case []uint32:
					err = gpu.Tile32(0, peer, 0, values, false)
				case []uint64:
					err = gpu.Tile64(0, peer, 0, values, false)
				}
				if err != nil {
					t.Fatal(err)
				}
			case "localcopy":
				if err := gpu.Reset(1, cols, rows, false); err != nil {
					t.Fatal(err)
				}
				column, _ := rankBand(cols, rank, ranks)
				expected[column] = bad[0]
				if err := storageTransfer(gpu, 1, first, bad, true); err != nil {
					t.Fatal(err)
				}
				if err := gpu.Local(1, 0); err != nil {
					t.Fatal(err)
				}
			}
			if err := storage.Apply(plan, ticket, 0, snapshot); err == nil {
				t.Fatal("raw write failed to invalidate canonical proof", mutation)
			}
			actual := make([]Word, words)
			storageIO(t, gpu, 0, actual, false)
			if !reflect.DeepEqual(actual, expected) {
				t.Fatal("rejected local Taylor changed owned data", mutation)
			}
			for _, v := range snapshot {
				if v != 0 {
					t.Fatal("local Taylor changed host snapshot")
				}
			}
		}
	}

	if err := gpu.Reset(0, rows, cols, false); err != nil {
		t.Fatal(err)
	}
	if err := gpu.Reset(1, rows, cols, false); err != nil {
		t.Fatal(err)
	}
	emit(map[string]any{"kind": "local-window-check", "mutation_kinds": []string{"upload", "scatter", "localcopy"}, "s": s, "dimension": dim, "canonical_proof_invalidated": true, "rejection_preserved": true, "both_layouts": true, "host_snapshot_preserved": true})
}

func TestLocalTaylorGPUProducts(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit local Taylor product gate")
	}
	t.Setenv("PIR_TAYLOR_LOCAL_GUARD", "1")
	withDistributedGPUStages(t, output, func(w *mpi.World, emit func(map[string]any)) {
		op := uint64(1)
		distributedProductGPUCases[uint32](t, w, 4, []int{4, 5, 8}, output, &op, emit, true, 4096, true)
		distributedProductGPUCases[uint64](t, w, 17, []int{5, 9, 12}, output, &op, emit, true, 4096, true)
		emit(map[string]any{"kind": "completed", "cases": 12})
	})
}
