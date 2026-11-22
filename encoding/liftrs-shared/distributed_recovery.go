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
	// Empty is the legacy full padded shard;prefix stores only its owned prefix.
	Layout     string   `json:"layout,omitempty"`
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
	first, words, _, extentErr := rankPrefixExtent(r.Layout, n, uint64(count), rank, ranks)
	if extentErr != nil {
		return state, extentErr
	}
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
	state.Values, err = readRankPrefixValues(filepath.Join(directory, fmt.Sprintf("%s-rank-%d-%s.bin", label, rank, stage)), s, r.Layout, n, uint64(count), rank, ranks, maxWords, r.SHA[rank])
	if err != nil {
		return state, err
	}
	if stage == "root" && first == 0 && len(state.Values) > 0 && state.Values[0] != 1 {
		return state, fmt.Errorf("root normalization")
	}
	return state, nil
}

// Prefix layout derives its exact payload extent from bound global metadata;
// a padding-only rank still creates an empty file with the SHA-256 of empty data.
func rankPrefixExtent(layout string, n, prefix uint64, rank, ranks int) (first, words, stored uint64, err error) {
	if (layout != "" && layout != "prefix") || n == 0 || prefix == 0 || prefix > n || ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks || n > uint64(^uint(0)>>1)/4 {
		return 0, 0, 0, fmt.Errorf("prefix layout/extent admission")
	}
	first, words = rankBand(n, rank, ranks)
	stored = words
	if layout == "prefix" {
		stored = min(words, prefix-min(prefix, first))
	}
	return first, words, stored, nil
}

// writeCompactRankPrefix omits only verified zero padding. Base coefficients
// retain canonical4-byte slots;bounded I/O avoids an extra payload-sized buffer.
// A partial file is never sufficient for a coordinated completed-prefix commit.
func writeCompactRankPrefix(path string, values []uint32, s int, n, prefix uint64, rank, ranks int) (digest [32]byte, err error) {
	_, words, stored, err := rankPrefixExtent("prefix", n, prefix, rank, ranks)
	if err != nil {
		return digest, err
	}
	if s < 2 || s > 20 || uint64(len(values)) != words {
		return digest, fmt.Errorf("compact prefix input")
	}
	for _, value := range values[stored:] {
		if value != 0 {
			return digest, fmt.Errorf("nonzero omitted prefix padding")
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return digest, err
	}
	hash := sha256.New()
	err = writeScalingCodeValues(io.MultiWriter(file, hash), 4, uint64(1)<<uint(s), values[:stored], nil)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return digest, err
	}
	if closeErr != nil {
		return digest, closeErr
	}
	copy(digest[:], hash.Sum(nil))
	return digest, nil
}

// Both layouts restore the same admitted resident shard and validate exact EOF,
// digest,field range and padding. Compact I/O streams directly into that array;
// legacy padded I/O retains its bulk buffer. Neither reload re-reads support.
func readRankPrefixValues(path string, s int, layout string, n, prefix uint64, rank, ranks int, maxWords uint64, digest string) ([]uint32, error) {
	first, words, stored, err := rankPrefixExtent(layout, n, prefix, rank, ranks)
	if err != nil {
		return nil, err
	}
	if s < 2 || s > 20 || words > maxWords || words > uint64(^uint(0)>>1)/4 {
		return nil, fmt.Errorf("prefix resident admission")
	}
	expected, err := expectedDigest(digest)
	if err != nil {
		return nil, err
	}
	// Preserve the validated padded reader's bulk-I/O behavior. Streaming is an
	// explicit compact-layout choice,whose large-file throughput is unmeasured.
	if layout == "" {
		data, err := readBoundedPilotFile(path, 4*words)
		if err != nil {
			return nil, err
		}
		return decodePaddedPrefix(data, s, first, prefix, words, digest)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := make([]uint32, words)
	hash := sha256.New()
	var block [65536]byte
	for offset := uint64(0); offset < stored; {
		take := min(uint64(len(block)/4), stored-offset)
		data := block[:4*take]
		if _, err = io.ReadFull(file, data); err != nil {
			return nil, err
		}
		hash.Write(data)
		for j := uint64(0); j < take; j++ {
			v := binary.LittleEndian.Uint32(data[4*j:])
			if uint64(v) >= uint64(1)<<uint(s) || (first+offset+j >= prefix && v != 0) {
				return nil, fmt.Errorf("prefix canonical value/padding")
			}
			values[offset+j] = v
		}
		offset += take
	}
	var extra [1]byte
	if count, err := file.Read(extra[:]); count != 0 || err != io.EOF {
		return nil, fmt.Errorf("prefix payload extent")
	}
	if !bytes.Equal(hash.Sum(nil), expected[:]) {
		return nil, fmt.Errorf("prefix payload digest")
	}
	return values, nil
}

// Legacy bulk reload keeps its timing/memory boundary;both fresh-process and
// in-process callers use the same range,padding,extent and digest checks.
func decodePaddedPrefix(data []byte, s int, first, prefix, words uint64, digest string) ([]uint32, error) {
	if s < 2 || s > 20 || words > uint64(^uint(0)>>1)/4 || uint64(len(data)) != 4*words {
		return nil, fmt.Errorf("padded prefix extent")
	}
	expected, err := expectedDigest(digest)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if !bytes.Equal(sum[:], expected) {
		return nil, fmt.Errorf("padded prefix digest")
	}
	values := make([]uint32, words)
	for j := range values {
		v := binary.LittleEndian.Uint32(data[4*j:])
		if uint64(v) >= uint64(1)<<uint(s) || (first+uint64(j) >= prefix && v != 0) {
			return nil, fmt.Errorf("padded prefix value/padding")
		}
		values[j] = v
	}
	return values, nil
}
