package liftrs

// factorProductBuilder retains a bounded flat leaf batch, then merges completed
// roots by binary carry. Polynomial products are exact, so regrouping preserves
// P while avoiding one heap object/header per factor for the whole code.
type factorProductBuilder struct {
	f       orbitField
	storage []uint32
	factors [][]uint32
	levels  [][]uint32
}

func newFactorProductBuilder(f orbitField, batch int) *factorProductBuilder {
	if batch < 1 {
		panic("invalid factor batch")
	}
	return &factorProductBuilder{f: f, storage: make([]uint32, 3*batch), factors: make([][]uint32, 0, batch)}
}
func (b *factorProductBuilder) add(c0, c1, c2 uint32, degree int) {
	slot := b.storage[3*len(b.factors) : 3*len(b.factors)+3]
	slot[0], slot[1], slot[2] = c0, c1, c2
	b.factors = append(b.factors, slot[:degree+1])
	if len(b.factors) == cap(b.factors) {
		b.flush()
	}
}
func (b *factorProductBuilder) flush() {
	if len(b.factors) == 0 {
		return
	}
	var root []uint32
	if len(b.factors) == 1 {
		// A singleton aliases the reusable flat leaf storage; retain its own copy.
		root = append([]uint32(nil), b.factors[0]...)
	} else {
		root = rootProduct(b.f, b.factors)
	}
	clear(b.factors)
	b.factors = b.factors[:0]
	level := 0
	for level < len(b.levels) && b.levels[level] != nil {
		root = polyMul(b.f, b.levels[level], root)
		b.levels[level] = nil
		level++
	}
	if level == len(b.levels) {
		b.levels = append(b.levels, root)
	} else {
		b.levels[level] = root
	}
}
func (b *factorProductBuilder) finish() []uint32 {
	b.flush()
	var root []uint32
	for i, p := range b.levels {
		if p == nil {
			continue
		}
		if root == nil {
			root = p
		} else {
			root = polyMul(b.f, root, p)
		}
		b.levels[i] = nil
	}
	if root == nil {
		return []uint32{1}
	}
	return root
}
