package liftrs_test

import (
	"fmt"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

func ExamplePreprocess() {
	// Tiny GF(4) fixture: code preprocessing knows no application data or keys.
	code, err := liftrs.Preprocess(liftrs.Spec{
		FieldDegree: 2,
		Dimension:   2,
		CurveDegree: 2,
	})
	if err != nil {
		panic(err)
	}
	message := make([]uint32, code.Dimension())
	for i := range message {
		message[i] = uint32(i % 4)
	}
	word, err := code.Encode(message)
	if err != nil {
		panic(err)
	}
	positions := code.SystematicPositions()
	recovered := make([]uint32, len(positions))
	for i, position := range positions {
		recovered[i] = word[position]
	}
	fmt.Println("Message:", message)
	fmt.Println("Recovered:", recovered)
	fmt.Println("Codeword length:", len(word))
	// Output:
	// Message: [0 1 2 3 0]
	// Recovered: [0 1 2 3 0]
	// Codeword length: 16
}
