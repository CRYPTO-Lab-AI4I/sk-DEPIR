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
	return partitionRootSupportWith(supportBitmap, support, n, q, workers)
}

// The explicit subset-validator choice changes storage/search cost only.
func partitionRootSupportWith(mode supportMembership, support []int, n int, q uint64, workers int) ([][]int, error) {
	if workers < 1 || workers > 64 {
		return nil, fmt.Errorf("root partition worker count")
	}
	parts := make([][]int, workers)
	err := routeRootSupport(mode, support, n, q, workers, 16384, func(owner int, packet []uint64) error {
		if parts[owner] == nil {
			parts[owner] = make([]int, 0, len(support)/workers+3)
		}
		for _, index := range packet {
			parts[owner] = append(parts[owner], int(index))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for j := range parts {
		sort.Ints(parts[j])
	}
	return parts, nil
}

// routeRootSupport validates the coordinator's source once, then emits bounded
// orbit-closed packets to deterministic least-loaded owners. Callback slices are
// borrowed only until return. Receiver shards require sorting before the shared
// sorted-membership root builder. Counters/buffers cost O(workers*packetWords),
// avoiding the extra coordinator K-index partition copy. The source itself is
// still resident: this is not distributed discovery or fulls18 admission.
func routeRootSupport(mode supportMembership, support []int, n int, q uint64, workers, packetWords int,
	emit func(int, []uint64) error) error {
	if workers < 1 || workers > 64 || packetWords < 2 || packetWords > 16384 || emit == nil {
		return fmt.Errorf("root routing bounds/callback")
	}
	if err := validateExponentSupportWith(mode, support, n, q); err != nil {
		return err
	}
	buffers := make([][]uint64, workers)
	loads := make([]int, workers)
	flush := func(owner int) error {
		if len(buffers[owner]) == 0 {
			return nil
		}
		if err := emit(owner, buffers[owner]); err != nil {
			return err
		}
		buffers[owner] = buffers[owner][:0]
		return nil
	}
	for _, index := range support {
		partner := int(uint64(index) * q % uint64(n-1))
		if partner < index {
			continue
		}
		owner := 0
		for j := 1; j < workers; j++ {
			if loads[j] < loads[owner] {
				owner = j
			}
		}
		degree := 1
		if partner != index {
			degree = 2
		}
		if len(buffers[owner])+degree > packetWords {
			if err := flush(owner); err != nil {
				return err
			}
		}
		if buffers[owner] == nil {
			buffers[owner] = make([]uint64, 0, minInt(packetWords, len(support)/workers+3))
		}
		buffers[owner] = append(buffers[owner], uint64(index))
		if degree == 2 {
			buffers[owner] = append(buffers[owner], uint64(partner))
		}
		loads[owner] += degree
	}
	for owner := range buffers {
		if err := flush(owner); err != nil {
			return err
		}
	}
	return nil
}
