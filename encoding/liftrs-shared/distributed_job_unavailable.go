//go:build !linux || !liftrs_cpp || !liftrs_cuda || !liftrs_mpi || !cgo

package liftrs

import (
	"fmt"
	"io"
)

// RunDistributedPreprocessing requires an explicit Linux CUDA/MPI/native build.
func RunDistributedPreprocessing(c DistributedPreprocessingConfig, output io.Writer, random io.Reader) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return fmt.Errorf("distributed preprocessing requires Linux,cgo and liftrs_cpp,liftrs_cuda,liftrs_mpi tags")
}
