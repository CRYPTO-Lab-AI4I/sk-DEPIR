package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRankPrefixResumeAdmission(t *testing.T) {
	directory := t.TempDir()
	support := []byte("trusted small fixture support")
	if err := os.WriteFile(filepath.Join(directory, "s4-support.bin"), support, 0600); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 128*4)
	binary.LittleEndian.PutUint32(data, 1)
	record := rankPrefixRecord{S: 4, K: 97, N: 256, Prefix: 8, Ranks: 2, SupportSHA: fmt.Sprintf("%x", sha256.Sum256(support)), Stage: "newton-8", SHA: []string{fmt.Sprintf("%x", sha256.Sum256(data)), fmt.Sprintf("%x", sha256.Sum256(data))}}
	path := filepath.Join(directory, "s4-rank-0-newton-8.bin")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	commit := filepath.Join(directory, "s4-newton-8-commit.json")
	save := func(r rankPrefixRecord) {
		b, _ := json.Marshal(r)
		if err := os.WriteFile(commit, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	save(record)
	state, err := loadRankPrefix(directory, "s4", 4, 0, 2, "newton-8", 128)
	if err != nil || len(state.Values) != 128 {
		t.Fatal(state, err)
	}
	if _, err = loadRankPrefix(directory, "s4", 4, 0, 2, "newton-8", 127); err == nil {
		t.Fatal("memory admission")
	}
	for _, change := range []func(*rankPrefixRecord){func(r *rankPrefixRecord) { r.SupportSHA = fmt.Sprintf("%064x", 0) }, func(r *rankPrefixRecord) { r.Ranks = 3 }, func(r *rankPrefixRecord) { r.Prefix = 7 }} {
		bad := record
		change(&bad)
		save(bad)
		if _, err = loadRankPrefix(directory, "s4", 4, 0, 2, "newton-8", 128); err == nil {
			t.Fatal("corrupt binding accepted", bad)
		}
	}
	save(record)
	data[4] ^= 1
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = loadRankPrefix(directory, "s4", 4, 0, 2, "newton-8", 128); err == nil {
		t.Fatal("bad shard accepted")
	}
}
