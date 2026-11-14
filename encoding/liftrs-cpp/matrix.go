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
	locks, err := lockPlans(rows, columns)
	if err != nil {
		return err
	}
	defer locks.release()
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
