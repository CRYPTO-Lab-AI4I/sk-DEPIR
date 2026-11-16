package liftrs

import "fmt"

// CUDACodeOptions selects same-node resident GPUs for verified code data.
// DeviceMemoryLimit caps explicit payload/metadata PER GPU, excluding contexts,
// driver/peer mappings. Devices must be1/2/4/8 distinct visible ordinals. Layout
// is explicitly peer or local; local changes vector routing/communication, not
// field parameters. CPU and legacy CUDA constructors remain independently usable.
type CUDACodeOptions struct {
	LibraryPath       string
	Devices           []int
	DeviceMemoryLimit uint64
	VectorLayout      string
}

// CUDACodeResources models principal resident arrays, not peak device/host RAM.
// CodeBytes counts copied P/spectrum; DeviceWorkspaceBytes includes arena,
// transpose/output scratch and chunk parity. Metadata/driver headroom is extra.
// MaxDeviceArrayBytes accounts for unequal prefix-constant placement on cards.
type CUDACodeResources struct {
	PrefixElements, ArenaElements, CodeBytes, DeviceWorkspaceBytes, MaxDeviceArrayBytes, HostOutputBytes uint64
}

// CUDAResources performs no GPU initialization/allocation. Full s17 capacity is
// a projection until complete encoding, transfers and concurrent lifetimes pass.
func (d *ScalingCodeData) CUDAResources(devices int) (CUDACodeResources, error) {
	h, err := canonicalScalingCodeData(d)
	if err != nil {
		return CUDACodeResources{}, err
	}
	if h.Spec.FieldDegree > 17 || devices < 1 || devices > 8 || devices&(devices-1) != 0 || uint64(devices) > h.N {
		return CUDACodeResources{}, fmt.Errorf("resident CUDA field/device count unsupported")
	}
	return cudaCodeResources(h, devices), nil
}

// Keep this pure admission model aligned with encoding::canonical_chunk.
const cudaCanonicalChunk uint64 = 4096

func cudaCodeResources(h ScalingCodeHeader, devices int) CUDACodeResources {
	n, k := h.N, h.K
	l := uint64(4)
	for l < 2*k-1 {
		l *= 2
	}
	arena := max(n, 2*l)
	width := uint64(h.ExtensionBytes)
	count := uint64(devices)
	prefixCount := min(count, l)
	prefixChunk := l / prefixCount
	parity := 4 * ((n/count+cudaCanonicalChunk-1)/cudaCanonicalChunk + 1)
	r := CUDACodeResources{PrefixElements: l, ArenaElements: arena, CodeBytes: 4*k + width*n, DeviceWorkspaceBytes: width*(arena+n) + parity*count, HostOutputBytes: 4 * n}
	for d := uint64(0); d < count; d++ {
		have := uint64(0)
		if d < prefixCount && d*prefixChunk < k {
			have = min(prefixChunk, k-d*prefixChunk)
		}
		r.MaxDeviceArrayBytes = max(r.MaxDeviceArrayBytes, width*(arena+n+n)/count+4*have+parity)
	}
	return r
}

// WithCUDA copies canonical code constants once and constructs the resident
// client schedule using the same verified data as WithWideCPP. Optional CUDA
// builds never replace the CPU path. Calls serialize and produce owning uint32
// words; no host code pointer is retained after construction. No full s17 run is
// validated by this constructor. Backend failures never select a slower fallback.
func (d *ScalingCodeData) WithCUDA(options CUDACodeOptions) (*ScalingCUDAEncoder, error) {
	r, err := d.CUDAResources(len(options.Devices))
	if err != nil {
		return nil, err
	}
	if options.LibraryPath == "" || options.DeviceMemoryLimit < r.MaxDeviceArrayBytes || (options.VectorLayout != "peer" && options.VectorLayout != "local") {
		return nil, fmt.Errorf("explicit CUDA library/layout and per-device array cap plus headroom required")
	}
	for i, id := range options.Devices {
		if id < 0 || uint64(id) > 1<<31-1 {
			return nil, fmt.Errorf("invalid CUDA device ordinal")
		}
		for _, old := range options.Devices[:i] {
			if id == old {
				return nil, fmt.Errorf("duplicate CUDA device")
			}
		}
	}
	backend, err := makeCUDACodeBackend(d, options)
	if err != nil {
		return nil, err
	}
	return &ScalingCUDAEncoder{backend: backend, k: int(d.header.K), n: int(d.header.N)}, nil
}
