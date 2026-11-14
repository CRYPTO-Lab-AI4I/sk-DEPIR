//go:build liftrs_cpp && cgo

// Package liftrscpp supplies the opt-in Go binding for immutable C++ transforms.
// Preprocessing descriptions are copied once; message buffers are borrowed only
// during execution. Matrix, online and disk bindings live in their corresponding
// files; both32/64-bit native encoders share the same mathematical schedule.
// Build with -tags liftrs_cpp and a C++17 compiler.
package liftrscpp

/*
#cgo CXXFLAGS: -std=c++17 -O3 -g -fno-omit-frame-pointer -pthread
#cgo LDFLAGS: -pthread
#include "c_api.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// Plan owns native metadata, not Go memory. Metadata is immutable after any
// explicit PrepareTaylor setup. Apply calls may run concurrently on distinct
// buffers; Close and one-time setup wait for active calls. Close is required.
type Plan struct {
	mu     sync.RWMutex
	handle *C.liftrs_plan
	bytes  uint64
}

// NewPlan copies a bounded trusted word description. Structural checks prevent
// unsafe shape/range access; they do not independently authenticate the code.
func NewPlan(words []uint32) (*Plan, error) {
	if len(words) == 0 {
		return nil, errors.New("empty C++ plan")
	}
	var handle *C.liftrs_plan
	var message [512]C.char
	code := C.liftrs_plan_create((*C.uint32_t)(unsafe.Pointer(&words[0])), C.size_t(len(words)), &handle, &message[0], C.size_t(len(message)))
	runtime.KeepAlive(words)
	if code != 0 {
		return nil, fmt.Errorf("C++ plan: %s", C.GoString(&message[0]))
	}
	return &Plan{handle: handle, bytes: uint64(C.liftrs_plan_owned_bytes(handle))}, nil
}

// Apply mutates canonical field elements in place. The pointer-free Go backing
// array is pinned for the cgo call and never retained by C++. Returned time is
// the native kernel interval, excluding Go/cgo dispatch and lock acquisition.
func (p *Plan) Apply(data []uint32, inverse bool) (time.Duration, error) {
	if p == nil || len(data) == 0 {
		return 0, errors.New("invalid C++ transform buffer")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.handle == nil {
		return 0, errors.New("C++ plan closed")
	}
	var message [512]C.char
	var elapsed C.uint64_t
	flag := C.int(0)
	if inverse {
		flag = 1
	}
	code := C.liftrs_plan_apply(p.handle, (*C.uint32_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), flag, &elapsed, &message[0], C.size_t(len(message)))
	runtime.KeepAlive(data)
	if code != 0 {
		return 0, fmt.Errorf("C++ transform: %s", C.GoString(&message[0]))
	}
	return time.Duration(elapsed), nil
}

// OwnedBytes counts native vector capacities and descriptors, excluding allocator
// bookkeeping and RSS. It is fixed after preparation and inspectable after Close.
func (p *Plan) OwnedBytes() uint64 { p.mu.RLock(); defer p.mu.RUnlock(); return p.bytes }

// Close releases the C++ plan; repeated calls are harmless. No input/output
// buffer is owned by the plan, so caller data remain valid after it closes.
func (p *Plan) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle != nil {
		C.liftrs_plan_destroy(p.handle)
		p.handle = nil
	}
	return nil
}

// ApplyAt evaluates/interpolates on the plan's ordered basis translated by offset.
// Offset is a packed field point. Plan metadata must describe that same field.
func (p *Plan) ApplyAt(data []uint32, offset uint32, inverse bool) (time.Duration, error) {
	if p == nil || len(data) == 0 {
		return 0, errors.New("invalid affine buffer")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.handle == nil {
		return 0, errors.New("C++ plan closed")
	}
	var message [512]C.char
	var elapsed C.uint64_t
	flag := C.int(0)
	if inverse {
		flag = 1
	}
	code := C.liftrs_plan_apply_at(p.handle, (*C.uint32_t)(unsafe.Pointer(&data[0])), C.size_t(len(data)), C.uint32_t(offset), flag, &elapsed, &message[0], C.size_t(len(message)))
	runtime.KeepAlive(data)
	if code != 0 {
		return 0, fmt.Errorf("C++ affine transform: %s", C.GoString(&message[0]))
	}
	return time.Duration(elapsed), nil
}
