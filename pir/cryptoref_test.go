package pir

import (
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/cryptoref"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/utils"
)

// Feed genuine Lift-RS messages through the new XOR-mask stage and the existing
// curve-query generator. Two separate databases and a fresh re-encoding share
// the same permutation/query but never a masking key. Storage does plain reads.
func TestCryptographicReferenceLiftedRSRecovery(t *testing.T) {
	const q = 16
	code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 2})
	if err != nil {
		t.Fatal(err)
	}
	permutationKey, err := cryptoref.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	permutation, err := cryptoref.NewPermutation(code.Length(), permutationKey)
	if err != nil {
		t.Fatal(err)
	}
	legacyPermutation, err := utils.NewPermutator(code.Length(), permutationKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	scheme := NewLiftedRS(Params{Q: q, M: 2, K: 2})
	scheme.permutator = &legacyPermutation
	gf := utils.NewGF2n(4)
	positions := code.SystematicPositions()
	messageA := make([]uint32, code.Dimension())
	messageB := make([]uint32, code.Dimension())
	for j, position := range positions {
		x, y := utils.SingleIndexToRowCol(q, position)
		messageA[j] = gf.Mul(gf.Pow(uint32(x), 8), gf.Pow(uint32(y), 8))
		messageB[j] = gf.Mul(3, messageA[j]) ^ 7
	}
	type database struct {
		message []uint32
		masker  *cryptoref.Masker
		stored  []uint32
	}
	var databases []database
	for _, message := range [][]uint32{messageA, messageB, messageA} {
		word, err := code.Encode(message)
		if err != nil {
			t.Fatal(err)
		}
		maskKey, err := cryptoref.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}
		masker, err := cryptoref.NewMasker(4, code.Length(), maskKey)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := masker.Encode(word, permutation)
		if err != nil {
			t.Fatal(err)
		}
		databases = append(databases, database{message, masker, stored})
	}
	recoverQuery := func(j int, query ClientQuery) {
		t.Helper()
		for d, db := range databases {
			response := make([]uint32, len(query.Query))
			// Each database receives its own full query, including repetitions.
			for i, address := range query.Query {
				response[i] = db.stored[address]
			}
			plain, err := db.masker.Unmask(query.Query, response)
			if err != nil {
				t.Fatal(err)
			}
			byPosition := make(map[int]uint32, len(query.Query))
			for i, address := range query.Query {
				byPosition[address] = plain[i]
			}
			var got uint32
			// Multiplicity belongs to the curve; never deduplicate this sum.
			for _, address := range query.RealPositions {
				got ^= byPosition[address]
			}
			if got != db.message[j] {
				t.Fatalf("database %d message %d: recovered %d want %d", d, j, got, db.message[j])
			}
		}
	}
	for j, position := range positions {
		recoverQuery(j, scheme.GenerateClientQuery(position, 128))
		// Explicit degree-2 curve x(z)=x_target+z^2+z, y(z)=y_target
		// ensures repeated curve positions instead of relying on random collisions.
		x, y := utils.SingleIndexToRowCol(q, position)
		curve := make([]int, 0, q-1)
		for z := uint32(1); z < q; z++ {
			cx := uint32(x) ^ gf.Mul(z, z) ^ z
			curve = append(curve, int(cx)+q*y)
		}
		real, err := permutation.Map(curve)
		if err != nil {
			t.Fatal(err)
		}
		query := append([]int(nil), real...)
		for i := 0; i < 128; i++ {
			query = append(query, i%code.Length())
		}
		recoverQuery(j, ClientQuery{Query: query, RealPositions: real})
	}
}
