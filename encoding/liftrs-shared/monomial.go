package liftrs

import "fmt"

// GoodMonomial tests membership of a reduced monomial in Lift-RS_q(m,q-1,t),
// using the same criterion as reference preprocessing (paper Lemma 4.8 and
// Algorithm 1, pp. 27–28). It accepts s=2..20, m=2..3, 1<=t<q-1, and m
// exponents in [0,q). It needs no field arithmetic or generator matrix.
//
// Unlike Preprocess, this operation supports the target-size code specifications.
// Its O(s*m*t) time and O(1) auxiliary space do not depend on q^m. The caller
// retains ownership of exponents. Statistical sampling is separate from this
// deterministic membership test.
func GoodMonomial(spec Spec, exponents []uint32) (bool, error) {
	if spec.FieldDegree < 2 || spec.FieldDegree > 20 || spec.Dimension < 2 || spec.Dimension > 3 {
		return false, fmt.Errorf("monomial test requires s=2..20 and m=2..3")
	}
	q := uint32(1) << uint(spec.FieldDegree)
	if spec.CurveDegree < 1 || spec.CurveDegree >= int(q)-1 || len(exponents) != spec.Dimension {
		return false, fmt.Errorf("invalid curve degree or exponent count")
	}
	for _, exponent := range exponents {
		if exponent >= q {
			return false, fmt.Errorf("monomial exponent is outside the reduced range [0,q)")
		}
	}
	return goodMonomialCarries(exponents, spec.FieldDegree, spec.CurveDegree), nil
}

// At bit k, Algorithm 1 permits any contribution a_k in [0,t*w_k], where
// w_k counts exponent bits. A monomial is bad iff some POSITIVE sum
// sum_k a_k*2^k is a multiple j*(2^s-1), 1<=j<=m*t.
//
// Add j to both sides: a binary addition starts with carry c_0=j and must
// finish with c_s=j, with c_k+a_k=2*c_(k+1). If reachable carries form [lo,hi],
// their successors form [ceil(lo/2), floor((hi+t*w_k)/2)]. This is an interval:
// all integer a_k in the permitted range are available, including when w_k=0
// (only even incoming carries survive). An empty interval stays unreachable.
// Thus only two bounds per j are needed, rather than 2^s residue states.
// Excluding j=0 is essential: the constant monomial is good.
func goodMonomialCarries(exponents []uint32, s, t int) bool {
	var weights [20]int
	for k := 0; k < s; k++ {
		for _, e := range exponents {
			weights[k] += int((e >> uint(k)) & 1)
		}
		weights[k] *= t
	}
	for j := 1; j <= len(exponents)*t; j++ {
		lo, hi := j, j
		for k := 0; k < s; k++ {
			lo = (lo + 1) / 2
			hi = (hi + weights[k]) / 2
			if lo > hi {
				break
			}
		}
		if lo <= j && j <= hi {
			return false
		}
	}
	return true
}
