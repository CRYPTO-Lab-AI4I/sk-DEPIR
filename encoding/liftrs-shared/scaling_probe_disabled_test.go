//go:build !liftrs_cpp || !cgo

package liftrs

import "testing"

func scalingNativeProbe(t *testing.T, s int, phase func(string, func() map[string]any), baseline *CyclicEncoder, plan *ScalingOnlinePlan) {
	t.Fatal("resource probe requires native build: -tags liftrs_cpp and cgo")
}
