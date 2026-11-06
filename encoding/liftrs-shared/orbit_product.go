package liftrs

// polynomialProducts reuses immutable extension-field transform constants for
// both base and extension coefficients. Ordinary calls own their scratch; a
// private code-preparation session may explicitly reuse single-caller buffers.
// Preparing/loading a code never caches message-dependent values. The cutoff
// applies to both operands after truncation and is a performance choice only.
type polynomialProducts struct {
	transform *additiveTransform
	cutoff    int
	// Non-nil only in a private, single-caller code-preparation session.
	workspace *productWorkspace
}

// multiply returns the first limit coefficients of the ordinary convolution.
// In characteristic two, pointwise multiplication on a binary subspace followed
// by interpolation recovers a product whose degree is below that subspace size.
// Padding is based on the FULL product degree even for a truncated output: using
// only limit points would alias discarded high coefficients into the prefix.
// If the full product exceeds the extension field, overlap-add with half-field
// chunks keeps every local product within the available evaluation domain.
func (p *polynomialProducts) multiply(a, b []uint32, limit int) []uint32 {
	out := make([]uint32, limit)
	if p.workspace != nil {
		if p.workspace.busy {
			panic("recursive preparation workspace use")
		}
		p.workspace.busy = true
		defer func() { p.workspace.busy = false }()
	}
	chunk := int(p.transform.f.q) / 2
	for i := 0; i < len(a) && i < limit; i += chunk {
		ai := a[i:minInt(i+chunk, len(a))]
		for j := 0; j < len(b) && i+j < limit; j += chunk {
			bj := b[j:minInt(j+chunk, len(b))]
			n := 1
			for n < len(ai)+len(bj)-1 {
				n *= 2
			}
			var x, y []uint32
			if p.workspace != nil {
				x, y = p.workspace.buffers(n)
			} else {
				x, y = make([]uint32, n), make([]uint32, n)
			}
			copy(x, ai)
			copy(y, bj)
			if p.workspace != nil && p.workspace.observe != nil {
				p.workspace.observe(n)
			}
			p.transform.applySubspace(x, false)
			p.transform.applySubspace(y, false)
			for k := range x {
				x[k] = p.transform.f.mul(x[k], y[k])
			}
			p.transform.applySubspace(x, true)
			count := minInt(limit-i-j, len(ai)+len(bj)-1)
			for k := 0; k < count; k++ {
				out[i+j+k] ^= x[k]
			}
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// polyMulLow computes a*b mod X^limit, without modifying either operand. The
// returned length is min(limit,len(a)+len(b)-1), including trailing zeroes.
// Coefficients at indices >=limit can never influence a lower coefficient.
func polyMulLow(f orbitField, a, b []uint32, limit int) []uint32 {
	if len(a) == 0 || len(b) == 0 || limit <= 0 {
		return nil
	}
	limit = minInt(limit, len(a)+len(b)-1)
	a = a[:minInt(len(a), limit)]
	b = b[:minInt(len(b), limit)]
	largest := len(a)
	if len(b) > largest {
		largest = len(b)
	}
	smallest := minInt(len(a), len(b))
	if f.products != nil && smallest >= f.products.cutoff && smallest >= 24 {
		return f.products.multiply(a, b, limit)
	}
	// Split very unequal operands before invoking balanced kernels. Otherwise a
	// short divisor can force padding to a huge FFT or Karatsuba domain. Each
	// block needs only O(len(short)) scratch, and blocks beyond limit are skipped.
	if smallest >= 24 && largest > 2*smallest {
		if len(a) < len(b) {
			a, b = b, a
		}
		chunk := 1
		for chunk < len(b) {
			chunk *= 2
		}
		out := make([]uint32, limit)
		for offset := 0; offset < len(a) && offset < limit; offset += chunk {
			product := polyMulLow(f, a[offset:minInt(offset+chunk, len(a))], b, limit-offset)
			for i, x := range product {
				out[offset+i] ^= x
			}
		}
		return out
	}
	if len(a) < 24 || len(b) < 24 {
		out := make([]uint32, limit)
		for i, x := range a {
			if x != 0 {
				for j := 0; j < len(b) && i+j < limit; j++ {
					out[i+j] ^= f.mul(x, b[j])
				}
			}
		}
		return out
	}
	n := 1
	for n < largest {
		n *= 2
	}
	x, y := make([]uint32, n), make([]uint32, n)
	copy(x, a)
	copy(y, b)
	return karatsubaLow(f, x, y, limit)
}

// karatsubaLow prunes recursive products that cannot reach the requested prefix.
// The high-half product is still required in the cross term even when its direct
// shift by 2*h lies outside the output. Omitting it would corrupt low coefficients.
func karatsubaLow(f orbitField, a, b []uint32, limit int) []uint32 {
	n := len(a)
	if limit >= 2*n-1 {
		return karatsuba(f, a, b)[:limit]
	}
	if n <= 32 {
		out := make([]uint32, limit)
		for i, x := range a {
			if i >= limit {
				break
			}
			if x != 0 {
				for j := 0; j < n && i+j < limit; j++ {
					out[i+j] ^= f.mul(x, b[j])
				}
			}
		}
		return out
	}
	h := n / 2
	if limit <= h {
		return karatsubaLow(f, a[:h], b[:h], limit)
	}
	cross := minInt(limit-h, 2*h-1)
	z0 := karatsubaLow(f, a[:h], b[:h], minInt(limit, 2*h-1))
	z2 := karatsubaLow(f, a[h:], b[h:], cross)
	x, y := make([]uint32, h), make([]uint32, h)
	for i := range x {
		x[i] = a[i] ^ a[h+i]
		y[i] = b[i] ^ b[h+i]
	}
	z1 := karatsubaLow(f, x, y, cross)
	out := make([]uint32, limit)
	copy(out, z0)
	for i := 0; i < cross; i++ {
		out[h+i] ^= z1[i] ^ z0[i] ^ z2[i]
	}
	for i := 0; i < len(z2) && 2*h+i < limit; i++ {
		out[2*h+i] ^= z2[i]
	}
	return out
}

// productWorkspace is mutable scratch, never part of a public prepared artifact.
// Private preparation owns it for one caller and releases it before returning.
// Ordinary encoders keep workspace nil and preserve concurrent-call ownership.
type productWorkspace struct {
	x, y    []uint32
	busy    bool
	observe func(int) // Diagnostic only; must not reenter multiplication.
}

func (w *productWorkspace) buffers(n int) ([]uint32, []uint32) {
	if cap(w.x) < n {
		w.x = make([]uint32, n)
		w.y = make([]uint32, n)
	}
	x, y := w.x[:n], w.y[:n]
	clear(x)
	clear(y)
	return x, y
}
func (w *productWorkspace) bytes() uint64 { return uint64(cap(w.x)+cap(w.y)) * 4 }
func (w *productWorkspace) release()      { w.x = nil; w.y = nil; w.observe = nil }
