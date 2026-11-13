package liftrs

import "fmt"

// WithCUDAHostBacked explicitly exchanges device memory for host arrays and
// repeated transfers. It retains this immutable plan's P/spectrum until Close.
// Two device tiles replace full device domains. Each Encode needs a retained
// 4N-byte host arena, a new 4N-byte owning result and at most 4T native staging,
// in addition to message/code/runtime memory. It can be much slower; WithCUDA
// remains the default GPU API. Callers with enough resident memory must select
// this alternative explicitly (for CLI programs, use an explicit mode option).
// Five transforms/products and orbit generation run on GPU; native CPU performs
// canonical scatter/XOR. No full orbit map, packed sink, or s17 support is added.
// Current schedule requires 2L<=N, where L is the next power of two >=2K-1.
func (p *ScalingOnlinePlan) WithCUDAHostBacked(options CUDAOptions, tileElements int) (*ScalingCUDAEncoder, error) {
	_, k, n, l, err := cudaShape(p)
	if err != nil {
		return nil, err
	}
	if l > n/2 || tileElements < 4 || tileElements > 1<<26 || options.LibraryPath == "" {
		return nil, fmt.Errorf("host-backed prefix, tile or library unsupported")
	}
	backend, err := newCUDAHostBackend(p, options, tileElements)
	if err != nil {
		return nil, err
	}
	return &ScalingCUDAEncoder{backend: backend, k: k, n: n}, nil
}
