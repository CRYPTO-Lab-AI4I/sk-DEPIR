package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func restartSupportMode(mode string) (string, error) {
	if mode == "" {
		return "local", nil
	}
	if mode != "local" && mode != "coordinator" {
		return "", fmt.Errorf("restart support mode")
	}
	return mode, nil
}

type supportReadProgress struct {
	io.Reader
	done, total uint64
	report      func(uint64, uint64)
}

func (r *supportReadProgress) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if n > 0 {
		r.done += uint64(n)
		r.report(r.done, r.total)
	}
	return n, err
}

// Hash the complete admitted support file with bounded memory. The cache is
// immutable trusted job input; checksums detect corruption, not malicious rewrites.
func hashRankSupport(path string, limit uint64, report func(uint64, uint64)) (digest [32]byte, size uint64, err error) {
	if limit > 1<<63-2 {
		return digest, 0, fmt.Errorf("support hash limit")
	}
	source, err := os.Open(path)
	if err != nil {
		return digest, 0, err
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil {
		return digest, 0, err
	}
	if info.Size() < 0 || uint64(info.Size()) > limit {
		return digest, 0, fmt.Errorf("support file admission")
	}
	var reader io.Reader = source
	if report != nil {
		report(0, uint64(info.Size()))
		reader = &supportReadProgress{Reader: source, total: uint64(info.Size()), report: report}
	}
	hasher := sha256.New()
	read, err := io.CopyBuffer(hasher, io.LimitReader(reader, int64(limit)+1), make([]byte, 65536))
	if err != nil {
		return digest, uint64(read), fmt.Errorf("support hash read: %w", err)
	}
	if read != info.Size() {
		return digest, uint64(read), fmt.Errorf("support changed size during hash")
	}
	if err = source.Close(); err != nil {
		return digest, uint64(read), err
	}
	copy(digest[:], hasher.Sum(nil))
	return digest, uint64(read), nil
}

// One binding belongs to one immutable cache and one coordinated recovery. Each
// checkpoint still requires collective agreement on its complete metadata. Only
// rank0 supplies the actual full-file hash; peers must contribute zero hash bytes.
// A failed agreement is not cached and collective callers must abort their peers.
type rankSupportBinding struct {
	directory, label                  string
	rank, ranks                       int
	checked                           bool
	digest                            [32]byte
	HashPasses, HashBytes, Agreements uint64
}

func newRankSupportBinding(directory, label string, rank, ranks int) (*rankSupportBinding, error) {
	if label == "" || ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks {
		return nil, fmt.Errorf("support binding ownership")
	}
	absolute, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	return &rankSupportBinding{directory: absolute, label: label, rank: rank, ranks: ranks}, nil
}
func (b *rankSupportBinding) frame(record rankPrefixRecord) ([64]byte, error) {
	var frame [64]byte
	if record.Ranks != b.ranks || record.S < 2 || record.S > 20 || record.N != 1<<uint(2*record.S) {
		return frame, fmt.Errorf("support binding parameters")
	}
	body, err := json.Marshal(struct {
		Context, Directory, Label string
		Record                    rankPrefixRecord
	}{"sk-DEPIR-rank-support-v1", b.directory, b.label, record})
	if err != nil {
		return frame, err
	}
	digest := sha256.Sum256(body)
	copy(frame[:32], digest[:])
	return frame, nil
}
func (b *rankSupportBinding) verify(record rankPrefixRecord, gather func([64]byte) ([][64]byte, error), report func(uint64, uint64)) error {
	if gather == nil {
		return fmt.Errorf("support agreement required")
	}
	decoded, err := expectedDigest(record.SupportSHA)
	if err != nil {
		return err
	}
	var expected [32]byte
	copy(expected[:], decoded)
	if b.checked && b.digest != expected {
		return fmt.Errorf("restart cross-record support binding")
	}
	own, err := b.frame(record)
	if err != nil {
		return err
	}
	if b.rank == 0 {
		observed := b.digest
		if !b.checked {
			var size uint64
			observed, size, err = hashRankSupport(filepath.Join(b.directory, b.label+"-support.bin"), 16*uint64(record.N)+65536, report)
			if err != nil {
				return err
			}
			b.HashPasses++
			b.HashBytes += size
		}
		if observed != expected {
			return fmt.Errorf("coordinator support checksum")
		}
		copy(own[32:], observed[:])
	}
	frames, err := gather(own)
	if err != nil {
		return err
	}
	if len(frames) != b.ranks {
		return fmt.Errorf("support agreement ownership")
	}
	for rank, frame := range frames {
		if !bytes.Equal(frame[:32], own[:32]) {
			return fmt.Errorf("support checkpoint metadata disagreement")
		}
		var wanted [32]byte
		if rank == 0 {
			wanted = expected
		}
		if !bytes.Equal(frame[32:], wanted[:]) {
			return fmt.Errorf("support coordinator digest disagreement")
		}
	}
	if frames[b.rank] != own {
		return fmt.Errorf("support self agreement")
	}
	b.digest = expected
	b.checked = true
	b.Agreements++
	return nil
}
