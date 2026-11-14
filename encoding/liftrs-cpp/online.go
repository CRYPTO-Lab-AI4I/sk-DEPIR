//go:build liftrs_cpp && cgo

package liftrscpp

/*
#include "c_api.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

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
	locks, err := lockPlans(pr, pc, fr, fc)
	if err != nil {
		return result, err
	}
	defer locks.release()
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
