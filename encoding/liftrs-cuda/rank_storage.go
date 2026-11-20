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

// RankStorage owns three reusable device arrays (slots 0..2), plus a bounded
// transfer tile, on a dedicated executor. It moves raw field-slot bits, without
// arithmetic or field validation. Logical I/O always uses row-major offsets;
// physical row/interleaved views are selected explicitly. This is experimental
// server storage, not an FFT or a caller-visible complete matrix. Calls are
// synchronous/serialized. An interrupted transpose leaves its target incomplete;
// reset that slot before reuse. Ordinary CPU builds need neither CUDA nor MPI.
type RankStorage struct {
	engine        *Preparation
	handle        unsafe.Pointer
	rank, ranks   int
	tile          uint64
	rows, columns uint64
}

// NewRankStorage allocates all three arrays for the larger local band of rows x
// columns or its transpose, with tileWords <= 1 Miword. MaxElements bounds the
// global grid; exactly one GPU is required. DeviceBytes includes every array,
// transfer tile and executor tables, but excludes driver/context and host memory.
func NewRankStorage(options PreparationOptions, rows, columns uint64, rank, ranks int, tileWords uint64) (*RankStorage, error) {
	if len(options.Devices) != 1 || rows == 0 || columns == 0 || rows > ^uint64(0)/columns || rows*columns > options.MaxElements || ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks || tileWords == 0 || tileWords > 1<<20 {
		return nil, fmt.Errorf("GPU rank storage shape/rank/tile admission")
	}
	p, err := NewPreparation(options)
	if err != nil {
		return nil, err
	}
	v := &RankStorage{engine: p, rank: rank, ranks: ranks, tile: tileWords, rows: rows, columns: columns}
	var text [512]C.char
	var handle unsafe.Pointer
	if err = preparationError(C.pg_bridge_rank_create(p.bridge, C.size_t(rows), C.size_t(columns), C.uint(rank), C.uint(ranks), C.size_t(tileWords), &handle, &text[0], 512), &text[0]); err != nil {
		p.Close()
		return nil, err
	}
	v.handle = handle
	return v, nil
}

// Close releases storage before its borrowed executor; repeated calls are safe.
func (v *RankStorage) Close() error {
	if v == nil || v.engine == nil {
		return nil
	}
	p := v.engine
	p.mu.Lock()
	if v.handle != nil {
		C.pg_bridge_rank_destroy(p.bridge, v.handle)
		v.handle = nil
	}
	p.mu.Unlock()
	return p.Close()
}

// OwnedBytes includes arrays/transport/tables, excluding CUDA context/driver.
func (v *RankStorage) OwnedBytes() uint64 {
	if v == nil || v.engine == nil {
		return 0
	}
	return v.engine.OwnedBytes()
}
func (v *RankStorage) call(fn func(*C.pg_bridge, *C.char) C.int) error {
	if v == nil || v.engine == nil {
		return fmt.Errorf("nil GPU rank storage")
	}
	p := v.engine
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.bridge == nil || v.handle == nil {
		return fmt.Errorf("closed GPU rank storage")
	}
	var text [512]C.char
	return preparationError(fn(p.bridge, &text[0]), &text[0])
}

// Reset clears one slot and selects either the original shape or its transpose.
// Interleaved layout is [column*ownedRows+row], suitable for existing vector kernels.
func (v *RankStorage) Reset(slot int, rows, columns uint64, interleaved bool) error {
	if slot < 0 || slot >= 3 {
		return fmt.Errorf("GPU rank storage slot")
	}
	layout := C.int(0)
	if interleaved {
		layout = 1
	}
	return v.call(func(b *C.pg_bridge, text *C.char) C.int {
		return C.pg_bridge_rank_view(b, v.handle, C.uint(slot), C.size_t(rows), C.size_t(columns), layout, text, 512)
	})
}

