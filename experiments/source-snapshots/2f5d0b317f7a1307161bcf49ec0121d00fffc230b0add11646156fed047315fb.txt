// Package liftrsgo implements only pure Go additive-transform execution.
// Code discovery, table construction and full encoder orchestration live in
// liftrs-shared. The plan borrows immutable prepared tables; scratch is caller-owned.
package liftrsgo

import (
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"
	"math/bits"
)

// Field describes base or packed quadratic arithmetic using common field types.
type Field struct {
	Base      field.BinaryField
	Extension *field.QuadraticField
}

func (f Field) mul(a, b uint32) uint32 {
	if f.Extension == nil {
		return f.Base.Mul(a, b)
	}
	s := uint(f.Base.Degree())
	mask := (uint32(1) << s) - 1
	x := f.Extension.Mul(field.QuadraticElement{A: a & mask, B: a >> s}, field.QuadraticElement{A: b & mask, B: b >> s})
	return x.A | x.B<<s
}

// Product borrows an immutable four-bit multiplication map. Empty rows use
// shared scalar field arithmetic; canonical input elements are required.
type Product struct {
	Coefficient uint32
	Rows        []uint32
}

// Mul evaluates this fixed map on a canonical field element.
func (p Product) Mul(f Field, x uint32) uint32 {
	if p.Coefficient == 0 {
		return 0
	}
	if p.Coefficient == 1 {
		return x
	}
	if len(p.Rows) == 0 {
		return f.mul(p.Coefficient, x)
	}
	var y uint32
	for j := 0; j < len(p.Rows); j += 16 {
		y ^= p.Rows[j+int(x&15)]
		x >>= 4
	}
	return y
}

// Term is a sparse linearized-polynomial term in the prepared transform.
type Term struct {
	Degree  int
	Product Product
}

// Transform contains immutable Go execution views of shared preprocessing.
// Recurse mutates the caller buffer only. The shared encoder validates dimensions;
// direct callers must supply a valid plan and 2^level canonical elements.
type Transform struct {
	Field                          Field
	Terms                          [][]Term
	Values                         [][]uint32
	DeltaProducts, InverseProducts []Product
}

func (a *Transform) offset(k int, offset uint32) uint32 {
	var out uint32
	for offset != 0 {
		j := bits.TrailingZeros32(offset)
		out ^= a.Values[k][j]
		offset &= offset - 1
	}
	return out
}

func (a *Transform) Recurse(buf []uint32, level int, offset uint32, inverse bool) {
	if level == 0 {
		return
	}
	k, half := level-1, len(buf)/2
	c := a.offset(k, offset)
	terms := a.Terms[k][:len(a.Terms[k])-1] // omit the monic leading term
	if !inverse {
		for i := len(buf) - 1; i >= half; i-- {
			v := buf[i]
			for _, t := range terms {
				buf[i-half+t.Degree] ^= t.Product.Mul(a.Field, v)
			}
		}
		for j := 0; j < half; j++ {
			v := buf[half+j]
			left := buf[j] ^ a.Field.mul(c, v)
			buf[j] = left
			buf[half+j] = left ^ a.DeltaProducts[k].Mul(a.Field, v)
		}
	}
	a.Recurse(buf[:half], k, offset, inverse)
	a.Recurse(buf[half:], k, offset^uint32(half), inverse)
	if inverse {
		for j := 0; j < half; j++ {
			v := a.InverseProducts[k].Mul(a.Field, buf[j]^buf[half+j])
			buf[j] ^= a.Field.mul(c, v)
			buf[half+j] = v
		}
		// Ascending j preserves each quotient coefficient before lower-degree terms
		// from later coefficients add into its already-consumed slot.
		for j := 0; j < half; j++ {
			v := buf[half+j]
			for _, t := range terms {
				buf[j+t.Degree] ^= t.Product.Mul(a.Field, v)
			}
		}
	}
}
