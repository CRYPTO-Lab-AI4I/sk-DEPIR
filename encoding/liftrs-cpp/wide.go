//go:build liftrs_cpp && cgo

package liftrscpp

/*
#include "wide_api.h"
*/
import "C"

import (
	"fmt"
	commonfield "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

func wideDelta(s int, polynomial uint32) (uint32, error) {
	if s < 2 || s > 20 {
		return 0, fmt.Errorf("wide field requires s=2..20")
	}
	f, err := commonfield.NewBinaryField(s, polynomial)
	if err != nil {
		return 0, err
	}
	tower, err := commonfield.NewQuadraticField(f)
	if err != nil {
		return 0, err
	}
	return tower.Delta(), nil
}

// WidePlan owns immutable64-bit transform metadata in C++. It is experimental,
// separate from32-bit Plan/cache formats, and requires liftrs_cpp/cgo. Packed
// values are a|(b<<s) with32-bit base components. Do not copy a live handle.
// Concurrent calls may use disjoint data; Close waits for all active calls.
type WidePlan struct {
	mu     sync.RWMutex
	handle *C.liftrs_wide_plan
	bytes  uint64
}

// NewWidePlan validates the field and copies an ordered independent basis.
// It allocates bounded code-only metadata, never a cardinality-sized grid.
// Field validity is not evidence that supplied code data matches a PIR code.
func NewWidePlan(s int, polynomial uint32, basis []uint64) (*WidePlan, error) {
	delta, err := wideDelta(s, polynomial)
	if err != nil {
		return nil, err
	}
	if len(basis) < 1 || len(basis) > 2*s {
		return nil, fmt.Errorf("invalid wide basis length")
	}
	var handle *C.liftrs_wide_plan
	var message [512]C.char
	code := C.liftrs_wide_plan_create(C.uint(s), C.uint32_t(polynomial), C.uint32_t(delta), (*C.uint64_t)(unsafe.Pointer(&basis[0])), C.size_t(len(basis)), &handle, &message[0], 512)
	runtime.KeepAlive(basis)
	if code != 0 {
		return nil, fmt.Errorf("wide plan: %s", C.GoString(&message[0]))
	}
	return &WidePlan{handle: handle, bytes: uint64(C.liftrs_wide_plan_bytes(handle))}, nil
}

// ApplyAt evaluates/interpolates monomial coefficients on start plus the first
// log2(len(data)) basis directions. It borrows data for this call only and checks
// canonical values before mutation. Returned native time includes validation.
// Invalid shape/range is rejected before mutation; execution errors may leave
// caller-owned data partial. No buffer pointer is retained by C++.
func (p *WidePlan) ApplyAt(data []uint64, start uint64, inverse bool) (time.Duration, error) {
	if p == nil || len(data) == 0 {
		return 0, fmt.Errorf("invalid wide transform buffer")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.handle == nil {
		return 0, fmt.Errorf("wide plan closed")
	}
	flag := C.int(0)
	if inverse {
		flag = 1
	}
	var ns C.uint64_t
	var message [512]C.char
	code := C.liftrs_wide_plan_apply(p.handle, (*C.uint64_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.uint64_t(start), flag, &ns, &message[0], 512)
	runtime.KeepAlive(data)
	if code != 0 {
		return 0, fmt.Errorf("wide transform: %s", C.GoString(&message[0]))
	}
	return time.Duration(ns), nil
}

// OwnedBytes counts metadata/vector capacities, excluding allocator/RSS overhead.
// It returns zero after Close; it does not count caller buffers.
func (p *WidePlan) OwnedBytes() uint64 {
	if p == nil {
		return 0
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.bytes
}

// Close releases native metadata after active calls; repeated/nil Close is safe.
func (p *WidePlan) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle != nil {
		C.liftrs_wide_plan_destroy(p.handle)
		p.handle = nil
		p.bytes = 0
	}
	return nil
}

// WideMatrix owns immutable64-bit canonical-grid matrix/Taylor plans. It borrows
// two disjoint N-element buffers per Apply and retains neither pointer. Calls
// may run concurrently on distinct buffers; Close waits. Do not copy a live handle.
type WideMatrix struct {
	mu     sync.RWMutex
	handle *C.liftrs_wide_matrix
	bytes  uint64
	n      int
}

// NewWideMatrix prepares N=2^dimension points, with2^split row coefficients.
// Dimensions must satisfy1<=split<dimension<=2s. It builds metadata only; input
// and full-sized scratch allocation remain explicit responsibilities of callers.
func NewWideMatrix(s int, polynomial uint32, dimension, split int) (*WideMatrix, error) {
	delta, err := wideDelta(s, polynomial)
	if err != nil {
		return nil, err
	}
	if split < 1 || split >= dimension || dimension > 2*s || uint64(^uint(0)>>1) < uint64(1)<<uint(dimension) {
		return nil, fmt.Errorf("invalid wide matrix dimension/host width")
	}
	var handle *C.liftrs_wide_matrix
	var message [512]C.char
	code := C.liftrs_wide_matrix_create(C.uint(s), C.uint32_t(polynomial), C.uint32_t(delta), C.uint(dimension), C.uint(split), &handle, &message[0], 512)
	if code != 0 {
		return nil, fmt.Errorf("wide matrix: %s", C.GoString(&message[0]))
	}
	return &WideMatrix{handle: handle, bytes: uint64(C.liftrs_wide_matrix_bytes(handle)), n: 1 << uint(dimension)}, nil
}

// WideMatrixStats separates full native time (including input checks) from
// Taylor/column/row execution clocks. Setup, Go dispatch and caller allocation
// are outside these clocks and must be included in whole-encoding measurements.
type WideMatrixStats struct {
	Native, Taylor, Columns, Rows time.Duration
}

// Apply evaluates/interpolates exactly N canonical coefficients on the binary
// coordinate grid. workers=1..64; scratch must have N elements and not overlap
// data. Shape/range errors precede mutation; runtime errors may leave buffers
// partial. The immutable plans can be reused after errors with valid buffers.
func (m *WideMatrix) Apply(data, scratch []uint64, workers int, inverse bool) (WideMatrixStats, error) {
	if m == nil {
		return WideMatrixStats{}, fmt.Errorf("nil wide matrix")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.handle == nil {
		return WideMatrixStats{}, fmt.Errorf("wide matrix closed")
	}
	if len(data) != m.n || len(scratch) != m.n || workers < 1 || workers > 64 {
		return WideMatrixStats{}, fmt.Errorf("invalid wide matrix buffers/workers")
	}
	flag := C.int(0)
	if inverse {
		flag = 1
	}
	var ns C.uint64_t
	var stages [3]C.uint64_t
	var message [512]C.char
	code := C.liftrs_wide_matrix_apply(m.handle, (*C.uint64_t)(unsafe.Pointer(&data[0])), (*C.uint64_t)(unsafe.Pointer(&scratch[0])), C.size_t(m.n), C.uint(workers), flag, &ns, &stages[0], &message[0], 512)
	runtime.KeepAlive(data)
	runtime.KeepAlive(scratch)
	if code != 0 {
		return WideMatrixStats{}, fmt.Errorf("wide matrix: %s", C.GoString(&message[0]))
	}
	return WideMatrixStats{time.Duration(ns), time.Duration(stages[0]), time.Duration(stages[1]), time.Duration(stages[2])}, nil
}

// OwnedBytes counts immutable metadata capacities; zero after Close, no buffers.
func (m *WideMatrix) OwnedBytes() uint64 {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bytes
}

// Close waits for active calls and releases all matrix metadata; it is idempotent.
func (m *WideMatrix) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handle != nil {
		C.liftrs_wide_matrix_destroy(m.handle)
		m.handle = nil
		m.bytes = 0
	}
	return nil
}