// Local copies the source/target ownership intersection, preserving source and
// the third slot. Slots must be distinct with transposed shapes. It does not
// exchange remote words; Pack/Scatter complete the shared caller schedule.
func (v *RankStorage) Local(source, target int) error {
	if source < 0 || source >= 3 || target < 0 || target >= 3 || source == target {
		return fmt.Errorf("GPU rank storage alias/slot")
	}
	return v.call(func(b *C.pg_bridge, text *C.char) C.int {
		return C.pg_bridge_rank_local(b, v.handle, C.uint(source), C.uint(target), text, 512)
	})
}
func (v *RankStorage) transfer(slot, peer int, offset uint64, data unsafe.Pointer, words, width int, upload, packet bool) error {
	if v == nil || v.engine == nil || slot < 0 || slot >= 3 || words <= 0 || data == nil || uint64(words) > v.tile || v.engine.width != width || (packet && (peer < 0 || peer >= v.ranks || peer == v.rank)) {
		return fmt.Errorf("GPU rank storage transfer admission")
	}
	direction := C.int(0)
	if upload {
		direction = 1
	}
	return v.call(func(b *C.pg_bridge, text *C.char) C.int {
		if packet {
			return C.pg_bridge_rank_tile(b, v.handle, C.uint(slot), C.uint(peer), C.size_t(offset), data, C.size_t(words), direction, text, 512)
		}
		return C.pg_bridge_rank_transfer(b, v.handle, C.uint(slot), C.size_t(offset), data, C.size_t(words), direction, text, 512)
	})
}

// Transfer32 copies a bounded logical row-major interval; upload selects host to
// device. No caller pointer is retained. Zero-length intervals are rejected.
func (v *RankStorage) Transfer32(slot int, offset uint64, data []uint32, upload bool) error {
	err := v.transfer(slot, 0, offset, preparationPointer(data), len(data), 4, upload, false)
	runtime.KeepAlive(data)
	return err
}

// Transfer64 is Transfer32 for eight-byte extension-field slots.
func (v *RankStorage) Transfer64(slot int, offset uint64, data []uint64, upload bool) error {
	err := v.transfer(slot, 0, offset, preparationPointer(data), len(data), 8, upload, false)
	runtime.KeepAlive(data)
	return err
}

// Tile32 packs a peer rectangle to host when pack is true; otherwise scatters a
// received rectangle. Offset is within that peer's wire stream: original column
// first, then original row. The source/target views determine the peer extent.
func (v *RankStorage) Tile32(slot, peer int, offset uint64, data []uint32, pack bool) error {
	err := v.transfer(slot, peer, offset, preparationPointer(data), len(data), 4, pack, true)
	runtime.KeepAlive(data)
	return err
}

// Tile64 is Tile32 for eight-byte extension-field slots.
func (v *RankStorage) Tile64(slot, peer int, offset uint64, data []uint64, pack bool) error {
	err := v.transfer(slot, peer, offset, preparationPointer(data), len(data), 8, pack, true)
	runtime.KeepAlive(data)
	return err
}

// AdmitsTranspose checks immutable global ownership and the transport bound.
// Source/target view compatibility and distinct slots are checked by Local.
func (v *RankStorage) AdmitsTranspose(rows, columns uint64, rank, ranks int, tileWords uint64) bool {
	return v != nil && rank == v.rank && ranks == v.ranks && tileWords > 0 && tileWords <= v.tile && ((rows == v.rows && columns == v.columns) || (rows == v.columns && columns == v.rows))
}

// Vectors transforms complete resident vectors in place. The selected slot must
// be interleaved, with canonical balanced global dimensions (transpose for
// columns). Values are range-checked on GPU before arithmetic or mutation; raw
// storage bits outside the field are rejected. A plan is retained on first use,
// included in OwnedBytes/cap. Calls reuse the existing bounded vector launcher.
func (v *RankStorage) Vectors(slot int, columns, inverse bool) error {
	if slot < 0 || slot >= 3 {
		return fmt.Errorf("GPU rank vector slot")
	}
	axis, direction := C.int(0), C.int(0)
	if columns {
		axis = 1
	}
	if inverse {
		direction = 1
	}
	return v.call(func(b *C.pg_bridge, text *C.char) C.int {
		return C.pg_bridge_rank_vectors(b, v.handle, C.uint(slot), axis, direction, text, 512)
	})
}
