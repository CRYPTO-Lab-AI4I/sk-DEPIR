//go:build liftrs_cpp && cgo

package liftrscpp

import (
	"fmt"
	field "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"
)

func ExampleNewWideMatrix() {
	polynomial, _ := field.BinaryPolynomial(17)
	plan, err := NewWideMatrix(17, polynomial, 4, 2) //16 points, no full s17 domain.
	if err != nil {
		panic(err)
	}
	defer plan.Close()
	input := make([]uint64, 16)
	scratch := make([]uint64, 16)
	input[0] = 1 << 33
	if _, err = plan.Apply(input, scratch, 2, false); err != nil {
		panic(err)
	}
	if _, err = plan.Apply(input, scratch, 2, true); err != nil {
		panic(err)
	}
	fmt.Println(input[0])
	// Output: 8589934592
}
