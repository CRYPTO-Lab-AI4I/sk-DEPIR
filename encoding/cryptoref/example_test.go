package cryptoref_test

import (
	"fmt"

	"rme/encoding/cryptoref"
	"rme/encoding/liftrs"
)

func ExampleMasker_Encode() {
	code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2})
	if err != nil {
		panic(err)
	}
	message := []uint32{0, 1, 2, 3, 0}
	word, err := code.Encode(message)
	if err != nil {
		panic(err)
	}
	permutationKey, err := cryptoref.GenerateKey()
	if err != nil {
		panic(err)
	}
	permutation, err := cryptoref.NewPermutation(code.Length(), permutationKey)
	if err != nil {
		panic(err)
	}
	maskKey, err := cryptoref.GenerateKey() // Independent key for this database encoding.
	if err != nil {
		panic(err)
	}
	masker, err := cryptoref.NewMasker(2, code.Length(), maskKey)
	if err != nil {
		panic(err)
	}
	stored, err := masker.Encode(word, permutation)
	if err != nil {
		panic(err)
	}
	// Read systematic positions here to demonstrate the encoding boundary only.
	// A private PIR retrieval instead reads a locally decodable curve plus dummies.
	addresses, err := permutation.Map(code.SystematicPositions())
	if err != nil {
		panic(err)
	}
	values := make([]uint32, len(addresses))
	for i, address := range addresses {
		values[i] = stored[address]
	}
	recovered, err := masker.Unmask(addresses, values)
	if err != nil {
		panic(err)
	}
	fmt.Println(recovered)
	// Output: [0 1 2 3 0]
}
