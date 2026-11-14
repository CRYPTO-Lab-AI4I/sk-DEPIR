//go:build liftrs_cpp && cgo

package liftrs

import (
	"fmt"
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/bits"
	"sync"
)

func orderedCPP(a *additiveTransform, basis []uint32) (*liftrscpp.Plan, error) {
	words := cppDescription(a)
	header := 6
	if words[0] == 3 {
		header = 7
	}
	words[0]++ // Canonical 1/3 -> ordered 2/4, with the same count width.
	description := append([]uint32(nil), words[:header]...)
	description = append(description, basis...)
	description = append(description, words[header:]...)
	return liftrscpp.NewPlan(description)
}

type matrixCPPPlan struct {
	rows, columns *liftrscpp.Plan
	split         int
}

func newMatrixCPP(f orbitField, dimension int) (*matrixCPPPlan, error) {
	m, err := buildMatrixBasis(f, dimension)
	if err != nil {
		return nil, err
	}
	split := m.split
	h, c, low, gamma := m.rows, m.columns, m.low, m.gamma
	r, err := orderedCPP(h, low)
	if err != nil {
		return nil, err
	}
	col, err := orderedCPP(c, gamma)
	if err != nil {
		r.Close()
		return nil, err
	}
	if err := r.PrepareTaylor(dimension - split); err != nil {
		r.Close()
		col.Close()
		return nil, err
	}
	return &matrixCPPPlan{r, col, split}, nil
}
func (p *matrixCPPPlan) close() { p.rows.Close(); p.columns.Close() }

// ScalingCPPEncoder is the opt-in parallel in-memory matrix path. Code arrays
// remain shared and immutable; each workspace adds N uint32 transpose scratch.
// It is not an external-memory encoder. Close waits for active Encode calls.
// workers is explicit (1..64), with no assumed universal hardware default.
type ScalingCPPEncoder struct {
	mu      sync.RWMutex
	source  *ScalingOnlinePlan
	plans   map[int]*matrixCPPPlan
	workers int
	closed  bool
}

// WithCPPMatrix prepares small immutable arbitrary-basis native plans for the
// full and prefix domains. P and spectrum are borrowed from the verified plan.
func (p *ScalingOnlinePlan) WithCPPMatrix(workers int) (*ScalingCPPEncoder, error) {
	if p == nil || workers < 1 || workers > 64 {
		return nil, fmt.Errorf("invalid matrix plan/workers")
	}
	l := 1
	for l < 2*p.k-1 {
		l *= 2
	}
	return newScalingCPPEncoder(p, workers, l)
}

// Both resident and file-backed code need the same immutable prefix/full plans.
// The caller selects prefix padding because historical file formats admit L>=4.
func newScalingCPPEncoder(p *ScalingOnlinePlan, workers, l int) (*ScalingCPPEncoder, error) {
	e := &ScalingCPPEncoder{source: p, workers: workers, plans: make(map[int]*matrixCPPPlan)}
	if uint64(l) > p.basis.extension.q {
		return nil, fmt.Errorf("prefix convolution exceeds matrix field domain")
	}
	for _, n := range []int{l, p.basis.Length()} {
		if e.plans[n] != nil {
			continue
		}
		v, err := newMatrixCPP(p.basis.extension, bits.Len(uint(n))-1)
		if err != nil {
			e.Close()
			return nil, err
		}
		e.plans[n] = v
	}
	return e, nil
}

