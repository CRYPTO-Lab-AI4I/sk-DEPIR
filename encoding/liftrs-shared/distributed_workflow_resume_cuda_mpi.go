//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	"path/filepath"
	"time"

	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// Recovery is explicit immutable job input. SourceSHA comes from trusted saved
// provenance; Known selects a committed Newton prefix without changing rank count.
type rankWorkflowRecovery struct {
	Directory, Label, SourceSHA string
	Known                       uint64
}

func admitRankWorkflowRecovery[Word transposeWord](w *mpi.World, o rankWorkflowOptions, r rankWorkflowRecovery, sinks rankWorkflowSinks, emit func(map[string]any)) error {
	if w == nil || w.Size < 2 || w.Size > 64 || w.Size&(w.Size-1) != 0 || w.Rank < 0 || w.Rank >= w.Size || o.Spec.FieldDegree < 2 || o.Spec.FieldDegree > 18 || o.Spec.Dimension != 2 || o.Spec.CurveDegree < 1 || o.Discovery.Device < 0 || o.Discovery.DeviceBytes == 0 || emit == nil || r.Directory == "" || r.Label != fmt.Sprintf("s%d", o.Spec.FieldDegree) || r.Known < 1 || r.Known > (uint64(1)<<uint(2*o.Spec.FieldDegree)) {
		return fmt.Errorf("workflow recovery admission")
	}
	var zero Word
	_, wide := any(zero).(uint64)
	n := uint64(1) << uint(2*o.Spec.FieldDegree)
	_, words := rankBand(n, w.Rank, w.Size)
	if wide != (o.Spec.FieldDegree > 16) || n > uint64(^uint(0)>>1) || words > o.MaxLocalWords || words == 0 || (w.Rank == 0 && (sinks.Code == nil || sinks.Generator == nil)) {
		return fmt.Errorf("workflow recovery shard/output admission")
	}
	if _, err := expectedDigest(r.SourceSHA); err != nil {
		return err
	}
	return nil
}

// resumeRankWorkflow verifies all recovery inputs before GPU allocation, skips
// discovery/root reconstruction and uses the cold path's numerical tail. Only the
// coordinator hashes support, once; every rank agrees both full checkpoint records.
// Caller aborts peers on any error and owns trusted immutable source/durability.
func resumeRankWorkflow[Word transposeWord](w *mpi.World, o rankWorkflowOptions, r rankWorkflowRecovery, sinks rankWorkflowSinks, emit func(map[string]any)) (result rankWorkflowResult, err error) {
	if err = admitRankWorkflowRecovery[Word](w, o, r, sinks, emit); err != nil {
		return result, err
	}
	source, err := readWorkflowCheckpointSource(r.Directory, r.Label, r.SourceSHA)
	if err != nil {
		return result, err
	}
	h := source.Header
	if h.Spec != o.Spec || source.Ranks != w.Size || r.Known > h.N-h.K || (r.Known != h.N-h.K && r.Known&(r.Known-1) != 0) {
		return result, fmt.Errorf("workflow recovery source binding")
	}
	if sinks.AdmitHost != nil {
		_, words := rankBand(h.N, w.Rank, w.Size)
		if err = sinks.AdmitHost(hostArrayPhase("checkpoint-reload", o.Spec.FieldDegree, h.N, 2*words, 0, 0, 0)); err != nil {
			return result, err
		}
	}
	binding, err := newRankSupportBinding(r.Directory, r.Label, w.Rank, w.Size)
	if err != nil {
		return result, err
	}
	stage := ""
	last := time.Now()
	report := func(done, total uint64) {
		if done == 0 || done == total || time.Since(last) >= 30*time.Second {
			last = time.Now()
			emit(map[string]any{"kind": "restart-support-progress", "case": r.Label, "mode": "coordinator", "checkpoint": stage, "phase": "support-checksum", "completed": done, "total": total, "unit": "bytes"})
		}
	}
	verify := func(record rankPrefixRecord) error {
		if record.SourceSHA != r.SourceSHA || record.K != int(h.K) {
			return fmt.Errorf("workflow recovery checkpoint/source binding")
		}
		stage = record.Stage
		return binding.verify(record, func(own [64]byte) ([][64]byte, error) {
			frames, e := w.ControlFrames(own)
			if e != nil {
				return nil, e
			}
			encoded := make([]string, len(frames))
			for rank, frame := range frames {
				encoded[rank] = fmt.Sprintf("%x", frame)
			}
			emit(map[string]any{"kind": "restart-support-agreement", "case": r.Label, "checkpoint": stage, "source": r.Directory, "frames": encoded})
			return frames, nil
		}, report)
	}
	root, err := loadRankPrefixWithSupport(r.Directory, r.Label, o.Spec.FieldDegree, w.Rank, w.Size, "root", o.MaxLocalWords, verify)
	if err != nil {
		return result, err
	}
	prefix, err := loadRankPrefixWithSupport(r.Directory, r.Label, o.Spec.FieldDegree, w.Rank, w.Size, fmt.Sprintf("newton-%d", r.Known), o.MaxLocalWords, verify)
	if err != nil {
		return result, err
	}
	emit(map[string]any{"kind": "restart-support", "case": r.Label, "mode": "coordinator", "hash_passes": binding.HashPasses, "hashed_bytes": binding.HashBytes, "agreements": binding.Agreements, "verified": true})
	if sinks.DiscoveryReady != nil {
		if err = sinks.DiscoveryReady(h, source.Samples); err != nil {
			return result, err
		}
	}
	base, err := NewBinaryField(o.Spec.FieldDegree, h.Polynomial)
	if err != nil {
		return result, err
	}
	field := orbitField{q: uint64(1) << uint(o.Spec.FieldDegree), base: base, lookup: newBoundedPreparationProducts(base, 18)}
	options := cuda.PreparationOptions{Path: o.Discovery.Library, S: o.Spec.FieldDegree, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: field.lookup.log, Powers: field.lookup.powers, Devices: []int{o.Discovery.Device}, MaxElements: h.N, DeviceBytes: o.Discovery.DeviceBytes}
	session, err := newRankGPUSession[Word](w, options, 2*o.Spec.FieldDegree, r.Label, emit)
	if err != nil {
		return result, err
	}
	defer func() {
		if e := session.Close(); err == nil && e != nil {
			result = rankWorkflowResult{}
			err = e
		}
	}()
	emit(map[string]any{"kind": "workflow-resume", "case": r.Label, "prefix": r.Known, "source": filepath.Clean(r.Directory), "source_sha256": r.SourceSHA, "discovery_skipped": true, "roots_skipped": true})
	return finishRankWorkflow(w, o, sinks, rankWorkflowResult{Header: h, DiscoverySamples: source.Samples}, field, session, root.Values, prefix.Values, r.Known, emit)
}
