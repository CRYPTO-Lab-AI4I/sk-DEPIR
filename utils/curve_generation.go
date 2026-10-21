package utils

import "math/rand"

// EvaluatePolynomial evaluates a polynomial using Horner's method.
// coeffs[k] is the coefficient of x^k (ascending order).
func EvaluatePolynomial(coeffs []int, x, q int) int {
	result := 0
	for k := len(coeffs) - 1; k >= 0; k-- {
		result = (result*x + coeffs[k]) % q
	}
	return result
}

// GenerateRandomPolynomialCoeffs generates a random degree-d polynomial over F_q
// whose value at t equals a.
func GenerateRandomPolynomialCoeffs(degree, q, t, a int) []int {
	coeffs := make([]int, degree+1)
	for i := range coeffs {
		coeffs[i] = rand.Intn(q)
	}
	// Adjust constant term so p(t) = a.
	coeffs[0] = (coeffs[0] - (EvaluatePolynomial(coeffs, t, q) - a) + q) % q
	return coeffs
}

// buildDiffTable computes the Newton forward difference table for polynomial p at x=0.
// Returns state where state[k] = Δ^k p(0).
// Initialization cost is O(degree²); used once per polynomial before the hot loop.
func buildDiffTable(coeffs []int, degree, q int) []int {
	state := make([]int, degree+1)
	// Evaluate p at 0, 1, …, degree via Horner.
	for x := 0; x <= degree; x++ {
		v := coeffs[degree]
		for k := degree - 1; k >= 0; k-- {
			v = (v*x + coeffs[k]) % q
		}
		state[x] = v
	}
	// Convert to forward differences in-place: state[k] becomes Δ^k p(0).
	for k := 1; k <= degree; k++ {
		for j := degree; j >= k; j-- {
			state[j] = (state[j] - state[j-1] + q) % q
		}
	}
	return state
}

// advanceDiffState advances the difference table by one evaluation step.
// After the call, state[0] = p(x+1) if state was at position x.
// Hot path: only additions, no multiplications or divisions.
func advanceDiffState(state []int, q int) {
	for k := 0; k < len(state)-1; k++ {
		v := state[k] + state[k+1]
		if v >= q {
			v -= q
		}
		state[k] = v
	}
}

// GenerateCurvePoints generates q-1 query positions for a 2D database of size q×q.
// Uses Newton forward differences: O(degree²) setup + O(q·degree) additions only (no mults).
func GenerateCurvePoints(degree, q, i int) []int {
	x, y := SingleIndexToRowCol(q, i)
	t := rand.Intn(q)
	uniX := GenerateRandomPolynomialCoeffs(degree, q, t, x)
	uniY := GenerateRandomPolynomialCoeffs(degree, q, t, y)

	stateX := buildDiffTable(uniX, degree, q)
	stateY := buildDiffTable(uniY, degree, q)

	points := make([]int, 0, q-1)
	for z := 0; z < q; z++ {
		if z != t {
			// RowColToSingleIndex(q, stateX[0], stateY[0]) = stateY[0]*q + stateX[0]
			points = append(points, stateY[0]*q+stateX[0])
		}
		if z < q-1 {
			advanceDiffState(stateX, q)
			advanceDiffState(stateY, q)
		}
	}
	return points
}

// GenerateCurvePoints3D generates q-1 query positions for a 3D database of size q³.
func GenerateCurvePoints3D(degree, q, i int) []int {
	x, y, z := SingleIndexToRowCol3D(q, i)
	t := rand.Intn(q)
	uniX := GenerateRandomPolynomialCoeffs(degree, q, t, x)
	uniY := GenerateRandomPolynomialCoeffs(degree, q, t, y)
	uniZ := GenerateRandomPolynomialCoeffs(degree, q, t, z)

	stateX := buildDiffTable(uniX, degree, q)
	stateY := buildDiffTable(uniY, degree, q)
	stateZ := buildDiffTable(uniZ, degree, q)

	q2 := q * q
	points := make([]int, 0, q-1)
	for ev := 0; ev < q; ev++ {
		if ev != t {
			// RowColToSingleIndex3D(q, stateX[0], stateY[0], stateZ[0]) = stateZ*q²+stateY*q+stateX
			points = append(points, stateZ[0]*q2+stateY[0]*q+stateX[0])
		}
		if ev < q-1 {
			advanceDiffState(stateX, q)
			advanceDiffState(stateY, q)
			advanceDiffState(stateZ, q)
		}
	}
	return points
}

