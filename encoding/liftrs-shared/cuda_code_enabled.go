//go:build liftrs_cuda && cgo

package liftrs

import cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"

type cudaCodeBackend struct{ engine *cuda.Preparation }

func makeCUDACodeBackend(d *ScalingCodeData, o CUDACodeOptions) (cudaOnlineBackend, error) {
	h := d.header
	base, err := NewBinaryField(h.Spec.FieldDegree, h.Polynomial)
	if err != nil {
		return nil, err
	}
	table := newPreparationProducts(base)
	engine, err := cuda.NewPreparation(cuda.PreparationOptions{Path: o.LibraryPath, S: h.Spec.FieldDegree, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: o.Devices, MaxElements: h.N, DeviceBytes: o.DeviceMemoryLimit})
	if err != nil {
		return nil, err
	}
	if err = engine.ConfigureEncoding(d.p, d.spectrum32, d.spectrum64, o.VectorLayout); err != nil {
		engine.Close()
		return nil, err
	}
	return &cudaCodeBackend{engine}, nil
}
func (p *cudaCodeBackend) EncodeOnline(input, word []uint32) ([10]float64, error) {
	return p.engine.EncodeCode(input, word)
}
func (p *cudaCodeBackend) OwnedBytes() uint64   { return p.engine.OwnedBytes() }
func (p *cudaCodeBackend) Close()               { _ = p.engine.Close() }
func (p *cudaCodeBackend) TimingSource() string { return "synchronized-host-wall" }
