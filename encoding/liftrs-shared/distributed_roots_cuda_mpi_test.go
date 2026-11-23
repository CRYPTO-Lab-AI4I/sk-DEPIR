//go:build linux && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"errors"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"testing"
)

// All cases must reject before MPI split or CUDA allocation. The fake world has
// ownership metadata only; no communicator/context is initialized by these tests.
func TestRankRootRuntimeEarlyErrors(t *testing.T) {
	w := &mpi.World{Rank: 0, Size: 4}
	options := cuda.PreparationOptions{S: 2}
	counts := []uint64{1, 1, 1, 1}
	emit := func(map[string]any) {}
	if _, _, err := mergeRankRootTree[uint32](nil, "case", 16, 8, options, counts, []uint32{1}, emit, nil); err == nil {
		t.Fatal("missing world")
	}
	if _, _, err := mergeRankRootTree[uint32](w, "case", 16, 1, options, counts, []uint32{1}, emit, nil); err == nil {
		t.Fatal("host cap")
	}
	if _, _, err := mergeRankRootTree[uint64](w, "case", 16, 8, options, counts, []uint32{1}, emit, nil); err == nil {
		t.Fatal("slot width")
	}
	if _, _, err := mergeRankRootTree[uint32](w, "case", 16, 8, options, counts, nil, emit, nil); err == nil {
		t.Fatal("input extent")
	}
	failure := errors.New("observer save failure")
	save := func(kind string, _ rankRootMergeStage, _ []uint32) error {
		if kind != "leaf" {
			t.Fatal("unexpected stage")
		}
		return failure
	}
	if _, _, err := mergeRankRootTree[uint32](w, "case", 16, 8, options, counts, []uint32{1}, emit, save); !errors.Is(err, failure) {
		t.Fatal("lost save error", err)
	}
}
