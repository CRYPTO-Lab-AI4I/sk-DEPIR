//go:build !liftrs_cuda || !cgo

package liftrs

import (
	"errors"
	"testing"
)

func TestCUDAUnavailable(t *testing.T) {
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
	if _, err = plan.WithCUDA(CUDAOptions{LibraryPath: "not-loaded", DeviceMemoryLimit: 1 << 20}); !errors.Is(err, ErrCUDAUnavailable) {
		t.Fatal("unavailable backend", err)
	}
	// The default CPU/Go encoder still works in the same build.
	input := make([]uint32, plan.Dimension())
	if _, err = plan.Encode(input, plan.NewWorkspace()); err != nil {
		t.Fatal(err)
	}
}
