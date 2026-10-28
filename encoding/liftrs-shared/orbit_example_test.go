package liftrs_test

import (
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
)

func ExampleOrbitBasis() {
	basis, err := liftrs.DiscoverOrbit(liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 1}, liftrs.OrbitOptions{})
	if err != nil {
		panic(err)
	}
	prepared, err := basis.PrepareCyclicFFT()
	if err != nil {
		panic(err)
	}
	public, err := prepared.MarshalBinary()
	if err != nil {
		panic(err)
	}
	// Publish public once; a fresh process loads it once and reuses its constants.
	encoder, err := liftrs.LoadCyclicEncoder(public)
	if err != nil {
		panic(err)
	}
	message := make([]uint32, basis.Dimension())
	for i := range message {
		message[i] = uint32(i % 4)
	}
	word, err := encoder.Encode(message)
	if err != nil {
		panic(err)
	}
	position, err := basis.SystematicPosition(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(word), word[position] == message[0])
	// Output: 16 true
}
