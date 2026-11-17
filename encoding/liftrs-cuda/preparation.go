//go:build liftrs_cuda && cgo

package liftrscuda

/*
#cgo LDFLAGS: -ldl
#include "preparation_bridge.h"
*/
import "C"

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// PreparationOptions selects a resident server engine. Devices are visible CUDA
// ordinals; choose1/2/4/8 distinct GPUs on one host with all-pairs peer access.
// DeviceBytes is a PER-GPU explicit allocation cap, excluding CUDA context and
// peer mappings. MaxElements admits kernel sizes, never allocates the full grid.
// Tables and field identifiers must derive from trusted verified code arithmetic.
type PreparationOptions struct {
	Path                     string
	S                        int
	Polynomial, Delta        uint32
	Omega                    uint64
	Logs, Powers             []uint32
	Devices                  []int
	MaxElements, DeviceBytes uint64
}

// Preparation is a synchronous, serialized, width-selected GPU preprocessing
// executor. It owns device payloads and no borrowed Go pointer. Results remain
// Go-owned after Close. Do not copy a live engine. CUDA contexts are process-
// scoped. Peer access is leased across live engines: Close releases payloads and
// the last owned link lease, while leaving externally enabled mappings alone.
type Preparation struct {
	mu      sync.Mutex
	bridge  *C.pg_bridge
	width   int
	maximum uint64
	degree  int
	codeK   int
}

// NewPreparation loads the explicit trusted library and builds bounded metadata
// on the selected GPUs. Peer access and allocation errors are returned; no
// host-backed mode or CPU arithmetic fallback is selected.
func NewPreparation(options PreparationOptions) (*Preparation, error) {
	s := options.S
	count := len(options.Devices)
	if options.Path == "" || s < 2 || s > 17 || options.Polynomial>>uint(s) != 1 || options.Delta == 0 || options.Delta >= 1<<uint(s) || len(options.Logs) != 1<<uint(s) || len(options.Powers) != 2*((1<<uint(s))-1) || count == 0 || count > 8 || count&(count-1) != 0 || options.MaxElements < 4 || options.MaxElements&(options.MaxElements-1) != 0 || options.MaxElements > 1<<uint(2*s) || options.DeviceBytes == 0 || unsafe.Sizeof(int(0)) != 8 {
		return nil, fmt.Errorf("invalid resident GPU preprocessing options")
	}
	devices := make([]C.int, count)
	for i, id := range options.Devices {
		if id < 0 || id > 1<<31-1 {
			return nil, fmt.Errorf("GPU device ordinal")
		}
		for _, previous := range options.Devices[:i] {
			if id == previous {
				return nil, fmt.Errorf("duplicate GPU device")
			}
		}
		devices[i] = C.int(id)
	}
	path := C.CString(options.Path)
	defer C.free(unsafe.Pointer(path))
	var text [512]C.char
	handle := C.pg_bridge_open(path, C.uint(s), C.uint32_t(options.Polynomial), C.uint32_t(options.Delta), C.uint64_t(options.Omega), (*C.uint32_t)(unsafe.Pointer(&options.Logs[0])), C.size_t(len(options.Logs)), (*C.uint32_t)(unsafe.Pointer(&options.Powers[0])), C.size_t(len(options.Powers)), &devices[0], C.size_t(count), C.size_t(options.MaxElements), C.uint64_t(options.DeviceBytes), &text[0], 512)
	runtime.KeepAlive(options)
	runtime.KeepAlive(devices)
	if handle == nil {
		return nil, fmt.Errorf("resident GPU preprocessing: %s", C.GoString(&text[0]))
	}
	width := 4
	if s > 16 {
		width = 8
	}
	return &Preparation{bridge: handle, width: width, maximum: options.MaxElements, degree: options.S}, nil
}
func (p *Preparation) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge != nil {
		C.pg_bridge_close(p.bridge)
		p.bridge = nil
	}
	return nil
}

