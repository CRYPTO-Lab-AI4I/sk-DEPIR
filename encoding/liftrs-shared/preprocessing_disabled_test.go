//go:build !liftrs_cpp || !cgo

package liftrs

import (
	"errors"
	"testing"
)

func TestPreprocessingNativeUnavailable(t *testing.T) {
	_, err := NewCodePreprocessor(Spec{4, 2, 2}, PreprocessingOptions{Backend: "cpp", Workers: 1, Limits: PreprocessingLimits{MaxElements: 256, MaxHostArrayBytes: 1 << 20, NativeBytes: 8 << 20}})
	if !errors.Is(err, ErrCPPUnavailable) {
		t.Fatal(err)
	}
}
