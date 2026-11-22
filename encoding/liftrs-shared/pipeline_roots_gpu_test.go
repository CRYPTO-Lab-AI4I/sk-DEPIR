//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"math/bits"
	"os"
	"path/filepath"
	"testing"
)

// Only local roots/coefficients and small integer counts enter this executor.
// Independent retained native files are checked by the controller after the arm;
// they are neither loaded here nor used as arithmetic inputs.
func ownedPipelineRoots[Word transposeWord](t *testing.T, w *mpi.World, directory, label string, fixture *rootFixture, options cuda.PreparationOptions, counts []uint64, local []uint32, emit func(map[string]any)) ([]uint32, rootShardLayout) {
	t.Helper()
	saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("%s-rank-%d-leaf.bin", label, w.Rank)), local)
	stages, err := rankRootMergeStages(w.Rank, counts, uint64(fixture.n))
	if err != nil {
		t.Fatal(err)
	}
	emit(map[string]any{"kind": "root-tree", "case": label, "leaf_counts": counts, "levels": len(stages), "support_count": discoveryTotal(counts) - uint64(w.Size), "matches": true})
	result, err := reduceShardedRoots(w.Rank, counts, uint64(fixture.n), uint64(1)<<uint(options.S), local, func(stage rankRootMergeStage, input []uint32) ([]uint32, error) {
		group, err := w.Split(w.Rank/stage.Ranks, w.Rank)
		if err != nil {
			return nil, err
		}
		closed := false
		defer func() {
			if !closed {
				w.Abort()
			}
		}()
		if group.Size != stage.Ranks || group.Rank != w.Rank-stage.First {
			return nil, fmt.Errorf("owned root subgroup mapping")
		}
		if err = group.EnableStaging(); err != nil {
			return nil, err
		}
		record := func(row map[string]any) {
			row["tree_case"], row["tree_level"], row["group_first"], row["group_size"], row["group_rank"] = label, stage.Level, stage.First, stage.Ranks, group.Rank
			emit(row)
		}
		dim := bits.Len64(stage.Domain) - 1
		driver := openPipelineGPU[Word](t, group, options, dim, fmt.Sprintf("%s/level-%d/group-%d", label, stage.Level, stage.First), record)
		link := operandMPI{world: group, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
			record(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
		}}
		record(map[string]any{"kind": "root-stage", "stage": stage, "dimension": dim, "event": "started"})
		driver.begin("root")
		merged, err := driver.Executor.MergeRoots(stage, input, link)
		owned := driver.Executor.GPU.OwnedBytes()
		closeErr := driver.Executor.GPU.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		driver.productDone(stage.Sources[0].Count, stage.Sources[1].Count, stage.Count)
		record(map[string]any{"kind": "root-stage", "stage": stage, "dimension": dim, "event": "completed", "owned_device_bytes": owned, "matches": true})
		saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("%s-level-%d-group-%d-rank-%d-output.bin", label, stage.Level, stage.First, group.Rank)), merged)
		if err = group.Close(); err != nil {
			return nil, err
		}
		closed = true
		emit(map[string]any{"kind": "root-group-closed", "case": label, "level": stage.Level, "first": stage.First, "matches": true})
		return merged, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	final := stages[len(stages)-1]
	saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("%s-rank-%d-final.bin", label, w.Rank)), result)
	emit(map[string]any{"kind": "root-final", "case": label, "domain": final.Domain, "count": final.Count, "matches": true})
	return result, rootShardLayout{0, w.Size, final.Domain, final.Count}
}

// Bounded synthetic field18 inputs also exercise the owned executor's64-bit
// instantiation. Fixture construction is diagnostic;the executor sees only local
// roots/shards. Retained native leaves/stages/final outputs are checked afterward.
func TestOwnedWideRoots(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit owned field18 root validation")
	}
	withRankGPUStages(t, output, 4, func(w *mpi.World, emit func(map[string]any)) {
		for _, count := range []int{16, 4092} {
			fixture, err := makeRootFixture(18, count, false)
			if err != nil {
				t.Fatal(err)
			}
			parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<18, w.Size)
			if err != nil {
				t.Fatal(err)
			}
			options := cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: 18, Polynomial: fixture.header.Polynomial, Delta: fixture.header.Delta, Omega: fixture.header.Omega, Logs: fixture.base.lookup.log, Powers: fixture.base.lookup.powers, Devices: []int{0}, MaxElements: uint64(fixture.maximum), DeviceBytes: 64 << 20}
			factor, err := cuda.NewPreparation(options)
			if err != nil {
				t.Fatal(err)
			}
			local, err := fixture.root(supportSorted, factor, parts[w.Rank], nil, nil)
			closeErr := factor.Close()
			if err != nil || closeErr != nil {
				t.Fatal(err, closeErr)
			}
			counts, err := w.Counts(uint64(len(local)))
			if err != nil {
				t.Fatal(err)
			}
			ownedPipelineRoots[uint64](t, w, output, fmt.Sprintf("roots-%d", count), fixture, options, counts, local, emit)
		}
		emit(map[string]any{"kind": "hierarchy-completed", "cases": 2})
	})
}
