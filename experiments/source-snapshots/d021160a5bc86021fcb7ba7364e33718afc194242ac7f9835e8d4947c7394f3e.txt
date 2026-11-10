//go:build liftrs_cpp && cgo

package liftrs

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func scalingNativeProbe(t *testing.T, s int, phase func(string, func() map[string]any), baseline *CyclicEncoder, plan *ScalingOnlinePlan) {
	var native cppPlan
	phase(fmt.Sprintf("native-plan-s%d", s), func() map[string]any {
		var err error
		var description uint64
		native, description, err = makeCPPPlan(plan.basis.extensionFFT)
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{"description_bytes": description, "native_owned_bytes": native.OwnedBytes()}
	})
	defer native.Close()
	var stats CPPEncodeStats
	plan.basis.extensionFFT.executor = &meteredCPP{native, &stats}
	reference, err := baseline.WithCPPTransforms()
	if err != nil {
		t.Fatal(err)
	}
	defer reference.Close()
	random := rand.New(rand.NewSource(903 + int64(s)))
	message := make([]uint32, plan.k)
	for i := range message {
		message[i] = uint32(random.Intn(int(plan.basis.base.q)))
	}
	var want []uint32
	phase(fmt.Sprintf("baseline-native-encode-s%d", s), func() map[string]any {
		var err error
		want, err = reference.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		return map[string]any{}
	})
	var workspace *ScalingWorkspace
	phase(fmt.Sprintf("workspace-allocation-s%d", s), func() map[string]any {
		workspace = plan.NewWorkspace()
		return map[string]any{"workspace_payload_bytes": workspace.Bytes(), "plan_array_bytes": plan.ResidentBytes(), "input_bytes": 4 * len(message), "output_bytes": 4 * plan.basis.Length()}
	})
	for repeat := 0; repeat < 3; repeat++ {
		stats = CPPEncodeStats{}
		phase(fmt.Sprintf("workspace-native-encode-s%d-r%d", s, repeat), func() map[string]any {
			got, err := plan.Encode(message, workspace)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("native workspace mismatch", err)
			}
			return map[string]any{"equal_codeword": true, "native": stats, "note": "time includes full word comparison; does not replace the s13 latency baseline"}
		})
	}
}
func TestScalingNativeWorkspace(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(904))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicScalingPrototype()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := e.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	native, _, err := makeCPPPlan(plan.basis.extensionFFT)
	if err != nil {
		t.Fatal(err)
	}
	defer native.Close()
	var stats CPPEncodeStats
	plan.basis.extensionFFT.executor = &meteredCPP{native, &stats}
	for repeat := 0; repeat < 3; repeat++ {
		message := make([]uint32, plan.k)
		for i := range message {
			message[i] = uint32((i + repeat) % 16)
		}
		want, err := e.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		got, err := plan.Encode(message, plan.NewWorkspace())
		if err != nil || !reflect.DeepEqual(want, got) {
			t.Fatal("native word", err)
		}
	}
	if stats.Calls != 15 {
		t.Fatal("expected five native calls per prototype encoding", stats)
	}
}
