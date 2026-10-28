//go:build !liftrs_cpp || !cgo

package liftrs

func makeCPPPlan(*additiveTransform) (cppPlan, uint64, error) {
	return nil, 0, ErrCPPUnavailable
}
