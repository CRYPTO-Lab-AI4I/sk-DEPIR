//go:build !liftrs_cuda || !cgo

package liftrs

import "io"

// EncodePackedTo requires the optional CUDA backend; CPU/Go builds stay usable.
func (e *ScalingCUDAEncoder) EncodePackedTo(message []uint32, destination io.Writer, bufferBytes int) (CUDAPackedStats, error) {
	return CUDAPackedStats{}, ErrCUDAUnavailable
}
