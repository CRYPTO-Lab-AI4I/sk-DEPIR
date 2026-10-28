// Package liftrs provides shared code specifications, preprocessing and complete
// encoding orchestration for Go and C++ transform execution.
package liftrs

import commonfield "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"

// BinaryField is the common polynomial-basis field implementation.
type BinaryField = commonfield.BinaryField

// QuadraticElement uses the shared a+b*u representation.
type QuadraticElement = commonfield.QuadraticElement

// QuadraticField is the common immutable quadratic extension.
type QuadraticField = commonfield.QuadraticField

// BinaryPolynomial returns the conventional irreducible polynomial.
func BinaryPolynomial(degree int) (uint32, error) { return commonfield.BinaryPolynomial(degree) }

// NewBinaryField validates and constructs a common binary field.
func NewBinaryField(degree int, polynomial uint32) (BinaryField, error) {
	return commonfield.NewBinaryField(degree, polynomial)
}

// NewQuadraticField constructs the common tower over base.
func NewQuadraticField(base BinaryField) (QuadraticField, error) {
	return commonfield.NewQuadraticField(base)
}
