package liftrs

import (
	"math/rand"
	"testing"
)

func TestCarryCriterionAgainstResidues(t *testing.T) {
	// The existing residue DP is deliberately retained as an independent oracle.
	// Its tiny memberships are themselves checked against exhaustive actual curves
	// in reference_test.go. Include m=3 and t close to q, beyond target parameters.
	for _, s := range []int{2, 3, 4} {
		for _, m := range []int{2, 3} {
			if s == 4 && m == 3 {
				continue
			}
			q := 1 << uint(s)
			for degree := 1; degree < q-1; degree++ {
				spec := Spec{s, m, degree}
				n := 1 << uint(s*m)
				for i := 0; i < n; i++ {
					exponents := coordinates(i, q, m)
					got, err := GoodMonomial(spec, exponents)
					if err != nil || got != goodMonomial(exponents, s, degree) {
						t.Fatalf("spec=%+v exponents=%v got=%v err=%v", spec, exponents, got, err)
					}
				}
			}
		}
	}
}

func TestCarryCriterionLargerFields(t *testing.T) {
	rng := rand.New(rand.NewSource(73))
	for _, s := range []int{5, 8, 10} {
		for _, m := range []int{2, 3} {
			for _, degree := range []int{1, 2, 5, 8} {
				for i := 0; i < 80; i++ {
					e := make([]uint32, m)
					for j := range e {
						e[j] = rng.Uint32() & ((1 << uint(s)) - 1)
					}
					got, err := GoodMonomial(Spec{s, m, degree}, e)
					if err != nil || got != goodMonomial(e, s, degree) {
						t.Fatalf("s=%d m=%d t=%d e=%v got=%v err=%v", s, m, degree, e, got, err)
					}
				}
			}
		}
	}
}

func TestGoodMonomialValidationAndEdges(t *testing.T) {
	for _, s := range []int{8, 12, 20} {
		for _, degree := range []int{1, 5, 8} {
			spec := Spec{s, 2, degree}
			q := uint32(1) << uint(s)
			for _, e := range [][]uint32{{0, 0}, {1, 0}, {q - 1, 0}, {q - 1, q - 1}} {
				got, err := GoodMonomial(spec, e)
				want := e[0] == 0 || e[0] == 1
				if err != nil || got != want {
					t.Fatalf("edge spec=%+v e=%v got=%v err=%v", spec, e, got, err)
				}
			}
		}
	}
	for _, test := range []struct {
		spec Spec
		e    []uint32
	}{
		{Spec{1, 2, 1}, []uint32{0, 0}}, {Spec{21, 2, 1}, []uint32{0, 0}},
		{Spec{4, 1, 1}, []uint32{0}}, {Spec{4, 4, 1}, []uint32{0, 0, 0, 0}},
		{Spec{4, 2, 0}, []uint32{0, 0}}, {Spec{4, 2, 15}, []uint32{0, 0}},
		{Spec{4, 2, 1}, []uint32{0}}, {Spec{4, 2, 1}, []uint32{16, 0}},
	} {
		if _, err := GoodMonomial(test.spec, test.e); err == nil {
			t.Fatalf("accepted %+v", test)
		}
	}
	e := []uint32{2, 2}
	_, _ = GoodMonomial(Spec{2, 2, 1}, e)
	if e[0] != 2 || e[1] != 2 {
		t.Fatal("mutated exponents")
	}
	if n := testing.AllocsPerRun(100, func() { _, _ = GoodMonomial(Spec{20, 2, 5}, []uint32{123, 456}) }); n != 0 {
		t.Fatalf("membership allocates: %v", n)
	}
}
