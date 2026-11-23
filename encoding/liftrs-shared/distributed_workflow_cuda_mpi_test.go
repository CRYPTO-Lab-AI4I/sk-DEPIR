//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRankWorkflowAdmission(t *testing.T) {
	w := &mpi.World{Rank: 0, Size: 4}
	emit := func(map[string]any) {}
	o := rankWorkflowOptions{Spec: Spec{8, 2, 5}, Discovery: rankDiscoveryOptions{DeviceBytes: 64 << 20, MaxLocalWords: 16384, MaxGoodWords: 65536, MaxRootWords: 65536, Random: rand.New(rand.NewSource(109))}, MaxLocalWords: 16384}
	if _, err := prepareRankWorkflow[uint32](w, o, rankWorkflowSinks{}, emit); err == nil {
		t.Fatal("missing output sinks admitted")
	}
	o.MaxLocalWords = 16383
	sinks := rankWorkflowSinks{Code: io.Discard, Generator: io.Discard, Archive: func([]uint64) error { return nil }}
	if _, err := prepareRankWorkflow[uint32](w, o, sinks, emit); err == nil {
		t.Fatal("host shard cap")
	}
}

// The runtime never receives a reference path. The controller independently
// compares these saved bytes/proofs with the fully reviewed pilot and native data.
func TestAssembledRankWorkflow(t *testing.T) {
	directory := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if directory == "" {
		t.Skip("explicit bounded assembled-workflow gate")
	}
	withRankGPUStages(t, directory, 4, func(w *mpi.World, emit func(map[string]any)) {
		began := time.Now()
		var code, generator, archive *os.File
		mode := os.Getenv("PIR_WORKFLOW_PUBLICATION")
		if mode == "" {
			mode = "direct"
		}
		if mode != "direct" && mode != "commit" {
			t.Fatal("workflow publication mode")
		}
		var publication *workflowPublication
		var codeWriter, generatorWriter, archiveWriter io.Writer
		open := func(name string) *os.File {
			f, err := os.OpenFile(filepath.Join(directory, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			return f
		}
		if w.Rank == 0 {
			if mode == "commit" {
				var err error
				publication, err = newWorkflowPublication(directory, "s8")
				if err != nil {
					t.Fatal(err)
				}
				defer publication.Close()
				codeWriter, generatorWriter, archiveWriter = publication.Code, publication.Generator, publication.Archive
			} else {
				code = open("s8-code.bin")
				defer code.Close()
				generator = open("s8-generator.bin")
				defer generator.Close()
				archive = open("s8-archive.raw")
				defer archive.Close()
				codeWriter, generatorWriter, archiveWriter = code, generator, archive
			}
		}
		sinks := rankWorkflowSinks{Code: codeWriter, Generator: generatorWriter, Archive: func(values []uint64) error { return writeDiagnosticWords(archiveWriter, values, nil) }}
		sinks.SupportObserver = func(values []int) error {
			indices := make([]uint64, len(values))
			for i, v := range values {
				indices[i] = uint64(v)
			}
			saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("s8-rank-%d-support.bin", w.Rank)), indices)
			return nil
		}
		sinks.RootObserver = func(kind string, stage rankRootMergeStage, values []uint32) error {
			name := fmt.Sprintf("s8-rank-%d-%s.bin", w.Rank, kind)
			if kind == "stage" {
				name = fmt.Sprintf("s8-level-%d-group-%d-rank-%d-output.bin", stage.Level, stage.First, w.Rank-stage.First)
			}
			saveMatrixStageWords(t, filepath.Join(directory, name), values)
			return nil
		}
		sinks.Checkpoint = func(stage string, keep uint64, values []uint32) ([]uint32, error) {
			_, err := writeCompactRankPrefix(filepath.Join(directory, fmt.Sprintf("s8-rank-%d-%s.bin", w.Rank, stage)), values, 8, 65536, keep, w.Rank, w.Size)
			return values, err
		}
		sinks.IdentityObserver = func(values []uint32) error {
			saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("s8-rank-%d-identity.bin", w.Rank)), values)
			return nil
		}
		options := rankWorkflowOptions{Spec: Spec{8, 2, 5}, Discovery: rankDiscoveryOptions{Library: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), ResultMode: "stream", DeviceBytes: 64 << 20, MaxLocalWords: 16384, MaxGoodWords: 65536, MaxRootWords: 65536, Random: rand.New(rand.NewSource(109))}, MaxLocalWords: 16384, PrefixInputs: true, StreamSpectrum: true}
		result, err := prepareRankWorkflow[uint32](w, options, sinks, emit)
		if err != nil {
			t.Fatal(err)
		}
		if w.Rank == 0 {
			if publication != nil {
				record, err := publication.Commit(result.Header, result.IdentityVerified, result.DiscoverySamples)
				if err != nil {
					t.Fatal(err)
				}
				emit(map[string]any{"kind": "workflow-publication", "case": "s8", "record": record})
			} else {
				for _, f := range []*os.File{code, generator, archive} {
					if err = f.Sync(); err != nil {
						t.Fatal(err)
					}
				}
				info, e := archive.Stat()
				if e != nil || info.Size() != int64(result.Header.K)*8 {
					t.Fatal("archive extent", e)
				}
			}
		}
		w.Barrier()
		emit(map[string]any{"kind": "workflow-result", "case": "s8", "n": result.Header.N, "k": result.Header.K, "samples": result.DiscoverySamples, "root_counts": result.RootCounts, "refinement": result.Refinement, "identity_verified": result.IdentityVerified, "seconds": w.Max(time.Since(began).Seconds()), "reference_input": false, "publication_mode": mode})
	})
}
