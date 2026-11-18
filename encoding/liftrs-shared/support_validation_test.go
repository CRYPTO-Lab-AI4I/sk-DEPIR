package liftrs

import (
	"reflect"
	"runtime"
	"sort"
	"testing"
)

func TestSupportMembershipEquivalence(t *testing.T) {
	// Exhaust every subset, including excluded N-1 and absent conjugates. This
	// checks rejection as well as acceptance without trusting either root builder.
	for mask := 0; mask < 1<<16; mask++ {
		var support []int
		for index := 0; index < 16; index++ {
			if mask&(1<<index) != 0 {
				support = append(support, index)
			}
		}
		a := validateExponentSupportWith(supportBitmap, support, 16, 4)
		b := validateExponentSupportWith(supportSorted, support, 16, 4)
		if (a == nil) != (b == nil) || (a != nil && a.Error() != b.Error()) {
			t.Fatal("membership disagreement", mask, a, b)
		}
	}
	for _, support := range [][]int{{-1}, {16}, {1, 1}, {4, 1}, {0, 4, 1}, {1}} {
		for _, mode := range []supportMembership{supportBitmap, supportSorted} {
			if err := validateExponentSupportWith(mode, support, 16, 4); err == nil {
				t.Fatal("invalid support accepted", support, mode)
			}
		}
	}
	if err := validateExponentSupportWith(99, []int{0}, 16, 4); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func TestSortedSupportWideMemoryAndProgress(t *testing.T) {
	const n = 1 << 34
	const q = 1 << 17
	const exponent = (1 << 32) + 5
	support := []int{0, exponent, int(uint64(exponent) * q % uint64(n-1))}
	sort.Ints(support)
	original := append([]int(nil), support...)
	var stages []string
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	err := validateExponentSupportWith(supportSorted, support, n, q, func(stage string, done, total uint64, unit string) {
		if done != uint64(len(support)) || done != total || unit != "exponents" {
			t.Fatal("progress changed")
		}
		stages = append(stages, stage)
	})
	runtime.ReadMemStats(&after)
	if err != nil {
		t.Fatal(err)
	}
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatal("sorted validation allocated domain-sized storage")
	}
	if !reflect.DeepEqual(original, support) {
		t.Fatal("borrowed support mutated")
	}
	if !reflect.DeepEqual(stages, []string{"support-order-validation", "support-Frobenius-validation"}) {
		t.Fatal("missing progress", stages)
	}
	parts, err := partitionRootSupportWith(supportSorted, support, n, q, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, part := range parts {
		if len(part) > 0 {
			if err := validateExponentSupportWith(supportSorted, part, n, q); err != nil {
				t.Fatal(err)
			}
		}
	}
}
