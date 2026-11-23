package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"sort"
)

type proofLogCount struct {
	records, bytes uint64
	digest         hash.Hash
}

// proofLog retains at most one hash/count per fixed proof kind. The caller must
// serialize Record/Summaries (the rank logger already holds its mutex). Digests
// describe ordered public experiment evidence,not authentication or verification.
// Runtime MPI/field/identity checks remain outside this recording policy.
type proofLog struct {
	mode   string
	counts map[string]*proofLogCount
	total  uint64
}

func newProofLog(mode string) (*proofLog, error) {
	if mode == "" || mode == "full" {
		return nil, nil
	}
	if mode != "trace" && mode != "summary" {
		return nil, fmt.Errorf("proof log mode")
	}
	return &proofLog{mode: mode, counts: map[string]*proofLogCount{}}, nil
}

func proofLogKind(kind string) bool {
	switch kind {
	case "tile-proof", "taylor-proof", "operand-proof", "discovery-wire", "support-proof":
		return true
	}
	return false
}

// Record hides selected rows only in summary mode. Trace additionally writes
// full rows for small-case independent replay. Full/default does no hash work.
func (p *proofLog) Record(row map[string]any) (bool, error) {
	if p == nil {
		return false, nil
	}
	kind, _ := row["kind"].(string)
	if !proofLogKind(kind) {
		return false, nil
	}
	body := make(map[string]any, len(row))
	for key, value := range row {
		switch key {
		case "utc", "rank", "elapsed_seconds", "peak_rss_bytes":
			continue
		}
		body[key] = value
	}
	data, err := json.Marshal(body)
	if err != nil {
		return false, err
	}
	count := p.counts[kind]
	if count == nil {
		count = &proofLogCount{digest: sha256.New()}
		p.counts[kind] = count
	}
	var length [8]byte
	binary.LittleEndian.PutUint64(length[:], uint64(len(data)))
	count.digest.Write(length[:])
	count.digest.Write(data)
	count.records++
	count.bytes += uint64(len(data))
	p.total++
	return p.mode == "summary", nil
}

func (p *proofLog) Summaries(complete bool) []map[string]any {
	if p == nil {
		return nil
	}
	keys := make([]string, 0, len(p.counts))
	for kind := range p.counts {
		keys = append(keys, kind)
	}
	sort.Strings(keys)
	rows := make([]map[string]any, 0, len(keys))
	for _, kind := range keys {
		count := p.counts[kind]
		rows = append(rows, map[string]any{"kind": "proof-summary", "proof_kind": kind, "records": count.records, "semantic_bytes": count.bytes, "sha256": hex.EncodeToString(count.digest.Sum(nil)), "complete": complete, "mode": p.mode, "codec": "go-json-v1-length64le"})
	}
	return rows
}
