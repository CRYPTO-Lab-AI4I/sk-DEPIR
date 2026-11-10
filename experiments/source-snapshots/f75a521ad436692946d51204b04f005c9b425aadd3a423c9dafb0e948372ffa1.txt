package liftrs_test

import (
	"fmt"
	liftrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"math/rand"
)

func ExampleOrbitBasis_PrepareCyclicScalingPrototype() {
	basis, err := liftrs.DiscoverOrbit(liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 1}, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(303))})
	if err != nil {
		panic(err)
	}
	prepared, err := basis.PrepareCyclicScalingPrototype()
	if err != nil {
		panic(err)
	}
	online, err := prepared.ScalingPlan()
	if err != nil {
		panic(err)
	}
	// Workspace ownership is explicit; use one workspace per concurrent call.
	workspace := online.NewWorkspace()
	word, err := online.Encode(make([]uint32, basis.Dimension()), workspace)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(word))
	// Output: 16
}
