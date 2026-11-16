//go:build !liftrs_cuda || !cgo

package liftrs

func makeCUDACodeBackend(*ScalingCodeData, CUDACodeOptions) (cudaOnlineBackend, error) {
	return nil, ErrCUDAUnavailable
}
