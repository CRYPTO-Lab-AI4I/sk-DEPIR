//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRankWorkflowRecoveryAdmission(t *testing.T) {
	w := &mpi.World{Rank: 0, Size: 4}
	o := rankWorkflowOptions{Spec: Spec{8, 2, 5}, MaxLocalWords: 16384, Discovery: rankDiscoveryOptions{DeviceBytes: 64 << 20}}
	r := rankWorkflowRecovery{Directory: "/trusted/cache", Label: "s8", Known: 8, SourceSHA: fmt.Sprintf("%064x", 1)}
	sinks := rankWorkflowSinks{Code: io.Discard, Generator: io.Discard}
	emit := func(map[string]any) {}
	if err := admitRankWorkflowRecovery[uint32](w, o, r, sinks, emit); err != nil {
		t.Fatal(err)
	}
	bad := r
	bad.SourceSHA = "bad"
	if err := admitRankWorkflowRecovery[uint32](w, o, bad, sinks, emit); err == nil {
		t.Fatal("invalid source digest")
	}
	// Reject an intact cache for a different code before collective/GPU work.
	directory := t.TempDir()
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	source := workflowCheckpointSource{Version: 1, Header: h, Ranks: 4, Samples: 2, SupportFormat: "uint64-exponents-le", SupportSHA: fmt.Sprintf("%064x", 0), SupportBytes: 8 * h.K}
	data, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeWorkflowMetadata(directory, "s8-checkpoint-source.json", data); err != nil {
		t.Fatal(err)
	}
	bad = r
	bad.Directory = directory
	bad.SourceSHA = fmt.Sprintf("%x", sha256.Sum256(data))
	if _, err = resumeRankWorkflow[uint32](w, o, bad, sinks, emit); err == nil || err.Error() != "workflow recovery source binding" {
		t.Fatal("wrong-code cache", err)
	}
	o.MaxLocalWords--
	if _, err := resumeRankWorkflow[uint32](w, o, r, sinks, emit); err == nil {
		t.Fatal("resident admission before reading source")
	}
	if err := admitRankWorkflowRecovery[uint64](w, o, r, sinks, emit); err == nil {
		t.Fatal("wrong slot width")
	}
}

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
		resumeFrom := os.Getenv("PIR_WORKFLOW_RESUME_FROM")
		var recovery rankWorkflowRecovery
		var recoveredSource workflowCheckpointSource
		if resumeFrom != "" {
			data, e := readBoundedPilotFile(filepath.Join(resumeFrom, "s8-checkpoint-source.json"), 65536)
			if e != nil {
				t.Fatal(e)
			}
			recovery = rankWorkflowRecovery{Directory: resumeFrom, Label: "s8", SourceSHA: fmt.Sprintf("%x", sha256.Sum256(data)), Known: 8}
			recoveredSource, e = readWorkflowCheckpointSource(resumeFrom, "s8", recovery.SourceSHA)
			if e != nil {
				t.Fatal(e)
			}
		}
		var code, generator, archive *os.File
		mode := os.Getenv("PIR_WORKFLOW_PUBLICATION")
		if mode == "" {
			mode = "direct"
		}
		if mode != "direct" && mode != "commit" {
			t.Fatal("workflow publication mode")
		}
		checkpoints := os.Getenv("PIR_WORKFLOW_CHECKPOINT") == "commit"
		if v := os.Getenv("PIR_WORKFLOW_CHECKPOINT"); v != "" && v != "commit" {
			t.Fatal("workflow checkpoint mode")
		}
		if checkpoints && mode != "commit" {
			t.Fatal("checkpoint mode requires sealed publication archive")
		}
		if resumeFrom != "" && !checkpoints {
			t.Fatal("assembled restart requires source-bound checkpoint policy")
		}
		var checkpointWriter *rankCheckpointWriter
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
		if checkpoints {
			sinks.DiscoveryReady = func(h ScalingCodeHeader, samples int) error {
				supportSHA := ""
				if w.Rank == 0 {
					var err error
					if resumeFrom == "" {
						supportSHA, err = publication.SealArchive(8 * h.K)
					} else {
						supportSHA = recoveredSource.SupportSHA
						err = publication.ReuseVerifiedArchive(filepath.Join(resumeFrom, "s8-support.bin"), 8*h.K, supportSHA)
					}
					if err != nil {
						return err
					}
				}
				var err error
				checkpointWriter, err = newRankCheckpointWriter(w, directory, "s8", h, samples, supportSHA, 16384, emit)
				return err
			}
		}
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
			if checkpoints {
				return values, checkpointWriter.Commit(stage, keep, values)
			}
			_, err := writeCompactRankPrefix(filepath.Join(directory, fmt.Sprintf("s8-rank-%d-%s.bin", w.Rank, stage)), values, 8, 65536, keep, w.Rank, w.Size)
			return values, err
		}
		sinks.IdentityObserver = func(values []uint32) error {
			saveMatrixStageWords(t, filepath.Join(directory, fmt.Sprintf("s8-rank-%d-identity.bin", w.Rank)), values)
			return nil
		}
		options := rankWorkflowOptions{Spec: Spec{8, 2, 5}, Discovery: rankDiscoveryOptions{Library: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), ResultMode: "stream", DeviceBytes: 64 << 20, MaxLocalWords: 16384, MaxGoodWords: 65536, MaxRootWords: 65536, Random: rand.New(rand.NewSource(109))}, MaxLocalWords: 16384, PrefixInputs: true, StreamSpectrum: true}
		var result rankWorkflowResult
		var err error
		if resumeFrom == "" {
			result, err = prepareRankWorkflow[uint32](w, options, sinks, emit)
		} else {
			result, err = resumeRankWorkflow[uint32](w, options, recovery, sinks, emit)
		}
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
		emit(map[string]any{"kind": "workflow-result", "case": "s8", "n": result.Header.N, "k": result.Header.K, "samples": result.DiscoverySamples, "root_counts": result.RootCounts, "refinement": result.Refinement, "identity_verified": result.IdentityVerified, "seconds": w.Max(time.Since(began).Seconds()), "reference_input": false, "publication_mode": mode, "checkpoint_mode": map[bool]string{true: "commit", false: "observer"}[checkpoints], "resumed": resumeFrom != "", "initial_prefix": map[bool]int{true: 8, false: 1}[resumeFrom != ""]})
	})
}