// GenerateCurvePoints4D generates q-1 query positions for a 4D database of size q⁴.
func GenerateCurvePoints4D(degree, q, i int) []int {
	x, y, z, v := SingleIndexToRowCol4D(q, i)
	t := rand.Intn(q)
	uniX := GenerateRandomPolynomialCoeffs(degree, q, t, x)
	uniY := GenerateRandomPolynomialCoeffs(degree, q, t, y)
	uniZ := GenerateRandomPolynomialCoeffs(degree, q, t, z)
	uniV := GenerateRandomPolynomialCoeffs(degree, q, t, v)

	stateX := buildDiffTable(uniX, degree, q)
	stateY := buildDiffTable(uniY, degree, q)
	stateZ := buildDiffTable(uniZ, degree, q)
	stateV := buildDiffTable(uniV, degree, q)

	q2 := q * q
	q3 := q2 * q
	points := make([]int, 0, q-1)
	for ev := 0; ev < q; ev++ {
		if ev != t {
			points = append(points, stateV[0]*q3+stateZ[0]*q2+stateY[0]*q+stateX[0])
		}
		if ev < q-1 {
			advanceDiffState(stateX, q)
			advanceDiffState(stateY, q)
			advanceDiffState(stateZ, q)
			advanceDiffState(stateV, q)
		}
	}
	return points
}

// GenerateCurvePointsWithNoise generates the lifted-RS query for the Lifted RS construction.
// Returns (allPoints, realPoints):
//   - allPoints: q-1+noiseCount positions shuffled together, sent to the server.
//   - realPoints: the q-1 genuine curve positions, used by the client for decoding.
func GenerateCurvePointsWithNoise(degree, q, i, noiseCount int) ([]int, []int) {
	x, y := SingleIndexToRowCol(q, i)
	t := rand.Intn(q)
	uniX := GenerateRandomPolynomialCoeffs(degree, q, t, x)
	uniY := GenerateRandomPolynomialCoeffs(degree, q, t, y)

	stateX := buildDiffTable(uniX, degree, q)
	stateY := buildDiffTable(uniY, degree, q)

	real := make([]int, 0, q-1)
	for z := 0; z < q; z++ {
		if z != t {
			real = append(real, stateY[0]*q+stateX[0])
		}
		if z < q-1 {
			advanceDiffState(stateX, q)
			advanceDiffState(stateY, q)
		}
	}

	// Append noiseCount uniform random positions.
	qSq := q * q
	all := make([]int, len(real)+noiseCount)
	copy(all, real)
	for k := len(real); k < len(all); k++ {
		all[k] = rand.Intn(qSq)
	}
	rand.Shuffle(len(all), func(a, b int) { all[a], all[b] = all[b], all[a] })

	return all, real
}

// GenerateCurvePointsGF2n generates q−1 query positions for a 2D GF(2^n) database.
// The curve is a pair of random degree-k polynomials (p_x, p_y) over GF(2^n) with
// p_x(t) = x_target and p_y(t) = y_target for a random evaluation parameter t.
func GenerateCurvePointsGF2n(gf *GF2n, degree, i int) []int {
	q := int(gf.Q)
	x, y := SingleIndexToRowCol(q, i)
	t := rand.Intn(q)

	uniX := gf.RandPolyWithValue(degree, uint32(t), uint32(x))
	uniY := gf.RandPolyWithValue(degree, uint32(t), uint32(y))

	points := make([]int, 0, q-1)
	for z := 0; z < q; z++ {
		if z == t {
			continue
		}
		fx := gf.EvalPoly(uniX, uint32(z))
		fy := gf.EvalPoly(uniY, uint32(z))
		points = append(points, int(fy)*q+int(fx))
	}
	return points
}

// GenerateCurvePointsGF2n3D generates q−1 query positions for a 3D GF(2^n) database.
func GenerateCurvePointsGF2n3D(gf *GF2n, degree, i int) []int {
	q := int(gf.Q)
	x, y, z := SingleIndexToRowCol3D(q, i)
	t := rand.Intn(q)

	uniX := gf.RandPolyWithValue(degree, uint32(t), uint32(x))
	uniY := gf.RandPolyWithValue(degree, uint32(t), uint32(y))
	uniZ := gf.RandPolyWithValue(degree, uint32(t), uint32(z))

	q2 := q * q
	points := make([]int, 0, q-1)
	for ev := 0; ev < q; ev++ {
		if ev == t {
			continue
		}
		fx := gf.EvalPoly(uniX, uint32(ev))
		fy := gf.EvalPoly(uniY, uint32(ev))
		fz := gf.EvalPoly(uniZ, uint32(ev))
		points = append(points, int(fz)*q2+int(fy)*q+int(fx))
	}
	return points
}
