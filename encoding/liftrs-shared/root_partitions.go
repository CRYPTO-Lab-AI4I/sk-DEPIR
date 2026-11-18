package liftrs

import (
	"fmt"
	"sort"
)

// partitionRootSupport is an experimental, in-memory scheduling helper, not a
// distributed preprocessing API. A complete Frobenius orbit stays on one worker,
// so each worker uses the same base-field root builder as whole preparation.
// Assign orbit degrees greedily for deterministic balance; return sorted copies
// without changing caller support. This adds one K-index copy, not a scalable
// sharded-support representation for full s18 discovery.
func partitionRootSupport(support []int, n int, q uint64, workers int) ([][]int, error) {
	if workers < 1 || workers > 64 {
		return nil, fmt.Errorf("root partition worker count")
	}
	if err := validateExponentSupport(support, n, q); err != nil {
		return nil, err
	}
	parts := make([][]int, workers)
	// Greedy weights are one or two, so final loads differ by at most two.
	// Reserve once to avoid retaining replaced append buffers during timing.
	for j := range parts {
		parts[j] = make([]int, 0, len(support)/workers+3)
	}
	for _, index := range support {
		partner := int(uint64(index) * q % uint64(n-1))
		if partner < index {
			continue
		}
		owner := 0
		for j := 1; j < workers; j++ {
			if len(parts[j]) < len(parts[owner]) {
				owner = j
			}
		}
		parts[owner] = append(parts[owner], index)
		if partner != index {
			parts[owner] = append(parts[owner], partner)
		}
	}
	for j := range parts {
		sort.Ints(parts[j])
	}
	return parts, nil
}
