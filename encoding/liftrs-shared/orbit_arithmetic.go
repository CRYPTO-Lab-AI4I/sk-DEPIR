package liftrs

// orbitField adapts the public table-free fields to the bounded transform's
// uint32 coefficients. Packing a+b*u as a+q*b preserves existing coordinates
// and version-1 artifacts. The bounded encoder uses s<=13 in this adapter;
// the public two-component quadratic API separately supports all s<=20.
type orbitField struct {
	q         uint32
	base      BinaryField
	extension *QuadraticField
	products  *polynomialProducts
}

func (f orbitField) mul(a, b uint32) uint32 {
	if f.extension == nil {
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
	field, err := NewBinaryField(s, polynomial)
	if err != nil {
		panic(err)
	}
	tower, err := NewQuadraticField(field)
	if err != nil {
		panic(err)
	}
	q := uint32(1) << uint(s)
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
	for omega = q; omega < q*q; omega++ {
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
