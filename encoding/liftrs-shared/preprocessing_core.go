package liftrs

import (
	"crypto/rand"
	"fmt"
	"io"
	"sort"
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
	evaluate func([]Word) error, notice preparationNotice, recoveries ...*PreprocessingRecovery) (support []int, samples int, err error) {
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
	var recovery *PreprocessingRecovery
	if len(recoveries) > 0 {
		recovery = recoveries[0]
	}
	var info SupportParameters
	if recovery != nil {
		h, _, e := canonicalScalingCodeHeader(spec.FieldDegree, spec.CurveDegree, 1, uint64(n))
		if e != nil {
			return nil, 0, e
		}
		info = SupportParameters{Spec: spec, K: uint64(len(good)), N: uint64(n), Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega}
	}
	if recovery != nil && recovery.Discovery != nil {
		c := recovery.Discovery
		if err = c.matches(info, "", "discovery"); err != nil {
			return nil, 0, err
		}
		found = append(exponentSet(nil), c.bitmap...)
		count = int(c.record.Count)
		samples = c.record.Parameters.Samples
		if samples > options.MaxSamples {
			return nil, 0, fmt.Errorf("saved samples exceed current limit")
		}
	}
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
		if recovery != nil && recovery.Save != nil {
			info.Samples = samples
			c := &PreprocessingCheckpoint{record: ProgressRecord{Kind: "discovery", Parameters: info, Count: uint64(count)}, bitmap: found}
			if err = recovery.Save(c); err != nil {
				return nil, samples, err
			}
		}
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
	if err = validateExponentSupport(support, n, uint64(q), notice); err != nil {
		return nil, samples, err
	}
	return support, samples, nil
}

// supportMembership selects only the lookup representation, never which
// checks run. Ordinary discovery, cache loading and preparation stay bitmap.
// Sorted membership borrows the validated ordered input: O(KlogK) searches,
// no N-sized allocation. Its dense-support latency trade-off is unmeasured.
type supportMembership uint8

const (
	supportBitmap supportMembership = iota
	supportSorted
)

// Cardinality and q-Frobenius arithmetic remain64-bit even when field values
// fit32-bit slots. The input index slice is immutable and sorted by discovery.
func validateExponentSupport(support []int, n int, q uint64, notices ...preparationNotice) error {
	return validateExponentSupportWith(supportBitmap, support, n, q, notices...)
}

// Cardinality, strict ordering and range checks finish before binary search.
// Both representations verify every q-Frobenius partner, including wide indices.
func validateExponentSupportWith(mode supportMembership, support []int, n int, q uint64, notices ...preparationNotice) error {
	notice := newPreparationReporter(nil).notice
	if len(notices) > 0 && notices[0] != nil {
		notice = notices[0]
	}
	order := n - 1
	if n < 4 || len(support) == 0 || len(support) > order {
		return fmt.Errorf("invalid degree-set size")
	}
	if mode != supportBitmap && mode != supportSorted {
		return fmt.Errorf("invalid support membership mode")
	}
	var seen exponentSet
	if mode == supportBitmap {
		seen = newExponentSet(order)
	}
	for first := 0; first < len(support); first += 65536 {
		end := minInt(first+65536, len(support))
		for j := first; j < end; j++ {
			i := support[j]
			if i < 0 || i >= order || (j > 0 && i <= support[j-1]) {
				return fmt.Errorf("degree set must be sorted, distinct and exclude N-1")
			}
			if mode == supportBitmap {
				seen.add(i)
			}
		}
		notice("support-order-validation", uint64(end), uint64(len(support)), "exponents")
	}
	for first := 0; first < len(support); first += 65536 {
		end := minInt(first+65536, len(support))
		for _, i := range support[first:end] {
			partner := int(uint64(i) * q % uint64(order))
			found := false
			if mode == supportBitmap {
				found = seen.has(partner)
			} else {
				index := sort.SearchInts(support, partner)
				found = index < len(support) && support[index] == partner
			}
			if !found {
				return fmt.Errorf("degree set is not Frobenius closed")
			}
		}
		notice("support-Frobenius-validation", uint64(end), uint64(len(support)), "exponents")
	}
	return nil
}

type rootFactorBatch func(indices []int, output [][4]uint32) error

