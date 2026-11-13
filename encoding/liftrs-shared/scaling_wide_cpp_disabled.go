//go:build !liftrs_cpp || !cgo

package liftrs

func makeWideCPPOnline(ScalingCodeHeader) (wideCPPOnline, error) { return nil, ErrCPPUnavailable }
