//go:build liftrs_cpp && cgo

package liftrs

import (
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

func TestCPPCompactOwnership(t *testing.T) {
	for _, spec := range []Spec{{6, 2, 4}, {8, 2, 5}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(3401))})
		if err != nil {
			t.Fatal(err)
		}
		code, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		p, err := code.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		e, err := p.WithCPPMatrix(4)
		if err != nil {
			t.Fatal(err)
		}
		w, err := e.NewCompactWorkspace()
		if err != nil {
			t.Fatal(err)
		}
		before := w.Bytes()
		firstPointer := &w.arena[0]
		var kept, saved []uint32
		for pass := 0; pass < 3; pass++ {
			input := make([]uint32, p.k)
			if pass != 1 {
				for i := range input {
					input[i] = uint32((17*i + pass + 1) % (1 << uint(spec.FieldDegree)))
				}
			}
			original := append([]uint32(nil), input...)
			want, err := p.Encode(input, p.NewWorkspace())
			if err != nil {
				t.Fatal(err)
			}
			got, st, err := e.EncodeCompactWithStats(input, w)
			if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(input, original) || st.Calls != 5 {
				t.Fatal("compact/reference word", spec, err)
			}
			if w.arena != nil {
				t.Fatal("returned arena retained by workspace")
			}
			if pass == 0 {
				kept = got
				saved = append([]uint32(nil), got...)
				if &got[0] != firstPointer || w.Bytes()+uint64(p.Length())*4 != before {
					t.Fatal("compact ownership accounting")
				}
			}
			if !reflect.DeepEqual(kept, saved) {
				t.Fatal("old compact output overwritten")
			}
		}
		bad := make([]uint32, p.k)
		bad[0] = 1 << uint(spec.FieldDegree)
		if _, _, err = e.EncodeCompactWithStats(bad, w); err == nil {
			t.Fatal("accepted field error")
		}
		if _, _, err = e.EncodeCompactWithStats(bad[:len(bad)-1], w); err == nil {
			t.Fatal("accepted length error")
		}
		input := make([]uint32, p.k)
		input[0] = 1
		want, err := p.Encode(input, p.NewWorkspace())
		if err != nil {
			t.Fatal(err)
		}

		got, _, err := e.EncodeCompactWithStats(input, w)
		if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(kept, saved) {
			t.Fatal("compact recovery/ownership", err)
		}
		if _, _, err = e.EncodeCompactWithStats(input, nil); err == nil {
			t.Fatal("nil compact workspace")
		}
		// A short tile must fail before native execution, not touch caller data.
		l := len(e.source.NewWorkspace().x)
		prefix, full := e.plans[l], e.plans[p.Length()]
		if _, err := cpp.EncodeOnlineCompact(prefix.rows, prefix.columns, full.rows, full.columns, prefix.split, full.split, 4, p.basis.omega, input, p.p, p.spectrum, make([]uint32, p.Length()), make([]uint32, 1), l); err == nil {
			t.Fatal("short column workspace")
		}
		alias := make([]uint32, p.Length())
		if _, err := cpp.EncodeOnlineCompact(prefix.rows, prefix.columns, full.rows, full.columns, prefix.split, full.split, 4, p.basis.omega, input, p.p, alias, alias, w.columns, l); err == nil {
			t.Fatal("compact accepted spectrum/arena overlap")
		}
		e.Close()
		if !reflect.DeepEqual(kept, saved) {
			t.Fatal("close invalidated output")
		}
	}
}

func TestCPPCompactShape(t *testing.T) {
	n, err := cpp.BoundedColumnElements(1<<32, 16, 32)
	if err != nil || uint64(n)*4 != 512<<20 {
		t.Fatal("full compact tile accounting", n, err)
	}
	if _, err = cpp.BoundedColumnElements(16, 4, 4); err == nil {
		t.Fatal("bad split")
	}
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(3402))})
	if err != nil {
		t.Fatal(err)
	}
	code, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	p, err := code.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	e, err := p.WithCPPMatrix(4)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if _, err = e.NewCompactWorkspace(); err == nil {
		t.Fatal("compact accepted prefix larger than arena")
	}
}

func TestCPUCompactCached(t *testing.T) {
	out, cache := os.Getenv("PIR_CPU_COMPACT_OUT"), os.Getenv("PIR_CPU_COMPACT_CACHE")
	if out == "" {
		t.Skip("opt-in bounded CPU compact workflow")
	}
	if cache == "" {
		t.Fatal("cache required")
	}
	spec := Spec{16, 2, 5}
	repeats := 1
	limit := int64(44 << 30)
	if os.Getenv("PIR_CPU_COMPACT_S14") == "1" {
		spec = Spec{14, 2, 5}
		repeats = 2
		limit = 8 << 30
	}
	if os.Getenv("PIR_CPU_COMPACT_SMOKE") == "1" {
		spec = Spec{6, 2, 4}
		repeats = 2
		limit = 1 << 30
	}
	compact := os.Getenv("PIR_CPU_COMPACT_MODE") == "compact"
	if !compact {
		limit = 64 << 30
	}
	cfg := scalingFirstUseConfig{spec: spec, out: out, cache: cache, workers: 32, repeats: repeats, memoryLimit: limit, onlineOnly: true}
	cfg.onlineSetup = func(p *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
		e, err := p.WithCPPMatrix(32)
		if err != nil {
			return nil, nil, nil, err
		}
		if compact {
			w, err := e.NewCompactWorkspace()
			if err != nil {
				e.Close()
				return nil, nil, nil, err
			}
			return func(input []uint32) ([]uint32, map[string]any, error) {
				before := w.Bytes()
				word, st, err := e.EncodeCompactWithStats(input, w)
				return word, map[string]any{"native_stages": st.ScalingCPPStats, "compaction_ns": st.CompactionNS, "expansion_ns": st.ExpansionNS, "workspace_before_bytes": before, "workspace_after_bytes": w.Bytes(), "arena_transferred": w.arena == nil}, err
			}, func() { e.Close() }, map[string]any{"workers": 32, "compact": true, "workspace_bytes": w.Bytes(), "column_bytes": uint64(len(w.columns)) * 4}, nil
		}
		w := e.NewWorkspace()
		return func(input []uint32) ([]uint32, map[string]any, error) {
			before := w.Bytes()
			word, st, err := e.EncodeWithStats(input, w)
			return word, map[string]any{"native_stages": st, "workspace_before_bytes": before, "workspace_after_bytes": w.Bytes(), "arena_transferred": false}, err
		}, func() { e.Close() }, map[string]any{"workers": 32, "compact": false, "workspace_bytes": w.Bytes()}, nil
	}
	runScalingFirstUse(t, cfg)
}
