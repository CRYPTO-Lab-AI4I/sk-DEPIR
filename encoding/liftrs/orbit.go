package liftrs

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

// OrbitOptions controls database-independent discovery of the paper's univariate
// degree set. Random defaults to crypto/rand.Reader. A supplied reader must yield
// independent uniform bytes; deterministic readers are for reproducible tests.
type OrbitOptions struct {
	Random     io.Reader
	MaxSamples int
}

// OrbitBasis contains only public, reusable code data, never a database or key.
// This bounded implementation accepts m=2 and s=2..13 (N<=67108864). Larger catalogue
// instances need a different memory/I/O backend. Its methods are concurrency-safe.
type OrbitBasis struct {
	spec                     Spec
	polynomial, delta, omega uint32
	base, extension          orbitField
	baseFFT, extensionFFT    *additiveTransform
	support                  []int
	samples                  int
}

// DiscoverOrbit implements paper Claim 4.2: sample genuine codewords, interpolate
// their univariate representations, and union supports until exactly K exponents
// have appeared. K is independently counted by exact good-monomial enumeration;
// failure to reach K is an error, never an estimated-capacity encoding artifact.
func DiscoverOrbit(spec Spec, options OrbitOptions) (*OrbitBasis, error) {
	polynomial, polynomialErr := BinaryPolynomial(spec.FieldDegree)
	ok := polynomialErr == nil && spec.FieldDegree <= 13
	if !ok || spec.Dimension != 2 {
		return nil, fmt.Errorf("orbit backend requires m=2 and s=2..13")
	}
	q := 1 << uint(spec.FieldDegree)
	if spec.CurveDegree < 1 || spec.CurveDegree >= q-1 {
		return nil, fmt.Errorf("invalid curve degree")
	}
	if options.MaxSamples < 0 {
		return nil, fmt.Errorf("negative sample limit")
	}
	if options.MaxSamples == 0 {
		options.MaxSamples = 2 + (128+spec.FieldDegree-1)/spec.FieldDegree
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	b := newOrbitBasis(spec, polynomial)
	var good []int
	for i := 0; i < q*q; i++ {
		yes, _ := GoodMonomial(spec, []uint32{uint32(i % q), uint32(i / q)})
		if yes {
			good = append(good, i)
		}
	}
	found := make([]bool, q*q)
	count := 0
	sampleBytes := 1
	if spec.FieldDegree > 8 {
		sampleBytes = 4
	}
	random := make([]byte, sampleBytes*len(good))
	for b.samples < options.MaxSamples && count < len(good) {
		if _, err := io.ReadFull(options.Random, random); err != nil {
			return nil, fmt.Errorf("degree-set sampling: %w", err)
		}
		grid := make([]uint32, q*q)
		for j, index := range good {
			offset := sampleBytes * j
			value := uint32(random[offset])
			if sampleBytes == 4 {
				value |= uint32(random[offset+1])<<8 | uint32(random[offset+2])<<16 | uint32(random[offset+3])<<24
			}
			grid[index] = value & uint32(q-1)
		}
		for y := 0; y < q; y++ {
			b.baseFFT.apply(grid[y*q:(y+1)*q], false)
		}
		column := make([]uint32, q)
		for x := 0; x < q; x++ {
			for y := 0; y < q; y++ {
				column[y] = grid[x+q*y]
			}
			b.baseFFT.apply(column, false)
			for y := 0; y < q; y++ {
				grid[x+q*y] = column[y]
			}
		}
		// The extension's packed integer is phi^-1(x0,x1); it has the same index
		// as the canonical grid, so interpolation needs no permutation table.
		b.extensionFFT.apply(grid, true)
		for i, c := range grid {
			if c != 0 && !found[i] {
				found[i] = true
				count++
			}
		}
		b.samples++
	}
	if count != len(good) {
		return nil, fmt.Errorf("degree discovery incomplete: %d of %d exponents after %d samples", count, len(good), b.samples)
	}
	for i, yes := range found {
		if yes {
			b.support = append(b.support, i)
		}
	}
	if err := b.validateSupport(); err != nil {
		return nil, err
	}
	return b, nil
}
func newOrbitBasis(spec Spec, polynomial uint32) *OrbitBasis {
	base, extension, delta, omega := orbitFields(spec.FieldDegree, polynomial)
	fft := newAdditiveTransform(extension)
	// The same extension transform multiplies base-field polynomials without
	// changing their coefficients or the serialized public artifact.
	base.products = &polynomialProducts{transform: fft, cutoff: 32768}
	extension.products = &polynomialProducts{transform: fft, cutoff: 1024}
	return &OrbitBasis{spec: spec, polynomial: polynomial, base: base, extension: extension, delta: delta, omega: omega, baseFFT: newAdditiveTransform(base), extensionFFT: fft}
}
func (b *OrbitBasis) validateSupport() error {
	n := int(b.extension.q) - 1
	if len(b.support) == 0 || len(b.support) > n {
		return fmt.Errorf("invalid degree-set size")
	}
	seen := make(map[int]bool, len(b.support))
	for j, i := range b.support {
		if i < 0 || i >= n || (j > 0 && i <= b.support[j-1]) {
			return fmt.Errorf("degree set must be sorted, distinct and exclude N-1")
		}
		seen[i] = true
	}
	for _, i := range b.support {
		if !seen[i*int(b.base.q)%n] {
			return fmt.Errorf("degree set is not Frobenius closed")
		}
	}
	return nil
}

// Dimension returns the exact message length established during preprocessing.
func (b *OrbitBasis) Dimension() int { return len(b.support) }

// Length returns the number of codeword elements, including the coordinate zero.
func (b *OrbitBasis) Length() int { return int(b.extension.q) }

// Samples returns how many random codewords discovery needed (zero after loading).
func (b *OrbitBasis) Samples() int { return b.samples }

// SystematicPosition computes phi(omega^(j+1)) flattened as x0+q*x1, without
// a database-sized lookup table. j is a zero-based message index.
func (b *OrbitBasis) SystematicPosition(j int) (int, error) {
	if j < 0 || j >= b.Dimension() {
		return 0, fmt.Errorf("message index outside [0,K)")
	}
	return int(b.extension.pow(b.omega, j+1)), nil
}
func (b *OrbitBasis) validateMessage(message []uint32) error {
	if len(message) != b.Dimension() {
		return fmt.Errorf("message has %d elements; need %d", len(message), b.Dimension())
	}
	for _, x := range message {
		if x >= b.base.q {
			return fmt.Errorf("message element outside base field")
		}
	}
	return nil
}

// PaperEncoder holds the paper's reusable transposed-Vandermonde solver data.
// Encode does no monomial discovery, transform setup or product-tree construction.
type PaperEncoder struct {
	basis      *OrbitBasis
	tree       *productTree
	p, weights []uint32
}

// PreparePaper builds a solver for the paper's fixed information set. Its
// polynomials and denominator inverses depend only on the code, not the message.
func (b *OrbitBasis) PreparePaper() (*PaperEncoder, error) {
	roots := make([]uint32, b.Dimension())
	for j, i := range b.support {
		roots[j] = b.extension.inv(b.extension.pow(b.omega, i))
	}
	tree := newProductTree(b.extension, roots)
	p := append([]uint32(nil), tree.polynomial...)
	scale := b.extension.inv(p[0])
	for i := range p {
		p[i] = b.extension.mul(p[i], scale)
	}
	derivative := make([]uint32, len(p)-1)
	for i := 1; i < len(p); i += 2 {
		derivative[i-1] = p[i]
	}
	weights := make([]uint32, len(roots))
	tree.evaluate(b.extension, derivative, weights)
	for i, x := range weights {
		if x == 0 {
			return nil, fmt.Errorf("singular Vandermonde nodes")
		}
		weights[i] = b.extension.inv(x)
	}
	tree.prepare(b.extension)
	return &PaperEncoder{b, tree, p, weights}, nil
}

// Encode solves the transposed Vandermonde system and evaluates over the full
// extension field using an additive FFT. Output is canonical-grid polynomial-
// basis GF(q) elements, with no masks, permutation or application conversion.
func (e *PaperEncoder) Encode(message []uint32) ([]uint32, error) {
	b := e.basis
	if err := b.validateMessage(message); err != nil {
		return nil, err
	}
	k := b.Dimension()
	h := polyMulLow(b.extension, e.p, message, k)
	coefficients := make([]uint32, k)
	e.tree.evaluate(b.extension, h, coefficients)
	word := make([]uint32, b.Length())
	// H/P has simple-pole residues -a_i. Characteristic two removes the sign.
	for j, i := range b.support {
		word[i] = b.extension.mul(coefficients[j], e.weights[j])
	}
	b.extensionFFT.apply(word, false)
	for _, x := range word {
		if x >= b.base.q {
			return nil, fmt.Errorf("encoding escaped the base field")
		}
	}
	return word, nil
}

// CyclicEncoder holds reusable base-field polynomials P and G, P*G=1+X^(N-1).
// It shares the paper information set and emits exactly the same canonical word.
type CyclicEncoder struct {
	basis    *OrbitBasis
	p, g     []uint32
	spectrum []uint32
}

// PrepareCyclic constructs all message-independent cyclic polynomial data. It
// does not retain the temporary extension-field product tree.
func (b *OrbitBasis) PrepareCyclic() (*CyclicEncoder, error) {
	roots := make([]uint32, b.Dimension())
	for j, i := range b.support {
		roots[j] = b.extension.inv(b.extension.pow(b.omega, i))
	}
	p := newProductTree(b.extension, roots).polynomial
	scale := b.extension.inv(p[0])
	for i := range p {
		p[i] = b.extension.mul(p[i], scale)
		if p[i] >= b.base.q {
			return nil, fmt.Errorf("cyclic check polynomial escaped the base field")
		}
	}
	dividend := make([]uint32, b.Length())
	dividend[0] = 1
	dividend[len(dividend)-1] = 1
	g, remainder := polyDiv(b.base, dividend, p)
	if len(remainder) != 0 || g[0] != 1 {
		return nil, fmt.Errorf("invalid cyclic generator")
	}
	return &CyclicEncoder{basis: b, p: p, g: g}, nil
}

// Encode constructs H=(P*M) mod X^K, Y=G*H over the base field, places the
// nonzero-coordinate values in the grid, and extends by the overall parity sum.
func (e *CyclicEncoder) Encode(message []uint32) ([]uint32, error) {
	b := e.basis
	if err := b.validateMessage(message); err != nil {
		return nil, err
	}
	h := polyMulLow(b.base, e.p, message, b.Dimension())
	var y []uint32
	if len(e.spectrum) == 0 {
		y = polyMul(b.base, e.g, h)
	} else {
		y = make([]uint32, b.Length())
		copy(y, h)
		b.extensionFFT.apply(y, false)
		for i := range y {
			y[i] = b.extension.mul(y[i], e.spectrum[i])
		}
		b.extensionFFT.apply(y, true)
		y = y[:b.Length()-1]
		for _, x := range y {
			if x >= b.base.q {
				return nil, fmt.Errorf("cyclic FFT product escaped the base field")
			}
		}
	}
	word := make([]uint32, b.Length())
	point := b.omega
	for _, x := range y {
		word[point] = x
		word[0] ^= x
		point = b.extension.mul(point, b.omega)
	}
	return word, nil
}

// MarshalBinary serializes public discovery data as versioned JSON. This is a
// compact reusable basis artifact, not a database, key, or complete solver plan.
// Loading and PreparePaper/PrepareCyclic are separate setup costs, amortized over
// encodings. Publish artifacts with their reproducible discovery provenance.
func (b *OrbitBasis) MarshalBinary() ([]byte, error) {
	return json.Marshal(struct {
		Version int   `json:"version"`
		Spec    Spec  `json:"spec"`
		Support []int `json:"support"`
	}{1, b.spec, b.support})
}

// LoadOrbit loads a trusted public basis artifact and rebuilds field/transform
// tables without random degree discovery. Validation checks format, bounded
// specifications and Frobenius closure; it does not prove the supplied degree set
// is the intended code. Use artifacts generated and independently verified by the
// project. Prepare a method-specific encoder once after loading, before databases.
func LoadOrbit(data []byte) (*OrbitBasis, error) {
	if len(data) > 2<<20 {
		return nil, fmt.Errorf("basis artifact exceeds 2 MiB")
	}
	var artifact struct {
		Version int   `json:"version"`
		Spec    Spec  `json:"spec"`
		Support []int `json:"support"`
	}
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, err
	}
	spec := artifact.Spec
	polynomial, polynomialErr := BinaryPolynomial(spec.FieldDegree)
	ok := polynomialErr == nil && spec.FieldDegree <= 13
	if artifact.Version != 1 || !ok || spec.Dimension != 2 || spec.CurveDegree < 1 || spec.CurveDegree >= (1<<uint(spec.FieldDegree))-1 {
		return nil, fmt.Errorf("unsupported basis artifact")
	}
	b := newOrbitBasis(spec, polynomial)
	b.support = artifact.Support
	if err := b.validateSupport(); err != nil {
		return nil, err
	}
	return b, nil
}

// PrepareCyclicFFT additionally caches the fixed generator's additive transform.
// Encode multiplies G*H using two transforms and pointwise extension products,
// with O(N log² N) arithmetic including monomial-basis conversion. The prefix
// H=(P*M) mod X^K uses the hybrid low-product backend: pruned Karatsuba below
// its cutoff, additive-FFT products above it. No message spectrum is cached.
// The public artifact grows by N extension-field elements, reused for all databases.
func (b *OrbitBasis) PrepareCyclicFFT() (*CyclicEncoder, error) {
	e, err := b.PrepareCyclic()
	if err != nil {
		return nil, err
	}
	e.spectrum = make([]uint32, b.Length())
	copy(e.spectrum, e.g)
	b.extensionFFT.apply(e.spectrum, false)
	return e, nil
}

// Basis returns the immutable public code data shared by this encoder, including
// its dimension and information-set mapping. It contains no database or keys.
func (e *PaperEncoder) Basis() *OrbitBasis { return e.basis }

// Basis returns the immutable public code data shared by this encoder. Preparing
// another encoder from it does not modify this encoder or its cached constants.
func (e *CyclicEncoder) Basis() *OrbitBasis { return e.basis }
