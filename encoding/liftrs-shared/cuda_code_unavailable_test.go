//go:build !liftrs_cuda || !cgo

package liftrs

import (
	"bytes"
	"errors"
	"testing"
)

func TestCUDACodeUnavailable(t *testing.T) {
	p, raw := smallScalingCode(t)
	d, err := LoadScalingCode(bytes.NewReader(raw), codeDigest(raw), codeLimits(p))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = d.WithCUDA(CUDACodeOptions{LibraryPath: "not-loaded", Devices: []int{0}, VectorLayout: "local", DeviceMemoryLimit: 1 << 30}); !errors.Is(err, ErrCUDAUnavailable) {
		t.Fatal(err)
	}
}
