package liftrs

import (
	liftrsgo "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
	"math/bits"
)

// constantProduct caches a GF(2)-linear map x -> c*x in four-bit rows. It works
// for either packed tower coordinates or base-field coordinates, preserving the
// existing basis. Rows are immutable and reused across databases. Tables are
// indexed by input symbols: this is not a constant-time arithmetic primitive.
// The general BinaryField and QuadraticField descriptions remain table-free.
type constantProduct struct {
	coefficient uint32
	rows        []uint32
}

// newConstantProduct spends at most budget bytes on table entries. Zero/identity
// maps need no table. With no budget, the same map uses general multiplication.
func newConstantProduct(f orbitField, c uint32, budget int) constantProduct {
	p := constantProduct{coefficient: c}
	count := (bits.Len32(f.q) - 1 + 3) / 4
	if c <= 1 || count*16*4 > budget {
		return p
	}
	p.rows = make([]uint32, count*16)
	for j := 0; j < count; j++ {
		row := p.rows[j*16 : (j+1)*16]
		for bit := 0; bit < 4; bit++ {
			position := uint(j*4 + bit)
			// The final row may contain unused high bits; callers use canonical values.
			if position < uint(bits.Len32(f.q)-1) {
				row[1<<bit] = f.mul(c, 1<<position)
			}
		}
		for n := 1; n < 16; n++ {
			low := n & -n
			if n != low {
				row[n] = row[low] ^ row[n^low]
			}
		}
	}
	return p
}

// mul exposes the Go executor's fixed-map operation to shared arithmetic tests
// and kernel fixtures. Native execution has its own C++ evaluator of these rows.
func (p constantProduct) mul(f orbitField, x uint32) uint32 {
	return (liftrsgo.Product{Coefficient: p.coefficient, Rows: p.rows}).Mul(liftrsgo.Field{Base: f.base, Extension: f.extension}, x)
}
