package liftrs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCUDAUnavailable means this build omits the optional CUDA binding. CPU
// encoders and prepared-code loading do not require a GPU or CUDA installation.
var ErrCUDAUnavailable = errors.New("CUDA encoding unavailable: build with liftrs_cuda and cgo")

// ErrCUDAClosed indicates an encode request after device state was released.
var ErrCUDAClosed = errors.New("CUDA encoder closed")

// CUDAOptions chooses an optional client encoding backend, independently of how
// code preprocessing was obtained. LibraryPath must identify a trusted CUDA
// library built for the installed device/runtime. DeviceMemoryLimit bounds all
// explicit device arrays/tables/plans; driver/context allocations are separate.
// Selection is explicit: initialization errors never silently switch to CPU.
type CUDAOptions struct {
	LibraryPath       string
	DeviceMemoryLimit uint64
}

type cudaOnlineBackend interface {
	EncodeOnline([]uint32, []uint32) ([10]float64, error)
	OwnedBytes() uint64
	Close()
}

// ScalingCUDAEncoder owns reusable device constants/workspace. Calls serialize;
// Close waits for an active encode. The host ScalingOnlinePlan is not retained,
// so callers may release it after construction or keep it for CPU execution.
// This is the same standalone mathematical encoder as the CPU path, not the
// application/cryptographic encoder. Current packed arithmetic supports s2..16.
type ScalingCUDAEncoder struct {
	mu      sync.Mutex
	backend cudaOnlineBackend
	k, n    int
	closed  bool
}

func cudaShape(p *ScalingOnlinePlan) (s, k, n, l int, err error) {
	if p == nil {
		return 0, 0, 0, 0, fmt.Errorf("nil online plan")
	}
	s = p.FieldDegree()
	k = p.k
	n = p.Length()
	if s < 2 || s > 16 || p.basis.spec.Dimension != 2 || uint64(n) != uint64(1)<<uint(2*s) || k < 1 || k > n/2 || len(p.p) != k+1 || len(p.spectrum) != n {
		return 0, 0, 0, 0, fmt.Errorf("CUDA online shape unsupported")
	}
	l = 4
	for l < 2*k-1 {
		l *= 2
	}
	return
}

// CUDAArrayBytes reports the code and workspace array payload for this plan.
// It excludes small tables/plans and driver/context overhead, so an allocation
// cap needs additional headroom. It performs no GPU initialization/allocation.
func (p *ScalingOnlinePlan) CUDAArrayBytes() (uint64, error) {
	_, k, n, _, err := cudaShape(p)
	if err != nil {
		return 0, err
	}
	return (4*uint64(n) + uint64(k) + 1) * 4, nil
}

// WithCUDA copies public code constants to device and creates a reusable GPU
// encoder. Build with liftrs_cuda,cgo; liftrs_cpp is independently optional.
// CPU use remains available through Encode or WithCPPMatrix. Code data/spec and
// canonical outputs are shared across backends; the caller chooses explicitly.
func (p *ScalingOnlinePlan) WithCUDA(options CUDAOptions) (*ScalingCUDAEncoder, error) {
	_, k, n, _, err := cudaShape(p)
	if err != nil {
		return nil, err
	}
	payload, _ := p.CUDAArrayBytes()
	if options.LibraryPath == "" || options.DeviceMemoryLimit < payload {
		return nil, fmt.Errorf("CUDA library path and device cap above %d array bytes required", payload)
	}
	backend, err := newCUDAOnlineBackend(p, options)
	if err != nil {
		return nil, err
	}
	return &ScalingCUDAEncoder{backend: backend, k: k, n: n}, nil
}

// CUDAEncodeStats separates transfer-inclusive native call and whole Go call
// clocks. Kernel stages use CUDA event durations. Their sum need not equal total
// elapsed (validation, allocation, synchronization and wrapper overhead remain).
type CUDAEncodeStats struct {
	HostToDeviceNS   int64  `json:"host_to_device_ns"`
	TaylorNS         int64  `json:"taylor_ns"`
	ColumnsNS        int64  `json:"columns_ns"`
	RowsNS           int64  `json:"rows_ns"`
	LayoutNS         int64  `json:"layout_ns"`
	DeviceToHostNS   int64  `json:"device_to_host_ns"`
	PointwiseNS      int64  `json:"pointwise_ns"`
	NativeTotalNS    int64  `json:"native_total_ns"`
	BufferNS         int64  `json:"buffer_ns"`
	AssemblyNS       int64  `json:"assembly_ns"`
	ElapsedNS        int64  `json:"elapsed_ns"`
	OwnedDeviceBytes uint64 `json:"owned_device_bytes"`
}

// Encode returns a new canonical uint32 codeword. Inputs remain unchanged;
// device constants/workspace are reused. All five transforms, products and
// canonical assembly stay on device; only the message and final word transfer.
func (e *ScalingCUDAEncoder) Encode(message []uint32) ([]uint32, error) {
	word, _, err := e.EncodeWithStats(message)
	return word, err
}

// EncodeWithStats includes host output allocation in ElapsedNS, separates the
// native call's transfer/kernel stages and reports explicit resident device bytes.
func (e *ScalingCUDAEncoder) EncodeWithStats(message []uint32) ([]uint32, CUDAEncodeStats, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var stats CUDAEncodeStats
	if e.closed || e.backend == nil {
		return nil, stats, ErrCUDAClosed
	}
	if len(message) != e.k {
		return nil, stats, fmt.Errorf("CUDA message length mismatch")
	}
	start := time.Now()
	word := make([]uint32, e.n)
	values, err := e.backend.EncodeOnline(message, word)
	stats.ElapsedNS = time.Since(start).Nanoseconds()
	stats.OwnedDeviceBytes = e.backend.OwnedBytes()
	ns := func(i int) int64 { return int64(values[i] * 1e9) }
	stats.HostToDeviceNS = ns(0)
	stats.TaylorNS = ns(1)
	stats.ColumnsNS = ns(2)
	stats.RowsNS = ns(3)
	stats.LayoutNS = ns(4)
	stats.DeviceToHostNS = ns(5)
	stats.PointwiseNS = ns(6)
	stats.NativeTotalNS = ns(7)
	stats.BufferNS = ns(8)
	stats.AssemblyNS = ns(9)
	if err != nil {
		return nil, stats, err
	}
	return word, stats, nil
}

// OwnedBytes counts explicit device code/constants/workspace; context/driver
// memory is excluded. It returns zero after Close.
func (e *ScalingCUDAEncoder) OwnedBytes() uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.backend == nil {
		return 0
	}
	return e.backend.OwnedBytes()
}

// Close releases device state once and waits for in-flight calls. CPU objects
// and the caller's immutable prepared plan are unaffected.
func (e *ScalingCUDAEncoder) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.closed && e.backend != nil {
		e.backend.Close()
	}
	e.closed = true
	return nil
}
