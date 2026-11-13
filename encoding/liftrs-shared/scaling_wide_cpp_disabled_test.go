//go:build !liftrs_cpp || !cgo

package liftrs

import (
	"bytes"
	"testing"
)

func TestScalingWideCPPUnavailable(t *testing.T) {
	plan, raw := smallScalingCode(t)
	data, err := LoadScalingCode(bytes.NewReader(raw), codeDigest(raw), codeLimits(plan))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = data.WithWideCPP(1); err != ErrCPPUnavailable {
		t.Fatal("ordinary build must need no native backend", err)
	}
	var empty ScalingWideCPPEncoder
	empty.Close()
	if empty.OwnedBytes() != 0 {
		t.Fatal("empty encoder")
	}
}
