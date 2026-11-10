//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"testing"
	"time"
)

// Opt-in bounded experiment: fixed s16 leaf work and a complete s10 run.
// The controller builds the same test with baseline and changed preparation.
func TestPreparationBenchmark(t *testing.T) {
	out := os.Getenv("PIR_PREPARATION_OUT")
	if out == "" {
		t.Skip("opt-in preparation arithmetic experiment")
	}
	runScalingFirstUse(t, scalingFirstUseConfig{spec: Spec{10, 2, 5},
		out: out, cache: os.Getenv("PIR_PREPARATION_CACHE"), workers: 4,
		repeats: 1, memoryLimit: 4 << 30})
}

func TestPreparationLeafProfile(t *testing.T) {
	out := os.Getenv("PIR_PREPARATION_OUT")
	if out == "" {
		t.Skip("opt-in leaf profile")
	}
	base, _, _, _ := orbitFields(16, 0x1002d)
	setup := time.Now()
	if os.Getenv("PIR_PREPARATION_LOOKUP") == "1" {
		base.lookup = newPreparationProducts(base.base)
	}
	setupSeconds := time.Since(setup).Seconds()
	factors := make([][]uint32, 32768)
	rng := rand.New(rand.NewSource(1301))
	for i := range factors {
		factors[i] = []uint32{rng.Uint32() & 65535, rng.Uint32() & 65535, 1}
	}
	f, err := os.Create(filepath.Join(out, "leaves.pprof"))
	if err != nil {
		t.Fatal(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	product := rootProduct(base, factors)
	elapsed := time.Since(start)
	pprof.StopCPUProfile()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	for _, x := range product {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], x)
		h.Write(b[:])
	}
	raw, _ := json.MarshalIndent(map[string]any{"seconds": elapsed.Seconds(),
		"setup_seconds": setupSeconds, "factors": len(factors), "sha256": fmt.Sprintf("%x", h.Sum(nil))}, "", "  ")
	if err := os.WriteFile(filepath.Join(out, "leaves.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
