//go:build liftrs_cpp && cgo && !liftrs_cuda

package liftrs

func makeCUDAPreparationBackend(ScalingCodeHeader, PreprocessingOptions, codePreparationBackend) (codePreparationBackend, error) {
	return nil, ErrCUDAUnavailable
}
