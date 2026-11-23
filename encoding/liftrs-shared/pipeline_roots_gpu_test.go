//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"os"
	"path/filepath"
	"testing"
)

// Native references and diagnostic file policy stay in this bounded test adapter.
func ownedPipelineRoots[Word transposeWord](t *testing.T, w *mpi.World, directory, label string, maxDomain uint64, options cuda.PreparationOptions, counts []uint64, local []uint32, emit func(map[string]any)) ([]uint32, rootShardLayout) {
	t.Helper()
	save := func(kind string, stage rankRootMergeStage, values []uint32) error {
		name := fmt.Sprintf("%s-rank-%d-%s.bin", label, w.Rank, kind)
		if kind == "stage" {
			name = fmt.Sprintf("%s-level-%d-group-%d-rank-%d-output.bin", label, stage.Level, stage.First, w.Rank-stage.First)
		}
		saveMatrixStageWords(t, filepath.Join(directory, name), values)
		return nil
	}
	result, layout, err := mergeRankRootTree[Word](w, label, maxDomain, 8<<20, options, counts, local, emit, save)
	if err != nil {
		t.Fatal(err)
	}
	return result, layout
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
			ownedPipelineRoots[uint64](t, w, output, fmt.Sprintf("roots-%d", count), uint64(fixture.n), options, counts, local, emit)
		}
		emit(map[string]any{"kind": "hierarchy-completed", "cases": 2})
	})
}
