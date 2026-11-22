//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// Synthetic closed supports isolate hierarchical root ownership from discovery.
// Small native complete references remain diagnostic only;intermediate GPU
// products flow between levels exclusively as local base-coefficient shards.
func TestHierarchicalRootMerge(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit hierarchical root integration")
	}
	sizes := []int{16, 4092}
	if os.Getenv("PIR_ROOT_HIERARCHY_SMALL") == "1" {
		sizes = sizes[:1]
	}
	withRankGPUStages(t, output, 4, func(world *mpi.World, emit func(map[string]any)) {
		for _, count := range sizes {
			began := time.Now()
			label := fmt.Sprintf("roots-%d", count)
			fixture, err := makeRootFixture(18, count, false)
			if err != nil {
				t.Fatal(err)
			}
			parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<18, world.Size)
			if err != nil {
				t.Fatal(err)
			}
			native, err := cpp.NewPreparation(18, fixture.header.Polynomial, fixture.header.Omega, 1, uint64(fixture.maximum), 32<<20)
			if err != nil {
				t.Fatal(err)
			}
			reference := func(indices []int) []uint32 {
				p, e := fixture.root(supportSorted, native, indices, nil, nil)
				if e != nil {
					t.Fatal(e)
				}
				return p
			}
			full := reference(fixture.support)
			for _, index := range []int{fixture.support[0], fixture.support[len(fixture.support)/2], fixture.support[len(fixture.support)-1]} {
				point := fixture.extension.pow(fixture.header.Omega, uint64(fixture.n-1-index))
				var value uint64
				for j := len(full) - 1; j >= 0; j-- {
					value = fixture.extension.mul(value, point) ^ uint64(full[j])
				}
				if value != 0 {
					t.Fatal("native root scalar witness")
				}
			}
			factor, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: 18, Polynomial: fixture.header.Polynomial, Delta: fixture.header.Delta, Omega: fixture.header.Omega, Logs: fixture.base.lookup.log, Powers: fixture.base.lookup.powers, Devices: []int{0}, MaxElements: uint64(fixture.maximum), DeviceBytes: 64 << 20})
			if err != nil {
				t.Fatal(err)
			}
			local, err := fixture.root(supportSorted, factor, parts[world.Rank], nil, func(phase string, done, total uint64, unit string) {
				if done == 0 || done == total {
					emit(map[string]any{"kind": "progress", "case": label, "phase": "leaf/" + phase, "completed": done, "total": total, "unit": unit})
				}
			})
			closeErr := factor.Close()
			if err != nil || closeErr != nil {
				t.Fatal(err, closeErr)
			}
			leafReference := reference(parts[world.Rank])
			if !reflect.DeepEqual(local, leafReference) {
				t.Fatal("local root native reference")
			}
			saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-leaf.bin", label, world.Rank)), local)
			saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-leaf-reference.bin", label, world.Rank)), leafReference)
			emit(map[string]any{"kind": "root-reference", "case": label, "matches": true, "witnesses": 3, "wide_indices": true})
			counts, err := world.Counts(uint64(len(local)))
			if err != nil {
				t.Fatal(err)
			}
			stages, err := rankRootMergeStages(world.Rank, counts, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			emit(map[string]any{"kind": "root-tree", "case": label, "leaf_counts": counts, "levels": len(stages), "support_count": count, "matches": true})
			result, err := reduceShardedRoots(world.Rank, counts, 1<<20, 1<<18, local, func(stage rankRootMergeStage, input []uint32) ([]uint32, error) {
				group, e := world.Split(world.Rank/stage.Ranks, world.Rank)
				if e != nil {
					return nil, e
				}
				// The child owns its MPI scratch;the parent remains valid between levels.
				defer func() {
					if group != nil {
						world.Abort()
					}
				}()
				if group.Size != stage.Ranks || group.Rank != world.Rank-stage.First {
					return nil, fmt.Errorf("subgroup mapping")
				}
				memberCounts, e := group.Counts(counts[world.Rank])
				if e != nil {
					return nil, e
				}
				if !reflect.DeepEqual(memberCounts, counts[stage.First:stage.First+stage.Ranks]) {
					return nil, fmt.Errorf("subgroup count isolation")
				}
				if e = group.EnableStaging(); e != nil {
					return nil, e
				}
				subdir := filepath.Join(output, fmt.Sprintf("%s-level-%d-group-%d", label, stage.Level, stage.First))
				if e = os.MkdirAll(subdir, 0700); e != nil {
					return nil, e
				}
				record := func(row map[string]any) {
					row["tree_case"] = label
					row["tree_level"] = stage.Level
					row["group_first"] = stage.First
					row["group_size"] = stage.Ranks
					row["group_rank"] = group.Rank
					emit(row)
				}
				// Exact native subtree oracle is confined to these bounded validation cases.
				var support []int
				for rank := stage.First; rank < stage.First+stage.Ranks; rank++ {
					support = append(support, parts[rank]...)
				}
				sort.Ints(support)
				expected := reference(support)
				dimension := bits.Len64(stage.Domain) - 1
				var merged []uint32
				operation := uint64(1)
				link := operandMPI{world: group, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
					record(map[string]any{"kind": "operand-proof", "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
				}}
				callback := func(_, _, _ int, first, words uint64) productCaseInput {
					var own [2][]uint32
					for j, source := range stage.Sources {
						if group.Rank >= source.First && group.Rank < source.First+source.Ranks {
							own[j] = input
						}
					}
					pair, e := routeShardedOperands(rankShardOperandPlan{1, stage.Domain, 127, 1 << 18, words, group.Rank, group.Size, stage.Sources}, own, link, func(operand int, done, total uint64) {
						if done == 0 || done == total {
							record(map[string]any{"kind": "progress", "phase": "root-operand-routing", "operand": operand, "completed": done, "total": total})
						}
					})
					if e != nil {
						t.Fatal(e)
					}
					want := make([]uint32, words)
					for j := range want {
						if index := first + uint64(j); index < uint64(len(expected)) {
							want[j] = expected[index]
						}
					}
					return productCaseInput{pair[0], pair[1], want, stage.Sources[0].Count, stage.Sources[1].Count, expected[len(expected)-1]}
				}
				record(map[string]any{"kind": "root-stage", "stage": stage, "dimension": dimension, "event": "started"})
				distributedProductGPUCases[uint64](t, group, 18, []int{dimension}, subdir, &operation, record, productTestOptions{Resident: true, LocalTaylor: true, TileWords: 127, Pattern: "roots", Shapes: 1, DeviceBytes: 64 << 20, Input: callback, Result: func(values []uint32) { merged = values }})
				record(map[string]any{"kind": "completed", "cases": 1})
				if e = group.Close(); e != nil {
					return nil, e
				}
				group = nil
				emit(map[string]any{"kind": "root-group-closed", "case": label, "level": stage.Level, "first": stage.First, "matches": true})
				return merged, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			final := stages[len(stages)-1]
			first, words := rankBand(final.Domain, world.Rank, world.Size)
			want := make([]uint32, words)
			for j := range want {
				if index := first + uint64(j); index < uint64(len(full)) {
					want[j] = full[index]
				}
			}
			if !reflect.DeepEqual(result, want) {
				t.Fatal("final hierarchical root native reference")
			}
			saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-final.bin", label, world.Rank)), result)
			saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-reference.bin", label, world.Rank)), want)
			if e := native.Close(); e != nil {
				t.Fatal(e)
			}
			world.Barrier()
			if shared := world.ShareCount(uint64(count)); shared != uint64(count) {
				t.Fatal("parent communicator after child close")
			}
			emit(map[string]any{"kind": "root-final", "case": label, "domain": final.Domain, "count": final.Count, "words": words, "matches": true, "levels": len(stages), "seconds": world.Max(time.Since(began).Seconds())})
		}
		emit(map[string]any{"kind": "hierarchy-completed", "cases": len(sizes)})
	})
}

