package liftrs

// field implements polynomial-basis arithmetic for the tiny reference fields.
// It deliberately uses shift/XOR multiplication rather than the production
// package's log tables, providing a simple independent arithmetic path.
type field struct {
	q          uint32
	polynomial uint32
}

func (f field) mul(a, b uint32) uint32 {
	var result uint32
	for b != 0 {
		if b&1 != 0 {
			result ^= a
		}
		b >>= 1
		a <<= 1
		if a&f.q != 0 {
			a ^= f.polynomial
		}
	}
	return result
}

func (f field) pow(a uint32, exponent int) uint32 {
	result := uint32(1) // Includes 0^0=1 for monomial evaluation.
	for exponent > 0 {
		if exponent&1 != 0 {
			result = f.mul(result, a)
		}
		a = f.mul(a, a)
		exponent >>= 1
	}
	return result
}
