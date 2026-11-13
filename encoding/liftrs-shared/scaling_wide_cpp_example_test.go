//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func ExampleScalingCodeData_WithWideCPP() {
	basis, _ := DiscoverOrbit(Spec{3, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(170058))})
	prepared, err := basis.PrepareCyclicFFT()
	if err != nil {
		panic(err)
	}
	plan, _ := prepared.ScalingPlan()
	var raw bytes.Buffer
	WriteScalingOnline(&raw, plan)
	digest := sha256.Sum256(raw.Bytes()) // Real loads use an independently trusted manifest.
	data, err := LoadScalingCode(bytes.NewReader(raw.Bytes()), hex.EncodeToString(digest[:]), ScalingCodeLimits{64, 4096})
	if err != nil {
		panic(err)
	}
	encoder, err := data.WithWideCPP(2)
	if err != nil {
		panic(err)
	}
	defer encoder.Close()
	workspace, err := encoder.NewWorkspace(encoder.Resources().WorkspaceBytes)
	if err != nil {
		panic(err)
	}
	word, err := encoder.Encode(make([]uint32, plan.Dimension()), workspace)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(word), encoder.Resources().OutputBackingBytes)
	// Output: 64 512
}
