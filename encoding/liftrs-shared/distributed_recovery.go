package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type rankPrefixRecord struct {
	S          int      `json:"s"`
	K          int      `json:"k"`
	N          int      `json:"n"`
	Prefix     int      `json:"prefix"`
	Ranks      int      `json:"ranks"`
	SupportSHA string   `json:"support_sha256"`
	Stage      string   `json:"stage"`
	SHA        []string `json:"sha256"`
}
type rankPrefixState struct {
	Record rankPrefixRecord
	Values []uint32
}

func readBoundedPilotFile(path string, limit uint64) ([]byte, error) {
	if limit > uint64(^uint(0)>>1)-1 {
		return nil, fmt.Errorf("file admission")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if uint64(len(data)) > limit {
		return nil, fmt.Errorf("file exceeds admitted bound")
	}
	return data, nil
}

// Load only an admitted canonical shard, binding shape, source support and hash
// before any GPU allocation. Manifests/checksums detect corruption, not malicious
// rewrites; callers must retain trusted experiment/source provenance separately.
func loadRankPrefix(directory, label string, s, rank, ranks int, stage string, maxWords uint64) (rankPrefixState, error) {
	var state rankPrefixState
	if s < 2 || s > 20 || ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks {
		return state, fmt.Errorf("prefix ownership")
	}
	data, err := readBoundedPilotFile(filepath.Join(directory, label+"-"+stage+"-commit.json"), 64<<10)
	if err != nil {
		return state, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&state.Record); err != nil {
		return state, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return state, fmt.Errorf("prefix trailing metadata")
	}
	r := state.Record
	n := uint64(1) << uint(2*s)
	if r.S != s || r.N != int(n) || r.Ranks != ranks || r.K < 1 || uint64(r.K) > n/2 || r.Stage != stage || len(r.SHA) != ranks {
		return state, fmt.Errorf("prefix parameter binding")
	}
	count := r.Prefix
	if stage == "root" {
		if count != r.K+1 {
			return state, fmt.Errorf("root degree")
		}
	} else {
		if stage != fmt.Sprintf("newton-%d", count) || count < 1 || count > r.N-r.K || (count != r.N-r.K && count&(count-1) != 0) {
			return state, fmt.Errorf("incomplete Newton prefix")
		}
	}
	first, words := rankBand(n, rank, ranks)
	if words > maxWords || words > uint64(^uint(0)>>1)/4 {
		return state, fmt.Errorf("prefix local memory admission")
	}
	for _, digest := range append(append([]string(nil), r.SHA...), r.SupportSHA) {
		b, e := hex.DecodeString(digest)
		if e != nil || len(b) != 32 {
			return state, fmt.Errorf("prefix digest format")
		}
	}
	source, err := os.Open(filepath.Join(directory, label+"-support.bin"))
	if err != nil {
		return state, err
	}
	hasher := sha256.New()
	limit := 16*n + 65536
	size, err := io.Copy(hasher, io.LimitReader(source, int64(limit)+1))
	closeErr := source.Close()
	if err != nil || closeErr != nil || uint64(size) > limit || fmt.Sprintf("%x", hasher.Sum(nil)) != r.SupportSHA {
		return state, fmt.Errorf("prefix support binding")
	}
	data, err = readBoundedPilotFile(filepath.Join(directory, fmt.Sprintf("%s-rank-%d-%s.bin", label, rank, stage)), 4*words)
	if err != nil {
		return state, err
	}
	if uint64(len(data)) != 4*words || fmt.Sprintf("%x", sha256.Sum256(data)) != r.SHA[rank] {
		return state, fmt.Errorf("prefix shard hash/extent")
	}
	state.Values = make([]uint32, words)
	for j := range state.Values {
		v := binary.LittleEndian.Uint32(data[4*j:])
		if uint64(v) >= uint64(1)<<uint(s) || (first+uint64(j) >= uint64(count) && v != 0) {
			return state, fmt.Errorf("prefix coefficient/padding")
		}
		state.Values[j] = v
	}
	if stage == "root" && first == 0 && state.Values[0] != 1 {
		return state, fmt.Errorf("root normalization")
	}
	return state, nil
}
