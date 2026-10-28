package liftrs

import (
	"fmt"
	"io"
	"os"
)

func ExampleLoadCyclicEncoderWithLimit() {
	// This is a trusted fixture; real local caches also retain provenance/checksums.
	file, err := os.Open("../../experiments/encoder-comparison/artifacts/s04-m2-t2-cyclic-fft.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	const limit int64 = 2 << 30
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		panic(err)
	}
	encoder, err := LoadCyclicEncoderWithLimit(raw, limit)
	if err != nil {
		panic(err)
	}
	fmt.Println(encoder.Basis().Length())
	// Output: 256
}
