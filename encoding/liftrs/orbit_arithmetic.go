package liftrs

// orbitField uses log tables only for bounded encoding experiments. The quadratic
// extension is represented as a+b*u, u²=u+delta; low s bits hold a, high bits b.
// Thus the public coordinate grid remains x0+q*x1, with no field-basis conversion.
type orbitField struct {
	q        uint32
	exp, log []uint32
}

func (f orbitField) mul(a, b uint32) uint32 {
	if a == 0 || b == 0 {
		return 0
	}
	return f.exp[f.log[a]+f.log[b]]
}
func (f orbitField) pow(a uint32, n int) uint32 {
	if n == 0 {
		return 1
	}
	if a == 0 {
		return 0
	}
	return f.exp[uint64(f.log[a])*uint64(n)%uint64(f.q-1)]
}
func (f orbitField) inv(a uint32) uint32 { return f.exp[f.q-1-f.log[a]] }
func tableField(q, generator uint32, multiply func(uint32, uint32) uint32) orbitField {
	f := orbitField{q: q, exp: make([]uint32, 2*(q-1)), log: make([]uint32, q)}
	x := uint32(1)
	for i := uint32(0); i < q-1; i++ {
		f.exp[i] = x
		f.exp[i+q-1] = x
		f.log[x] = i
		x = multiply(x, generator)
	}
	return f
}
func orbitFields(s int, polynomial uint32) (base, extension orbitField, delta, omega uint32) {
	q := uint32(1) << uint(s)
	slow := field{q, polynomial}
	base = tableField(q, 2, slow.mul)
	for delta = 1; delta < q; delta++ {
		x, tr := delta, uint32(0)
		for i := 0; i < s; i++ {
			tr ^= x
			x = base.mul(x, x)
		}
		if tr == 1 {
			break
		}
	}
	multiply := func(x, y uint32) uint32 {
		a, b, c, d := x&(q-1), x>>uint(s), y&(q-1), y>>uint(s)
		ac, bd := base.mul(a, c), base.mul(b, d)
		return (ac ^ base.mul(delta, bd)) | ((base.mul(a^b, c^d) ^ ac) << uint(s))
	}
	power := func(x uint32, n int) uint32 {
		r := uint32(1)
		for n > 0 {
			if n&1 != 0 {
				r = multiply(r, x)
			}
			x = multiply(x, x)
			n >>= 1
		}
		return r
	}
	order := int(q*q) - 1
	var factors []int
	rem := order
	for p := 2; p*p <= rem; p++ {
		if rem%p == 0 {
			factors = append(factors, p)
			for rem%p == 0 {
				rem /= p
			}
		}
	}
	if rem > 1 {
		factors = append(factors, rem)
	}
	for omega = q; omega < q*q; omega++ {
		primitive := true
		for _, p := range factors {
			if power(omega, order/p) == 1 {
				primitive = false
				break
			}
		}
		if primitive {
			break
		}
	}
	extension = tableField(q*q, omega, multiply)
	return
}
