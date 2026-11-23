package liftrs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflowCheckpointSourceBinding(t *testing.T) {
	directory := t.TempDir()
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	support := make([]byte, 8*h.K)
	supportSHA := fmt.Sprintf("%x", sha256.Sum256(support))
	if err = os.WriteFile(filepath.Join(directory, "s4-support.bin"), support, 0600); err != nil {
		t.Fatal(err)
	}
	source := workflowCheckpointSource{Version: 1, Header: h, Ranks: 4, Samples: 2, SupportFormat: "uint64-exponents-le", SupportSHA: supportSHA, SupportBytes: uint64(len(support))}
	data, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	sourceSHA := fmt.Sprintf("%x", sha256.Sum256(data))
	if err = writeWorkflowMetadata(directory, "s4-checkpoint-source.json", data); err != nil {
		t.Fatal(err)
	}
	values := make([]uint32, 64)
	values[0] = 1
	digest, err := writeCompactRankPrefix(filepath.Join(directory, "s4-rank-0-root.bin"), values, 4, 256, 98, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	record := rankPrefixRecord{Layout: "prefix", S: 4, K: 97, N: 256, Prefix: 98, Ranks: 4, SupportSHA: supportSHA, Stage: "root", SHA: []string{fmt.Sprintf("%x", digest), fmt.Sprintf("%064x", 0), fmt.Sprintf("%064x", 0), fmt.Sprintf("%064x", 0)}, SourceSHA: sourceSHA}
	commit, _ := json.Marshal(record)
	if err = writeWorkflowMetadata(directory, "s4-root-commit.json", commit); err != nil {
		t.Fatal(err)
	}
	if _, err = loadRankPrefix(directory, "s4", 4, 0, 4, "root", 64); err != nil {
		t.Fatal(err)
	}
	if err = writeWorkflowMetadata(directory, "s4-root-commit.json", commit); err == nil {
		t.Fatal("checkpoint overwritten")
	}
	if err = os.WriteFile(filepath.Join(directory, "s4-checkpoint-source.json"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = loadRankPrefix(directory, "s4", 4, 0, 4, "root", 64); err == nil {
		t.Fatal("changed source metadata accepted")
	}
}
