package planner

import "sort"

// blockCounts enumerates every count when it fits the budget. Otherwise sample
// zero, one, filling counts, then powers of two from either end of the range.
// A final count may pad one block. Count sampling has no approximation bound.
func blockCounts(rows, capacity uint64, limit int) ([]uint64, bool) {
	full := rows / capacity
	last := full
	if rows%capacity != 0 {
		last++
	}
	if last < uint64(limit) {
		values := make([]uint64, last+1)
		for i := range values {
			values[i] = uint64(i)
		}
		return values, false
	}
	seen := map[uint64]bool{}
	add := func(n uint64) {
		if n <= last && len(seen) < limit {
			seen[n] = true
		}
	}
	add(0)
	add(1)
	add(full)
	add(last)
	for step := uint64(1); step <= full && len(seen) < limit; {
		add(full - step)
		add(step)
		if step > full/2 {
			break
		}
		step *= 2
	}
	values := make([]uint64, 0, len(seen))
	for n := range seen {
		values = append(values, n)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	return values, true
}

// appendBatch stores a full repeated run compactly, plus at most one padded
// final block. Work and metadata do not grow with the repetition count.
func appendBatch(prefix *partialPlan, count, total uint64, choice blockOption) (*partialPlan, error) {
	if count == 0 {
		return prefix, nil
	}
	remaining := total - prefix.end
	full := minWidth(count, remaining/choice.capacity)
	next := prefix
	if full > 0 {
		cost, err := scaleModelCosts(choice.costs, full)
		if err == nil {
			cost, err = addModelCosts(prefix.costs, cost)
		}
		if err != nil {
			return nil, err
		}
		end := prefix.end + full*choice.capacity
		next = &partialPlan{end: end, costs: cost, previous: prefix, block: RowBlock{RowStart: prefix.end, RowEnd: end, ConfigurationID: choice.id, Repeat: full}}
	}
	if count > full {
		cost, err := addModelCosts(next.costs, choice.costs)
		if err != nil {
			return nil, err
		}
		next = &partialPlan{end: total, costs: cost, previous: next, block: RowBlock{RowStart: next.end, RowEnd: total, ConfigurationID: choice.id}}
	}
	return next, nil
}

func stateKeys(states map[uint64][]*partialPlan) []uint64 {
	keys := make([]uint64, 0, len(states))
	for remaining := range states {
		keys = append(keys, remaining)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// boundStates bounds all live partial plans at a configuration stage, not the
// number of historical row positions. Sample remaining-row keys evenly, then
// round-robin their cost endpoints. This is lossy, not cross-state dominance.
func boundStates(states map[uint64][]*partialPlan, limit int) (map[uint64][]*partialPlan, bool) {
	count := 0
	for _, nodes := range states {
		count += len(nodes)
	}
	if count <= limit {
		return states, false
	}
	keys := stateKeys(states)
	if len(keys) > limit {
		selected := make([]uint64, 0, limit)
		for i := 0; i < limit; i++ {
			index := 0
			if limit > 1 {
				index = int(float64(i) * float64(len(keys)-1) / float64(limit-1))
			}
			selected = append(selected, keys[index])
		}
		keys = selected
	}
	result := map[uint64][]*partialPlan{}
	used := 0
	for rank := 0; used < limit; rank++ {
		added := false
		for _, key := range keys {
			nodes := states[key]
			if rank >= len(nodes) {
				continue
			}
			index := rank / 2
			if rank%2 != 0 {
				index = len(nodes) - 1 - rank/2
			}
			result[key] = append(result[key], nodes[index])
			used++
			added = true
			if used == limit {
				break
			}
		}
		if !added {
			break
		}
	}
	return result, true
}
