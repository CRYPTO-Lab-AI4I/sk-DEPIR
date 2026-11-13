//go:build liftrs_cpp && cgo

package liftrs

import (
	"os"
	"strconv"
	"testing"
)

// Table-gap runs change experiment parameters only. Public constructors and
// historical harness defaults retain their existing admission rules.
func scalingTableDegree(t *testing.T, fallback int) int {
	t.Helper()
	raw := os.Getenv("PIR_TABLE_S")
	if raw == "" {
		return fallback
	}
	s, err := strconv.Atoi(raw)
	if err != nil || (s != 10 && s != 14 && s != 15 && s != 16) {
		t.Fatal("unsupported table experiment size", raw)
	}
	return s
}

func scalingTableConfig(t *testing.T, cfg scalingFirstUseConfig) scalingFirstUseConfig {
	if os.Getenv("PIR_TABLE_S") == "" {
		return cfg
	}
	cfg.spec = Spec{scalingTableDegree(t, cfg.spec.FieldDegree), 2, 5}
	cfg.workers, cfg.repeats, cfg.memoryLimit = 32, 2, 64<<30
	if cfg.spec.FieldDegree == 16 {
		cfg.repeats = 1
	}
	return cfg
}

func scalingTableDeviceCap(fallback uint64) uint64 {
	if os.Getenv("PIR_TABLE_S") != "" {
		return 40 << 30
	}
	return fallback
}

// Validate the experiment-only admission boundary without allocating a domain.
func TestScalingTableControls(t *testing.T) {
	t.Setenv("PIR_TABLE_S", "")
	cfg := scalingFirstUseConfig{spec: Spec{6, 2, 4}, workers: 4, repeats: 3, memoryLimit: 123}
	if got := scalingTableConfig(t, cfg); got.spec != cfg.spec || got.workers != 4 || got.repeats != 3 || got.memoryLimit != 123 {
		t.Fatal("historical defaults changed")
	}
	for _, s := range []int{10, 14, 15, 16} {
		t.Setenv("PIR_TABLE_S", strconv.Itoa(s))
		got := scalingTableConfig(t, cfg)
		if got.spec != (Spec{s, 2, 5}) || got.workers != 32 || got.repeats < 1 || got.memoryLimit != 64<<30 {
			t.Fatal(got)
		}
	}
}
