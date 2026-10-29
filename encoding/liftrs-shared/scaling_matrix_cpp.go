//go:build liftrs_cpp && cgo

package liftrs

import (
	"fmt"
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/bits"
	"sync"
)

// orderedTransform constructs sparse S_k for an arbitrary independent basis.
// values still evaluates canonical packed offsets; the native executor advances
// children using basis[k], rather than confusing basis indices with field points.
func orderedTransform(f orbitField, basis []uint32) (*additiveTransform, error) {
	width := bits.Len32(f.q) - 1
	levels := len(basis)
	if levels < 1 || levels > width {
		return nil, fmt.Errorf("invalid ordered basis length")
	}
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{1}
	eval := func(x uint32) uint32 {
		var y uint32
		for j, c := range p {
			y ^= f.mul(c, f.pow(x, 1<<uint(j)))
		}
		return y
	}
	for k, x := range basis {
		if x >= f.q {
			return nil, fmt.Errorf("basis outside field")
		}
		for j, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{degree: 1 << uint(j), coefficient: c})
			}
		}
		a.values[k] = make([]uint32, width)
		for j := 0; j < width; j++ {
			a.values[k][j] = eval(1 << uint(j))
		}
		a.delta[k] = eval(x)
		if a.delta[k] == 0 {
			return nil, fmt.Errorf("dependent basis")
		}
		a.inverse[k] = f.inv(a.delta[k])
		next := make([]uint32, len(p)+1)
		for j, c := range p {
			next[j+1] ^= f.mul(c, c)
			next[j] ^= f.mul(a.delta[k], c)
		}
		p = next
	}
	return finishAdditiveTransform(a), nil
}
func orderedCPP(a *additiveTransform, basis []uint32) (*liftrscpp.Plan, error) {
	words := cppDescription(a)
	words[0] = 2
	description := append([]uint32(nil), words[:6]...)
	description = append(description, basis...)
	description = append(description, words[6:]...)
	return liftrscpp.NewPlan(description)
}

type matrixCPPPlan struct {
	rows, columns *liftrscpp.Plan
	split         int
}

func newMatrixCPP(f orbitField, dimension int) (*matrixCPPPlan, error) {
	split := dimension / 2
	low := make([]uint32, split+1)
	for j := range low {
		low[j] = 1 << uint(j)
	}
	h, err := orderedTransform(f, low)
	if err != nil {
		return nil, err
	}
	gamma := make([]uint32, dimension-split)
	for j := range gamma {
		for _, term := range h.terms[split] {
			gamma[j] ^= f.mul(term.coefficient, f.pow(1<<uint(split+j), term.degree))
		}
	}
	c, err := orderedTransform(f, gamma)
	if err != nil {
		return nil, err
	}
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
	e := &ScalingCPPEncoder{source: p, workers: workers, plans: make(map[int]*matrixCPPPlan)}
	l := 1
	for l < 2*p.k-1 {
		l *= 2
	}
	if uint64(l) > uint64(p.basis.extension.q) {
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
// belong to this encoder and may not be used by simultaneous calls.
func (e *ScalingCPPEncoder) NewWorkspace() *ScalingCPPWorkspace {
	return &ScalingCPPWorkspace{owner: e, work: e.source.NewWorkspace(), scratch: make([]uint32, e.source.basis.Length())}
}

// ScalingCPPWorkspace owns all mutable transform scratch; the output is separate.
type ScalingCPPWorkspace struct {
	owner   *ScalingCPPEncoder
	work    *ScalingWorkspace
	scratch []uint32
}

// Bytes counts working-array payload, excluding input, output and plan metadata.
func (w *ScalingCPPWorkspace) Bytes() uint64 { return w.work.Bytes() + uint64(len(w.scratch))*4 }

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
		m.stats.TaylorNS += stages.TaylorNS
		m.stats.ColumnsNS += stages.ColumnsNS
		m.stats.RowsNS += stages.RowsNS
	}
}

// Encode preserves canonical codewords and uses five matrix transforms. Taylor
// conversion, transposes and native thread scheduling are inside the measured call.
func (e *ScalingCPPEncoder) Encode(message []uint32, w *ScalingCPPWorkspace) ([]uint32, error) {
	word, _, err := e.EncodeWithStats(message, w)
	return word, err
}

// ScalingCPPStats accounts for native elapsed stages across five transforms.
// Columns includes transposes. Other encoding arithmetic/assembly remains Go.
type ScalingCPPStats struct {
	Calls     uint64 `json:"calls"`
	TaylorNS  uint64 `json:"taylor_ns"`
	ColumnsNS uint64 `json:"columns_transposes_ns"`
	RowsNS    uint64 `json:"rows_ns"`
}

// EncodeWithStats returns the same word plus per-call native stage accounting.
func (e *ScalingCPPEncoder) EncodeWithStats(message []uint32, w *ScalingCPPWorkspace) (word []uint32, stats ScalingCPPStats, err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
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
