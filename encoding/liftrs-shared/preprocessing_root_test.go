package liftrs

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
)

// Partition complete Frobenius orbits, not arbitrary exponent ranges: both
// members of a quadratic base-field factor must belong to the same worker.
func TestPreparationRootPartitions(t *testing.T) {
	for _, s := range []int{3, 4, 6} {
		basis, err := DiscoverOrbit(Spec{FieldDegree: s, Dimension: 2, CurveDegree: 2}, OrbitOptions{Random: rand.New(rand.NewSource(106))})
		if err != nil {
			t.Fatal(err)
		}
		noop := func(string) {}
		expected, err := prepareCyclicRoot(basis.base, basis.support, basis.Length(), basis.base.q, basis.rootFactors, nil, nil, noop, nil)
		if err != nil {
			t.Fatal(err)
		}
		original := append([]int(nil), basis.support...)
		for _, workers := range []int{1, 2, 4} {
			for _, mode := range []supportMembership{supportBitmap, supportSorted} {
				partitions, err := partitionRootSupportWith(mode, basis.support, basis.Length(), basis.base.q, workers)
				if err != nil {
					t.Fatal(err)
				}
				low, high := len(partitions[0]), len(partitions[0])
				for _, part := range partitions {
					low = minInt(low, len(part))
					high = max(high, len(part))
				}
				if high-low > 2 {
					t.Fatal("unbalanced orbit degrees")
				}

				combined := []uint32{1}
				for _, support := range partitions {
					if len(support) == 0 {
						continue
					}
					part, err := prepareCyclicRootWith(mode, basis.base, support, basis.Length(), basis.base.q, basis.rootFactors, nil, nil, noop, nil)
					if err != nil {
						t.Fatal(s, workers, err)
					}
					if part[0] != 1 {
						t.Fatal("partial product not normalized")
					}
					combined = polyMul(basis.base, combined, part)
				}
				if !reflect.DeepEqual(combined, expected) {
					t.Fatal("partition/merge changes root", s, workers)
				}
			}
		}
		if !reflect.DeepEqual(original, basis.support) {
			t.Fatal("caller support mutated")
		}
		// A failed durable-root callback must prevent reciprocal/identity execution.
		stop := errors.New("root publication failed")
		generatorStarted := false
		_, _, err = prepareCyclicPolynomials(basis.base, basis.support, basis.Length(), basis.base.q, basis.rootFactors, nil,
			func([]uint32) error { return stop }, func(stage string) {
				if stage == "before-reciprocal" {
					generatorStarted = true
				}
			}, nil)
		if !errors.Is(err, stop) || generatorStarted {
			t.Fatal("root publication failure crossed boundary", err)
		}
	}
}

func TestRootPartitionAdmission(t *testing.T) {
	for _, workers := range []int{0, 65} {
		if _, err := partitionRootSupport([]int{0}, 16, 4, workers); err == nil {
			t.Fatal("invalid worker count admitted")
		}
	}
	if _, err := partitionRootSupport([]int{1}, 16, 4, 2); err == nil {
		t.Fatal("split Frobenius orbit admitted")
	}
}
