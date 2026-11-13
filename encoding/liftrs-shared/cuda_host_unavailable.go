//go:build !liftrs_cuda || !cgo

package liftrs

func newCUDAHostBackend(*ScalingOnlinePlan, CUDAOptions, int) (cudaOnlineBackend, error) {
	return nil, ErrCUDAUnavailable
}