// OwnedBytes sums explicit allocations across all selected cards, not host RAM.
func (p *Preparation) OwnedBytes() uint64 {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return uint64(C.pg_bridge_owned(p.bridge))
}
func preparationError(code C.int, text *C.char) error {
	if code != 0 {
		return fmt.Errorf("GPU preprocessing: %s", C.GoString(text))
	}
	return nil
}
func preparationPointer[Word ~uint32 | ~uint64](data []Word) unsafe.Pointer {
	if len(data) == 0 {
		return nil
	}
	return unsafe.Pointer(&data[0])
}
func (p *Preparation) transform(data unsafe.Pointer, n, width int, inverse, grid bool) error {
	if p == nil || data == nil || n < 4 || n&(n-1) != 0 {
		return fmt.Errorf("GPU transform buffer/shape")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || p.width != width || uint64(n) > p.maximum {
		return fmt.Errorf("GPU preprocessing width/count/state")
	}
	var text [512]C.char
	var code C.int
	if grid {
		code = C.pg_bridge_grid(p.bridge, data, C.size_t(n), C.uint(width), &text[0], 512)
	} else {
		direction := C.int(0)
		if inverse {
			direction = 1
		}
		code = C.pg_bridge_apply(p.bridge, data, C.size_t(n), C.uint(width), direction, &text[0], 512)
	}
	return preparationError(code, &text[0])
}
func (p *Preparation) Apply32(data []uint32, inverse bool) error {
	err := p.transform(preparationPointer(data), len(data), 4, inverse, false)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) Apply64(data []uint64, inverse bool) error {
	err := p.transform(preparationPointer(data), len(data), 8, inverse, false)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) BaseGrid32(data []uint32) error {
	err := p.transform(preparationPointer(data), len(data), 4, false, true)
	runtime.KeepAlive(data)
	return err
}
func (p *Preparation) BaseGrid64(data []uint64) error {
	err := p.transform(preparationPointer(data), len(data), 8, false, true)
	runtime.KeepAlive(data)
	return err
}

// Product returns an owned uint32 base-field prefix. n covers the full product,
// even for truncation. Device operands/scratch are reused without host backing.
func (p *Preparation) Product(a, b []uint32, limit, n int) ([]uint32, error) {
	if p == nil {
		return nil, fmt.Errorf("nil GPU preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || n < 4 || n&(n-1) != 0 || uint64(n) > p.maximum || len(a) == 0 || len(b) == 0 || len(a) > n || len(b) > n || len(a) > n-len(b)+1 || limit < 1 || limit > len(a)+len(b)-1 {
		return nil, fmt.Errorf("GPU product shape/state")
	}
	out := make([]uint32, limit)
	var text [512]C.char
	code := C.pg_bridge_product(p.bridge, (*C.uint32_t)(preparationPointer(a)), C.size_t(len(a)), (*C.uint32_t)(preparationPointer(b)), C.size_t(len(b)), (*C.uint32_t)(preparationPointer(out)), C.size_t(limit), C.size_t(n), &text[0], 512)
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	runtime.KeepAlive(out)
	if err := preparationError(code, &text[0]); err != nil {
		return nil, err
	}
	return out, nil
}
func (p *Preparation) spectrum(g []uint32, output unsafe.Pointer, n, width int) error {
	if p == nil || len(g) == 0 || output == nil {
		return fmt.Errorf("GPU spectrum buffers")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || width != p.width || n < 4 || uint64(n) > p.maximum || len(g) > n {
		return fmt.Errorf("GPU spectrum shape/state")
	}
	var text [512]C.char
	code := C.pg_bridge_spectrum(p.bridge, (*C.uint32_t)(preparationPointer(g)), C.size_t(len(g)), output, C.size_t(n), C.uint(width), &text[0], 512)
	runtime.KeepAlive(g)
	return preparationError(code, &text[0])
}
func (p *Preparation) Spectrum32(g, out []uint32) error {
	err := p.spectrum(g, preparationPointer(out), len(out), 4)
	runtime.KeepAlive(out)
	return err
}
func (p *Preparation) Spectrum64(g []uint32, out []uint64) error {
	err := p.spectrum(g, preparationPointer(out), len(out), 8)
	runtime.KeepAlive(out)
	return err
}

// SpectrumTo writes the forward transform of zero-padded base coefficients as
// canonical little-endian extension slots. n must be a power of two within the
// admitted domain; use 1<<(2*s) for a complete generator spectrum. tileBytes is
// an explicit host payload cap (8 bytes..64 MiB), rounded down to an 8-byte boundary.
// It allocates one tile rather than an n-element host spectrum. Device workspace
// and transform count match resident execution; smaller transfers/writes may cost
// time. This explicit server prototype does not change preprocessing defaults.
// A writer failure leaves partial output: the caller must publish atomically only
// after success. The writer must not reenter this serialized engine. No pointer
// is retained after return; a later call can reuse the engine after an error.
func (p *Preparation) SpectrumTo(g []uint32, n int, destination io.Writer, tileBytes int) error {
	if p == nil || len(g) == 0 || destination == nil || tileBytes < 8 || tileBytes > 64<<20 {
		return fmt.Errorf("GPU streamed spectrum buffers/tile cap")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || n < 4 || n&(n-1) != 0 || uint64(n) > p.maximum || len(g) > n {
		return fmt.Errorf("GPU streamed spectrum shape/state")
	}
	var text [512]C.char
	status := C.pg_bridge_spectrum_begin(p.bridge, (*C.uint32_t)(preparationPointer(g)), C.size_t(len(g)), C.size_t(n), &text[0], 512)
	runtime.KeepAlive(g)
	if err := preparationError(status, &text[0]); err != nil {
		return err
	}
	// One aligned allocation is viewed as bytes only while converting/writing it.
	words := min(n, (tileBytes/8*8)/p.width)
	tile := make([]uint64, (words*p.width+7)/8)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(&tile[0])), words*p.width)
	began, last := time.Now(), time.Now()
	progress := func(event string, done int) {
		if os.Getenv("LIFTRS_CPU_PROGRESS") == "1" {
			_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"utc": time.Now().UTC().Format(time.RFC3339Nano), "phase": "GPU-spectrum-stream", "event": event, "completed": done, "total": n, "unit": "elements", "elapsed_seconds": time.Since(began).Seconds()})
		}
	}
	progress("started", 0)
	for first := 0; first < n; {
		take := min(words, n-first)
		status = C.pg_bridge_spectrum_read(p.bridge, unsafe.Pointer(&tile[0]), C.size_t(first), C.size_t(take), C.uint(p.width), &text[0], 512)
		runtime.KeepAlive(tile)
		if err := preparationError(status, &text[0]); err != nil {
			progress("failed", first)
			return err
		}
		// Convert host slots explicitly; canonical disk bytes are little-endian.
		for j := 0; j < take; j++ {
			if p.width == 4 {
				value := *(*uint32)(unsafe.Pointer(&raw[4*j]))
				binary.LittleEndian.PutUint32(raw[4*j:], value)
			} else {
				binary.LittleEndian.PutUint64(raw[8*j:], tile[j])
			}
		}
		block := raw[:take*p.width]
		written, err := destination.Write(block)
		if err == nil && written != len(block) {
			err = io.ErrShortWrite
		}
		if err != nil {
			progress("failed", first)
			return err
		}
		first += take
		if time.Since(last) >= 30*time.Second {
			progress("progress", first)
			last = time.Now()
		}
	}
	progress("completed", n)
	return nil
}

