//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"encoding/binary"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"math/rand"
	"testing"
)

// These test only policy/count/wire admission, without CUDA or MPI initialization.
func TestRankDiscoveryRuntimeAdmission(t *testing.T) {
	w := &mpi.World{Rank: 0, Size: 4}
	spec := Spec{8, 2, 5}
	emit := func(map[string]any) {}
	o := rankDiscoveryOptions{DeviceBytes: 64 << 20, MaxLocalWords: 16384, MaxGoodWords: 65536, MaxRootWords: 65536, MaxArchiveWords: 65536, Random: rand.New(rand.NewSource(109))}
	if err := admitRankDiscovery[uint32](w, spec, o, emit); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*rankDiscoveryOptions){func(o *rankDiscoveryOptions) { o.Random = nil }, func(o *rankDiscoveryOptions) { o.MaxLocalWords = 16383 }, func(o *rankDiscoveryOptions) { o.ResultMode = "bad" }, func(o *rankDiscoveryOptions) { o.MaxGoodWords = 0 }, func(o *rankDiscoveryOptions) { o.DeviceBytes = 0 }} {
		bad := o
		change(&bad)
		if _, err := discoverRankSupport[uint32](w, spec, bad, emit, nil); err == nil {
			t.Fatal("invalid policy reached runtime")
		}
	}
	peer := &mpi.World{Rank: 1, Size: 4}
	peerOptions := o
	peerOptions.Random = nil
	if err := admitRankDiscovery[uint32](peer, spec, peerOptions, emit); err != nil {
		t.Fatal("non-coordinator reader required", err)
	}
	if err := admitRankDiscovery[uint64](w, spec, o, emit); err == nil {
		t.Fatal("wrong slots")
	}
	if err := discoveryWire(w, 1, 1, 65536, 0, 0, 0, []uint32{1}, emit); err == nil {
		t.Fatal("self wire")
	}
	if err := discoveryWire[uint64](w, 2, 0, 16, 16, 0, 1, []uint64{1}, emit); err == nil {
		t.Fatal("wire interval")
	}
}
func TestRankDiscoveryCountFrames(t *testing.T) {
	frames := make([][64]byte, 4)
	for rank := range frames {
		for i, v := range []uint64{uint64(rank + 1), 2, 9, uint64(rank)} {
			binary.LittleEndian.PutUint64(frames[rank][8*i:], v)
		}
	}
	counts, err := validateDiscoveryCounts(frames, 4, 2, 9, 10)
	if err != nil || discoveryTotal(counts) != 10 {
		t.Fatal(counts, err)
	}
	if _, err := validateDiscoveryCounts(frames, 4, 2, 9, 9); err == nil {
		t.Fatal("sum exceeds bound")
	}
	for _, offset := range []int{8, 16, 24, 32} {
		bad := append([][64]byte(nil), frames...)
		bad[1][offset] ^= 1
		if _, err := validateDiscoveryCounts(bad, 4, 2, 9, 10); err == nil {
			t.Fatal("bad phase/sequence/rank/reserved", offset)
		}
	}
	if _, err := validateDiscoveryCounts(frames[:3], 4, 2, 9, 10); err == nil {
		t.Fatal("missing count frame")
	}
}
