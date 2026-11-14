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
	locks, err := lockPlans(rows, columns)
	if err != nil {
		return result, err
	}
	defer locks.release()
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
	locks, err := lockPlans(pr, pc, fr, fc)
	if err != nil {
		return result, err
	}
	defer locks.release()
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
