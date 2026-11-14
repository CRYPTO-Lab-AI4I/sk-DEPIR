//go:build liftrs_cpp && cgo

package liftrs

import cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"

type nativePreparationBackend struct{ *cpp.Preparation }

func (p nativePreparationBackend) NativeOwnedBytes() uint64 { return p.OwnedBytes() }
func (p nativePreparationBackend) DeviceOwnedBytes() uint64 { return 0 }
func makeCodePreparationBackend(h ScalingCodeHeader, options PreprocessingOptions) (codePreparationBackend, error) {
	native, err := cpp.NewPreparation(h.Spec.FieldDegree, h.Polynomial, h.Omega, options.Workers, h.N, options.Limits.NativeBytes)
	if err != nil {
		return nil, err
	}
	backend := nativePreparationBackend{native}
	if options.Backend == "cpp" {
		return backend, nil
	}
	gpu, err := makeCUDAPreparationBackend(h, options, backend)
	if err != nil {
		native.Close()
		return nil, err
	}
	return gpu, nil
}
