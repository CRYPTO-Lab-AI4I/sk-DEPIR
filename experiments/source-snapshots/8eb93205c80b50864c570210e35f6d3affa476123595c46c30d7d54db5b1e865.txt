//go:build !liftrs_cpp || !cgo

package liftrs

import (
	"errors"
	"math/rand"
	"testing"
)

func TestCPPOptionalBuild(t *testing.T) {
	b, err := DiscoverOrbit(Spec{2, 2, 1}, OrbitOptions{Random: rand.New(rand.NewSource(1205))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.WithCPPTransforms(); !errors.Is(err, ErrCPPUnavailable) {
		t.Fatal("optional C++ backend did not report unavailable", err)
	}
	if _, err := e.Encode(make([]uint32, b.Dimension())); err != nil {
		t.Fatal("pure-Go fallback failed", err)
	}
}
