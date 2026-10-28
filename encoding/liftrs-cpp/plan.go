//go:build liftrs_cpp && cgo

// Package liftrscpp supplies the opt-in Go binding for immutable C++ transforms.
// Preprocessing descriptions are copied once; message buffers are borrowed only
// during Apply. Build with -tags liftrs_cpp and a C++17 compiler.
package liftrscpp

/*
#cgo CXXFLAGS: -std=c++17 -O3 -g -fno-omit-frame-pointer
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

// Plan owns immutable native metadata, not Go memory. Close is required. Apply
// calls may run concurrently on distinct buffers; Close waits for active calls.
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
// bookkeeping and RSS. It is immutable and remains inspectable after Close.
func (p *Plan) OwnedBytes() uint64 { return p.bytes }

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
