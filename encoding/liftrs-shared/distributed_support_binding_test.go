package liftrs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRankSupportAgreement(t *testing.T) {
	directory := t.TempDir()
	data := []byte("immutable support fixture")
	if err := os.WriteFile(filepath.Join(directory, "s4-support.bin"), data, 0600); err != nil {
		t.Fatal(err)
	}
	expected := sha256.Sum256(data)
	record := rankPrefixRecord{Layout: "prefix", S: 4, N: 256, K: 97, Prefix: 98, Ranks: 4, Stage: "root", SupportSHA: hex.EncodeToString(expected[:]), SHA: make([]string, 4)}
	for i := range record.SHA {
		record.SHA[i] = fmt.Sprintf("%064x", i)
	}
	for rank := 0; rank < 4; rank++ {
		binding, err := newRankSupportBinding(directory, "s4", rank, 4)
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		gather := func(own [64]byte) ([][64]byte, error) {
			calls++
			frames := make([][64]byte, 4)
			for i := range frames {
				copy(frames[i][:32], own[:32])
			}
			copy(frames[0][32:], expected[:])
			return frames, nil
		}
		if err = binding.verify(record, gather, nil); err != nil {
			t.Fatal(err)
		}
		next := record
		next.Stage = "newton-8"
		next.Prefix = 8
		if err = binding.verify(next, gather, nil); err != nil {
			t.Fatal(err)
		}
		wantedPasses := uint64(0)
		if rank == 0 {
			wantedPasses = 1
		}
		if calls != 2 || binding.Agreements != 2 || binding.HashPasses != wantedPasses {
			t.Fatal("hash/agreement counts", binding)
		}
		if rank == 0 && binding.HashBytes != uint64(len(data)) {
			t.Fatal("byte count")
		}
		bad := next
		bad.SupportSHA = fmt.Sprintf("%064x", 0)
		if err = binding.verify(bad, gather, nil); err == nil {
			t.Fatal("cross-record mismatch accepted")
		}
	}
	for _, change := range []func([][64]byte){func(f [][64]byte) { f[2][0] ^= 1 }, func(f [][64]byte) { f[0][32] ^= 1 }, func(f [][64]byte) { f[3][32] = 1 }} {
		b, _ := newRankSupportBinding(directory, "s4", 1, 4)
		if err := b.verify(record, func(own [64]byte) ([][64]byte, error) {
			frames := make([][64]byte, 4)
			for i := range frames {
				copy(frames[i][:32], own[:32])
			}
			copy(frames[0][32:], expected[:])
			change(frames)
			return frames, nil
		}, nil); err == nil || b.checked {
			t.Fatal("invalid collective accepted")
		}
	}
	b, _ := newRankSupportBinding(directory, "s4", 0, 4)
	if err := os.WriteFile(filepath.Join(directory, "s4-support.bin"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	contacted := false
	if err := b.verify(record, func(own [64]byte) ([][64]byte, error) { contacted = true; return nil, nil }, nil); err == nil || contacted {
		t.Fatal("corrupt support reached agreement")
	}
	if _, _, err := hashRankSupport(filepath.Join(directory, "s4-support.bin"), 2, nil); err == nil {
		t.Fatal("hash budget accepted")
	}
	for _, mode := range []string{"", "local", "coordinator"} {
		if _, err := restartSupportMode(mode); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := restartSupportMode("bad"); err == nil {
		t.Fatal("bad mode")
	}
}

// Independent Python/Go fixture binds typed metadata field order and full record
// contents even though persisted commit maps use a different key order.
func TestRankSupportMetadataGolden(t *testing.T) {
	b, e := newRankSupportBinding("/shared/checkpoint", "s4", 0, 2)
	if e != nil {
		t.Fatal(e)
	}
	r := rankPrefixRecord{Layout: "prefix", S: 4, K: 97, N: 256, Prefix: 98, Ranks: 2, SupportSHA: "0000000000000000000000000000000000000000000000000000000000000000", Stage: "root", SHA: []string{"0000000000000000000000000000000000000000000000000000000000000000", "0101010101010101010101010101010101010101010101010101010101010101"}}
	frame, e := b.frame(r)
	if e != nil || hex.EncodeToString(frame[:32]) != "59d6e0931e823a7c6d5502a2114947d81df298945f2373bc7689d61af281f35a" {
		t.Fatal("metadata codec", e, hex.EncodeToString(frame[:32]))
	}
}
