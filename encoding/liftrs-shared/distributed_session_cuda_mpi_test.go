//go:build linux && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"testing"
)

// Admission and terminal-state failures must happen before any CUDA/MPI call.
func TestRankGPUSessionAdmission(t *testing.T) {
	options := cuda.PreparationOptions{S: 18}
	world := &mpi.World{Rank: 0, Size: 4}
	emit := func(map[string]any) {}
	if _, err := newRankGPUSession[uint64](nil, options, 12, "case", emit); err == nil {
		t.Fatal("nil communicator")
	}
	for _, dim := range []int{-1, 1, 37, 1000} {
		if _, err := newRankGPUSession[uint64](world, options, dim, "case", emit); err == nil {
			t.Fatal("dimension", dim)
		}
	}
	if _, err := newRankGPUSession[uint64](world, options, 12, "case", nil); err == nil {
		t.Fatal("missing observer")
	}
	if _, err := newRankGPUSession[uint32](world, options, 12, "case", emit); err == nil {
		t.Fatal("narrow instantiation of wide storage")
	}
	options.S = 8
	if _, err := newRankGPUSession[uint64](world, options, 12, "case", emit); err == nil {
		t.Fatal("wide instantiation of narrow storage")
	}

}
func TestRankGPUSessionFailureAndClose(t *testing.T) {
	var absent *rankGPUSession[uint32]
	if err := absent.Close(); err != nil {
		t.Fatal(err)
	}
	// Invalid product bounds reject before touching this uninitialized storage.
	// The session then rejects even a well-shaped retry until peers abort.
	d := &rankGPUSession[uint32]{Executor: rankGPUExecutor[uint32]{GPU: &cuda.RankStorage{}}}
	if _, err := d.Product("bad", nil, nil, 0, 1, 1); err == nil || !d.failed {
		t.Fatal("failed session not poisoned")
	}
	if _, err := d.ProductPrefixInputs("retry", nil, nil, 1, 1, 1); err == nil {
		t.Fatal("reused failed session")
	}
	if _, err := d.Spectrum("retry", nil, 1); err == nil {
		t.Fatal("spectrum reused failed session")
	}
	if err := d.PrepareSpectrum("retry", nil, 1); err == nil {
		t.Fatal("prepare reused failed session")
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if err := d.Close(); err != nil {
		t.Fatal(err)
	}
	if !d.closed {
		t.Fatal("close state")
	}
}
