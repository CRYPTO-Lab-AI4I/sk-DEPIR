package liftrs

import "fmt"

// wideOrbitField is the scalar reference for the s>=17 port. A packed a+b*u
// uses a|(b<<s) in uint64, while both base-field components remain uint32.
// It does not admit larger encoder plans or reinterpret existing uint32 caches.
// Inner operations require canonical inputs below cardinality; the future
// loader/native boundaries must validate that invariant before bulk execution.
type wideOrbitField struct {
	tower       QuadraticField
	cardinality uint64
	shift       uint
	mask        uint64
}

func newWideOrbitField(s int, polynomial uint32) (wideOrbitField, error) {
	if s < 2 || s > 20 {
		return wideOrbitField{}, fmt.Errorf("wide orbit reference requires s=2..20")
	}
	base, err := NewBinaryField(s, polynomial)
	if err != nil {
		return wideOrbitField{}, err
	}
	tower, err := NewQuadraticField(base)
	if err != nil {
		return wideOrbitField{}, err
	}
	return wideOrbitField{tower: tower, cardinality: uint64(1) << uint(2*s), shift: uint(s), mask: (uint64(1) << uint(s)) - 1}, nil
}

func (f wideOrbitField) unpack(x uint64) QuadraticElement {
	return QuadraticElement{A: uint32(x & f.mask), B: uint32(x >> f.shift)}
}
func (f wideOrbitField) pack(x QuadraticElement) uint64 {
	return uint64(x.A) | uint64(x.B)<<f.shift
}
func (f wideOrbitField) mul(a, b uint64) uint64 {
	return f.pack(f.tower.Mul(f.unpack(a), f.unpack(b)))
}
func (f wideOrbitField) pow(a, n uint64) uint64 {
	result := uint64(1)
	for n != 0 {
		if n&1 != 0 {
			result = f.mul(result, a)
		}
		n >>= 1
		if n != 0 {
			a = f.mul(a, a)
		}
	}
	return result
}
func (f wideOrbitField) inverse(a uint64) (uint64, error) {
	x, err := f.tower.Inverse(f.unpack(a))
	return f.pack(x), err
}

// Match the existing first-primitive-element ordering, including the same tower
// delta. Factor only q²-1; no grid-sized search table or coordinate array exists.
func (f wideOrbitField) generator() (uint64, error) {
	order := f.cardinality - 1
	remaining := order
	var primes []uint64
	for p := uint64(2); p <= remaining/p; p++ {
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
	for candidate := uint64(1) << f.shift; candidate < f.cardinality; candidate++ {
		primitive := true
		for _, p := range primes {
			if f.pow(candidate, order/p) == 1 {
				primitive = false
				break
			}
		}
		if primitive {
			return candidate, nil
		}
	}
	return 0, fmt.Errorf("primitive extension generator not found")
}