// Fresh-process validation of persisted files only. It does not resume arithmetic
// or open MPI/CUDA;the controller separately retains all native comparisons.
func TestReloadWorkflowCheckpoints(t *testing.T) {
	directory := os.Getenv("PIR_WORKFLOW_RELOAD_SOURCE")
	if directory == "" {
		t.Skip("explicit saved checkpoint source")
	}
	sourceBytes, err := readBoundedPilotFile(filepath.Join(directory, "s8-checkpoint-source.json"), 65536)
	if err != nil {
		t.Fatal(err)
	}
	sourceSHA := fmt.Sprintf("%x", sha256.Sum256(sourceBytes))
	source, err := readWorkflowCheckpointSource(directory, "s8", sourceSHA)
	if err != nil {
		t.Fatal(err)
	}
	for rank := 0; rank < source.Ranks; rank++ {
		for _, stage := range []string{"root", "newton-8"} {
			state, err := loadRankPrefix(directory, "s8", 8, rank, source.Ranks, stage, 16384)
			if err != nil || state.Record.SourceSHA != sourceSHA || state.Record.SupportSHA != source.SupportSHA {
				t.Fatal("fresh checkpoint reload", rank, stage, err)
			}
			t.Logf("verified rank=%d stage=%s words=%d source=%s", rank, stage, len(state.Values), sourceSHA)
		}
	}
}
