package pir

import (
	"path/filepath"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/utils"
)

func TestLiftedRSDecodeThroughPIRInterface(t *testing.T) {
	const q = 31
	var scheme PIR = NewLiftedRS(Params{Q: q, K: 2, M: 2})
	scheme.Gen()

	lifted := scheme.(*LiftedRS)
	query := []int{3, 7}
	want := 11

	values := []int{
		lifted.encryptor.Encrypt(query[0], 5),
		lifted.encryptor.Encrypt(query[1], q-want-5),
	}
	got, err := scheme.Decode(QueryResult{Values: values, Query: query, RealPositions: query})
	if err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("Decode() = %v, want [%d]", got, want)
	}
}

func TestPIRDecodeRejectsMissingRealPositions(t *testing.T) {
	var scheme PIR = NewLiftedRS(Params{Q: 31, K: 2, M: 2})
	if _, err := scheme.Decode(QueryResult{Values: []int{1}, Query: []int{1}}); err == nil {
		t.Fatal("Decode() returned nil error without real positions")
	}
}

func TestPIRRMEncodingRequiresKeys(t *testing.T) {
	var scheme PIR = NewLiftedRS(Params{Q: 31, K: 2, M: 2})
	if _, err := scheme.Encode("unused-input", "unused-output"); err == nil {
		t.Fatal("Encode() returned nil error before key generation")
	}
}

func TestLiftedRSEndToEndWithGF2nCodeword(t *testing.T) {
	const (
		n           = 4
		q           = 1 << n
		curveDegree = 2
	)
	gf := utils.NewGF2n(n)
	// A total-degree-2 polynomial has degree at most four on a degree-2
	// query curve, below q-1. The Lift-RS line-sum identity therefore applies.
	evalCodeword := func(x, y uint32) uint32 {
		return 7 ^ gf.Mul(3, x) ^ gf.Mul(5, y) ^
			gf.Mul(11, gf.Mul(x, x)) ^ gf.Mul(13, gf.Mul(x, y))
	}

	codeword := make([][]int, q)
	for x := 0; x < q; x++ {
		codeword[x] = make([]int, q)
		for y := 0; y < q; y++ {
			codeword[x][y] = int(evalCodeword(uint32(x), uint32(y)))
		}
	}

	scheme := NewLiftedRS(Params{Q: q, K: curveDegree, M: 2})
	scheme.Gen()
	encoded := &Matrix2D{data: codeword, q: q}
	if err := scheme.permuteAndEncryptMatrix(0, encoded); err != nil {
		t.Fatalf("permute and mask Lift-RS codeword: %v", err)
	}
	serverDatabase := filepath.Join(t.TempDir(), "lift-rs-encoded.csv")
	if err := utils.Write2DMatrixToFile(codeword, serverDatabase); err != nil {
		t.Fatalf("write encoded server database: %v", err)
	}

	indices := []int{0, 1, 7, 3*q + 5, q*q - 1}
	for _, index := range indices {
		values, query, realPositions := scheme.QueryLocalLiftedRS(index, serverDatabase)
		got, err := scheme.Decode(QueryResult{Values: values, Query: query, RealPositions: realPositions})
		if err != nil {
			t.Fatalf("index %d: Decode() returned error: %v", index, err)
		}
		x, y := utils.SingleIndexToRowCol(q, index)
		want := int(evalCodeword(uint32(x), uint32(y)))
		if len(got) != 1 || got[0] != want {
			t.Fatalf("index %d: Decode() = %v, want [%d]", index, got, want)
		}
	}
}

func TestLiftedRSRejectsUnsupportedCompositeFieldSize(t *testing.T) {
	scheme := NewLiftedRS(Params{Q: 15, K: 2, M: 2})
	scheme.Gen()
	defer func() {
		if recover() == nil {
			t.Fatal("GenerateClientQuery accepted composite q=15")
		}
	}()
	scheme.GenerateClientQuery(0, 0)
}
