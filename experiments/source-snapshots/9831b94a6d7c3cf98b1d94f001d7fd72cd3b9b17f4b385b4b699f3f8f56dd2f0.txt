//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"
	"unsafe"
)

// CPUFilePackedStats includes direct packed scatter and verified code-file reads.
// ComputeNS includes workspace allocation, metadata checks and native execution;
// ElapsedNS also includes bounded buffer allocation, writer and hash work.
// WorkspaceBytes counts both full N transform arrays, not caller/file-cache RAM.
// SHA256 describes the full emitted result only on success, not durable storage.
type CPUFilePackedStats struct {
	ScalingCPPFileStats
	ComputeNS, ElapsedNS, WriteNS, HashNS uint64
	WrittenBytes                          uint64
	BufferBytes                           int
	WorkspaceBytes                        uint64
	SHA256                                string
}

// EncodePackedTo explicitly emits canonical s-bit little-endian output using
// dead transform scratch. It retains/reuses both N arrays; there is no detached
// canonical uint32 result or extra full output array. Packing has atomic-write
// costs, so owning methods remain unchanged. bufferBytes must be1..64MiB.
// The writer must consume bytes synchronously and must not reenter this encoder.
// Calls/Close serialize through writing. Errors may leave a partial destination;
// file callers must Sync/Close and publish separately after success.
func (e *ScalingCPPFileEncoder) EncodePackedTo(message []uint32, destination io.Writer, bufferBytes int) (result CPUFilePackedStats, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return result, ErrCPPClosed
	}
	if destination == nil || bufferBytes < 1 || bufferBytes > 64<<20 {
		return result, fmt.Errorf("packed writer/buffer bound")
	}
	begin := time.Now()
	defer func() { result.ElapsedNS = uint64(time.Since(begin).Nanoseconds()) }()
	compute := time.Now()
	word, stats, err := e.encodeFileLocked(message, true)
	result.ComputeNS = uint64(time.Since(compute).Nanoseconds())
	result.ScalingCPPFileStats = stats
	result.WorkspaceBytes = 4 * uint64(len(e.arena)+len(e.scratch))
	if err != nil {
		return result, err
	}
	count := (len(word)*e.cpu.source.FieldDegree() + 7) / 8
	// Native atomic byte stores already produced the endian-independent packed
	// format. Borrow only its initialized prefix while this call holds the lock.
	packed := unsafe.Slice((*byte)(unsafe.Pointer(&word[0])), count)
	buffer := make([]byte, minInt(bufferBytes, count))
	result.BufferBytes = len(buffer)
	hash := sha256.New()
	for first := 0; first < count; {
		take := minInt(len(buffer), count-first)
		copy(buffer, packed[first:first+take])
		start := time.Now()
		n, writeErr := destination.Write(buffer[:take])
		result.WriteNS += uint64(time.Since(start).Nanoseconds())
		if n < 0 || n > take {
			return result, fmt.Errorf("invalid packed writer count")
		}
		start = time.Now()
		hash.Write(buffer[:n])
		result.HashNS += uint64(time.Since(start).Nanoseconds())
		result.WrittenBytes += uint64(n)
		if writeErr != nil {
			return result, writeErr
		}
		if n != take {
			return result, io.ErrShortWrite
		}
		first += take
	}
	result.SHA256 = hex.EncodeToString(hash.Sum(nil))
	return result, nil
}
