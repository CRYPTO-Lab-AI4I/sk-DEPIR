//go:build liftrs_cpp && cgo

package liftrs

import (
	"os"
	"testing"
)

// This opt-in first-use case is admitted only by the Linux resource controller.
// Normal suites skip it; the smoke mode checks persistence on a tiny known code.
func TestS16FirstUse(t *testing.T) {
	out, cache := os.Getenv("PIR_S16_OUT"), os.Getenv("PIR_S16_CACHE")
	if out == "" {
		t.Skip("opt-in guarded s16 first use")
	}
	if cache == "" {
		t.Fatal("cache required")
	}
	cfg := scalingFirstUseConfig{spec: Spec{16, 2, 5}, out: out, cache: cache,
		workers: 32, repeats: 1, memoryLimit: 192 << 30, onlineOnly: os.Getenv("PIR_S16_ONLINE_ONLY") == "1"}
	if os.Getenv("PIR_S16_SMOKE") == "1" {
		cfg.spec, cfg.workers, cfg.memoryLimit = Spec{6, 2, 4}, 2, 1<<30
	}
	runScalingFirstUse(t, cfg)
}

func TestScalingPreparationAdmissionShape(t *testing.T) {
	if !validScalingPreparationShape(Spec{16, 2, 5}, 32) {
		t.Fatal("s16 shape rejected")
	}
	for _, c := range []struct {
		spec    Spec
		workers int
	}{
		{Spec{17, 2, 5}, 32}, {Spec{16, 3, 5}, 32}, {Spec{16, 2, 0}, 32},
		{Spec{16, 2, 65535}, 32}, {Spec{16, 2, 5}, 0}, {Spec{16, 2, 5}, 65},
	} {
		if validScalingPreparationShape(c.spec, c.workers) {
			t.Fatal("invalid preparation shape", c)
		}
	}
}
