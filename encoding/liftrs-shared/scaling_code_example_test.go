package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func ExampleLoadScalingCode() {
	// This fixture uses the existing verified small-code preparation path.
	basis, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(170057))})
	if err != nil {
		panic(err)
	}
	encoder, err := basis.PrepareCyclicFFT()
	if err != nil {
		panic(err)
	}
	plan, err := encoder.ScalingPlan()
	if err != nil {
		panic(err)
	}
	var raw bytes.Buffer
	if err = WriteScalingOnline(&raw, plan); err != nil {
		panic(err)
	}
	hash := sha256.Sum256(raw.Bytes()) // Real loads take this from a trusted manifest.
	limits := ScalingCodeLimits{MaxElements: 256, MaxPayloadBytes: 4096}
	code, err := LoadScalingCode(bytes.NewReader(raw.Bytes()), hex.EncodeToString(hash[:]), limits)
	if err != nil {
		panic(err)
	}
	h := code.Header()
	fmt.Println(h.Spec.FieldDegree, h.N, h.ExtensionBytes, code.ResidentBytes() == plan.ResidentBytes())
	// Output: 4 256 4 true
}