// RootFactorsInto fills normalized-input-independent[constant,linear,quadratic,
// degree] rows. Indices must be sorted/distinct; smaller partner rows stay zero.
func (p *Preparation) RootFactorsInto(indices []int, out [][4]uint32) error {
	if p == nil {
		return fmt.Errorf("nil GPU preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || len(indices) == 0 || len(indices) > 65536 || len(out) != len(indices) {
		return fmt.Errorf("GPU root batch shape/state")
	}
	var text [512]C.char
	code := C.pg_bridge_roots(p.bridge, (*C.int64_t)(unsafe.Pointer(&indices[0])), C.size_t(len(indices)), (*C.uint32_t)(unsafe.Pointer(&out[0])), &text[0], 512)
	runtime.KeepAlive(indices)
	runtime.KeepAlive(out)
	return preparationError(code, &text[0])
}

// GoodMonomialCount returns the exact count without allocating/transferring a
// host index array. Executor-owned membership buffers remain until reuse/Close.
func (p *Preparation) GoodMonomialCount(t int) (uint64, error) {
	if p == nil {
		return 0, fmt.Errorf("nil preparation")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.goodCountLocked(t)
}
func (p *Preparation) goodCountLocked(t int) (uint64, error) {
	if p.bridge == nil || t < 1 || t >= (1<<uint(p.degree))-1 {
		return 0, fmt.Errorf("GPU monomial state/degree")
	}
	var count C.uint64_t
	var text [512]C.char
	if err := preparationError(C.pg_bridge_good_count(p.bridge, C.uint(t), &count, &text[0], 512), &text[0]); err != nil {
		return 0, err
	}
	k := uint64(count)
	if k == 0 || k >= p.maximum {
		return 0, fmt.Errorf("invalid exact monomial count")
	}
	return k, nil
}

// GoodMonomials evaluates ternary weight classes and grid membership on GPUs,
// then streams the bitmap into a capped sorted Go index array. It is exact, not
// a sampled dimension estimate; the temporary bitmap is released after filling.
func (p *Preparation) GoodMonomials(t int, hostArrayBytes uint64) ([]int, error) {
	if p == nil {
		return nil, fmt.Errorf("nil GPU preprocessor")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	k, err := p.goodCountLocked(t)
	if err != nil {
		return nil, err
	}
	var text [512]C.char
	if k > uint64(^uint(0)>>1)/8 || k*8 > hostArrayBytes {
		return nil, fmt.Errorf("GPU monomial host-array cap")
	}
	out := make([]int, int(k))
	code := C.pg_bridge_good_fill(p.bridge, (*C.int64_t)(unsafe.Pointer(&out[0])), C.size_t(k), &text[0], 512)
	runtime.KeepAlive(out)
	if err := preparationError(code, &text[0]); err != nil {
		return nil, err
	}
	return out, nil
}

// ReleaseProducts drops dead product/root buffers, retaining transform data and
// transpose scratch for the final spectrum. It adds no transforms or transfers.
func (p *Preparation) ReleaseProducts() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil {
		return fmt.Errorf("GPU preprocessor closed")
	}
	var text [512]C.char
	return preparationError(C.pg_bridge_release(p.bridge, &text[0], 512), &text[0])
}

// ConfigureEncoding copies canonical P/spectrum to device once and configures a
// dedicated resident client. Spectrum slot width follows s. No Go pointer is
// retained. The same resident executor/kernel arithmetic serves preprocessing;
// this engine must have no prior workspace use and cannot be reconfigured.
func (p *Preparation) ConfigureEncoding(code, s32 []uint32, s64 []uint64, layout string) error {
	if p == nil {
		return fmt.Errorf("nil resident engine")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	n := int(p.maximum)
	mode := 0
	if layout == "local" {
		mode = 1
	} else if layout != "peer" {
		return fmt.Errorf("explicit peer/local vector layout required")
	}
	if p.bridge == nil || p.codeK != 0 || len(code) < 2 || len(code)-1 > n/2 || p.width == 4 && (len(s32) != n || len(s64) != 0) || p.width == 8 && (len(s64) != n || len(s32) != 0) {
		return fmt.Errorf("resident encoding code shape/state")
	}
	spectrum := preparationPointer(s32)
	if p.width == 8 {
		spectrum = preparationPointer(s64)
	}
	var text [512]C.char
	status := C.pg_bridge_set_code(p.bridge, (*C.uint32_t)(preparationPointer(code)), C.size_t(len(code)), spectrum, C.size_t(n), C.uint(p.width), C.int(mode), &text[0], 512)
	runtime.KeepAlive(code)
	runtime.KeepAlive(s32)
	runtime.KeepAlive(s64)
	if err := preparationError(status, &text[0]); err != nil {
		return err
	}
	p.codeK = len(code) - 1
	return nil
}

// EncodeCode borrows host input/output only for this synchronous complete call.
// Output must be disjoint from input. Constants and five transforms stay on the
// selected GPUs; stats are synchronized host-wall phase seconds (not CUDA events).
func (p *Preparation) EncodeCode(input, output []uint32) (stats [10]float64, err error) {
	if p == nil {
		return stats, fmt.Errorf("nil resident engine")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || p.codeK == 0 || len(input) != p.codeK || uint64(len(output)) != p.maximum {
		return stats, fmt.Errorf("resident encoding input/state")
	}
	var text [512]C.char
	status := C.pg_bridge_encode_code(p.bridge, (*C.uint32_t)(preparationPointer(input)), C.size_t(len(input)), (*C.uint32_t)(preparationPointer(output)), C.size_t(len(output)), (*C.double)(unsafe.Pointer(&stats[0])), &text[0], 512)
	runtime.KeepAlive(input)
	runtime.KeepAlive(output)
	return stats, preparationError(status, &text[0])
}
