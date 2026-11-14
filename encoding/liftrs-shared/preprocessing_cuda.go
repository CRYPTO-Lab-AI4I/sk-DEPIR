//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/bits"
)

// Existing resident GPU products/transforms are execution adapters, not a second
// discovery/root/reciprocal procedure. Native code supplies roots/base-grid work.
// Wider/multi-GPU support is not silently replaced by a slower host-backed mode.
type cudaPreparationBackend struct {
	codePreparationBackend
	engine *cuda.Engine
	field  orbitField
	plans  map[int]bool
}

func makeCUDAPreparationBackend(h ScalingCodeHeader, options PreprocessingOptions, native codePreparationBackend) (codePreparationBackend, error) {
	if len(options.GPUDevices) != 0 {
		return makeResidentCUDAPreparationBackend(h, options, native)
	}
	if h.Spec.FieldDegree > 16 {
		return nil, fmt.Errorf("s17 CUDA preprocessing requires explicit GPUDevices and the resident preparation library")
	}
	_, extension, _, _ := orbitFields(h.Spec.FieldDegree, h.Polynomial)
	base, _ := NewBinaryField(h.Spec.FieldDegree, h.Polynomial)
	table := newPreparationProducts(base)
	engine, err := cuda.NewBounded(options.GPUPath, h.Spec.FieldDegree, h.Polynomial, h.Delta, table.log, table.powers, int(h.N), options.GPUBytes)
	if err != nil {
		return nil, err
	}
	return &cudaPreparationBackend{codePreparationBackend: native, engine: engine, field: extension, plans: make(map[int]bool)}, nil
}
func (p *cudaPreparationBackend) plan(n int) error {
	if p.plans[n] {
		return nil
	}
	description, err := cudaMatrixDescription(p.field, bits.Len(uint(n))-1)
	if err != nil {
		return err
	}
	if err = p.engine.AddPlan(description); err != nil {
		return err
	}
	p.plans[n] = true
	return nil
}
func (p *cudaPreparationBackend) Apply32(data []uint32, inverse bool) error {
	if len(data) < 4 {
		return p.codePreparationBackend.Apply32(data, inverse)
	}
	if err := p.plan(len(data)); err != nil {
		return err
	}
	_, err := p.engine.Apply(data, inverse)
	return err
}
func (p *cudaPreparationBackend) Product(a, b []uint32, limit, n int) ([]uint32, error) {
	if n < 4 {
		return p.codePreparationBackend.Product(a, b, limit, n)
	}
	if err := p.plan(n); err != nil {
		return nil, err
	}
	out, _, err := p.engine.Product(a, b, limit, n)
	return out, err
}
func (p *cudaPreparationBackend) Spectrum32(g, out []uint32) error {
	clear(out)
	copy(out, g)
	return p.Apply32(out, false)
}
func (p *cudaPreparationBackend) DeviceOwnedBytes() uint64 { return p.engine.OwnedBytes() }
func (p *cudaPreparationBackend) Close() error {
	p.engine.Close()
	return p.codePreparationBackend.Close()
}

// The resident engine owns every heavy public-code kernel; the embedded native
// metadata remains only for common bookkeeping/reference and has no grid arrays.
type residentCUDAPreparationBackend struct {
	*cuda.Preparation
	native codePreparationBackend
}

func makeResidentCUDAPreparationBackend(h ScalingCodeHeader, options PreprocessingOptions, native codePreparationBackend) (codePreparationBackend, error) {
	base, err := NewBinaryField(h.Spec.FieldDegree, h.Polynomial)
	if err != nil {
		return nil, err
	}
	table := newPreparationProducts(base)
	engine, err := cuda.NewPreparation(cuda.PreparationOptions{Path: options.GPUPath, S: h.Spec.FieldDegree, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: options.GPUDevices, MaxElements: h.N, DeviceBytes: options.GPUBytes})
	if err != nil {
		return nil, err
	}
	return &residentCUDAPreparationBackend{Preparation: engine, native: native}, nil
}
func (p *residentCUDAPreparationBackend) NativeOwnedBytes() uint64 {
	return p.native.NativeOwnedBytes()
}
func (p *residentCUDAPreparationBackend) DeviceOwnedBytes() uint64 { return p.Preparation.OwnedBytes() }
func (p *residentCUDAPreparationBackend) ReleaseProducts() {
	if err := p.Preparation.ReleaseProducts(); err != nil {
		panic(cppFailure{err})
	}
	p.native.ReleaseProducts()
}
func (p *residentCUDAPreparationBackend) Close() error {
	p.Preparation.Close()
	return p.native.Close()
}
