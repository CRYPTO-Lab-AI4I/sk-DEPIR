package liftrs

import (
	"fmt"
	"math/bits"
)

// newSparseTransform constructs the same S_k constants without N-sized dense
// polynomials. Coefficient i represents X^(2^i). S_(k+1)=S_k²+delta_k*S_k;
// equal degrees must XOR before the next level. No unused final polynomial is made.
// Experimental only: the ordinary constructors retain their measured behavior.
func newSparseTransform(f orbitField) *additiveTransform {
	levels := bits.Len32(f.q) - 1
	a := &additiveTransform{f: f, terms: make([][]linearTerm, levels), values: make([][]uint32, levels), delta: make([]uint32, levels), inverse: make([]uint32, levels)}
	p := []uint32{1}
	for k := 0; k < levels; k++ {
		for i, c := range p {
			if c != 0 {
				a.terms[k] = append(a.terms[k], linearTerm{degree: 1 << uint(i), coefficient: c})
			}
		}
		a.values[k] = make([]uint32, levels)
		for j := 0; j < levels; j++ {
			for _, term := range a.terms[k] {
				a.values[k][j] ^= f.mul(term.coefficient, f.pow(1<<uint(j), term.degree))
			}
		}
		a.delta[k] = a.values[k][k]
		a.inverse[k] = f.inv(a.delta[k])
		if k+1 < levels {
			next := make([]uint32, len(p)+1)
			for i, c := range p {
				next[i+1] ^= f.mul(c, c)
				next[i] ^= f.mul(a.delta[k], c)
			}
			p = next
		}
	}
	return finishAdditiveTransform(a)
}

// SparseTransformProbe constructs code-only sparse transform constants and returns
// their native-description words for independent comparison/retention. It does not
// enable a larger encoder: packed extension arithmetic is limited here to s<=15.
// For s>=14 this tests only setup, not discovery, encoding or security parameters.
func SparseTransformProbe(s int, extension bool) ([]uint32, error) {
	if s < 2 || s > 15 {
		return nil, fmt.Errorf("sparse transform probe requires s=2..15")
	}
	polynomial, err := BinaryPolynomial(s)
	if err != nil {
		return nil, err
	}
	base, ext, _, _ := orbitFields(s, polynomial)
	f := base
	if extension {
		f = ext
	}
	return cppDescription(newSparseTransform(f)), nil
}

// rootProduct consumes a balanced factor list. Only completed sibling roots are
// retained, not all tree nodes. Go GC can retain dead allocations, so this is a
// logical ownership reduction rather than an RSS guarantee.
func rootProduct(f orbitField, factors [][]uint32) []uint32 {
	if len(factors) == 0 {
		return []uint32{1}
	}
	if len(factors) == 1 {
		return factors[0]
	}
	h := len(factors) / 2
	left := rootProduct(f, factors[:h])
	right := rootProduct(f, factors[h:])
	out := polyMul(f, left, right)
	return out
}

// PrepareCyclicScalingPrototype is an opt-in setup experiment. It groups roots
// into Frobenius orbits, constructs P over the base field without retaining a
// product tree, and obtains G from the truncated reciprocal 1/P. It uses sparse
// FFT setup. The result preserves the original information set/artifact format.
// It does not establish bounded-RAM preparation: factors and multiplication
// scratch still allocate, discovery/support are retained, and no disk spill exists.
func (b *OrbitBasis) PrepareCyclicScalingPrototype() (*CyclicEncoder, error) {
	clone := *b
	fft := newSparseTransform(b.extension)
	clone.extensionFFT = fft
	clone.baseFFT = newSparseTransform(b.base)
	clone.base.products = &polynomialProducts{fft, b.base.products.cutoff}
	clone.extension.products = &polynomialProducts{fft, b.extension.products.cutoff}
	// Frobenius closure is checked independently of random discovery on every basis.
	if err := b.validateSupport(); err != nil {
		return nil, err
	}
	var factors [][]uint32
	order := b.Length() - 1
	for _, i := range b.support {
		partner := int(uint64(i) * uint64(b.base.q) % uint64(order))
		if partner < i {
			continue
		}
		root := b.extension.inv(b.extension.pow(b.omega, i))
		if partner == i {
			if root >= b.base.q {
				return nil, fmt.Errorf("fixed root outside base field")
			}
			factors = append(factors, []uint32{root, 1})
			continue
		}
		conjugate := b.extension.pow(root, int(b.base.q))
		trace, norm := root^conjugate, b.extension.mul(root, conjugate)
		if trace >= b.base.q || norm >= b.base.q {
			return nil, fmt.Errorf("paired root factor outside base field")
		}
		factors = append(factors, []uint32{norm, trace, 1})
	}
	p := rootProduct(clone.base, factors)
	if len(p) != b.Dimension()+1 {
		return nil, fmt.Errorf("paired product degree mismatch")
	}
	scale := clone.base.inv(p[0])
	for i := range p {
		p[i] = clone.base.mul(p[i], scale)
	}
	g := polyInverse(clone.base, p, b.Length()-b.Dimension())
	// Verify the defining identity independently of reciprocal truncation. This
	// one-time full convolution is deliberately included in preparation timings.
	identity := polyMul(clone.base, p, g)
	if len(identity) != b.Length() {
		return nil, fmt.Errorf("generator identity degree mismatch")
	}
	for i, x := range identity {
		want := uint32(0)
		if i == 0 || i == len(identity)-1 {
			want = 1
		}
		if x != want {
			return nil, fmt.Errorf("generator identity mismatch at %d", i)
		}
	}
	spectrum := make([]uint32, b.Length())
	copy(spectrum, g)
	fft.apply(spectrum, false)
	return &CyclicEncoder{basis: &clone, p: p, g: g, spectrum: spectrum}, nil
}

