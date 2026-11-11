//go:build liftrs_cpp && cgo

package liftrs

import (
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

// Retaining old results is valid even when later calls (including failed or
// hybrid calls) reuse a workspace. Only caller-owned words add retained memory.
func TestCPPOutputOwnership(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}} {
		basis, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(3001))})
		if err != nil {
			t.Fatal(err)
		}
		code, err := basis.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		plan, err := code.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		native, err := plan.WithCPPMatrix(4)
		if err != nil {
			t.Fatal(err)
		}
		work := native.NewWorkspace()
		initial := work.Bytes()
		message := make([]uint32, plan.k)
		for i := range message {
			message[i] = uint32(i % (1 << uint(spec.FieldDegree)))
		}
		original := append([]uint32(nil), message...)
		scratch := &work.scratch[0]
		word, err := native.Encode(message, work)
		if err != nil {
			t.Fatal(err)
		}
		saved := append([]uint32(nil), word...)
		if &word[0] != scratch || work.scratch != nil || work.Bytes()+uint64(plan.Length())*4 != initial {
			t.Fatal("output not transferred from workspace")
		}
		for _, hybrid := range []bool{false, true, false} {
			input := make([]uint32, plan.k)
			var next []uint32
			if hybrid {
				next, _, err = native.EncodeHybridWithStats(input, work)
			} else {
				next, err = native.Encode(input, work)
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(word, saved) {
				t.Fatal("previous output overwritten")
			}
			if !reflect.DeepEqual(next, make([]uint32, plan.Length())) {
				t.Fatal("zero message output")
			}
		}
		bad := append([]uint32(nil), message...)
		bad[0] = 1 << uint(spec.FieldDegree)
		if _, err = native.Encode(bad, work); err == nil {
			t.Fatal("field error not rejected")
		}
		if _, err = native.Encode(message[:len(message)-1], work); err == nil {
			t.Fatal("length error not rejected")
		}
		next, err := native.Encode(message, work)
		if err != nil || !reflect.DeepEqual(next, saved) || !reflect.DeepEqual(word, saved) || !reflect.DeepEqual(message, original) {
			t.Fatal("recovery/ownership/input mutation", err)
		}
		native.Close()
		if !reflect.DeepEqual(word, saved) {
			t.Fatal("close invalidated owning output")
		}
	}
}

// Retain the prior separate-output schedule for a matched comparison; arithmetic
// and all five native transforms are identical to the candidate path.
func cpuSeparateOutput(e *ScalingCPPEncoder, input []uint32, w *ScalingCPPWorkspace) ([]uint32, ScalingCPPStats, error) {
	p := e.source
	l := len(w.work.x)
	prefix, full := e.plans[l], e.plans[p.Length()]
	word := make([]uint32, p.Length())
	st, err := liftrscpp.EncodeOnline(prefix.rows, prefix.columns, full.rows, full.columns, prefix.split, full.split, e.workers, p.basis.omega,
		input, p.p, p.spectrum, w.work.x[:cap(w.work.x)], w.work.h, w.scratch, word, l)
	return word, ScalingCPPStats{Calls: 5, BoundaryCalls: 1, TaylorNS: st.TaylorNS, ColumnsNS: st.ColumnsNS, RowsNS: st.RowsNS, BufferNS: st.BufferNS, PointwiseNS: st.PointwiseNS, AssemblyNS: st.AssemblyNS}, err
}

func TestCPUMemoryCached(t *testing.T) {
	out, cache := os.Getenv("PIR_CPU_MEMORY_OUT"), os.Getenv("PIR_CPU_MEMORY_CACHE")
	if out == "" {
		t.Skip("opt-in CPU output/scratch reuse workflow")
	}
	if cache == "" {
		t.Fatal("persistent cache required")
	}
	spec := Spec{16, 2, 5}
	limit := int64(64 << 30)
	if os.Getenv("PIR_CPU_MEMORY_SMALL") == "1" {
		spec = Spec{10, 2, 5}
		limit = 1 << 30
	}
	if os.Getenv("PIR_CPU_MEMORY_SMOKE") == "1" {
		spec = Spec{6, 2, 4}
		limit = 1 << 30
	}
	separate := os.Getenv("PIR_CPU_MEMORY_BASELINE") == "1"
	cfg := scalingFirstUseConfig{spec: spec, out: out, cache: cache, workers: 32, repeats: 2, memoryLimit: limit, onlineOnly: true}
	cfg.onlineSetup = func(p *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
		encoder, err := p.WithCPPMatrix(32)
		if err != nil {
			return nil, nil, nil, err
		}
		w := encoder.NewWorkspace()
		encode := func(input []uint32) ([]uint32, map[string]any, error) {
			before := w.Bytes()
			var word []uint32
			var st ScalingCPPStats
			var err error
			if separate {
				word, st, err = cpuSeparateOutput(encoder, input, w)
			} else {
				word, st, err = encoder.EncodeWithStats(input, w)
			}
			return word, map[string]any{"native_stages": st, "workspace_before_bytes": before, "workspace_after_bytes": w.Bytes(), "scratch_transferred": !separate && w.scratch == nil}, err
		}
		return encode, func() { encoder.Close() }, map[string]any{"workers": 32, "separate_output": separate, "workspace_bytes": w.Bytes()}, nil
	}
	runScalingFirstUse(t, cfg)
}