// Bound native setup and every planned shape before admitting the GPU job.
// This uses no MPI/CUDA initialization and never allocates the full s18 grid.
func TestRootHierarchyShapes(t *testing.T) {
	for _, count := range []int{16, 4092} {
		fixture, err := makeRootFixture(18, count, false)
		if err != nil {
			t.Fatal(err)
		}
		parts, err := partitionRootSupportWith(supportSorted, fixture.support, fixture.n, 1<<18, 4)
		if err != nil {
			t.Fatal(err)
		}
		counts := make([]uint64, 4)
		for rank, part := range parts {
			counts[rank] = uint64(len(part) + 1)
		}
		want := []uint64{5, 5, 5, 5}
		if count == 4092 {
			want = []uint64{1025, 1025, 1023, 1023}
		}
		if !reflect.DeepEqual(counts, want) {
			t.Fatal("leaf ownership", counts)
		}
		native, err := cpp.NewPreparation(18, fixture.header.Polynomial, fixture.header.Omega, 1, uint64(fixture.maximum), 32<<20)
		if err != nil {
			t.Fatal(err)
		}
		result, err := fixture.root(supportSorted, native, fixture.support, nil, nil)
		ce := native.Close()
		if err != nil || ce != nil || len(result) != count+1 || result[0] != 1 || result[len(result)-1] == 0 {
			t.Fatal("native root", err, ce)
		}
		for rank := 0; rank < 4; rank++ {
			plan, e := rankRootMergeStages(rank, counts, 1<<20)
			if e != nil || len(plan) != 2 {
				t.Fatal("stage admission", e)
			}
			t.Logf("roots=%d rank=%d leaf=%d child domain=%d final domain=%d", count, rank, counts[rank], plan[0].Domain, plan[1].Domain)
		}
	}
}
