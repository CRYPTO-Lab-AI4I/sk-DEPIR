package liftrs

// orbitField adapts the public table-free fields to the bounded transform's
// uint32 coefficients. Packing a+b*u as a+q*b preserves existing coordinates
// and version-1 artifacts. Counts need one more bit than coordinates at s=16;
// widening this small field must not widen the coefficient arrays.
type orbitField struct {
	q         uint64
	base      BinaryField
	extension *QuadraticField
	products  *polynomialProducts
	// Private code-preparation lookup; never serialized or used by public fields.
	lookup *preparationProducts
}

func (f orbitField) mul(a, b uint32) uint32 {
	if f.extension == nil {
		if f.lookup != nil {
			return f.lookup.mul(a, b)
		}
		return f.base.Mul(a, b)
	}
	s := uint(f.base.Degree())
	mask := (uint32(1) << s) - 1
	x := f.extension.Mul(QuadraticElement{A: a & mask, B: a >> s}, QuadraticElement{A: b & mask, B: b >> s})
	return x.A | x.B<<s
}
func (f orbitField) pow(a uint32, n int) uint32 {
	r := uint32(1)
	for n > 0 {
		if n&1 != 0 {
			r = f.mul(r, a)
		}
		n >>= 1
		if n != 0 {
			a = f.mul(a, a)
		}
	}
	return r
}
func (f orbitField) inv(a uint32) uint32 {
	if f.extension == nil {
		x, err := f.base.Inverse(a)
		if err != nil {
			panic(err)
		}
		return x
	}
	s := uint(f.base.Degree())
	mask := (uint32(1) << s) - 1
	x, err := f.extension.Inverse(QuadraticElement{A: a & mask, B: a >> s})
	if err != nil {
		panic(err)
	}
	return x.A | x.B<<s
}

// orbitFields is called only after bounded encoder specifications are checked.
// Generator search factors q²-1, then picks the first primitive packed element
// at least q. It allocates no arithmetic table and preserves version-1 ordering.
func orbitFields(s int, polynomial uint32) (base, extension orbitField, delta, omega uint32) {
	if s < 2 || s > 16 || uint64(^uint(0)>>1) < uint64(1)<<uint(2*s) {
		panic("packed orbit fields require s=2..16 and a sufficient host integer width")
	}
	field, err := NewBinaryField(s, polynomial)
	if err != nil {
		panic(err)
	}
	tower, err := NewQuadraticField(field)
	if err != nil {
		panic(err)
	}
	q := uint64(1) << uint(s)
	base = orbitField{q: q, base: field}
	extension = orbitField{q: q * q, base: field, extension: &tower}
	order := int(q*q) - 1
	rem := order
	var factors []int
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
	for candidate := q; candidate < q*q; candidate++ {
		omega = uint32(candidate)
		primitive := true
		for _, p := range factors {
			if extension.pow(omega, order/p) == 1 {
				primitive = false
				break
			}
		}
		if primitive {
			return base, extension, tower.Delta(), omega
		}
	}
	panic("primitive extension generator not found")
}
