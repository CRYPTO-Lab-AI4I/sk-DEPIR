//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"
)

func TestPreprocessingPolynomialStage(t *testing.T) {
	spec := Spec{4, 2, 3}
	session, err := NewCodePreprocessor(spec, preprocessingOptions(4, 2))
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	support, err := session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(171001))})
	if err != nil {
		t.Fatal(err)
	}
	indices := append([]int(nil), support.exponents...)
	poly, err := session.PreparePolynomials(support, PreparationReuse{})
	if err != nil {
		t.Fatal(err)
	}
	resident, err := session.Prepare(support, PreparationReuse{})
	if err != nil {
		t.Fatal(err)
	}
	if poly.Header() != resident.CodeData().Header() || !reflect.DeepEqual(poly.p, resident.code.p) || !reflect.DeepEqual(poly.generator, resident.generator) {
		t.Fatal("polynomial/resident stages differ")
	}
	if !reflect.DeepEqual(indices, support.exponents) {
		t.Fatal("support mutated")
	}
	var a, b bytes.Buffer
	if err = poly.WriteGenerator(&a); err != nil {
		t.Fatal(err)
	}
	if err = resident.WriteGenerator(&b); err != nil || !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatal("generator byte identity", err)
	}
	if err = session.WriteCode(io.Discard, poly, 8); err == nil {
		t.Fatal("implicit CPU export fallback")
	}
	session.Close()
	a.Reset()
	if err = poly.WriteGenerator(&a); err != nil || !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatal("polynomial data did not survive session", err)
	}
	if _, err = session.PreparePolynomials(support, PreparationReuse{}); err == nil {
		t.Fatal("closed session")
	}
	var absent *PreparedPolynomials
	if absent.Header() != (ScalingCodeHeader{}) || absent.WriteGenerator(io.Discard) == nil {
		t.Fatal("nil polynomials")
	}
}
