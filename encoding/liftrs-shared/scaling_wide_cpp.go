package liftrs

import (
	"fmt"
	"sync"
	"time"
)

// WideCPPResources counts initial workspace and output backing payloads. Code
// arrays, input, native metadata, retained previous words and runtime/headroom
// are extra. At s<=16 this optional wide path uses wider work arrays than the
// existing32-bit encoder; it does not replace that default.
type WideCPPResources struct {
	PrefixElements, ArenaElements, WorkspaceBytes, OutputBackingBytes, TransientBytes uint64
}

// WideCPPEncodingStats exposes native clocks including checks, stage clocks and
// backing sizes. Setup, Go dispatch and scratch allocation are outside Native.
// The logical uint32 word is4N bytes but OutputBackingBytes is8N.
type WideCPPEncodingStats struct {
	Native, Taylor, Columns, Rows, Buffer, Pointwise, Assembly time.Duration
	Calls, OutputBackingBytes, TransientBytes                  uint64
}
type wideCPPOnline interface {
	EncodeBorrowed([]uint32, []uint32, []uint32, []uint64, []uint64, []uint32, []uint64, int) ([]uint32, WideCPPEncodingStats, error)
	Resources() WideCPPResources
	OwnedBytes() uint64
	Close() error
}

// ScalingWideCPPEncoder is an explicit optional complete cyclic encoding path
// using64-bit extension workspace and unchanged uint32 messages/results. It
// consumes immutable verified code data but does not perform preprocessing or
// authenticate its code identities. No full s17 encoding has yet been validated.
// Calls may run concurrently on distinct workspaces; Close waits for active calls.
type ScalingWideCPPEncoder struct {
	mu        sync.RWMutex
	source    *ScalingCodeData
	backend   wideCPPOnline
	resources WideCPPResources
	workers   int
	n, k      int
	closed    bool
}

// ScalingWideCPPWorkspace owns reusable arena/H/scratch. It is not safe for
// concurrent reuse or inspection during Encode. Do not copy a live workspace:
// copies would share its backing arrays. Successful calls transfer
// scratch to the returned word, so the next call allocates new scratch. Older
// retained words count as extra caller memory beyond this workspace's cap.
type ScalingWideCPPWorkspace struct {
	owner          *ScalingWideCPPEncoder
	arena, scratch []uint64
	h              []uint32
}

// WithWideCPP prepares explicit wide native metadata, requiring liftrs_cpp/cgo.
// Ordinary builds return ErrCPPUnavailable; no CUDA dependency is introduced.
// Existing32-bit CPU/GPU encoders remain separate. Resource admission for code
// loading and the workspace is explicit, and larger arithmetic is not proof of
// valid larger preprocessing or PIR functionality.
func (d *ScalingCodeData) WithWideCPP(workers int) (*ScalingWideCPPEncoder, error) {
	h, err := canonicalScalingCodeData(d)
	if err != nil {
		return nil, err
	}
	if workers < 1 || workers > 64 {
		return nil, fmt.Errorf("invalid wide workers")
	}
	native, err := makeWideCPPOnline(h)
	if err != nil {
		return nil, err
	}
	return &ScalingWideCPPEncoder{source: d, backend: native, resources: native.Resources(), workers: workers, n: int(h.N), k: int(h.K)}, nil
}

// Resources reports immutable allocation requirements; no buffers are allocated.
func (e *ScalingWideCPPEncoder) Resources() WideCPPResources {
	if e == nil {
		return WideCPPResources{}
	}
	return e.resources
}

// OwnedBytes counts only native metadata, returning zero after Close.
func (e *ScalingWideCPPEncoder) OwnedBytes() uint64 {
	if e == nil {
		return 0
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.backend == nil {
		return 0
	}
	return e.backend.OwnedBytes()
}

// NewWorkspace admits principal workspace bytes before allocation. maxBytes is
// mandatory; its cap excludes source/input, returned older words, transient
// assembly/allocator/runtime memory. Allow all of those in a host allocation.
func (e *ScalingWideCPPEncoder) NewWorkspace(maxBytes uint64) (*ScalingWideCPPWorkspace, error) {
	if e == nil {
		return nil, fmt.Errorf("nil wide encoder")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.backend == nil {
		return nil, ErrCPPClosed
	}
	r := e.resources
	if maxBytes == 0 || r.WorkspaceBytes > maxBytes {
		return nil, fmt.Errorf("wide workspace byte cap exceeded")
	}
	return &ScalingWideCPPWorkspace{owner: e, arena: make([]uint64, int(r.ArenaElements)), h: make([]uint32, e.k), scratch: make([]uint64, e.n)}, nil
}

// Bytes reports arrays currently held, excluding scratch already transferred to
// a caller word. Add each returned word's OutputBackingBytes to live RAM counts.
func (w *ScalingWideCPPWorkspace) Bytes() uint64 {
	if w == nil {
		return 0
	}
	return 8*uint64(len(w.arena)+len(w.scratch)) + 4*uint64(len(w.h))
}

// Encode returns an owning canonical uint32 word. Its scratch backing is
// transferred rather than reused, so subsequent calls/Close cannot alter it.
func (e *ScalingWideCPPEncoder) Encode(message []uint32, w *ScalingWideCPPWorkspace) ([]uint32, error) {
	word, _, err := e.EncodeWithStats(message, w)
	return word, err
}

// EncodeWithStats separates native stages from allocation/Go dispatch. Invalid
// messages may leave private work partial; retry with valid inputs is supported.
func (e *ScalingWideCPPEncoder) EncodeWithStats(message []uint32, w *ScalingWideCPPWorkspace) ([]uint32, WideCPPEncodingStats, error) {
	if e == nil {
		return nil, WideCPPEncodingStats{}, fmt.Errorf("nil wide encoder")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.backend == nil {
		return nil, WideCPPEncodingStats{}, ErrCPPClosed
	}
	if w == nil || w.owner != e || len(message) != e.k {
		return nil, WideCPPEncodingStats{}, fmt.Errorf("wide workspace/message mismatch")
	}
	if w.scratch == nil {
		w.scratch = make([]uint64, e.n)
	}
	d := e.source
	word, stats, err := e.backend.EncodeBorrowed(message, d.p, d.spectrum32, d.spectrum64, w.arena, w.h, w.scratch, e.workers)
	if err != nil {
		return nil, WideCPPEncodingStats{}, err
	}
	w.scratch = nil // Every previous returned word keeps this old allocation alive.
	return word, stats, nil
}

// Close releases code references/native metadata after active calls finish.
// Workspace arrays and returned words are Go-owned and remain caller-owned.
func (e *ScalingWideCPPEncoder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.closed {
		e.closed = true
		if e.backend != nil {
			e.backend.Close()
		}
		e.backend = nil
		e.source = nil
	}
	return nil
}