// Root construction and generator preparation share arithmetic across backends,
// but have separate ownership boundaries: generator work needs K, not support.
func prepareCyclicPolynomials(base orbitField, support []int, n int, q uint64,
	roots rootFactorBatch, savedRoot []uint32, saveRoot func([]uint32) error,
	checkpoint func(string), notice preparationNotice, recoveries ...*cyclicRecovery) (p, g []uint32, err error) {
	k := len(support)
	var recovery *cyclicRecovery
	var progress *PreprocessingCheckpoint
	if len(recoveries) > 0 {
		recovery = recoveries[0]
	}
	// Snapshot the input recovery before root/checkpoint callbacks run, preserving
	// the original resume boundary even if a callback updates caller-owned options.
	if recovery != nil && recovery.recovery != nil {
		progress = recovery.recovery.Preparation
	}
	p, err = prepareCyclicRoot(base, support, n, q, roots, savedRoot, saveRoot, checkpoint, notice, recoveries...)
	if err != nil {
		return nil, nil, err
	}
	var prefix []uint32
	if progress != nil && progress.record.Kind == "reciprocal" {
		prefix = progress.sections[0]
	}
	g, err = prepareCyclicGenerator(base, p, n, k, prefix, checkpoint, notice, recovery)
	if err != nil {
		return nil, nil, err
	}
	return p, g, nil
}

// prepareCyclicRoot validates a sorted Frobenius-closed support and returns its
// normalized product. Independent closed subsets can use this same builder;
// multiplying their normalized products recovers the full normalized root.
// Caller-owned support and saved roots remain immutable; checkpoints keep their
// original source/support binding and are validated before any reuse.
func prepareCyclicRoot(base orbitField, support []int, n int, q uint64,
	roots rootFactorBatch, savedRoot []uint32, saveRoot func([]uint32) error,
	checkpoint func(string), notice preparationNotice, recoveries ...*cyclicRecovery) (p []uint32, err error) {
	return prepareCyclicRootWith(supportBitmap, base, support, n, q, roots, savedRoot, saveRoot, checkpoint, notice, recoveries...)
}

// Experimental explicit validation storage choice; all root arithmetic and
// checkpoint validation remain in this shared implementation.
func prepareCyclicRootWith(mode supportMembership, base orbitField, support []int, n int, q uint64,
	roots rootFactorBatch, savedRoot []uint32, saveRoot func([]uint32) error,
	checkpoint func(string), notice preparationNotice, recoveries ...*cyclicRecovery) (p []uint32, err error) {
	checkpoint("before-validation")
	if err = validateExponentSupportWith(mode, support, n, q, notice); err != nil {
		return nil, err
	}
	var recovery *cyclicRecovery
	if len(recoveries) > 0 {
		recovery = recoveries[0]
	}
	var progress *PreprocessingCheckpoint
	if recovery != nil && recovery.recovery != nil {
		progress = recovery.recovery.Preparation
		if progress != nil {
			if progress.record.Kind != "roots" && progress.record.Kind != "reciprocal" {
				return nil, fmt.Errorf("invalid preparation checkpoint kind")
			}
			if err = progress.matches(recovery.info, recovery.supportSHA, progress.record.Kind); err != nil {
				return nil, err
			}
			if progress.record.Kind == "reciprocal" && savedRoot == nil {
				return nil, fmt.Errorf("reciprocal requires completed root")
			}
		}
	}
	p = savedRoot
	if p == nil {
		checkpoint("before-root-product")
		builder := newFactorProductBuilder(base, minInt(32768, len(support)))
		batch := make([][4]uint32, minInt(65536, len(support)))
		firstSaved := 0
		if progress != nil && progress.record.Kind == "roots" {
			builder, err = progress.restoreBuilder(base, support, n, q, notice)
			if err != nil {
				return nil, err
			}
			firstSaved = int(progress.record.Next)
		}
		for first := firstSaved; first < len(support); first += len(batch) {
			take := minInt(len(batch), len(support)-first)
			if err = roots(support[first:first+take], batch[:take]); err != nil {
				return nil, err
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
					return nil, fmt.Errorf("invalid root factor")
				}
				builder.add(factor[0], factor[1], factor[2], int(degree))
			}
			if notice != nil {
				notice("root-product", uint64(first+take), uint64(len(support)), "exponents")
			}
			if recovery != nil && recovery.recovery != nil && recovery.recovery.Save != nil {
				sections := make([][]uint32, 1, 1+len(builder.levels))
				sections[0] = builder.storage[:3*len(builder.factors)]
				sections = append(sections, builder.levels...)
				if err = recovery.save("roots", uint64(first+take), sections); err != nil {
					return nil, err
				}
			}
		}
		p = builder.finish()
		// finish may return one retained level directly; normalization must not mutate
		// the caller's loaded checkpoint if it is reused again.
		if progress != nil && progress.record.Kind == "roots" {
			for _, level := range progress.sections[1:] {
				if len(level) > 0 && len(p) > 0 && &p[0] == &level[0] {
					p = append([]uint32(nil), p...)
					break
				}
			}
		}
		checkpoint("before-normalization")
		if len(p) != len(support)+1 {
			return nil, fmt.Errorf("paired product degree mismatch")
		}
		scale := base.inv(p[0])
		for first := 0; first < len(p); first += 65536 {
			end := minInt(first+65536, len(p))
			block := p[first:end]
			for i := range block {
				block[i] = base.mul(block[i], scale)
			}
			if notice != nil {
				notice("normalization", uint64(end), uint64(len(p)), "coefficients")
			}
		}
		if saveRoot != nil {
			if err = saveRoot(p); err != nil {
				return nil, err
			}
		}
	} else {
		checkpoint("reuse-normalized-root")
		if len(p) != len(support)+1 || p[0] != 1 {
			return nil, fmt.Errorf("root checkpoint shape/normalization")
		}
		for first := 0; first < len(p); first += 65536 {
			end := minInt(first+65536, len(p))
			for _, value := range p[first:end] {
				if uint64(value) >= q {
					return nil, fmt.Errorf("root checkpoint field")
				}
			}
			if notice != nil {
				notice("root-field-validation", uint64(end), uint64(len(p)), "coefficients")
			}
		}
	}
	return p, nil
}

