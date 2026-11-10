//go:build liftrs_cpp && cgo

package liftrs_test

import (
	"fmt"
	liftrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"math/rand"
)

// Code-only preparation is reused; each encoding workspace belongs to one
// concurrent caller. This tiny fixture demonstrates the opt-in API, not security.
func ExampleScalingOnlinePlan_WithCPPMatrix() {
	basis, err := liftrs.DiscoverOrbit(liftrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 2}, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(1505))})
	if err != nil {
		panic(err)
	}
	prepared, err := basis.PrepareCyclicFFT()
	if err != nil {
		panic(err)
	}
	plan, err := prepared.ScalingPlan()
	if err != nil {
		panic(err)
	}
	encoder, err := plan.WithCPPMatrix(4)
	if err != nil {
		panic(err)
	}
	defer encoder.Close()
	work := encoder.NewWorkspace()
	message := make([]uint32, plan.Dimension())
	message[0] = 7
	word, err := encoder.Encode(message, work)
	if err != nil {
		panic(err)
	}
	pos, err := plan.SystematicPosition(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(word), word[pos])
	// Output: 256 7
}
