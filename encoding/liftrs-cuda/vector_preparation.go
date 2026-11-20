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

// VectorPreparation is an experimental single-GPU vector/Taylor stage owner. It owns a
// dedicated executor so vector storage cannot alias resident products/encoding.
// The global domain fixes metadata; localWords bounds one reusable device tile.
// Calls are serialized, synchronous and copy caller-owned interleaved tiles in
// and out. No global transpose or complete FFT is implicit; Taylor source-window
// updates are explicit and require caller-orchestrated phase synchronization.
// Ordinary CPU builds do not import CUDA, and older plugins remain usable for
// other Preparation operations; this constructor requires the optional stage ABI.
type VectorPreparation struct {
	engine           *Preparation
	domain, capacity uint64
}

// NewVectorPreparation creates metadata only; the first stage operation allocates the
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

// TaylorTile identifies a target interval within a 2*half block, where
// half=2^(floor(log2(domain)/2)+Level). First is relative to the selected half;
// quotient updates the upper half, remainder the lower. Count is target words.
type TaylorTile struct {
	Level               int
	Quotient            bool
	Block, First, Count uint64
}

// TaylorDegrees returns the sparse source displacements for one level. It can
// initialize the bounded device tile and host metadata, but no full grid. Source
// windows must follow this exact term order; coefficients remain in the GPU plan.
func (v *VectorPreparation) TaylorDegrees(level int) ([]uint64, error) {
	if v == nil || v.engine == nil || level < 0 {
		return nil, fmt.Errorf("GPU Taylor owner/level")
	}
	p := v.engine
	p.mu.Lock()
	defer p.mu.Unlock()
	var dimension int
	for n := v.domain; n > 1; n >>= 1 {
		dimension++
	}
	if p.bridge == nil || level >= dimension-dimension/2 {
		return nil, fmt.Errorf("GPU Taylor state/level")
	}
	var terms [21]C.uint64_t
	var count C.size_t
	var text [512]C.char
	code := C.pg_bridge_taylor_degrees(p.bridge, C.uint(p.width), C.size_t(v.domain), C.size_t(v.capacity), C.uint(level), &terms[0], 21, &count, &text[0], 512)
	if err := preparationError(code, &text[0]); err != nil {
		return nil, err
	}
	if count > 21 {
		return nil, fmt.Errorf("GPU Taylor term bound")
	}
	out := make([]uint64, int(count))
	for j := range out {
		out[j] = uint64(terms[j])
	}
	return out, nil
}
func (v *VectorPreparation) taylor(data unsafe.Pointer, words, width int, tile TaylorTile) error {
	if v == nil || v.engine == nil || data == nil || words <= 0 || tile.Level < 0 || tile.Count == 0 {
		return fmt.Errorf("GPU Taylor owner/tile")
	}
	p := v.engine
	p.mu.Lock()
	defer p.mu.Unlock()
	var dimension int
	for n := v.domain; n > 1; n >>= 1 {
		dimension++
	}
	if p.bridge == nil || p.width != width || uint64(words) > v.capacity || tile.Level >= dimension-dimension/2 || tile.Block >= v.domain || tile.First >= v.domain || tile.Count > v.capacity {
		return fmt.Errorf("GPU Taylor state/extent")
	}
	quotient := C.int(0)
	if tile.Quotient {
		quotient = 1
	}
	var text [512]C.char
	return preparationError(C.pg_bridge_taylor_stage(p.bridge, data, C.size_t(words), C.uint(width), C.size_t(v.domain), C.size_t(v.capacity), C.uint(tile.Level), quotient, C.size_t(tile.Block), C.size_t(tile.First), C.size_t(tile.Count), &text[0], 512), &text[0])
}

// Taylor32 updates a packed target/source snapshot using the shared recurrence.
// Layout is Count target words then Count source words per TaylorDegrees term;
// source windows extending beyond the block end must be zero-padded. Only the
// first Count words change. Other source words are preserved, not retained.
func (v *VectorPreparation) Taylor32(data []uint32, tile TaylorTile) error {
	err := v.taylor(preparationPointer(data), len(data), 4, tile)
	runtime.KeepAlive(data)
	return err
}

// Taylor64 is Taylor32 for eight-byte extension-field slots.
func (v *VectorPreparation) Taylor64(data []uint64, tile TaylorTile) error {
	err := v.taylor(preparationPointer(data), len(data), 8, tile)
	runtime.KeepAlive(data)
	return err
}

func (v *VectorPreparation) pointwise(left, right unsafe.Pointer, words, rightWords, width int) error {
	if v == nil || v.engine == nil || left == nil || right == nil || words <= 0 || words != rightWords {
		return fmt.Errorf("GPU stage product operands")
	}
	p := v.engine
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || p.width != width || uint64(words) > v.capacity {
		return fmt.Errorf("GPU stage product state/width/extent")
	}
	var text [512]C.char
	return preparationError(C.pg_bridge_stage_pointwise(p.bridge, left, right, C.size_t(words), C.uint(width), C.size_t(v.domain), C.size_t(v.capacity), &text[0], 512), &text[0])
}

// Pointwise32 replaces left with its pointwise extension-field product with right.
// Operands are disjoint with equal nonzero length <=localWords; right is preserved. This is
// spectrum multiplication, not polynomial multiplication. A second bounded
// device tile is allocated lazily; owned-memory caps include both tiles.
func (v *VectorPreparation) Pointwise32(left, right []uint32) error {
	err := v.pointwise(preparationPointer(left), preparationPointer(right), len(left), len(right), 4)
	runtime.KeepAlive(left)
	runtime.KeepAlive(right)
	return err
}

// Pointwise64 is Pointwise32 for eight-byte extension-field slots.
func (v *VectorPreparation) Pointwise64(left, right []uint64) error {
	err := v.pointwise(preparationPointer(left), preparationPointer(right), len(left), len(right), 8)
	runtime.KeepAlive(left)
	runtime.KeepAlive(right)
	return err
}
