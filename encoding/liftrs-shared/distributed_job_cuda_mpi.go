//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// RunDistributedPreprocessing owns one MPI world on its calling OS thread. Every
// rank must call with identical config and abort together on error. Random is used
// only by rank0; production must provide a suitable secure reader. Output receives
// flushed JSON lines. Caller owns CPU profiling, deadlines and periodic guards.
// This experimental entry does not admit fullsize resources or authenticate caches.
func RunDistributedPreprocessing(c DistributedPreprocessingConfig, output io.Writer, random io.Reader) (err error) {
	if err = c.Validate(); err != nil {
		return err
	}
	if output == nil || random == nil {
		return fmt.Errorf("job log/random source required")
	}
	w, err := mpi.Open()
	if err != nil {
		return err
	}
	defer func() {
		if problem := recover(); problem != nil {
			err = fmt.Errorf("distributed job: %v", problem)
		}
		if err != nil {
			w.Abort()
		} else {
			err = w.Close()
		}
	}()
	if w.Size != c.Ranks {
		return fmt.Errorf("MPI size/config disagreement")
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(data)
	var frame [64]byte
	copy(frame[:32], digest[:])
	copy(frame[32:], digest[:])
	frames, err := w.ControlFrames(frame)
	if err != nil {
		return err
	}
	for _, v := range frames {
		if v != frame {
			return fmt.Errorf("distributed job configuration disagreement")
		}
	}
	w.ChunkWords = 1024
	if err = w.EnableStaging(); err != nil {
		return err
	}
	began := time.Now()
	recorder, err := newProofLog(c.ProofLog)
	if err != nil {
		return err
	}
	lastProof := time.Now()
	var rss uint64
	last := time.Time{}
	emit := func(row map[string]any) {
		if time.Since(last) >= time.Second {
			var usage syscall.Rusage
			if e := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); e != nil {
				panic(e)
			}
			rss = uint64(usage.Maxrss) * 1024
			last = time.Now()
			if rss > c.RSSBytes {
				panic("rank RSS budget")
			}
		}
		hide, e := recorder.Record(row)
		if e != nil {
			panic(e)
		}
		if hide {
			if time.Since(lastProof) < 30*time.Second {
				return
			}
			lastProof = time.Now()
			row = map[string]any{"kind": "proof-progress", "phase": "transport-proof-recording", "completed": recorder.total, "unit": "proof records", "total_known": false}
		}
		row["rank"], row["utc"], row["elapsed_seconds"], row["peak_rss_bytes"] = w.Rank, time.Now().UTC().Format(time.RFC3339Nano), time.Since(began).Seconds(), rss
		if e = json.NewEncoder(output).Encode(row); e != nil {
			panic(e)
		}
	}
	complete := false
	defer func() {
		for _, row := range recorder.Summaries(complete) {
			emit(row)
		}
	}()
	host, err := os.Hostname()
	if err != nil {
		return err
	}
	emit(map[string]any{"kind": "job-start", "host": host, "config": c})
	if c.Spec.FieldDegree <= 16 {
		err = runRankPreprocessingJob[uint32](w, c, random, emit)
	} else {
		err = runRankPreprocessingJob[uint64](w, c, random, emit)
	}
	if err != nil {
		return err
	}
	w.Barrier()
	emit(map[string]any{"kind": "job-complete", "seconds": time.Since(began).Seconds()})
	complete = true
	return nil
}

func runRankPreprocessingJob[Word transposeWord](w *mpi.World, c DistributedPreprocessingConfig, random io.Reader, emit func(map[string]any)) error {
	began := time.Now()
	label := fmt.Sprintf("s%d", c.Spec.FieldDegree)
	var publication *workflowPublication
	var err error
	if w.Rank == 0 {
		publication, err = newWorkflowPublication(c.Directory, label)
		if err != nil {
			return err
		}
		defer publication.Close()
	}
	var checkpoint *rankCheckpointWriter
	sinks := rankWorkflowSinks{}
	if w.Rank == 0 {
		sinks.Code, sinks.Generator = publication.Code, publication.Generator
		sinks.Archive = func(values []uint64) error { return writeDiagnosticWords(publication.Archive, values, nil) }
	}
	sinks.DiscoveryReady = func(h ScalingCodeHeader, samples int) error {
		if e := c.admitFiles(h); e != nil {
			return e
		}
		emit(map[string]any{"kind": "job-admission", "phase": "root-and-reciprocal", "header": h, "ranks": c.Ranks})
		supportSHA := ""
		if w.Rank == 0 {
			if c.ResumeDirectory == "" {
				supportSHA, err = publication.SealArchive(8 * h.K)
			} else {
				source, e := readWorkflowCheckpointSource(c.ResumeDirectory, label, c.SourceSHA)
				if e != nil {
					return e
				}
				supportSHA = source.SupportSHA
				err = publication.ReuseVerifiedArchive(filepath.Join(c.ResumeDirectory, label+"-support.bin"), 8*h.K, supportSHA)
			}
			if err != nil {
				return err
			}
		}
		checkpoint, err = newRankCheckpointWriter(w, c.Directory, label, h, samples, supportSHA, c.MaxLocalWords, emit)
		return err
	}
	sinks.Checkpoint = func(stage string, keep uint64, values []uint32) ([]uint32, error) {
		return values, checkpoint.Commit(stage, keep, values)
	}
	options := rankWorkflowOptions{Spec: c.Spec, MaxLocalWords: c.MaxLocalWords, PrefixInputs: c.NewtonInput == "prefix", StreamSpectrum: c.Spectrum == "stream", Discovery: rankDiscoveryOptions{Library: c.Library, Device: c.Device, DeviceBytes: c.DeviceBytes, MaxLocalWords: c.MaxLocalWords, MaxGoodWords: c.MaxGoodWords, MaxRootWords: c.MaxRootWords, MaxArchiveWords: c.MaxArchiveWords, ResultMode: c.DiscoveryResult, Random: random}}
	options.Discovery.AdmitCode = c.admitFiles
	var result rankWorkflowResult
	if c.ResumeDirectory == "" {
		result, err = prepareRankWorkflow[Word](w, options, sinks, emit)
	} else {
		result, err = resumeRankWorkflow[Word](w, options, rankWorkflowRecovery{c.ResumeDirectory, label, c.SourceSHA, c.Known}, sinks, emit)
	}
	if err != nil {
		return err
	}
	if w.Rank == 0 {
		record, e := publication.Commit(result.Header, result.IdentityVerified, result.DiscoverySamples)
		if e != nil {
			return e
		}
		emit(map[string]any{"kind": "workflow-publication", "case": label, "record": record})
	}
	emit(map[string]any{"kind": "workflow-result", "entry": "command", "case": label, "n": result.Header.N, "k": result.Header.K, "root_counts": result.RootCounts, "samples": result.DiscoverySamples, "refinement": result.Refinement, "identity_verified": result.IdentityVerified, "resumed": c.ResumeDirectory != "", "initial_prefix": max(uint64(1), c.Known), "reference_input": false, "publication_mode": "commit", "checkpoint_mode": "commit", "seconds": time.Since(began).Seconds()})
	return nil
}
