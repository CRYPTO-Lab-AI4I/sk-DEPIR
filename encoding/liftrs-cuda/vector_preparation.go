//go:build liftrs_cuda && cgo

package liftrscuda

/*
#cgo LDFLAGS: -ldl
#include "preparation_bridge.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// VectorPreparation is an experimental single-GPU server stage owner. It owns a
// dedicated executor so vector storage cannot alias resident products/encoding.
// The global domain fixes metadata; localWords bounds one reusable device tile.
// Calls are serialized, synchronous and copy caller-owned interleaved tiles in
// and out. This is not a complete FFT: no transpose or Taylor conversion occurs.
// Ordinary CPU builds do not import CUDA, and older plugins remain usable for
// other Preparation operations; this constructor requires the optional stage ABI.
type VectorPreparation struct {
	engine           *Preparation
	domain, capacity uint64
}

// NewVectorPreparation creates metadata only; the first Apply allocates the
// bounded tile. options.MaxElements is the global domain; exactly one GPU is
// required. localWords must hold a complete longest row/column vector. Include
// caller buffers, transfer packing and CUDA context separately in job admission.
func NewVectorPreparation(options PreparationOptions, localWords uint64) (*VectorPreparation, error) {
	domain := options.MaxElements
	var dimension uint
	for n := domain; n > 1; n >>= 1 {
		dimension++
	}
	if len(options.Devices) != 1 || localWords == 0 || localWords > domain || localWords < uint64(1)<<((dimension+1)/2) {
		return nil, fmt.Errorf("GPU vector stage capacity/device count")
	}
	p, err := NewPreparation(options)
	if err != nil {
		return nil, err
	}
	if C.pg_bridge_has_vector_stage(p.bridge) == 0 {
		p.Close()
		return nil, fmt.Errorf("GPU vector stage ABI unavailable")
	}
	return &VectorPreparation{engine: p, domain: domain, capacity: localWords}, nil
}

// Close releases the dedicated executor; repeated calls are safe.
func (v *VectorPreparation) Close() error {
	if v == nil {
		return nil
	}
	return v.engine.Close()
}

// OwnedBytes reports retained device payloads, excluding context/driver memory.
func (v *VectorPreparation) OwnedBytes() uint64 {
	if v == nil {
		return 0
	}
	return v.engine.OwnedBytes()
}
func (v *VectorPreparation) apply(data unsafe.Pointer, words, width int, columns, inverse bool, first, count uint64) error {
	if v == nil || v.engine == nil || data == nil || words <= 0 {
		return fmt.Errorf("nil GPU vector stage/tile")
	}
	p := v.engine
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || p.width != width || uint64(words) > v.capacity {
		return fmt.Errorf("GPU vector stage state/width/capacity")
	}
	var dimension uint
	for n := v.domain; n > 1; n >>= 1 {
		dimension++
	}
	length := uint64(1) << (dimension / 2)
	if columns {
		length = v.domain / length
	}
	jobs := v.domain / length
	if count == 0 || first > jobs || count > jobs-first || count > v.capacity/length || count*length != uint64(words) {
		return fmt.Errorf("GPU vector stage origin/extent")
	}
	axis, direction := C.int(0), C.int(0)
	if columns {
		axis = 1
	}
	if inverse {
		direction = 1
	}
	var text [512]C.char
	return preparationError(C.pg_bridge_vector_stage(p.bridge, data, C.size_t(words), C.uint(width), C.size_t(v.domain), C.size_t(v.capacity), axis, direction, C.size_t(first), C.size_t(count), &text[0], 512), &text[0])
}

// Apply32 transforms count complete vectors in [coefficient*count+vector] layout.
// first is their global row/column origin; it never offsets local storage. For
// columns the global affine point is zero, but the ownership range is checked.
func (v *VectorPreparation) Apply32(data []uint32, columns, inverse bool, first, count uint64) error {
	err := v.apply(preparationPointer(data), len(data), 4, columns, inverse, first, count)
	runtime.KeepAlive(data)
	return err
}

// Apply64 is Apply32 for s17's eight-byte extension-field slots.
func (v *VectorPreparation) Apply64(data []uint64, columns, inverse bool, first, count uint64) error {
	err := v.apply(preparationPointer(data), len(data), 8, columns, inverse, first, count)
	runtime.KeepAlive(data)
	return err
}
