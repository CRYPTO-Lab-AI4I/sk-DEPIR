//go:build liftrs_cuda && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"io"
	"time"
)

// EncodePackedTo emits the canonical s-bit database directly to destination,
// without allocating a full host canonical word. The writer must not reenter
// this encoder. Errors can leave a partial sink; publish files only after success.
// This explicit path may be slower. It preserves ordinary owning Encode methods.
func (e *ScalingCUDAEncoder) EncodePackedTo(message []uint32, destination io.Writer, bufferBytes int) (CUDAPackedStats, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var result CUDAPackedStats
	if e.closed || e.backend == nil {
		return result, ErrCUDAClosed
	}
	if len(message) != e.k {
		return result, fmt.Errorf("CUDA message length mismatch")
	}
	backend, ok := e.backend.(interface {
		EncodePackedTo([]uint32, io.Writer, int) ([10]float64, cuda.PackedStats, error)
	})
	if !ok {
		return result, fmt.Errorf("packed output backend unavailable")
	}
	start := time.Now()
	v, st, err := backend.EncodePackedTo(message, destination, bufferBytes)
	result.ElapsedNS = time.Since(start).Nanoseconds()
	result.OwnedDeviceBytes = e.backend.OwnedBytes()
	ns := func(i int) int64 { return int64(v[i] * 1e9) }
	result.HostToDeviceNS = ns(0)
	result.TaylorNS = ns(1)
	result.ColumnsNS = ns(2)
	result.RowsNS = ns(3)
	result.LayoutNS = ns(4)
	result.DeviceToHostNS = ns(5)
	result.PointwiseNS = ns(6)
	result.NativeTotalNS = ns(7)
	result.BufferNS = ns(8)
	result.AssemblyNS = ns(9)
	result.PackedBytes = st.WrittenBytes
	result.PackedSHA256 = st.SHA256
	result.HostBufferBytes = st.BufferBytes
	result.WriteNS = st.WriteNS
	result.HashNS = st.HashNS
	result.DownloadNS = st.DownloadNS
	return result, err
}
func (e *cudaStreamedBackend) EncodePackedTo(input []uint32, destination io.Writer, bufferBytes int) ([10]float64, cuda.PackedStats, error) {
	return e.Engine.EncodePackedStreamedTo(input, e.p, e.spectrum, destination, bufferBytes, e.tile)
}
