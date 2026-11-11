package liftrs

import (
	"errors"
	"testing"
)

type mockCUDAOnline struct{ calls, closed int }

func (m *mockCUDAOnline) EncodeOnline(input, word []uint32) ([10]float64, error) {
	m.calls++
	for i := range word {
		word[i] = input[0]
	}
	return [10]float64{0.001}, nil
}
func (m *mockCUDAOnline) OwnedBytes() uint64 {
	if m.closed > 0 {
		return 0
	}
	return 4096
}
func (m *mockCUDAOnline) Close() { m.closed++ }

func TestCUDAEncoderLifecycle(t *testing.T) {
	backend := &mockCUDAOnline{}
	encoder := &ScalingCUDAEncoder{backend: backend, k: 2, n: 16}
	input := []uint32{2, 3}
	if _, _, err := encoder.EncodeWithStats(input[:1]); err == nil || backend.calls != 0 {
		t.Fatal("invalid length reached backend")
	}
	word, stats, err := encoder.EncodeWithStats(input)
	if err != nil || len(word) != 16 || word[0] != 2 || stats.OwnedDeviceBytes != 4096 || stats.HostToDeviceNS != 1_000_000 {
		t.Fatal("encode/stats", err)
	}
	if input[0] != 2 || input[1] != 3 {
		t.Fatal("mutated input")
	}
	encoder.Close()
	encoder.Close()
	if backend.closed != 1 || encoder.OwnedBytes() != 0 {
		t.Fatal("close ownership")
	}
	if _, err = encoder.Encode(input); !errors.Is(err, ErrCUDAClosed) {
		t.Fatal("closed encode", err)
	}
}

func TestCUDAAdmission(t *testing.T) {
	if _, err := (*ScalingOnlinePlan)(nil).CUDAArrayBytes(); err == nil {
		t.Fatal("nil plan admitted")
	}
	basis, err := DiscoverOrbit(Spec{3, 2, 2}, OrbitOptions{})
	if err != nil {
		t.Fatal(err)
	}
	code, err := basis.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := code.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := plan.CUDAArrayBytes()
	if err != nil || payload != (4*uint64(plan.Length())+uint64(plan.Dimension())+1)*4 {
		t.Fatal("array accounting", err)
	}
	if _, err = plan.WithCUDA(CUDAOptions{LibraryPath: "unused", DeviceMemoryLimit: payload - 1}); err == nil {
		t.Fatal("underfunded device admitted")
	}
}
