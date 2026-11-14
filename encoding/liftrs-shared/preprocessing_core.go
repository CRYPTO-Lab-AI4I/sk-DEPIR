package liftrs

import (
	"crypto/rand"
	"fmt"
	"io"
)

// Discovery uses the same sampling/support-union algorithm at either storage
// width. Executors evaluate the m2 base grid and interpolate the extension grid;
// monomial counting, random sampling and acceptance do not depend on backend.
type preparationWord interface{ ~uint32 | ~uint64 }
type preparationNotice func(stage string, completed, total uint64, unit string)

func normalizedOrbitOptions(options OrbitOptions) (OrbitOptions, error) {
	if options.MaxSamples < 0 {
		return options, fmt.Errorf("negative sample limit")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	return options, nil
}

func discoverSupport[Word preparationWord](spec Spec, good []int, options OrbitOptions,
	evaluate func([]Word) error, notice preparationNotice) (support []int, samples int, err error) {
	options, err = normalizedOrbitOptions(options)
	if err != nil {
		return nil, 0, err
	}
	if options.MaxSamples == 0 {
		options.MaxSamples = 2 + (128+spec.FieldDegree-1)/spec.FieldDegree
	}
	q := 1 << uint(spec.FieldDegree)
	n := q * q
	if len(good) == 0 || len(good) >= n {
		return nil, 0, fmt.Errorf("invalid exact monomial dimension")
	}
	found := newExponentSet(n)
	count := 0
	grid := make([]Word, n)
	sampleBytes := 1
	if spec.FieldDegree > 8 {
		sampleBytes = 4
	}
	// Streaming preserves the original coefficient/random-byte order. Reuse one
	// grid and a bounded random buffer instead of retaining a4K-byte random array.
	var random [65536]byte
	for samples < options.MaxSamples && count < len(good) {
		clear(grid)
		for first := 0; first < len(good); {
			take := minInt(len(random)/sampleBytes, len(good)-first)
			if _, err = io.ReadFull(options.Random, random[:sampleBytes*take]); err != nil {
				return nil, samples, fmt.Errorf("degree-set sampling: %w", err)
			}
			for j := 0; j < take; j++ {
				offset := j * sampleBytes
				value := uint32(random[offset])
				if sampleBytes == 4 {
					value |= uint32(random[offset+1])<<8 | uint32(random[offset+2])<<16 | uint32(random[offset+3])<<24
				}
				grid[good[first+j]] = Word(value & uint32(q-1))
			}
			first += take
			if notice != nil {
				notice("discovery-sample", uint64(first), uint64(len(good)), "coefficients")
			}
		}
		// Packed a+b*u has index a+q*b, matching the canonical grid.
		// Extension interpolation therefore needs no permutation table.
		if err = evaluate(grid); err != nil {
			return nil, samples, err
		}
		for first := 0; first < n; first += 65536 {
			for i := first; i < minInt(n, first+65536); i++ {
				if grid[i] != 0 && !found.has(i) {
					found.add(i)
					count++
				}
			}
			if notice != nil {
				notice("discovery-support-union", uint64(minInt(n, first+65536)), uint64(n), "positions")
			}
		}
		samples++
	}
	if count != len(good) {
		return nil, samples, fmt.Errorf("degree discovery incomplete: %d of %d exponents after %d samples", count, len(good), samples)
	}
	support = make([]int, 0, count)
	for first := 0; first < n; first += 65536 {
		for i := first; i < minInt(n, first+65536); i++ {
			if found.has(i) {
				support = append(support, i)
			}
		}
		if notice != nil {
			notice("discovery-support-output", uint64(minInt(n, first+65536)), uint64(n), "positions")
		}
	}
	if err = validateExponentSupport(support, n, uint64(q)); err != nil {
		return nil, samples, err
	}
	return support, samples, nil
}

// Cardinality and q-Frobenius arithmetic remain64-bit even when field values
// fit32-bit slots. The input index slice is immutable and sorted by discovery.
func validateExponentSupport(support []int, n int, q uint64) error {
	order := n - 1
	if n < 4 || len(support) == 0 || len(support) > order {
		return fmt.Errorf("invalid degree-set size")
	}
	seen := newExponentSet(order)
	for j, i := range support {
		if i < 0 || i >= order || (j > 0 && i <= support[j-1]) {
			return fmt.Errorf("degree set must be sorted, distinct and exclude N-1")
		}
		seen.add(i)
	}
	for _, i := range support {
		if !seen.has(int(uint64(i) * q % uint64(order))) {
			return fmt.Errorf("degree set is not Frobenius closed")
		}
	}
	return nil
}

type rootFactorBatch func(indices []int, output [][4]uint32) error

// One root/reciprocal/identity procedure serves the existing GPU-assisted path
// and the wider server executor. Polynomial coefficients always stay uint32.
// Callers supply only extension-root arithmetic, products and stage observation.
func prepareCyclicPolynomials(base orbitField, support []int, n int, q uint64,
	roots rootFactorBatch, savedRoot []uint32, saveRoot func([]uint32) error,
	checkpoint func(string), notice preparationNotice) (p, g []uint32, err error) {
	checkpoint("before-validation")
	if err = validateExponentSupport(support, n, q); err != nil {
		return nil, nil, err
	}
	p = savedRoot
	if p == nil {
		checkpoint("before-root-product")
		builder := newFactorProductBuilder(base, minInt(32768, len(support)))
		batch := make([][4]uint32, minInt(65536, len(support)))
		for first := 0; first < len(support); first += len(batch) {
			take := minInt(len(batch), len(support)-first)
			if err = roots(support[first:first+take], batch[:take]); err != nil {
				return nil, nil, err
			}
			for j, index := range support[first : first+take] {
				partner := int(uint64(index) * q % uint64(n-1))
				factor := batch[j]
				if partner < index {
					continue
				}
				degree := uint32(2)
				if partner == index {
					degree = 1
				}
				if factor[3] != degree || uint64(factor[0]) >= q || uint64(factor[1]) >= q || factor[2] != (degree-1) {
					return nil, nil, fmt.Errorf("invalid root factor")
				}
				builder.add(factor[0], factor[1], factor[2], int(degree))
			}
			if notice != nil {
				notice("root-product", uint64(first+take), uint64(len(support)), "exponents")
			}
		}
		p = builder.finish()
		checkpoint("before-normalization")
		if len(p) != len(support)+1 {
			return nil, nil, fmt.Errorf("paired product degree mismatch")
		}
		scale := base.inv(p[0])
		for i := range p {
			p[i] = base.mul(p[i], scale)
		}
		if saveRoot != nil {
			if err = saveRoot(p); err != nil {
				return nil, nil, err
			}
		}
	} else {
		checkpoint("reuse-normalized-root")
		if len(p) != len(support)+1 || p[0] != 1 {
			return nil, nil, fmt.Errorf("root checkpoint shape/normalization")
		}
		for _, value := range p {
			if uint64(value) >= q {
				return nil, nil, fmt.Errorf("root checkpoint field")
			}
		}
	}
	checkpoint("before-reciprocal")
	g = polyInverse(base, p, n-len(support))
	checkpoint("before-identity")
	// This full, untruncated product is independent of Newton's prefix iteration.
	// Verify before publishing code data, even when a root checkpoint was reused.
	identity := polyMul(base, p, g)
	if len(identity) != n {
		return nil, nil, fmt.Errorf("generator identity degree mismatch")
	}
	for first := 0; first < n; first += 65536 {
		for i := first; i < minInt(n, first+65536); i++ {
			want := uint32(0)
			if i == 0 || i == n-1 {
				want = 1
			}
			if identity[i] != want {
				return nil, nil, fmt.Errorf("generator identity mismatch at %d", i)
			}
		}
		if notice != nil {
			notice("identity-verification", uint64(minInt(n, first+65536)), uint64(n), "coefficients")
		}
	}
	return p, g, nil
}

// Independent scalar root arithmetic remains the small-code reference and
// current32-bit GPU-harness adapter; it shares the polynomial preparation flow.
func (b *OrbitBasis) rootFactors(indices []int, out [][4]uint32) error {
	for j, i := range indices {
		out[j] = [4]uint32{}
		partner := int(uint64(i) * b.base.q % uint64(b.Length()-1))
		if partner < i {
			continue
		}
		root := b.extension.inv(b.extension.pow(b.omega, i))
		if partner == i {
			if uint64(root) >= b.base.q {
				return fmt.Errorf("fixed root outside base field")
			}
			out[j] = [4]uint32{root, 1, 0, 1}
		} else {
			conjugate := b.extension.pow(root, int(b.base.q))
			trace, norm := root^conjugate, b.extension.mul(root, conjugate)
			if uint64(trace) >= b.base.q || uint64(norm) >= b.base.q {
				return fmt.Errorf("paired root factor outside base field")
			}
			out[j] = [4]uint32{norm, trace, 1, 2}
		}
	}
	return nil
}
