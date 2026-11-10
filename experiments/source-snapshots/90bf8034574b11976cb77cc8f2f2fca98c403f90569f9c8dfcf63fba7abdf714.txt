package liftrs_test

import (
	"fmt"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
)

func ExampleNewQuadraticField() {
	polynomial, err := liftrs.BinaryPolynomial(20)
	if err != nil {
		panic(err)
	}
	base, err := liftrs.NewBinaryField(20, polynomial)
	if err != nil {
		panic(err)
	}
	extension, err := liftrs.NewQuadraticField(base)
	if err != nil {
		panic(err)
	}
	x := liftrs.QuadraticElement{A: 12345, B: 67890}
	inverse, err := extension.Inverse(x)
	if err != nil {
		panic(err)
	}
	fmt.Println(extension.Mul(x, inverse) == (liftrs.QuadraticElement{A: 1}))
	// Output: true
}
