//go:build liftrs_cuda && cgo

// Package gpumatrix preserves historical experiment imports. New client code
// uses encoding/liftrs-cuda through the shared encoder's WithCUDA constructor.
package gpumatrix

import cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"

type Engine = cuda.Engine

func New(path string, s int, polynomial, delta uint32, logs, powers []uint32) (*Engine, error) {
	return cuda.New(path, s, polynomial, delta, logs, powers)
}
func NewBounded(path string, s int, polynomial, delta uint32, logs, powers []uint32, maxPoints int, deviceBytes uint64) (*Engine, error) {
	return cuda.NewBounded(path, s, polynomial, delta, logs, powers, maxPoints, deviceBytes)
}
