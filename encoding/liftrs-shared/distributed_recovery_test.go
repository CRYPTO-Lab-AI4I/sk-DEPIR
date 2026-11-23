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
	// External support agreement must never bypass the shared shard checksum.
	if _, err = loadRankPrefixWithSupport(directory, "s4", 4, 0, 2, "newton-8", 128, func(rankPrefixRecord) error { return nil }); err == nil {
		t.Fatal("agreed support bypassed corrupt shard")
	}
}

// Compact roots include a partially owned rank and completely empty payloads.
// All formats restore the same canonical resident ownership,without disk padding.
func TestRankPrefixCompactRoundTrip(t *testing.T) {
	directory := t.TempDir()
	support := []byte("trusted compact checkpoint support")
	if err := os.WriteFile(filepath.Join(directory, "s4-support.bin"), support, 0600); err != nil {
		t.Fatal(err)
	}
	record := rankPrefixRecord{Layout: "prefix", S: 4, K: 97, N: 256, Prefix: 98, Ranks: 4, SupportSHA: fmt.Sprintf("%x", sha256.Sum256(support)), Stage: "root", SHA: make([]string, 4)}
	inputs := make([][]uint32, 4)
	for rank := range inputs {
		inputs[rank] = make([]uint32, 64)
		for j := range inputs[rank] {
			index := rank*64 + j
			if index < 98 {
				inputs[rank][j] = uint32(index%15 + 1)
			}
		}
		path := filepath.Join(directory, fmt.Sprintf("s4-rank-%d-root.bin", rank))
		digest, err := writeCompactRankPrefix(path, inputs[rank], 4, 256, 98, rank, 4)
		if err != nil {
			t.Fatal(err)
		}
		record.SHA[rank] = fmt.Sprintf("%x", digest)
		info, err := os.Stat(path)
		if err != nil || info.Size() != []int64{256, 136, 0, 0}[rank] {
			t.Fatal("stored extent", info, err)
		}
	}
	commit := filepath.Join(directory, "s4-root-commit.json")
	save := func(r rankPrefixRecord) {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(commit, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	save(record)
	for rank, want := range inputs {
		state, err := loadRankPrefix(directory, "s4", 4, rank, 4, "root", 64)
		if err != nil {
			t.Fatal(err)
		}
		for j, v := range want {
			if state.Values[j] != v {
				t.Fatal("restored coefficient", rank, j)
			}
		}
	}
	for _, layout := range []string{"unknown", ""} {
		bad := record
		bad.Layout = layout
		save(bad)
		if _, err := loadRankPrefix(directory, "s4", 4, 1, 4, "root", 64); err == nil {
			t.Fatal("wrong layout admitted")
		}
	}
	save(record)
	path := filepath.Join(directory, "s4-rank-1-root.bin")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data[0] ^= 1
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRankPrefix(directory, "s4", 4, 1, 4, "root", 64); err == nil {
		t.Fatal("corrupt compact data")
	}
	data[0] ^= 1
	data = append(data, 0)
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRankPrefix(directory, "s4", 4, 1, 4, "root", 64); err == nil {
		t.Fatal("extra compact bytes")
	}
	bad := make([]uint32, 64)
	bad[0] = 1
	if _, err := writeCompactRankPrefix(filepath.Join(directory, "bad.bin"), bad, 4, 256, 98, 2, 4); err == nil {
		t.Fatal("nonzero omitted padding")
	}
}

func TestRankPrefixCompactBoundsAndCanonicalValues(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "prefix.bin")
	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data, 16)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	if _, err := readRankPrefixValues(path, 4, "prefix", 256, 2, 0, 4, 64, digest); err == nil {
		t.Fatal("out-of-field payload")
	}
	binary.LittleEndian.PutUint32(data, 1)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	digest = fmt.Sprintf("%x", sha256.Sum256(data))
	if _, err := readRankPrefixValues(path, 4, "prefix", 256, 2, 0, 4, 63, digest); err == nil {
		t.Fatal("resident admission")
	}
	if _, _, _, err := rankPrefixExtent("prefix", 256, 257, 0, 4); err == nil {
		t.Fatal("prefix exceeds domain")
	}
	// Excess ranks have empty bands after the first N owners.
	empty := filepath.Join(directory, "empty.bin")
	if err := os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	values, err := readRankPrefixValues(empty, 2, "prefix", 16, 2, 63, 64, 0, fmt.Sprintf("%x", sha256.Sum256(nil)))
	if err != nil || len(values) != 0 {
		t.Fatal("empty ownership band", err)
	}
}

// Metadata-only checks cross the2^32 boundary without allocating large arrays.
func TestRankPrefixWideExtents(t *testing.T) {
	if uint64(^uint(0)>>1) < 1<<40 {
		t.Skip("wide host extents require64-bit architecture")
	}
	for s := 16; s <= 20; s++ {
		n := uint64(1) << uint(2*s)
		prefix := n - 7
		var total uint64
		for rank := 0; rank < 16; rank++ {
			first, words, stored, err := rankPrefixExtent("prefix", n, prefix, rank, 16)
			if err != nil || first != uint64(rank)*(n/16) || words != n/16 || stored > words {
				t.Fatal("wide extent", s, rank, err)
			}
			total += stored
		}
		if total != prefix {
			t.Fatal("wide prefix count truncated", s, total, prefix)
		}
	}
}