// NewWorkspace allocates reusable convolution and transpose storage. Instances
// belong to this encoder and may not be used by simultaneous calls. Nil or
// closed encoders return nil without allocating.
func (e *ScalingCPPEncoder) NewWorkspace() *ScalingCPPWorkspace {
	if e == nil {
		return nil
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.source == nil {
		return nil
	}
	return &ScalingCPPWorkspace{owner: e, work: e.source.NewWorkspace(), scratch: make([]uint32, e.source.basis.Length())}
}

// ScalingCPPWorkspace owns transform scratch until complete encoding transfers it
// to the caller as output. Later calls allocate fresh scratch, preserving old words.
type ScalingCPPWorkspace struct {
	owner   *ScalingCPPEncoder
	work    *ScalingWorkspace
	scratch []uint32
}

// Bytes counts currently owned working arrays, excluding input/output and plan
// metadata. After complete Encode returns, transferred scratch is no longer owned.
func (w *ScalingCPPWorkspace) Bytes() uint64 {
	if w == nil || w.work == nil {
		return 0
	}
	return w.work.Bytes() + uint64(len(w.scratch))*4
}

type matrixCPPExecutor struct {
	encoder *ScalingCPPEncoder
	scratch []uint32
	stats   *ScalingCPPStats
}

func (m *matrixCPPExecutor) apply(data []uint32, inverse bool) {
	p := m.encoder.plans[len(data)]
	if p == nil {
		panic(cppFailure{fmt.Errorf("unprepared matrix length")})
	}
	var stages liftrscpp.MatrixStats
	if err := liftrscpp.ApplyMatrixWithStats(p.rows, p.columns, data, m.scratch[:len(data)], p.split, m.encoder.workers, inverse, &stages); err != nil {
		panic(cppFailure{err})
	}
	if m.stats != nil {
		m.stats.Calls++
		m.stats.BoundaryCalls++
		m.stats.TaylorNS += stages.TaylorNS
		m.stats.ColumnsNS += stages.ColumnsNS
		m.stats.RowsNS += stages.RowsNS
	}
}

// Encode runs the complete database-dependent path in one native call. Taylor
// conversion, transposes and native thread scheduling are inside the measured call.
func (e *ScalingCPPEncoder) Encode(message []uint32, w *ScalingCPPWorkspace) ([]uint32, error) {
	word, _, err := e.EncodeWithStats(message, w)
	return word, err
}

// ScalingCPPStats accounts for native elapsed stages across five transforms.
// Columns includes transposes. Additional stages cover non-transform work.
type ScalingCPPStats struct {
	Calls         uint64 `json:"calls"`
	BoundaryCalls uint64 `json:"boundary_calls"`
	BufferNS      uint64 `json:"buffer_ns"`
	PointwiseNS   uint64 `json:"pointwise_ns"`
	AssemblyNS    uint64 `json:"assembly_ns"`
	TaylorNS      uint64 `json:"taylor_ns"`
	ColumnsNS     uint64 `json:"columns_transposes_ns"`
	RowsNS        uint64 `json:"rows_ns"`
}

// EncodeWithStats reuses transpose scratch for canonical assembly, then transfers
// ownership to the caller. Later calls allocate fresh scratch; earlier returned
// words stay valid across workspace reuse and Close. Replacement allocation is
// included in caller timing;
// no large input, preprocessing or output copies cross the language boundary.
func (e *ScalingCPPEncoder) EncodeWithStats(message []uint32, w *ScalingCPPWorkspace) ([]uint32, ScalingCPPStats, error) {
	if e == nil {
		return nil, ScalingCPPStats{}, ErrCPPClosed
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	var stats ScalingCPPStats
	if e.closed || e.source == nil {
		return nil, stats, ErrCPPClosed
	}
	if w == nil || w.owner != e || len(message) != e.source.k {
		return nil, stats, fmt.Errorf("matrix workspace/message mismatch")
	}
	p := e.source
	l := len(w.work.x)
	prefix, full := e.plans[l], e.plans[p.basis.Length()]
	if w.scratch == nil {
		w.scratch = make([]uint32, p.basis.Length())
	}
	word := w.scratch
	// NewWorkspace owns one arena; x and z partition its first 2L elements.
	arena := w.work.x[:cap(w.work.x)]
	native, err := liftrscpp.EncodeOnline(prefix.rows, prefix.columns, full.rows, full.columns, prefix.split, full.split, e.workers, p.basis.omega,
		message, p.p, p.spectrum, arena, w.work.h, w.scratch, word, l)
	if err != nil {
		return nil, stats, err
	}
	// Scratch is dead after the inverse. Detach only on success; no subsequent
	// call may mutate the caller-owned canonical word through this workspace.
	w.scratch = nil
	stats = ScalingCPPStats{Calls: 5, BoundaryCalls: 1, TaylorNS: native.TaylorNS, ColumnsNS: native.ColumnsNS, RowsNS: native.RowsNS, BufferNS: native.BufferNS, PointwiseNS: native.PointwiseNS, AssemblyNS: native.AssemblyNS}
	return word, stats, nil
}

// EncodeHybridWithStats retains the earlier Go orchestration/scalar arithmetic
// path with five native transform calls for reference and measured comparisons.
func (e *ScalingCPPEncoder) EncodeHybridWithStats(message []uint32, w *ScalingCPPWorkspace) (word []uint32, stats ScalingCPPStats, err error) {
	if e == nil {
		return nil, stats, ErrCPPClosed
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed || e.source == nil {
		return nil, stats, ErrCPPClosed
	}
	if w == nil || w.owner != e {
		return nil, stats, fmt.Errorf("matrix workspace owner mismatch")
	}
	defer func() {
		if x := recover(); x != nil {
			if v, ok := x.(cppFailure); ok {
				word = nil
				err = v.err
			} else {
				panic(x)
			}
		}
	}()
	if w.scratch == nil {
		w.scratch = make([]uint32, e.source.Length())
	}
	p := *e.source
	b := p.basis
	fft := *b.extensionFFT
	fft.executor = &matrixCPPExecutor{e, w.scratch, &stats}
	b.extensionFFT = &fft
	p.basis = b
	// Only this stack-local workspace view changes ownership for the copied plan.
	work := *w.work
	work.owner = &p
	word, err = p.Encode(message, &work)
	return
}

// Close releases only native metadata. Shared Go preprocessing remains valid.
func (e *ScalingCPPEncoder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.closed {
		for _, p := range e.plans {
			p.close()
		}
		e.closed = true
	}
	return nil
}