// ScalingOnlinePlan is an experimental minimal online state with explicit
// reusable workspace. It omits G/support and cannot independently authenticate a
// code; construct it only from a trusted verified prepared encoder. No format or
// loader validation has been weakened. Workspace instances must not be shared
// concurrently; the immutable plan can be shared with distinct workspaces.
type ScalingOnlinePlan struct {
	basis       OrbitBasis
	k           int
	p, spectrum []uint32
}

// ScalingWorkspace owns reusable prefix and full-domain buffers; it is not safe
// for concurrent reuse. Buffers are allocated by NewWorkspace, not each Encode.
type ScalingWorkspace struct {
	owner      *ScalingOnlinePlan
	x, z, h, y []uint32
}

// ScalingPlan copies only online P/spectrum and scalar code data so the source
// G/support can be released. Its Go default tables are rebuilt sparsely.
func (e *CyclicEncoder) ScalingPlan() (*ScalingOnlinePlan, error) {
	if e == nil || e.basis == nil || len(e.spectrum) != e.basis.Length() {
		return nil, fmt.Errorf("verified cached encoder required")
	}
	l := 1
	for l < 2*e.basis.Dimension()-1 {
		l *= 2
	}
	if uint64(l) > uint64(e.basis.extension.q) {
		return nil, fmt.Errorf("online probe requires full prefix convolution to fit one transform domain")
	}
	b := *e.basis
	b.support = nil
	b.extensionFFT = newSparseTransform(b.extension)
	b.baseFFT = nil
	// Online uses a dedicated product schedule, with no generic products dispatch.
	b.base.products = nil
	b.extension.products = nil
	return &ScalingOnlinePlan{basis: b, k: e.basis.Dimension(), p: append([]uint32(nil), e.p...), spectrum: append([]uint32(nil), e.spectrum...)}, nil
}

// ResidentBytes counts the plan's two coefficient arrays, excluding small tables
// and object metadata, allocator capacity and process memory.
func (p *ScalingOnlinePlan) ResidentBytes() uint64 { return uint64(len(p.p)+len(p.spectrum)) * 4 }

// NewWorkspace aliases two length-L prefix FFT arrays with the full-domain
// convolution buffer in one arena, retaining H separately before reusing it.
// L pads the full ordinary product even though the desired output is truncated.
func (p *ScalingOnlinePlan) NewWorkspace() *ScalingWorkspace {
	l := 1
	for l < 2*p.k-1 {
		l *= 2
	}
	arena := make([]uint32, maxIntScaling(2*l, p.basis.Length()))
	return &ScalingWorkspace{p, arena[:l], arena[l : 2*l], make([]uint32, p.k), arena[:p.basis.Length()]}
}

// Bytes counts reusable coefficient payload, not peak RAM or the caller's input
// and returned canonical word. Each call still allocates that final word.
func (w *ScalingWorkspace) Bytes() uint64 {
	return uint64(maxIntScaling(len(w.x)+len(w.z), len(w.y))+len(w.h)) * 4
}

// Encode uses the existing transform executor (including native whole-buffer
// calls when supplied by the package's experiment) and preserves canonical order.
// Input/output must not alias internal workspace, which is not exposed publicly.
func (p *ScalingOnlinePlan) Encode(message []uint32, w *ScalingWorkspace) ([]uint32, error) {
	if w == nil || w.owner != p || len(message) != p.k {
		return nil, fmt.Errorf("workspace/message mismatch")
	}
	for _, v := range message {
		if v >= p.basis.base.q {
			return nil, fmt.Errorf("message outside field")
		}
	}
	clear(w.x)
	clear(w.z)
	copy(w.x, p.p[:p.k])
	copy(w.z, message)
	fft := p.basis.extensionFFT
	fft.applySubspace(w.x, false)
	fft.applySubspace(w.z, false)
	for i := range w.x {
		w.x[i] = p.basis.extension.mul(w.x[i], w.z[i])
	}
	fft.applySubspace(w.x, true)
	copy(w.h, w.x[:p.k])
	clear(w.y) // y aliases both prefix arrays; H must be saved before clearing.
	copy(w.y, w.h)
	fft.apply(w.y, false)
	for i := range w.y {
		w.y[i] = p.basis.extension.mul(w.y[i], p.spectrum[i])
	}
	fft.apply(w.y, true)
	word := make([]uint32, p.basis.Length())
	point := p.basis.omega
	for _, x := range w.y[:len(w.y)-1] {
		if x >= p.basis.base.q {
			return nil, fmt.Errorf("product escaped base field")
		}
		word[point] = x
		word[0] ^= x
		point = p.basis.extension.mul(point, p.basis.omega)
	}
	return word, nil
}

func maxIntScaling(a, b int) int {
	if a > b {
		return a
	}
	return b
}
