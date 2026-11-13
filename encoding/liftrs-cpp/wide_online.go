//go:build liftrs_cpp && cgo

package liftrscpp

/*
#include "wide_api.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// WideOnlineLayout describes principal caller allocations. Scratch becomes a
// uint32 output view but retains its8N-byte allocation. Caller-held older words,
// code data, input, native metadata/allocator overhead are additional live RAM.
type WideOnlineLayout struct {
	K, N, Prefix, Arena                                uint64
	WorkspaceBytes, OutputBackingBytes, TransientBytes uint64
}

// WideOnline owns immutable native metadata for the complete cyclic schedule,
// not code coefficients or buffers. It is experimental; data authenticity and
// code identities remain with verified preparation. Close waits for active calls.
// Do not copy a live handle; concurrent calls require distinct mutable buffers.
type WideOnline struct {
	mu     sync.RWMutex
	handle *C.liftrs_wide_online
	bytes  uint64
	layout WideOnlineLayout
}

// NewWideOnline prepares small prefix/full matrix plans and checks primitive
// omega. N=2^(2s),1<=K<=N/2. No domain buffers are allocated here; actual s17
// encoding still requires independently verified code data and admitted budgets.
func NewWideOnline(s int, polynomial uint32, omega, k uint64) (*WideOnline, error) {
	delta, err := wideDelta(s, polynomial)
	if err != nil {
		return nil, err
	}
	n := uint64(1) << uint(2*s)
	if k < 1 || k > n/2 || n > uint64(^uint(0)>>1)/16 {
		return nil, fmt.Errorf("invalid wide online shape/host width")
	}
	l := uint64(1)
	for l < 2*k-1 {
		l *= 2
	}
	arena := n
	if 2*l > arena {
		arena = 2 * l
	}
	var handle *C.liftrs_wide_online
	var message [512]C.char
	code := C.liftrs_wide_online_create(C.uint(s), C.uint32_t(polynomial), C.uint32_t(delta), C.uint64_t(omega), C.uint64_t(k), &handle, &message[0], 512)
	if code != 0 {
		return nil, fmt.Errorf("wide online: %s", C.GoString(&message[0]))
	}
	return &WideOnline{handle: handle, bytes: uint64(C.liftrs_wide_online_bytes(handle)),
		layout: WideOnlineLayout{K: k, N: n, Prefix: l, Arena: arena, WorkspaceBytes: 8*arena + 4*k + 8*n, OutputBackingBytes: 8 * n, TransientBytes: 4 * ((n - 1 + 4095) / 4096)}}, nil
}

// Layout returns immutable allocation requirements, also after Close.
func (e *WideOnline) Layout() WideOnlineLayout {
	if e == nil {
		return WideOnlineLayout{}
	}
	return e.layout
}

// OwnedBytes counts native metadata capacities only; zero after Close.
func (e *WideOnline) OwnedBytes() uint64 {
	if e == nil {
		return 0
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.bytes
}

// Close releases native metadata; nil and repeated calls are safe.
func (e *WideOnline) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.handle != nil {
		C.liftrs_wide_online_destroy(e.handle)
		e.handle = nil
		e.bytes = 0
	}
	return nil
}

// WideOnlineStats counts five transforms (two when K=1) in one native boundary.
// Native includes validation. Buffer includes input checks and preparation;
// Taylor/Columns/Rows, Pointwise and Assembly are disjoint stage clocks.
// OutputBackingBytes and TransientBytes are payload counts, not RSS bounds.
type WideOnlineStats struct {
	Native, Taylor, Columns, Rows, Buffer, Pointwise, Assembly time.Duration
	Calls, OutputBackingBytes, TransientBytes                  uint64
}

func widePointer32(a []uint32) *C.uint32_t {
	if len(a) == 0 {
		return nil
	}
	return (*C.uint32_t)(unsafe.Pointer(&a[0]))
}
func widePointer64(a []uint64) *C.uint64_t {
	if len(a) == 0 {
		return nil
	}
	return (*C.uint64_t)(unsafe.Pointer(&a[0]))
}

// EncodeBorrowed reads immutable message/P/spectrum data and borrows disjoint
// arena/H/scratch. Exactly one spectrum is present with canonical field values.
// The shared code-data adapter chooses the source's size-dependent slot width. The result aliases scratch's first4N bytes and has len/cap N; its backing
// allocation remains8N. To retain it, the caller MUST transfer ownership of
// scratch before any later reuse (the shared encoder does this automatically).
// Native relaxed atomic byte stores avoid aliasing/shared-location races; the pointer-free Go backing
// remains live through the returned view. Errors leave private work partial;
// shape/range errors precede workspace writes. No Go pointer is retained by C++.
func (e *WideOnline) EncodeBorrowed(message, p, spectrum32 []uint32, spectrum64, arena []uint64, h []uint32, scratch []uint64, workers int) ([]uint32, WideOnlineStats, error) {
	if e == nil {
		return nil, WideOnlineStats{}, fmt.Errorf("nil wide online encoder")
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.handle == nil {
		return nil, WideOnlineStats{}, fmt.Errorf("wide online closed")
	}
	q := e.layout
	if uint64(len(message)) != q.K || uint64(len(p)) != q.K+1 || uint64(len(arena)) != q.Arena || uint64(len(h)) != q.K || uint64(len(scratch)) != q.N || workers < 1 || workers > 64 ||
		!((uint64(len(spectrum32)) == q.N && len(spectrum64) == 0) || (uint64(len(spectrum64)) == q.N && len(spectrum32) == 0)) {
		return nil, WideOnlineStats{}, fmt.Errorf("wide online buffers/width/workers")
	}
	var stats [7]C.uint64_t
	var errorText [512]C.char
	code := C.liftrs_wide_online_encode(e.handle, C.uint(workers), widePointer32(message), widePointer32(p), widePointer32(spectrum32), widePointer64(spectrum64), widePointer64(arena), C.size_t(len(arena)), widePointer32(h), C.size_t(len(h)), widePointer64(scratch), C.size_t(len(scratch)), &stats[0], &errorText[0], 512)
	runtime.KeepAlive(message)
	runtime.KeepAlive(p)
	runtime.KeepAlive(spectrum32)
	runtime.KeepAlive(spectrum64)
	runtime.KeepAlive(arena)
	runtime.KeepAlive(h)
	runtime.KeepAlive(scratch)
	if code != 0 {
		return nil, WideOnlineStats{}, fmt.Errorf("wide online: %s", C.GoString(&errorText[0]))
	}
	// Converting to smaller pointer-free elements is confined to this final
	// ownership boundary. No uint64 access to the result backing may follow.
	word := unsafe.Slice((*uint32)(unsafe.Pointer(&scratch[0])), int(q.N))
	calls := uint64(5)
	if q.Prefix == 1 {
		calls = 2
	}
	return word, WideOnlineStats{Native: time.Duration(stats[0]), Taylor: time.Duration(stats[1]), Columns: time.Duration(stats[2]), Rows: time.Duration(stats[3]), Buffer: time.Duration(stats[4]), Pointwise: time.Duration(stats[5]), Assembly: time.Duration(stats[6]), Calls: calls, OutputBackingBytes: q.OutputBackingBytes, TransientBytes: q.TransientBytes}, nil
}
