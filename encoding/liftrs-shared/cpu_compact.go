//go:build liftrs_cpp && cgo

package liftrs

import (
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
)

// ScalingCPPCompactWorkspace is an explicit experimental lower-memory schedule.
// Bounded column batches replace full scratch; packed canonical assembly reuses
// the arena. Calls may be slower. One caller owns a workspace at a time.
type ScalingCPPCompactWorkspace struct {
	owner          *ScalingCPPEncoder
	arena, columns []uint32
}

// NewCompactWorkspace opts into bounded columns and compact output. It requires
// both prefix operands to fit one full arena (2L<=N); normal NewWorkspace remains
// available for other shapes. Returned words retain ordinary owning semantics.
func (e *ScalingCPPEncoder) NewCompactWorkspace() (*ScalingCPPCompactWorkspace, error) {
	if e == nil {
		return nil, ErrCPPClosed
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.source == nil {
		return nil, ErrCPPClosed
	}
	n := e.source.Length()
	l := 1
	for l < 2*e.source.k-1 {
		l *= 2
	}
	if l > n/2 {
		return nil, fmt.Errorf("compact prefix operands exceed arena")
	}
	pre, err := cpp.BoundedColumnElements(l, e.plans[l].split, e.workers)
	if err != nil {
		return nil, err
	}
	full, err := cpp.BoundedColumnElements(n, e.plans[n].split, e.workers)
	if err != nil {
		return nil, err
	}
	return &ScalingCPPCompactWorkspace{e, make([]uint32, n), make([]uint32, maxIntScaling(pre, full))}, nil
}

// Bytes counts currently owned arrays, excluding the detached caller-owned word.
func (w *ScalingCPPCompactWorkspace) Bytes() uint64 {
	if w == nil {
		return 0
	}
	return uint64(len(w.arena)+len(w.columns)) * 4
}

// ScalingCPPCompactStats separates extra compaction/move and expansion stages.
// Normal columns timing includes bounded gather/transform/scatter in this mode.
type ScalingCPPCompactStats struct {
	ScalingCPPStats
	CompactionNS uint64 `json:"compaction_ns"`
	ExpansionNS  uint64 `json:"expansion_ns"`
}

// EncodeCompactWithStats explicitly selects lower-memory experimental execution.
// It compacts base values, scatters between disjoint arena halves, then expands
// backwards. On success it detaches the arena so later calls cannot alter old
// words. Replacement allocation and extra passes are included in caller latency.
func (e *ScalingCPPEncoder) EncodeCompactWithStats(message []uint32, w *ScalingCPPCompactWorkspace) ([]uint32, ScalingCPPCompactStats, error) {
	if e == nil {
		return nil, ScalingCPPCompactStats{}, ErrCPPClosed
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	var stats ScalingCPPCompactStats
	if e.closed || e.source == nil {
		return nil, stats, ErrCPPClosed
	}
	if w == nil || w.owner != e || len(message) != e.source.k {
		return nil, stats, fmt.Errorf("compact workspace/message mismatch")
	}
	p := e.source
	n := p.Length()
	l := 1
	for l < 2*p.k-1 {
		l *= 2
	}
	if w.arena == nil {
		w.arena = make([]uint32, n)
	}
	prefix, full := e.plans[l], e.plans[n]
	st, err := cpp.EncodeOnlineCompact(prefix.rows, prefix.columns, full.rows, full.columns, prefix.split, full.split, e.workers, p.basis.omega, message, p.p, p.spectrum, w.arena, w.columns, l)
	if err != nil {
		return nil, stats, err
	}
	word := w.arena
	w.arena = nil
	stats.ScalingCPPStats = ScalingCPPStats{Calls: 5, BoundaryCalls: 1, TaylorNS: st.TaylorNS, ColumnsNS: st.ColumnsNS, RowsNS: st.RowsNS, BufferNS: st.BufferNS, PointwiseNS: st.PointwiseNS, AssemblyNS: st.AssemblyNS}
	stats.CompactionNS = st.CompactionNS
	stats.ExpansionNS = st.ExpansionNS
	return word, stats, nil
}
