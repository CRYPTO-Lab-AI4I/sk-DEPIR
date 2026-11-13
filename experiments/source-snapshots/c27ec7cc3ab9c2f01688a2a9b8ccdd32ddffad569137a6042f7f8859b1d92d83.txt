//go:build liftrs_cpp && cgo

// Package liftrscpp supplies the opt-in Go binding for immutable C++ transforms.
// Preprocessing descriptions are copied once; message buffers are borrowed only
// during Apply. Build with -tags liftrs_cpp and a C++17 compiler.
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

// ApplyMatrix runs the canonical matrix decomposition with independent native
// workers and blocked transposes. rows must include S_split and columns must use
// its image basis. scratch owns at least len(data) disjoint elements and can be
// reused after return. No message pointers are retained; Close waits for this call.
// Partial overlap is rejected, including slices into the same backing array.
func ApplyMatrix(rows, columns *Plan, data, scratch []uint32, split, workers int, inverse bool) error {
	return applyMatrix(rows, columns, data, scratch, split, workers, inverse, nil)
}

// MatrixStats separates native Taylor conversion, columns (including both blocked
// transposes), and affine row transforms. These are elapsed intervals, not CPU time.
type MatrixStats struct {
	TaylorNS  uint64 `json:"taylor_ns"`
	ColumnsNS uint64 `json:"columns_transposes_ns"`
	RowsNS    uint64 `json:"rows_ns"`
}

// ApplyMatrixWithStats has the same ownership rules as ApplyMatrix and records
// native elapsed stage intervals. stats must be private to this call.
func ApplyMatrixWithStats(rows, columns *Plan, data, scratch []uint32, split, workers int, inverse bool, stats *MatrixStats) error {
	return applyMatrix(rows, columns, data, scratch, split, workers, inverse, stats)
}
func applyMatrix(rows, columns *Plan, data, scratch []uint32, split, workers int, inverse bool, stats *MatrixStats) error {
	if rows == nil || columns == nil || rows == columns || len(data) == 0 || len(scratch) < len(data) || split < 1 || workers < 1 || workers > 64 {
		return errors.New("invalid matrix request")
	}
	start := uintptr(unsafe.Pointer(&data[0]))
	other := uintptr(unsafe.Pointer(&scratch[0]))
	size := uintptr(len(data)) * 4
	if start < other+size && other < start+size {
		return errors.New("overlapping matrix scratch")
	}
	rows.mu.RLock()
	defer rows.mu.RUnlock()
	columns.mu.RLock()
	defer columns.mu.RUnlock()
	if rows.handle == nil || columns.handle == nil {
		return errors.New("matrix plan closed")
	}
	var message [512]C.char
	flag := C.int(0)
	if inverse {
		flag = 1
	}
	var stages [3]C.uint64_t
	code := C.liftrs_matrix_apply(rows.handle, columns.handle, (*C.uint32_t)(unsafe.Pointer(&data[0])), (*C.uint32_t)(unsafe.Pointer(&scratch[0])), C.size_t(len(data)), C.uint(split), C.uint(workers), flag, &stages[0], &message[0], C.size_t(len(message)))
	runtime.KeepAlive(data)
	runtime.KeepAlive(scratch)
	if code != 0 {
		return fmt.Errorf("C++ matrix transform: %s", C.GoString(&message[0]))
	}
	if stats != nil {
		stats.TaylorNS = uint64(stages[0])
		stats.ColumnsNS = uint64(stages[1])
		stats.RowsNS = uint64(stages[2])
	}
	return nil
}

