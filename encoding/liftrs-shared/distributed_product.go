package liftrs

import (
	"fmt"
	"unsafe"
)

// rankProductBackend composes the validated FFT operations and one local
// pointwise multiply; it carries no storage and retains no caller pointers.
type rankProductBackend[Word transposeWord] struct {
	Transpose transposeLink[Word]
	Windows   taylorLink[Word]
	Degrees   func(int) ([]uint64, error)
	Vectors   rankVectorApply[Word]
	Taylor    rankTaylorApply[Word]
	Multiply  func(left, right []Word) error
}

// rankPolynomialProduct consumes disjoint zero-padded coefficient shards and
// returns an owning product shard. Caller must choose a domain larger than the
// sum of operand degrees: otherwise this computes a product modulo the domain's
// vanishing polynomial. No operand degree or PIR identity is inferred here.
// This host-shard integration is not a selected GPU-resident production layout.
func rankPolynomialProduct[Word transposeWord](p rankTransposePlan, left, right []Word, capacity uint64, b rankProductBackend[Word], notice func(string, uint64, uint64)) ([]Word, error) {
	if b.Multiply == nil || uint64(len(left)) != p.InputWords || len(left) != len(right) || capacity == 0 {
		return nil, fmt.Errorf("distributed product operands/backend")
	}
	if len(left) > 0 {
		a, c := uintptr(unsafe.Pointer(&left[0])), uintptr(unsafe.Pointer(&right[0]))
		var zero Word
		size := uintptr(len(left)) * unsafe.Sizeof(zero)
		if a < c+size && c < a+size {
			return nil, fmt.Errorf("distributed product overlapping operands")
		}
	}
	plans, err := rankProductPlans(p, capacity)
	if err != nil {
		return nil, err
	}
	data := [2][]Word{left, right}
	err = rankProductSchedule(plans, p.InputWords,
		func(plan rankTransposePlan, operand int, inverse bool) error {
			var e error
			data[operand], e = rankAdditiveTransform(plan, data[operand], capacity, inverse, b.Transpose, b.Windows, b.Degrees, b.Vectors, b.Taylor, notice)
			return e
		},
		func() error {
			for offset := uint64(0); offset < uint64(len(data[0])); {
				count := min(capacity, uint64(len(data[0]))-offset)
				if err := b.Multiply(data[0][offset:offset+count], data[1][offset:offset+count]); err != nil {
					return err
				}
				offset += count
				if offset < uint64(len(data[0])) && notice != nil {
					notice("rank-product-pointwise", offset, uint64(len(data[0])))
				}
			}
			return nil
		}, notice)
	if err != nil {
		return nil, err
	}
	return data[0], nil
}

// rankProductPlans admits all three FFT spans before either operand is consumed.
func rankProductPlans(p rankTransposePlan, capacity uint64) (plans [3]rankTransposePlan, err error) {
	if _, _, err = rankAdditivePlans(p, capacity, false); err != nil {
		return plans, err
	}
	levels := 0
	for rows := p.Rows; rows > 1; rows >>= 1 {
		levels++
	}
	span := uint64(2 + 2*levels)
	if p.Operation > ^uint64(0)-3*span {
		return plans, fmt.Errorf("distributed product operation wrap")
	}
	for j := range plans {
		plans[j] = p
		plans[j].Operation += uint64(j) * span
	}
	return plans, nil
}

// rankProductSchedule shares forward-left/forward-right/pointwise/inverse-left
// ordering across storage backends. FFT arithmetic and storage are injected.
// On failure the result is incomplete; callers must discard/reset consumed data.
func rankProductSchedule(plans [3]rankTransposePlan, words uint64, transform func(rankTransposePlan, int, bool) error, multiply func() error, notice func(string, uint64, uint64)) error {
	stage := func(phase string, total uint64, run func() error) error {
		if notice != nil {
			notice(phase, 0, total)
		}
		if err := run(); err != nil {
			return err
		}
		if notice != nil {
			notice(phase, total, total)
		}
		return nil
	}
	if err := stage("rank-product-left-forward", 1, func() error { return transform(plans[0], 0, false) }); err != nil {
		return err
	}
	if err := stage("rank-product-right-forward", 1, func() error { return transform(plans[1], 1, false) }); err != nil {
		return err
	}
	if err := stage("rank-product-pointwise", words, multiply); err != nil {
		return err
	}
	return stage("rank-product-inverse", 1, func() error { return transform(plans[2], 0, true) })
}
