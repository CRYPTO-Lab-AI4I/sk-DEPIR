//go:build liftrs_cpp && cgo

package liftrs

import (
	"errors"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ScalingCPPDiskEncoder explicitly exchanges private transform RAM for two
// disk workspace files and repeated I/O. Immutable code remains file-streamed.
// An owning N-word output stays in host RAM; file cache/runtime are additional.
// Calls/Close serialize. Close preserves workspace files for caller management.
type ScalingCPPDiskEncoder struct {
	mu            sync.Mutex
	source        *ScalingCPPFileEncoder
	data, scratch *os.File
	tile          int
	closed        bool
}

// OpenScalingCPPDisk verifies trusted code as OpenScalingCPPFile does, then
// creates exclusive data.bin/scratch.bin files in workspaceDirectory, each4N
// bytes. The directory must not contain those names. No existing file is replaced.
// Caller must exclude external writes/truncation of code and workspace files.
// tileElements must fit complete rows/columns (and be4..2^26). Workspaces survive
// Close, including partial initialization on error; no checkpoint resume implied.
// Slower disk mode must be explicitly selected; ordinary CPU APIs stay default.
func OpenScalingCPPDisk(path, digest string, maxElements, workers, tileElements int, workspaceDirectory string) (*ScalingCPPDiskEncoder, error) {
	if tileElements < 4 || tileElements > 1<<26 || workspaceDirectory == "" {
		return nil, fmt.Errorf("disk tile/workspace required")
	}
	source, err := OpenScalingCPPFile(path, digest, maxElements, workers)
	if err != nil {
		return nil, err
	}
	ok := false
	var data, scratch *os.File
	defer func() {
		if !ok {
			source.Close()
			if data != nil {
				data.Close()
			}
			if scratch != nil {
				scratch.Close()
			}
		}
	}()
	n := source.Length()
	full := source.cpu.plans[n]
	if tileElements < max(1<<uint(full.split), n>>uint(full.split)) {
		return nil, fmt.Errorf("disk tile shorter than full vector")
	}
	if err = os.MkdirAll(workspaceDirectory, 0700); err != nil {
		return nil, err
	}
	data, err = os.OpenFile(filepath.Join(workspaceDirectory, "data.bin"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	scratch, err = os.OpenFile(filepath.Join(workspaceDirectory, "scratch.bin"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	for _, file := range []*os.File{data, scratch} {
		if err = file.Truncate(4 * int64(n)); err != nil {
			return nil, err
		}
	}
	ok = true
	return &ScalingCPPDiskEncoder{source: source, data: data, scratch: scratch, tile: tileElements}, nil
}

// ScalingCPPDiskStats adds logical file traffic and bounded native private
// payloads to phase timings. ElapsedNS includes owning output allocation and
// metadata checks; NativeTotalNS excludes admission. Buffers exclude output,
// input, plans/runtime and potentially large filesystem cache. No fsync is timed.
type ScalingCPPDiskStats struct {
	ScalingCPPStats
	ReadNS, WriteNS, ReadBytes, WriteBytes, ReadCalls, WriteCalls    uint64
	BufferBytes, BackingBytes, OutputBytes, NativeTotalNS, ElapsedNS uint64
}

// EncodeWithStats returns a new detached owning uint32 canonical word. No full
// host transform arrays or code arrays are allocated. Error recovery starts the
// complete schedule again; workspace files are not committed checkpoints.
func (e *ScalingCPPDiskEncoder) EncodeWithStats(message []uint32) (word []uint32, result ScalingCPPDiskStats, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, result, ErrCPPClosed
	}
	source := e.source
	p := source.cpu.source
	n := p.Length()
	if len(message) != p.k {
		return nil, result, fmt.Errorf("disk message length")
	}
	start := time.Now()
	defer func() { result.ElapsedNS = uint64(time.Since(start).Nanoseconds()) }()
	before, err := source.file.Stat()
	if err != nil {
		return nil, result, err
	}
	if !sameCodeFile(source.identity, before) {
		return nil, result, fmt.Errorf("disk code backing changed")
	}
	word = make([]uint32, n)
	pre, full := source.cpu.plans[source.prefix], source.cpu.plans[n]
	st, err := cpp.EncodeOnlineDisk(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, source.cpu.workers, p.basis.omega, message, word, source.prefix, e.tile, int(e.data.Fd()), int(e.scratch.Fd()), cpp.OnlineFile{Descriptor: int(source.file.Fd()), POffset: source.pOffset, SpectrumOffset: source.spectrumOffset})
	runtime.KeepAlive(e)
	result = ScalingCPPDiskStats{ScalingCPPStats: ScalingCPPStats{Calls: st.Transforms, BoundaryCalls: 1, TaylorNS: st.TaylorNS, ColumnsNS: st.ColumnsNS, RowsNS: st.RowsNS, BufferNS: st.BufferNS, PointwiseNS: st.PointwiseNS, AssemblyNS: st.AssemblyNS}, ReadNS: st.ReadNS, WriteNS: st.WriteNS, ReadBytes: st.ReadBytes, WriteBytes: st.WriteBytes, ReadCalls: st.ReadCalls, WriteCalls: st.WriteCalls, BufferBytes: st.BufferBytes, BackingBytes: st.BackingBytes, OutputBytes: 4 * uint64(n), NativeTotalNS: st.TotalNS}
	if err != nil {
		return nil, result, err
	}
	after, err := source.file.Stat()
	if err != nil {
		return nil, result, err
	}
	if !sameCodeFile(source.identity, after) {
		return nil, result, fmt.Errorf("disk code changed during encoding")
	}
	return word, result, nil
}

// Close releases files/metadata, preserves workspace artifacts and waits for
// encoding. Returned owning words remain valid; code preprocessing is unchanged.
func (e *ScalingCPPDiskEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil
	}
	e.closed = true
	return errors.Join(e.source.Close(), e.data.Close(), e.scratch.Close())
}