// PrepareTaylor constructs code-only quotient maps once, before matrix execution.
// It takes an exclusive lock; repeated calls are rejected. Plans used only for
// ordinary/affine transforms need no Taylor preparation.
func (p *Plan) PrepareTaylor(highLevels int) error {
	if p == nil || highLevels < 1 || highLevels > 26 {
		return errors.New("invalid Taylor levels")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.handle == nil {
		return errors.New("C++ plan closed")
	}
	var message [512]C.char
	if C.liftrs_plan_prepare_taylor(p.handle, C.uint(highLevels), &message[0], C.size_t(len(message))) != 0 {
		return fmt.Errorf("C++ Taylor preparation: %s", C.GoString(&message[0]))
	}
	p.bytes = uint64(C.liftrs_plan_owned_bytes(p.handle))
	return nil
}

// OnlineStats records elapsed native stages, excluding allocation and cgo entry.
// Matrix stages aggregate all five transforms; products and assembly are parallel.
type OnlineStats struct {
	MatrixStats
	BufferNS    uint64
	PointwiseNS uint64
	AssemblyNS  uint64
}

// EncodeOnline executes complete cyclic codeword encoding in one cgo call.
// P/spectrum and message are immutable. Writable arrays are disjoint except that
// word may exactly equal scratch: assembly starts after its last transform use.
// Partial overlap is rejected. arena contains intentional prefix/full aliasing.
// All backing arrays are borrowed only
// until return; no native plan retains Go pointers or copies large constants.
// The caller supplies trusted code parameters and canonical transform plans.
func EncodeOnline(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, p, spectrum, arena, h, scratch, word []uint32, prefixLength int) (OnlineStats, error) {
	result, err := encodeOnline(pr, pc, fr, fc, prefixSplit, fullSplit, workers, omega, message, p, spectrum, arena, h, scratch, word, prefixLength, false, nil, false)
	return result.OnlineStats, err
}

// CompactOnlineStats adds narrowing/move and expansion timings to the normal
// native stages. It is experimental; the extra work is not assumed cost-free.
type CompactOnlineStats struct {
	OnlineStats
	CompactionNS, ExpansionNS uint64
}

// BoundedColumnElements counts experimental column workspace in uint32 slots.
// It uses up to 64 columns per worker, matching the native bounded batch contract.
func BoundedColumnElements(count, split, workers int) (int, error) {
	if count < 4 || uint64(count) > 1<<32 || count&(count-1) != 0 || split < 1 || split >= 63 || workers < 1 || workers > 64 {
		return 0, errors.New("invalid bounded column shape")
	}
	n1 := 1 << uint(split)
	if n1 >= count {
		return 0, errors.New("invalid bounded split")
	}
	width := min(64, n1)
	active := min(workers, n1/width)
	return active * width * (count / n1), nil
}

// EncodeOnlineCompact explicitly borrows one arena as both working data and
// owning-result storage. Output equals arena; no h is needed. All other arrays
// must be disjoint; bounded scratch covers both domains. Requires 2L<=N and s<=16.
func EncodeOnlineCompact(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, p, spectrum, arena, columns []uint32, prefixLength int) (CompactOnlineStats, error) {
	if len(arena) != len(spectrum) {
		return CompactOnlineStats{}, errors.New("compact arena length")
	}
	r, err := encodeOnline(pr, pc, fr, fc, prefixSplit, fullSplit, workers, omega, message, p, spectrum, arena, nil, columns, arena, prefixLength, true, nil, false)
	return r.CompactOnlineStats, err
}

// OnlineFile describes immutable verified little-endian code on a borrowed file
// descriptor. Reads borrow dead transpose scratch; no file pointer is retained.
type OnlineFile struct {
	Descriptor              int
	POffset, SpectrumOffset uint64
	TileElements            int
}

// FileOnlineStats separates logical read/range-check time and bytes. Prefix read
// time also appears in BufferNS; do not add overlapping counters to total.
type FileOnlineStats struct {
	OnlineStats
	ReadNS, ReadBytes, ReadChunks uint64
}
type onlineExecutionStats struct {
	CompactOnlineStats
	ReadNS, ReadBytes, ReadChunks uint64
}

// EncodeOnlineFile selects explicit fast file streaming, 2L<=N. Arena and scratch
// are full N arrays; scratch is transferred as output only after the inverse.
func EncodeOnlineFile(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, arena, scratch []uint32, prefixLength int, file OnlineFile) (FileOnlineStats, error) {
	r, err := encodeOnline(pr, pc, fr, fc, prefixSplit, fullSplit, workers, omega, message, nil, nil, arena, nil, scratch, scratch, prefixLength, false, &file, false)
	return FileOnlineStats{r.OnlineStats, r.ReadNS, r.ReadBytes, r.ReadChunks}, err
}

// EncodeOnlineFilePacked selects direct packed canonical scatter into dead scratch.
// Its first ceil(N*s/8) bytes are the little-endian s-bit result; the remaining
// scratch bytes are unspecified. Both N-word arrays remain required by transforms.
func EncodeOnlineFilePacked(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, arena, scratch []uint32, prefixLength int, file OnlineFile) (FileOnlineStats, error) {
	r, err := encodeOnline(pr, pc, fr, fc, prefixSplit, fullSplit, workers, omega, message, nil, nil, arena, nil, scratch, scratch, prefixLength, false, &file, true)
	return FileOnlineStats{r.OnlineStats, r.ReadNS, r.ReadBytes, r.ReadChunks}, err
}

func encodeOnline(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, p, spectrum, arena, h, scratch, word []uint32, prefixLength int, compact bool, file *OnlineFile, packed bool) (onlineExecutionStats, error) {
	var result onlineExecutionStats
	if packed && (file == nil || compact) {
		return result, errors.New("packed output requires fast file schedule")
	}
	k, n, l := len(message), len(spectrum), prefixLength
	if file != nil {
		n = len(word)
	}
	if pr == nil || pc == nil || fr == nil || fc == nil || k == 0 || k > n/2 || (file == nil && len(p) != k+1) ||
		l < 2*k-1 || l > n || l&(l-1) != 0 || len(arena) < max(2*l, n) || (!compact && ((file == nil && len(h) != k) || len(scratch) != n)) || len(scratch) == 0 || len(word) != n || (compact && (len(h) != 0 || len(arena) != n || l > n/2)) ||
		prefixSplit < 1 || fullSplit < 1 || workers < 1 || workers > 64 {
		return result, errors.New("invalid online arrays/plans")
	}
	if file != nil && (compact || file.Descriptor < 0 || uint64(file.Descriptor) > 1<<31-1 || file.TileElements < 1 || file.TileElements > min(n, 1<<26) || l > n/2 || len(arena) != n || len(h) != 0 || file.POffset > uint64(1<<63-1)-uint64(k+1)*4 || file.SpectrumOffset > uint64(1<<63-1)-uint64(n)*4) {
		return result, errors.New("invalid file encoding shape/offsets")
	}
	if compact {
		pre, err := BoundedColumnElements(l, prefixSplit, workers)
		if err != nil {
			return result, err
		}
		full, err := BoundedColumnElements(n, fullSplit, workers)
		if err != nil {
			return result, err
		}
		if len(scratch) < max(pre, full) {
			return result, errors.New("short bounded column workspace")
		}
	}
	arrays := [][]uint32{message, p, spectrum, arena, h, scratch, word}
	// Constants may share backing storage with each other. All writable arrays
	// must be disjoint, except exact scratch/word identity after the final transform.
	for i := 3; i < len(arrays); i++ {
		if len(arrays[i]) == 0 {
			continue
		}
		a := uintptr(unsafe.Pointer(&arrays[i][0]))
		size := uintptr(len(arrays[i])) * 4
		for j := 0; j < i; j++ {
			if len(arrays[j]) == 0 {
				continue
			}
			b := uintptr(unsafe.Pointer(&arrays[j][0]))
			other := uintptr(len(arrays[j])) * 4
			if i == 6 && ((!compact && j == 5) || (compact && j == 3)) && a == b {
				continue // Exact permitted n-sized identity; schedule ensures safe reuse.
			}
			if a < b+other && b < a+size {
				return result, errors.New("overlapping online arrays")
			}
		}
	}
	// Prefix/full domains can share a plan. Lock each distinct handle once so a
	// waiting Close cannot deadlock a recursive read-lock acquisition.
	seen := map[*Plan]bool{}
	for _, plan := range []*Plan{pr, pc, fr, fc} {
		if !seen[plan] {
			seen[plan] = true
			plan.mu.RLock()
			defer plan.mu.RUnlock()
		}
		if plan.handle == nil {
			return result, errors.New("online plan closed")
		}
	}
	ptr := func(a []uint32) *C.uint32_t {
		if len(a) == 0 {
			return nil
		}
		return (*C.uint32_t)(unsafe.Pointer(&a[0]))
	}
	var stages [11]C.uint64_t
	var errorText [512]C.char
	var code C.int
	if file != nil && packed {
		code = C.liftrs_online_encode_file_packed(C.int(file.Descriptor), C.uint64_t(file.POffset), C.uint64_t(file.SpectrumOffset), C.size_t(file.TileElements), pr.handle, pc.handle, fr.handle, fc.handle, C.uint(prefixSplit), C.uint(fullSplit), C.uint(workers), C.uint32_t(omega), ptr(message), C.size_t(k), C.size_t(n), C.size_t(l), ptr(arena), ptr(scratch), &stages[0], &errorText[0], C.size_t(len(errorText)))
	} else if file != nil {
		code = C.liftrs_online_encode_file(C.int(file.Descriptor), C.uint64_t(file.POffset), C.uint64_t(file.SpectrumOffset), C.size_t(file.TileElements), pr.handle, pc.handle, fr.handle, fc.handle, C.uint(prefixSplit), C.uint(fullSplit), C.uint(workers), C.uint32_t(omega), ptr(message), C.size_t(k), C.size_t(n), C.size_t(l), ptr(arena), ptr(scratch), &stages[0], &errorText[0], C.size_t(len(errorText)))
	} else if compact {
		code = C.liftrs_online_encode_compact(C.size_t(len(scratch)), pr.handle, pc.handle, fr.handle, fc.handle, C.uint(prefixSplit), C.uint(fullSplit), C.uint(workers), C.uint32_t(omega),
			ptr(message), ptr(p), ptr(spectrum), C.size_t(k), C.size_t(n), C.size_t(l), ptr(arena), C.size_t(len(arena)), ptr(h), ptr(scratch), ptr(word), &stages[0], &errorText[0], C.size_t(len(errorText)))
	} else {
		code = C.liftrs_online_encode(pr.handle, pc.handle, fr.handle, fc.handle, C.uint(prefixSplit), C.uint(fullSplit), C.uint(workers), C.uint32_t(omega),
			ptr(message), ptr(p), ptr(spectrum), C.size_t(k), C.size_t(n), C.size_t(l), ptr(arena), C.size_t(len(arena)), ptr(h), ptr(scratch), ptr(word), &stages[0], &errorText[0], C.size_t(len(errorText)))
	}

	runtime.KeepAlive(arrays)
	if code != 0 {
		return result, fmt.Errorf("C++ online encoding: %s", C.GoString(&errorText[0]))
	}
	result.CompactOnlineStats = CompactOnlineStats{OnlineStats: OnlineStats{MatrixStats{uint64(stages[0]), uint64(stages[1]), uint64(stages[2])}, uint64(stages[3]), uint64(stages[4]), uint64(stages[5])}, CompactionNS: uint64(stages[6]), ExpansionNS: uint64(stages[7])}
	result.ReadNS = uint64(stages[8])
	result.ReadBytes = uint64(stages[9])
	result.ReadChunks = uint64(stages[10])
	return result, nil
}

// DiskStats separates transform phases and nested logical file I/O. ReadNS and
// WriteNS overlap phase times; bytes/syscalls are not physical disk traffic.
// BufferBytes excludes immutable plan metadata, runtime/thread stacks and OS cache.
type DiskStats struct {
	TaylorNS    uint64 `json:"taylor_ns"`
	ColumnsNS   uint64 `json:"columns_ns"`
	RowsNS      uint64 `json:"rows_ns"`
	ReadNS      uint64 `json:"read_ns"`
	WriteNS     uint64 `json:"write_ns"`
	ReadBytes   uint64 `json:"read_bytes"`
	WriteBytes  uint64 `json:"write_bytes"`
	ReadCalls   uint64 `json:"read_calls"`
	WriteCalls  uint64 `json:"write_calls"`
	BufferBytes uint64 `json:"buffer_bytes"`
	TotalNS     uint64 `json:"total_ns"`
	Transposes  uint64 `json:"transposes"`
}

// ApplyMatrixFile explicitly transforms an exclusive caller-owned little-endian
// file array using a distinct pre-sized scratch file and two bounded RAM tiles.
// Tile must fit a complete row/column. File cache can use additional RAM up to
// the backing files; this is a private-buffer bound, not a total-memory bound.
// The caller must keep descriptors open and exclude concurrent access/truncation.
// No descriptor/pointer is retained, no fsync is performed, and errors may leave
// both files partially modified. Restore verified input before retrying.
func ApplyMatrixFile(rows, columns *Plan, dataFile, scratchFile int, dataOffset, scratchOffset uint64, count, tile, split, workers int, inverse bool) (DiskStats, error) {
	var result DiskStats
	if rows == nil || columns == nil || dataFile < 0 || scratchFile < 0 || uint64(dataFile) > 1<<31-1 || uint64(scratchFile) > 1<<31-1 || count < 4 || count&(count-1) != 0 || tile < 4 || tile > 1<<26 || split < 1 || split > 32 || workers < 1 || workers > 64 {
		return result, errors.New("disk transform arguments")
	}
	rows.mu.RLock()
	defer rows.mu.RUnlock()
	if columns != rows {
		columns.mu.RLock()
		defer columns.mu.RUnlock()
	}
	if rows.handle == nil || columns.handle == nil {
		return result, errors.New("disk transform closed plan")
	}
	flag := 0
	if inverse {
		flag = 1
	}
	var values [12]C.uint64_t
	var detail [512]C.char
	code := C.liftrs_matrix_apply_file(rows.handle, columns.handle, C.int(dataFile), C.int(scratchFile), C.uint64_t(dataOffset), C.uint64_t(scratchOffset), C.size_t(count), C.size_t(tile), C.uint(split), C.uint(workers), C.int(flag), &values[0], &detail[0], 512)
	result = DiskStats{uint64(values[0]), uint64(values[1]), uint64(values[2]), uint64(values[3]), uint64(values[4]), uint64(values[5]), uint64(values[6]), uint64(values[7]), uint64(values[8]), uint64(values[9]), uint64(values[10]), uint64(values[11])}
	if code != 0 {
		return result, fmt.Errorf("C++ disk transform: %s", C.GoString(&detail[0]))
	}
	return result, nil
}

// DiskOnlineStats counts nested logical I/O and conservative private buffers.
// The owning N-word host output, message, plan metadata, runtime and OS cache
// are additional. I/O times overlap phase timings. No fsync is included.
type DiskOnlineStats struct {
	OnlineStats
	ReadNS, WriteNS, ReadBytes, WriteBytes, ReadCalls, WriteCalls uint64
	BufferBytes, TotalNS, Transforms, BackingBytes                uint64
}

// EncodeOnlineDisk executes five transforms with two distinct workspace files
// and streamed immutable code. Callers keep all descriptors open/exclusive;
// error paths can leave both workspace files/output partially modified.
// Message and owning output must be disjoint. No pointer/descriptor is retained.
func EncodeOnlineDisk(pr, pc, fr, fc *Plan, prefixSplit, fullSplit, workers int, omega uint32, message, word []uint32, prefixLength, tile, dataFile, scratchFile int, codeFile OnlineFile) (DiskOnlineStats, error) {
	var result DiskOnlineStats
	k, n, l := len(message), len(word), prefixLength
	if pr == nil || pc == nil || fr == nil || fc == nil || k < 1 || k > n/2 || l < 2*k-1 || l > n/2 || l&(l-1) != 0 || prefixSplit < 1 || prefixSplit > 32 || fullSplit < 1 || fullSplit > 32 || workers < 1 || workers > 64 || tile < 4 || tile > 1<<26 {
		return result, errors.New("disk online arguments")
	}
	for _, fd := range []int{codeFile.Descriptor, dataFile, scratchFile} {
		if fd < 0 || uint64(fd) > 1<<31-1 {
			return result, errors.New("disk online descriptor")
		}
	}
	a, b := uintptr(unsafe.Pointer(&message[0])), uintptr(unsafe.Pointer(&word[0]))
	if a < b+uintptr(n)*4 && b < a+uintptr(k)*4 {
		return result, errors.New("disk online input/output overlap")
	}
	seen := map[*Plan]bool{}
	for _, plan := range []*Plan{pr, pc, fr, fc} {
		if !seen[plan] {
			seen[plan] = true
			plan.mu.RLock()
			defer plan.mu.RUnlock()
		}
		if plan.handle == nil {
			return result, errors.New("disk online closed plan")
		}
	}
	var values [16]C.uint64_t
	var detail [512]C.char
	code := C.liftrs_online_encode_disk(pr.handle, pc.handle, fr.handle, fc.handle, C.uint(prefixSplit), C.uint(fullSplit), C.uint(workers), C.uint32_t(omega), C.int(codeFile.Descriptor), C.uint64_t(codeFile.POffset), C.uint64_t(codeFile.SpectrumOffset), C.int(dataFile), C.int(scratchFile), C.size_t(tile), (*C.uint32_t)(unsafe.Pointer(&message[0])), C.size_t(k), C.size_t(n), C.size_t(l), (*C.uint32_t)(unsafe.Pointer(&word[0])), &values[0], &detail[0], 512)
	runtime.KeepAlive(message)
	runtime.KeepAlive(word)
	result = DiskOnlineStats{OnlineStats: OnlineStats{MatrixStats{uint64(values[0]), uint64(values[1]), uint64(values[2])}, uint64(values[3]), uint64(values[4]), uint64(values[5])}, ReadNS: uint64(values[6]), WriteNS: uint64(values[7]), ReadBytes: uint64(values[8]), WriteBytes: uint64(values[9]), ReadCalls: uint64(values[10]), WriteCalls: uint64(values[11]), BufferBytes: uint64(values[12]), TotalNS: uint64(values[13]), Transforms: uint64(values[14]), BackingBytes: uint64(values[15])}
	if code != 0 {
		return result, fmt.Errorf("C++ disk online: %s", C.GoString(&detail[0]))
	}
	return result, nil
}
