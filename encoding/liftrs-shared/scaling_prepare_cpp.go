//go:build liftrs_cpp && cgo

package liftrs

import (
	"fmt"
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/bits"
)

// scalingPreparationCPP is private, single-caller code-preprocessing state.
// Matrix plans cover every convolution power of two; one transpose buffer is
// reused. Smaller transforms borrow its prefix. It never retains database data.
// Close only after discovery/preparation; a basis referencing it is not usable
// afterward except as serialized public code data. Default constructors are intact.
type scalingPreparationCPP struct {
	plans       map[int]*matrixCPPPlan
	small, base *liftrscpp.Plan
	scratch     []uint32
	workers     int
}

func (c *scalingPreparationCPP) apply(data []uint32, inverse bool) {
	if len(data) < 4 {
		if _, err := c.small.Apply(data, inverse); err != nil {
			panic(cppFailure{err})
		}
		return
	}
	p := c.plans[len(data)]
	if p == nil {
		panic(cppFailure{fmt.Errorf("unprepared preprocessing transform length")})
	}
	if err := liftrscpp.ApplyMatrix(p.rows, p.columns, data, c.scratch[:len(data)], p.split, c.workers, inverse); err != nil {
		panic(cppFailure{err})
	}
}
func (c *scalingPreparationCPP) close() {
	for _, p := range c.plans {
		p.close()
	}
	if c.small != nil {
		c.small.Close()
	}
	if c.base != nil {
		c.base.Close()
	}
	c.scratch = nil
}
func newScalingPreparationCPP(spec Spec, workers int) (*OrbitBasis, *scalingPreparationCPP, error) {
	if spec.FieldDegree < 2 || spec.FieldDegree > 14 || spec.Dimension != 2 || spec.CurveDegree < 1 || spec.CurveDegree >= (1<<uint(spec.FieldDegree))-1 || workers < 1 || workers > 64 {
		return nil, nil, fmt.Errorf("experimental preprocessing requires m=2,s=2..14 and valid t/workers")
	}
	polynomial, err := BinaryPolynomial(spec.FieldDegree)
	if err != nil {
		return nil, nil, err
	}
	base, extension, delta, omega := orbitFields(spec.FieldDegree, polynomial)
	fft := newSparseTransform(extension)
	baseFFT := newSparseTransform(base)
	c := &scalingPreparationCPP{plans: make(map[int]*matrixCPPPlan), workers: workers}
	fail := func(err error) (*OrbitBasis, *scalingPreparationCPP, error) { c.close(); return nil, nil, err }
	c.small, err = liftrscpp.NewPlan(cppDescription(fft))
	if err != nil {
		return fail(err)
	}
	c.base, err = liftrscpp.NewPlan(cppDescription(baseFFT))
	if err != nil {
		return fail(err)
	}
	for dimension := 2; dimension <= bits.Len32(extension.q)-1; dimension++ {
		p, err := newMatrixCPP(extension, dimension)
		if err != nil {
			return fail(err)
		}
		c.plans[1<<uint(dimension)] = p
	}
	c.scratch = make([]uint32, int(extension.q))
	fft.executor = c
	baseFFT.executor = &meteredCPP{plan: c.base, stats: &CPPEncodeStats{}}
	base.products = &polynomialProducts{fft, 32768}
	extension.products = &polynomialProducts{fft, 1024}
	return &OrbitBasis{spec: spec, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: extension, baseFFT: baseFFT, extensionFFT: fft}, c, nil
}
