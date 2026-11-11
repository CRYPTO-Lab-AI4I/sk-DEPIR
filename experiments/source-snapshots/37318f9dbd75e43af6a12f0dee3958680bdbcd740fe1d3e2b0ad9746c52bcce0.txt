//go:build !liftrs_cuda || !cgo

package liftrs

func newCUDAOnlineBackend(*ScalingOnlinePlan, CUDAOptions) (cudaOnlineBackend, error) {
	return nil, ErrCUDAUnavailable
}
