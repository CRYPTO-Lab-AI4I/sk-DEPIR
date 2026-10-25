package pir

import (
	"testing"

	"rme/encoding/liftrs"
	"rme/utils"
)

// Exercise the real online client at every systematic position of a genuine
// Lift-RS codeword. x^8*y^8 is beyond degree-7 RM over GF(16), but on a
// degree-2 curve its reachable exponents are 0,8,16,24,32, none a positive
// multiple of 15. Thus it satisfies the q-1 local sum constraint.
func TestReferenceLiftedRSEncoderOnlineRecovery(t *testing.T) {
	const q = 16
	code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 2})
	if err != nil {
		t.Fatal(err)
	}
	gf := utils.NewGF2n(4)
	evaluate := func(position int) uint32 {
		x, y := utils.SingleIndexToRowCol(q, position)
		return gf.Mul(gf.Pow(uint32(x), 8), gf.Pow(uint32(y), 8))
	}
	positions := code.SystematicPositions()
	message := make([]uint32, code.Dimension())
	for j, position := range positions {
		message[j] = evaluate(position)
	}
	word, err := code.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	matrix := &Matrix2D{data: make([][]int, q), q: q}
	for x := range matrix.data {
		matrix.data[x] = make([]int, q)
		for y := range matrix.data[x] {
			position := x + q*y
			if word[position] != evaluate(position) {
				t.Fatalf("position %d: systematic encoder did not reproduce lifted monomial", position)
			}
			matrix.data[x][y] = int(word[position])
		}
	}
	scheme := NewLiftedRS(Params{Q: q, M: 2, K: 2})
	scheme.Gen()
	if err := scheme.permuteAndEncryptMatrix(0, matrix); err != nil {
		t.Fatal(err)
	}
	for j, position := range positions {
		query := scheme.GenerateClientQuery(position, 128)
		response := make([]int, len(query.Query))
		// Storage performs plain indexed reads, retaining duplicate positions.
		for i, address := range query.Query {
			response[i] = matrix.GetByIndex(address)
		}
		got, err := scheme.Decode(QueryResult{
			Values: response, Query: query.Query, RealPositions: query.RealPositions,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0] != int(message[j]) {
			t.Fatalf("message row %d at position %d: got %v, want %d", j, position, got, message[j])
		}
	}
}
