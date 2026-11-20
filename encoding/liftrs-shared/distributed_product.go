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
	levels := 0
	for rows := p.Rows; rows > 1; rows >>= 1 {
		levels++
	}
	span := uint64(2 + 2*levels)
	if p.Operation > ^uint64(0)-3*span {
		return nil, fmt.Errorf("distributed product operation wrap")
	}
	transform := func(plan rankTransposePlan, data []Word, inverse bool) ([]Word, error) {
		return rankAdditiveTransform(plan, data, capacity, inverse, b.Transpose, b.Windows, b.Degrees, b.Vectors, b.Taylor, notice)
	}
	emit := func(phase string, done, total uint64) {
		if notice != nil {
			notice(phase, done, total)
		}
	}
	emit("rank-product-left-forward", 0, 1)
	first, err := transform(p, left, false)
	if err != nil {
		return nil, err
	}
	emit("rank-product-left-forward", 1, 1)
	next := p
	next.Operation += span
	emit("rank-product-right-forward", 0, 1)
	second, err := transform(next, right, false)
	if err != nil {
		return nil, err
	}
	emit("rank-product-right-forward", 1, 1)
	emit("rank-product-pointwise", 0, uint64(len(first)))
	for offset := uint64(0); offset < uint64(len(first)); {
		count := min(capacity, uint64(len(first))-offset)
		if err = b.Multiply(first[offset:offset+count], second[offset:offset+count]); err != nil {
			return nil, err
		}
		offset += count
		emit("rank-product-pointwise", offset, uint64(len(first)))
	}
	next.Operation += span
	emit("rank-product-inverse", 0, 1)
	out, err := transform(next, first, true)
	if err != nil {
		return nil, err
	}
	emit("rank-product-inverse", 1, 1)
	return out, nil
}
