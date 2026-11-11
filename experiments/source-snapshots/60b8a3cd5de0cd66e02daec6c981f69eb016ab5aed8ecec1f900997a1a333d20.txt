//go:build liftrs_cpp && cgo

package liftrs

import (
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"reflect"
	"testing"
)

// Cross the gather admission boundary and compare complete canonical transforms
// with the unchanged one-worker recurrence at small and full-width fields.
func TestCPPTaylorParallelMatrix(t *testing.T) {
	for _, s := range []int{8, 16} {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		for _, dimension := range []int{14, 15, 16} {
			plan, err := newMatrixCPP(f, dimension)
			if err != nil {
				t.Fatal(err)
			}
			n := 1 << uint(dimension)
			input := make([]uint32, n)
			scratch := make([]uint32, n)
			random := rand.New(rand.NewSource(int64(2302 + s + dimension)))
			for i := range input {
				input[i] = uint32(uint64(random.Uint32()) % f.q)
			}
			want := append([]uint32(nil), input...)
			if err = liftrscpp.ApplyMatrix(plan.rows, plan.columns, want, scratch, plan.split, 1, false); err != nil {
				t.Fatal(err)
			}
			for _, workers := range []int{2, 8, 32} {
				got := append([]uint32(nil), input...)
				if err = liftrscpp.ApplyMatrix(plan.rows, plan.columns, got, scratch, plan.split, workers, false); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatal("parallel forward differs", s, dimension, workers)
				}
				if err = liftrscpp.ApplyMatrix(plan.rows, plan.columns, got, scratch, plan.split, workers, true); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, input) {
					t.Fatal("parallel inverse differs", s, dimension, workers)
				}
			}
			plan.close()
		}
	}
}
