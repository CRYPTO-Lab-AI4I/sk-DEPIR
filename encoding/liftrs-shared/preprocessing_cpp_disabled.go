//go:build !liftrs_cpp || !cgo

package liftrs

func makeCodePreparationBackend(ScalingCodeHeader, PreprocessingOptions) (codePreparationBackend, error) {
	return nil, ErrCPPUnavailable
}