// The reciprocal and independent full identity need only normalized P, K and N.
// Recovery validation happens at the root boundary before this internal call.
// This separation permits future consuming callers to release their own support
// references; it does not force GC or release data owned by a public caller.
func prepareCyclicGenerator(base orbitField, p []uint32, n, k int, prefix []uint32,
	checkpoint func(string), notice preparationNotice, recovery *cyclicRecovery) ([]uint32, error) {
	checkpoint("before-reciprocal")
	g := polyInverseRecover(base, p, n-k, prefix, func(values []uint32) error {
		if notice != nil {
			notice("reciprocal-prefix", uint64(len(values)), uint64(n-k), "coefficients")
		}
		if recovery == nil {
			return nil
		}
		return recovery.save("reciprocal", uint64(len(values)), [][]uint32{values})
	})
	checkpoint("before-identity")
	// This full, untruncated product is independent of Newton's prefix iteration.
	// Verify before publishing code data, even when a root checkpoint was reused.
	identity := polyMul(base, p, g)
	if len(identity) != n {
		return nil, fmt.Errorf("generator identity degree mismatch")
	}
	for first := 0; first < n; first += 65536 {
		for i := first; i < minInt(n, first+65536); i++ {
			want := uint32(0)
			if i == 0 || i == n-1 {
				want = 1
			}
			if identity[i] != want {
				return nil, fmt.Errorf("generator identity mismatch at %d", i)
			}
		}
		if notice != nil {
			notice("identity-verification", uint64(minInt(n, first+65536)), uint64(n), "coefficients")
		}
	}
	return g, nil
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

// Optional recovery hooks surround the shared arithmetic; ordinary callers retain
// the same algorithm and allocation ownership. Errors unwind through Prepare.
type cyclicRecovery struct {
	recovery   *PreprocessingRecovery
	info       SupportParameters
	supportSHA string
}

func (r *cyclicRecovery) save(kind string, next uint64, sections [][]uint32) error {
	if r == nil || r.recovery == nil || r.recovery.Save == nil {
		return nil
	}
	return r.recovery.Save(&PreprocessingCheckpoint{record: ProgressRecord{Kind: kind, Parameters: r.info, SupportSHA256: r.supportSHA, Next: next}, sections: sections})
}
