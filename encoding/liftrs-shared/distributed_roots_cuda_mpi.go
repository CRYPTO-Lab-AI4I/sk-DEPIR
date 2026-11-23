//go:build linux && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"math/bits"
)

// Diagnostic saves are optional observations, not checkpoint commits. Callers
// own filesystem policy and must abort peers on any returned error. Once a
// subgroup exists, failures abort the world to avoid blocked collective cleanup.
// Only local roots/coefficients and small integer counts enter this executor.
// Independent retained native files are checked by the controller after the arm;
// they are neither loaded here nor used as arithmetic inputs.
func mergeRankRootTree[Word transposeWord](w *mpi.World, label string, maxDomain, maxLocalWords uint64, options cuda.PreparationOptions, counts []uint64, local []uint32, emit func(map[string]any), save func(string, rankRootMergeStage, []uint32) error) ([]uint32, rootShardLayout, error) {
	if w == nil || len(counts) != w.Size || options.S < 2 || options.S > 18 || label == "" || emit == nil || maxDomain > uint64(1)<<uint(2*options.S) {
		return nil, rootShardLayout{}, fmt.Errorf("owned root runtime admission")
	}
	var zero Word
	_, wide := any(zero).(uint64)
	if wide != (options.S > 16) {
		return nil, rootShardLayout{}, fmt.Errorf("root runtime slot width")
	}
	stages, err := admitRankRootStages(w.Rank, counts, maxDomain, maxLocalWords)
	if err != nil {
		return nil, rootShardLayout{}, err
	}
	if uint64(len(local)) != counts[w.Rank] {
		return nil, rootShardLayout{}, fmt.Errorf("root runtime input shard extent")
	}
	largest := counts[w.Rank]
	for _, stage := range stages {
		largest = max(largest, stage.Domain/uint64(stage.Ranks))
	}
	emit(map[string]any{"kind": "root-admission", "case": label, "leaf_words": counts[w.Rank], "largest_shard_words": largest, "max_local_words": maxLocalWords, "max_domain": maxDomain, "base_word_bytes": 4, "stages": len(stages)})
	if save != nil {
		if err := save("leaf", rankRootMergeStage{}, local); err != nil {
			return nil, rootShardLayout{}, err
		}
	}
	emit(map[string]any{"kind": "root-tree", "case": label, "leaf_counts": counts, "levels": len(stages), "support_count": stages[len(stages)-1].Count - 1, "matches": true})
	result, err := reduceShardedRoots(w.Rank, counts, maxDomain, uint64(1)<<uint(options.S), local, func(stage rankRootMergeStage, input []uint32) (merged []uint32, err error) {
		group, err := w.Split(w.Rank/stage.Ranks, w.Rank)
		if err != nil {
			return nil, err
		}
		closed := false
		defer func() {
			if !closed {
				// Even a broken observer must not prevent collective abort.
				defer w.Abort()
				reason := "root subgroup did not complete"
				if err != nil {
					reason = err.Error()
				}
				emit(map[string]any{"kind": "root-stage-error", "case": label, "level": stage.Level, "first": stage.First, "error": reason})
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
		driver, err := newRankGPUSession[Word](group, options, dim, fmt.Sprintf("%s/level-%d/group-%d", label, stage.Level, stage.First), record)
		if err != nil {
			return nil, err
		}
		link := operandMPI{world: group, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
			record(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
		}}
		record(map[string]any{"kind": "root-stage", "stage": stage, "dimension": dim, "event": "started"})
		driver.begin("root")
		merged, err = driver.Executor.MergeRoots(stage, input, link)
		owned := driver.Executor.GPU.OwnedBytes()
		closeErr := driver.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		driver.productDone(stage.Sources[0].Count, stage.Sources[1].Count, stage.Count)
		record(map[string]any{"kind": "root-stage", "stage": stage, "dimension": dim, "event": "completed", "owned_device_bytes": owned, "matches": true})
		if save != nil {
			if err = save("stage", stage, merged); err != nil {
				return nil, err
			}
		}
		if err = group.Close(); err != nil {
			return nil, err
		}
		closed = true
		emit(map[string]any{"kind": "root-group-closed", "case": label, "level": stage.Level, "first": stage.First, "matches": true})
		return merged, nil
	})
	if err != nil {
		return nil, rootShardLayout{}, err
	}
	final := stages[len(stages)-1]
	if save != nil {
		if err = save("final", final, result); err != nil {
			return nil, rootShardLayout{}, err
		}
	}
	emit(map[string]any{"kind": "root-final", "case": label, "domain": final.Domain, "count": final.Count, "matches": true})
	return result, rootShardLayout{0, w.Size, final.Domain, final.Count}, nil
}
