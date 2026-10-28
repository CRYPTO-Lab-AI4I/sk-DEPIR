package liftrs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCPPUnavailable means the optional backend was not built. The default Go
// encoder does not need cgo or a C++ compiler; opt in with -tags liftrs_cpp.
var ErrCPPUnavailable = errors.New("C++ transforms unavailable")

// ErrCPPClosed indicates an encoding request after native resources were closed.
var ErrCPPClosed = errors.New("C++ encoder closed")

type cppPlan interface {
	Apply([]uint32, bool) (time.Duration, error)
	OwnedBytes() uint64
	Close() error
}
type transformExecutor interface{ apply([]uint32, bool) }
type cppFailure struct{ err error }

// CPPPlanInfo describes one-time copied code metadata, not database data or RSS.
// These immutable construction values remain inspectable after Close.
type CPPPlanInfo struct {
	DescriptionBytes uint64 `json:"description_bytes"`
	NativeOwnedBytes uint64 `json:"native_owned_bytes"`
}

// CPPEncodeStats describes whole-transform boundary calls during one encoding.
// CallTime includes native time, cgo dispatch, locks, validation and timing;
// NativeTime is the C++ kernel's steady-clock interval. Their difference is an
// observed boundary interval, not a calibrated measurement of cgo overhead.
type CPPEncodeStats struct {
	Calls      uint64        `json:"native_calls"`
	Elements   uint64        `json:"transformed_elements"`
	CallTime   time.Duration `json:"native_call_ns"`
	NativeTime time.Duration `json:"native_kernel_ns"`
	// MessageCopyBytes counts interface copies only. It is zero by construction;
	// the ordinary Go encoder's working-array copies are unchanged and excluded.
	MessageCopyBytes uint64 `json:"interface_message_copy_bytes"`
}

type meteredCPP struct {
	plan  cppPlan
	stats *CPPEncodeStats
}

func (m *meteredCPP) apply(data []uint32, inverse bool) {
	start := time.Now()
	native, err := m.plan.Apply(data, inverse)
	elapsed := time.Since(start)
	m.stats.Calls++
	m.stats.Elements += uint64(len(data))
	m.stats.CallTime += elapsed
	m.stats.NativeTime += native
	if err != nil {
		panic(cppFailure{err})
	}
}

// CPPCyclicEncoder is an experimental encoding adapter: Go retains all hybrid
// multiplication/truncation/blocking/overlap-add and cached-spectrum logic, while
// complete transforms run in C++. Plans are immutable; per-message scratch stays
// in Go. Close is required. Concurrent Encode calls use independent buffers and
// stats; Close waits for them before freeing native plans.
type CPPCyclicEncoder struct {
	mu     sync.RWMutex
	source *CyclicEncoder
	plan   cppPlan
	info   CPPPlanInfo
	closed bool
}

// WithCPPTransforms copies code-only transform metadata into one C++-owned plan.
// It shares immutable P/G/spectrum/support with e, without modifying e or changing
// its polynomial cutoffs. Preprocessing and artifact loading remain in Go. It
// does not add a spectrum when e lacks one, or silently change algorithm choices.
func (e *CyclicEncoder) WithCPPTransforms() (*CPPCyclicEncoder, error) {
	if e == nil || e.basis == nil {
		return nil, fmt.Errorf("uninitialized cyclic encoder")
	}
	plan, bytes, err := makeCPPPlan(e.basis.extensionFFT)
	if err != nil {
		return nil, err
	}
	return &CPPCyclicEncoder{source: e, plan: plan, info: CPPPlanInfo{bytes, plan.OwnedBytes()}}, nil
}

// Basis returns the original Go code description. Further preparation from it
// stays in Go and does not depend on this adapter's native resource lifetime.
func (e *CPPCyclicEncoder) Basis() *OrbitBasis { return e.source.Basis() }

// PlanInfo returns immutable initialization accounting, not current allocation.
func (e *CPPCyclicEncoder) PlanInfo() CPPPlanInfo { return e.info }

// Encode returns the same canonical codeword as the source encoder.
func (e *CPPCyclicEncoder) Encode(message []uint32) ([]uint32, error) {
	word, _, err := e.EncodeWithStats(message)
	return word, err
}

// EncodeWithStats encodes with per-call boundary accounting. Go data are borrowed
// by C++ only during each transform; no Go pointer or message buffer is retained.
func (e *CPPCyclicEncoder) EncodeWithStats(message []uint32) (word []uint32, stats CPPEncodeStats, err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.closed {
		return nil, stats, ErrCPPClosed
	}
	defer func() {
		if failure := recover(); failure != nil {
			if native, ok := failure.(cppFailure); ok {
				word = nil
				err = native.err
			} else {
				panic(failure)
			}
		}
	}()
	engine := e.source.withTransformExecutor(&meteredCPP{e.plan, &stats})
	word, err = engine.Encode(message)
	return
}

// withTransformExecutor copies only small orchestration structs. All immutable
// code arrays are shared; both polynomial FFT paths point to the same executor.
func (e *CyclicEncoder) withTransformExecutor(executor transformExecutor) *CyclicEncoder {
	b := *e.basis
	fft := *b.extensionFFT
	fft.executor = executor
	b.extensionFFT = &fft
	baseProducts := *b.base.products
	baseProducts.transform = &fft
	b.base.products = &baseProducts
	extensionProducts := *b.extension.products
	extensionProducts.transform = &fft
	b.extension.products = &extensionProducts
	return &CyclicEncoder{basis: &b, p: e.p, g: e.g, spectrum: e.spectrum}
}

// Close waits for active encodings and releases native metadata. It is idempotent
// and does not close the source Go encoder or invalidate any returned codewords.
func (e *CPPCyclicEncoder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	return e.plan.Close()
}

// cppDescription is a trusted in-process ABI description (not an artifact file).
// Every uint32 value is copied by C++ before this temporary slice is released.
func cppDescription(a *additiveTransform) []uint32 {
	delta := uint32(0)
	if a.f.extension != nil {
		delta = a.f.extension.Delta()
	}
	words := []uint32{1, uint32(a.f.base.Degree()), a.f.base.Polynomial(), delta, a.f.q, uint32(len(a.terms))}
	appendProduct := func(p constantProduct) {
		words = append(words, p.coefficient, uint32(len(p.rows)))
		words = append(words, p.rows...)
	}
	for k := range a.terms {
		words = append(words, uint32(len(a.terms[k])))
		for _, term := range a.terms[k] {
			words = append(words, uint32(term.degree))
			appendProduct(term.product)
		}
		words = append(words, a.values[k]...)
		words = append(words, a.delta[k], a.inverse[k])
		appendProduct(a.deltaProducts[k])
		appendProduct(a.inverseProducts[k])
	}
	return words
}
