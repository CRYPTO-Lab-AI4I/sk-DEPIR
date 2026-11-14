//go:build liftrs_cpp && cgo

package liftrscpp

/*
#include "preparation_api.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Preparation owns single-caller native public-code workspace. WordBytes is4
// through s16 and8 thereafter; messages/base polynomial coefficients stay32-bit.
// No Go pointer is retained. Close waits for current calls. Do not copy a handle.
type Preparation struct {
	mu        sync.Mutex
	handle    *C.liftrs_preparation
	wordBytes int
	maximum   uint64
}

// NewPreparation builds bounded metadata only. maxPoints and nativeBytes are
// mandatory admission limits; nativeBytes covers native payloads/metadata, not
// caller arrays, allocator/runtime/file cache. Full-domain discovery needs q²
// admitted points; small high-width kernel checks can admit fewer points.
func NewPreparation(s int, polynomial uint32, omega uint64, workers int, maxPoints, nativeBytes uint64) (*Preparation, error) {
	delta, err := wideDelta(s, polynomial)
	if err != nil {
		return nil, err
	}
	if workers < 1 || workers > 64 || maxPoints == 0 || maxPoints&(maxPoints-1) != 0 || maxPoints > 1<<uint(2*s) || maxPoints > uint64(^uint(0)>>1)/8 || nativeBytes == 0 || unsafe.Sizeof(int(0)) != 8 {
		return nil, fmt.Errorf("invalid native preparation limits/host")
	}
	var handle *C.liftrs_preparation
	var message [512]C.char
	code := C.liftrs_preparation_create(C.uint(s), C.uint32_t(polynomial), C.uint32_t(delta), C.uint64_t(omega), C.uint(workers), C.size_t(maxPoints), C.uint64_t(nativeBytes), &handle, &message[0], 512)
	if code != 0 {
		return nil, fmt.Errorf("native preparation: %s", C.GoString(&message[0]))
	}
	width := 4
	if s > 16 {
		width = 8
	}
	return &Preparation{handle: handle, wordBytes: width, maximum: maxPoints}, nil
}
func (p *Preparation) WordBytes() int {
	if p == nil {
		return 0
	}
	return p.wordBytes
}
func (p *Preparation) OwnedBytes() uint64 {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return uint64(C.liftrs_preparation_bytes(p.handle))
}
func (p *Preparation) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle != nil {
		C.liftrs_preparation_destroy(p.handle)
		p.handle = nil
	}
	return nil
}

// GoodMonomials enumerates the exact m2 good monomials once via ternary weight
// classes, then transfers sorted indices to an admitted host array. The temporary
// native bitset/table is released after filling; no full host duplicate exists.
func (p *Preparation) GoodMonomials(t int, maxHostArrayBytes uint64) ([]int, error) {
	if p == nil {
		return nil, fmt.Errorf("nil preparation")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil || t < 1 {
		return nil, fmt.Errorf("closed/invalid preparation")
	}
	var count C.uint64_t
	var text [512]C.char
	if C.liftrs_preparation_good_count(p.handle, C.uint(t), &count, &text[0], 512) != 0 {
		return nil, fmt.Errorf("good monomials: %s", C.GoString(&text[0]))
	}
	k := uint64(count)
	if k == 0 || k >= p.maximum || k > uint64(^uint(0)>>1)/8 || maxHostArrayBytes == 0 || k*8 > maxHostArrayBytes {
		return nil, fmt.Errorf("good monomial host count/byte cap")
	}
	indices := make([]int, int(k))
	code := C.liftrs_preparation_good_fill(p.handle, (*C.int64_t)(unsafe.Pointer(&indices[0])), C.size_t(k), &text[0], 512)
	runtime.KeepAlive(indices)
	if code != 0 {
		return nil, fmt.Errorf("good monomial indices: %s", C.GoString(&text[0]))
	}
	return indices, nil
}

// GoodMonomialRange checks bounded witnesses independently of enumeration tables.
func GoodMonomialRange(s, t int, first uint64, count, workers int) ([]byte, error) {
	if s < 2 || s > 20 || t < 1 || t >= (1<<uint(s))-1 || count < 1 || workers < 1 || workers > 64 || first >= 1<<uint(2*s) || uint64(count) > (1<<uint(2*s))-first {
		return nil, fmt.Errorf("invalid good monomial range")
	}
	out := make([]byte, count)
	var text [512]C.char
	code := C.liftrs_preparation_good_range(C.uint(s), C.uint(t), C.uint64_t(first), (*C.uchar)(unsafe.Pointer(&out[0])), C.size_t(count), C.uint(workers), &text[0], 512)
	runtime.KeepAlive(out)
	if code != 0 {
		return nil, fmt.Errorf("good monomial range: %s", C.GoString(&text[0]))
	}
	return out, nil
}

func (p *Preparation) transform(data unsafe.Pointer, n, width int, inverse, grid bool) error {
	if p == nil || data == nil || n < 1 {
		return fmt.Errorf("missing preparation data")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil || width != p.wordBytes || uint64(n) > p.maximum {
		return fmt.Errorf("preparation width/count/state")
	}
	var text [512]C.char
	var code C.int
	if grid {
		code = C.liftrs_preparation_grid(p.handle, data, C.size_t(n), C.uint(width), &text[0], 512)
	} else {
		flag := C.int(0)
		if inverse {
			flag = 1
		}
		code = C.liftrs_preparation_apply(p.handle, data, C.size_t(n), C.uint(width), flag, &text[0], 512)
	}
	if code != 0 {
		return fmt.Errorf("preparation transform: %s", C.GoString(&text[0]))
	}
	return nil
}
func (p *Preparation) Apply32(data []uint32, inverse bool) error {
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	}
	err := p.transform(ptr, len(data), 4, inverse, false)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) Apply64(data []uint64, inverse bool) error {
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	}
	err := p.transform(ptr, len(data), 8, inverse, false)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) BaseGrid32(data []uint32) error {
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	}
	err := p.transform(ptr, len(data), 4, false, true)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) BaseGrid64(data []uint64) error {
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = unsafe.Pointer(&data[0])
	}
	err := p.transform(ptr, len(data), 8, false, true)
	runtime.KeepAlive(data)
	return err
}

// Product returns a base-field prefix. n covers the FULL convolution, including
// when limit truncates it. Inputs are immutable; native workspace is reused.
func (p *Preparation) Product(a, b []uint32, limit, n int) ([]uint32, error) {
	if p == nil {
		return nil, fmt.Errorf("nil preparation")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil || len(a) == 0 || len(b) == 0 || limit < 1 || n < 1 || n&(n-1) != 0 || uint64(n) > p.maximum || len(a) > n || len(b) > n || len(a) > n-len(b)+1 || limit > len(a)+len(b)-1 {
		return nil, fmt.Errorf("preparation product shape/state")
	}
	out := make([]uint32, limit)
	var text [512]C.char
	code := C.liftrs_preparation_product(p.handle, widePointer32(a), C.size_t(len(a)), widePointer32(b), C.size_t(len(b)), widePointer32(out), C.size_t(limit), C.size_t(n), &text[0], 512)
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	runtime.KeepAlive(out)
	if code != 0 {
		return nil, fmt.Errorf("preparation product: %s", C.GoString(&text[0]))
	}
	return out, nil
}

// RootFactors returns bounded batches of[constant,linear,quadratic,degree].
// Rows whose q-Frobenius partner is smaller are zero/ignored. Go int must be64-bit.
func (p *Preparation) RootFactors(indices []int) ([][4]uint32, error) {
	if len(indices) < 1 || len(indices) > 65536 {
		return nil, fmt.Errorf("root batch shape")
	}
	out := make([][4]uint32, len(indices))
	if err := p.RootFactorsInto(indices, out); err != nil {
		return nil, err
	}
	return out, nil
}

// RootFactorsInto fills a reusable bounded caller batch, avoiding allocation per
// batch in large preparations. Output must have one row per sorted coordinate.
func (p *Preparation) RootFactorsInto(indices []int, out [][4]uint32) error {
	if p == nil {
		return fmt.Errorf("nil preparation")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil || len(indices) < 1 || len(indices) > 65536 || len(out) != len(indices) {
		return fmt.Errorf("root batch shape/state")
	}
	var text [512]C.char
	code := C.liftrs_preparation_roots(p.handle, (*C.int64_t)(unsafe.Pointer(&indices[0])), C.size_t(len(indices)), (*C.uint32_t)(unsafe.Pointer(&out[0])), &text[0], 512)
	runtime.KeepAlive(indices)
	runtime.KeepAlive(out)
	if code != 0 {
		return fmt.Errorf("preparation roots: %s", C.GoString(&text[0]))
	}
	return nil
}

// ReleaseProducts drops dead operands/column staging before allocating final
// spectrum output. Scratch/metadata remain reusable; there is no extra transform.
func (p *Preparation) ReleaseProducts() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	C.liftrs_preparation_release_products(p.handle)
}
func (p *Preparation) spectrum(g []uint32, output unsafe.Pointer, n, width int) error {
	if p == nil || len(g) == 0 || output == nil {
		return fmt.Errorf("spectrum buffers")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil || width != p.wordBytes || n < 1 || uint64(n) > p.maximum || len(g) > n {
		return fmt.Errorf("spectrum shape/state")
	}
	var text [512]C.char
	code := C.liftrs_preparation_spectrum(p.handle, widePointer32(g), C.size_t(len(g)), output, C.size_t(n), C.uint(width), &text[0], 512)
	runtime.KeepAlive(g)
	if code != 0 {
		return fmt.Errorf("preparation spectrum: %s", C.GoString(&text[0]))
	}
	return nil
}
func (p *Preparation) Spectrum32(g, out []uint32) error {
	var ptr unsafe.Pointer
	if len(out) > 0 {
		ptr = unsafe.Pointer(&out[0])
	}
	err := p.spectrum(g, ptr, len(out), 4)
	runtime.KeepAlive(out)
	return err
}
func (p *Preparation) Spectrum64(g []uint32, out []uint64) error {
	var ptr unsafe.Pointer
	if len(out) > 0 {
		ptr = unsafe.Pointer(&out[0])
	}
	err := p.spectrum(g, ptr, len(out), 8)
	runtime.KeepAlive(out)
	return err
}
