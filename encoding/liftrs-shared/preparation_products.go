package liftrs

// preparationProducts accelerates public code coefficients, not secret data.
// Its input-indexed tables are not constant-time. Only the base field is tabulated:
// at s=16 the immutable log/duplicated-power arrays total less than 768 KiB.
// The polynomial coordinates and all persisted coefficients remain unchanged.
type preparationProducts struct {
	log, powers []uint32
}

func newPreparationProducts(f BinaryField) *preparationProducts {
	if f.Degree() > 16 {
		return nil // Retain bounded memory if larger base fields are admitted later.
	}
	order := (1 << uint(f.Degree())) - 1
	remaining := order
	var primes []int
	for p := 2; p*p <= remaining; p++ {
		if remaining%p == 0 {
			primes = append(primes, p)
			for remaining%p == 0 {
				remaining /= p
			}
		}
	}
	if remaining > 1 {
		primes = append(primes, remaining)
	}
	var generator uint32
	for generator = 2; generator <= uint32(order); generator++ {
		primitive := true
		for _, p := range primes {
			if f.Pow(generator, uint64(order/p)) == 1 {
				primitive = false
				break
			}
		}
		if primitive {
			break
		}
	}
	if generator > uint32(order) {
		panic("no base-field generator")
	}
	t := &preparationProducts{log: make([]uint32, order+1), powers: make([]uint32, 2*order)}
	x := uint32(1)
	for i := 0; i < order; i++ {
		t.log[x], t.powers[i], t.powers[i+order] = uint32(i), x, x
		x = f.Mul(x, generator)
	}
	return t
}

func (t *preparationProducts) mul(a, b uint32) uint32 {
	if a == 0 || b == 0 {
		return 0
	}
	return t.powers[t.log[a]+t.log[b]]
}
