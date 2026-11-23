//go:build linux && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

type rankCheckpointWriter struct {
	world                 *mpi.World
	directory, label      string
	header                ScalingCodeHeader
	supportSHA, sourceSHA string
	maxWords              uint64
	emit                  func(map[string]any)
}

// The coordinator supplies a sealed archive digest;all workers agree the entire
// code/source metadata before publication. Callers abort peers on every error.
// Namespace is exclusive and immutable;checksums are not authentication.
func newRankCheckpointWriter(w *mpi.World, directory, label string, h ScalingCodeHeader, samples int, supportSHA string, maxWords uint64, emit func(map[string]any)) (*rankCheckpointWriter, error) {
	if w == nil || w.Size < 2 || w.Size > 64 || w.Size&(w.Size-1) != 0 || w.Rank < 0 || w.Rank >= w.Size || label == "" || label == "." || label == ".." || filepath.Base(label) != label || maxWords == 0 || emit == nil || samples < 1 || h.Spec.FieldDegree < 2 || h.Spec.FieldDegree > 20 {
		return nil, fmt.Errorf("checkpoint writer admission")
	}
	canonical, _, err := canonicalScalingCodeHeader(h.Spec.FieldDegree, h.Spec.CurveDegree, h.K, h.N)
	if err != nil || canonical != h || h.N > uint64(^uint(0)>>1) || samples > discoverySampleLimit(h.Spec.FieldDegree) {
		return nil, fmt.Errorf("checkpoint header binding")
	}
	_, words := rankBand(h.N, w.Rank, w.Size)
	if words > maxWords {
		return nil, fmt.Errorf("checkpoint host-shard admission")
	}
	metadata := workflowCheckpointSource{Version: 1, Header: h, Ranks: w.Size, Samples: samples, SupportFormat: "uint64-exponents-le", SupportBytes: 8 * h.K}
	body, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	tag := sha256.Sum256(body)
	var own [64]byte
	copy(own[:32], tag[:])
	if w.Rank == 0 {
		digest, e := expectedDigest(supportSHA)
		if e != nil {
			return nil, e
		}
		copy(own[32:], digest)
	} else if supportSHA != "" {
		return nil, fmt.Errorf("checkpoint archive digest ownership")
	}
	frames, err := w.ControlFrames(own)
	if err != nil {
		return nil, err
	}
	if len(frames) != w.Size {
		return nil, fmt.Errorf("checkpoint source agreement extent")
	}
	for rank, frame := range frames {
		if !bytes.Equal(frame[:32], own[:32]) {
			return nil, fmt.Errorf("checkpoint source metadata disagreement")
		}
		if rank != 0 && !bytes.Equal(frame[32:], make([]byte, 32)) {
			return nil, fmt.Errorf("checkpoint source digest ownership")
		}
	}
	if frames[w.Rank] != own {
		return nil, fmt.Errorf("checkpoint source self disagreement")
	}
	metadata.SupportSHA = hex.EncodeToString(frames[0][32:])
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	if w.Rank == 0 {
		if err = writeWorkflowMetadata(directory, label+"-checkpoint-source.json", data); err != nil {
			return nil, err
		}
	}
	w.Barrier()
	return &rankCheckpointWriter{w, directory, label, h, metadata.SupportSHA, hex.EncodeToString(digest[:]), maxWords, emit}, nil
}

// Commit validates the common stage first,then writes/syncs each compact shard,
// collects every checksum and publishes the coordinated commit marker last.
// This does not resume arithmetic or reshard a cache;loader validation remains.
func (p *rankCheckpointWriter) Commit(stage string, keep uint64, values []uint32) error {
	if p == nil || p.world == nil {
		return fmt.Errorf("checkpoint writer unavailable")
	}
	w, h := p.world, p.header
	if (stage == "root" && keep != h.K+1) || (stage != "root" && (stage != fmt.Sprintf("newton-%d", keep) || keep < 1 || keep > h.N-h.K || (keep != h.N-h.K && keep&(keep-1) != 0))) {
		return fmt.Errorf("checkpoint stage admission")
	}
	_, words, _, err := rankPrefixExtent("prefix", h.N, keep, w.Rank, w.Size)
	if err != nil || words > p.maxWords || uint64(len(values)) != words {
		return fmt.Errorf("checkpoint shard admission")
	}
	record := rankPrefixRecord{Layout: "prefix", S: h.Spec.FieldDegree, K: int(h.K), N: int(h.N), Prefix: int(keep), Ranks: w.Size, SupportSHA: p.supportSHA, Stage: stage, SourceSHA: p.sourceSHA}
	metadata, err := json.Marshal(record)
	if err != nil {
		return err
	}
	tag := sha256.Sum256(metadata)
	var frame [64]byte
	copy(frame[:32], tag[:])
	copy(frame[32:], tag[:])
	frames, err := w.ControlFrames(frame)
	if err != nil {
		return err
	}
	if len(frames) != w.Size {
		return fmt.Errorf("checkpoint metadata agreement extent")
	}
	for _, remote := range frames {
		if remote != frame {
			return fmt.Errorf("checkpoint stage metadata disagreement")
		}
	}
	name := p.label + "-" + stage + "-commit.json"
	if _, e := os.Lstat(filepath.Join(p.directory, name)); !os.IsNotExist(e) {
		return fmt.Errorf("checkpoint already committed")
	}
	own, err := writeCompactRankPrefix(filepath.Join(p.directory, fmt.Sprintf("%s-rank-%d-%s.bin", p.label, w.Rank, stage)), values, h.Spec.FieldDegree, h.N, keep, w.Rank, w.Size)
	if err != nil {
		return err
	}
	copy(frame[:32], own[:])
	copy(frame[32:], own[:])
	frames, err = w.ControlFrames(frame)
	if err != nil {
		return err
	}
	record.SHA = make([]string, w.Size)
	if len(frames) != w.Size {
		return fmt.Errorf("checkpoint digest agreement extent")
	}
	for rank, remote := range frames {
		if !bytes.Equal(remote[:32], remote[32:]) {
			return fmt.Errorf("checkpoint shard digest disagreement")
		}
		record.SHA[rank] = hex.EncodeToString(remote[:32])
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	if w.Rank == 0 {
		if err = writeWorkflowMetadata(p.directory, name, data); err != nil {
			return err
		}
	}
	w.Barrier()
	p.emit(map[string]any{"kind": "workflow-checkpoint", "case": p.label, "record": record})
	return nil
}
