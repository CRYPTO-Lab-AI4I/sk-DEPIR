package planner

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// TradeoffOptions bounds capacity-ordered search over full-width row blocks.
// LogicalRecordBits follows ApplicationOptions. Zero selects defaults:
// MaxFrontier=32 per remaining-row state and exploratory completed set;
// MaxStates=256 live partial plans per configuration stage; MaxCounts=8 sampled
// block counts per transition; MaxBaselineBlocks=256 blocks per greedy baseline.
// Protected witnesses may increase the returned set beyond MaxFrontier.
// MaxFrontier must be >=2, MaxCounts >=4, other limits positive.
type TradeoffOptions struct {
	MaxCounts         int      `json:"max_counts,omitempty"`
	MaxBaselineBlocks int      `json:"max_baseline_blocks,omitempty"`
	LogicalRecordBits []uint64 `json:"logical_record_bits,omitempty"`
	MaxFrontier       int      `json:"max_frontier,omitempty"`
	MaxStates         int      `json:"max_states,omitempty"`
}

// Alternative pairs a validated application layout with its cost and evidence.
type Alternative struct {
	Plan       ApplicationPlan `json:"plan"`
	Evaluation *Evaluation     `json:"evaluation"`
}

// BaselineWitness identifies a successfully constructed feasible baseline and an
// alternative no worse in storage and download. Strict improvement is
// not required. AlternativeIndex addresses the final sorted Alternatives slice.
type BaselineWitness struct {
	LogicalRecordBits uint64 `json:"logical_record_bits"`
	FullDownloadOnly  bool   `json:"full_download_only"`
	Costs             Costs  `json:"costs"`
	AlternativeIndex  int    `json:"alternative_index"`
}

// TradeoffResult contains nondominated alternatives ordered by storage then
// communication. FrontierTrimmed, StateLimitReached and CountChoicesTrimmed
// report lossy approximations; BaselineLimitReached reports unfinished baselines.
// StatesExplored sums processed partial plans across configuration stages/widths.
// Without truncation, search is exhaustive within the specified width/full-block
// family plus one arbitrary final padded block or full-download suffix.
type TradeoffResult struct {
	CountChoicesTrimmed  bool              `json:"count_choices_trimmed"`
	BaselineWitnesses    []BaselineWitness `json:"baseline_witnesses"`
	Alternatives         []Alternative     `json:"alternatives"`
	FrontierTrimmed      bool              `json:"frontier_trimmed"`
	BaselineLimitReached bool              `json:"baseline_limit_reached"`
	StateLimitReached    bool              `json:"state_limit_reached"`
	StatesExplored       uint64            `json:"states_explored"`
}

type partialPlan struct {
	end      uint64
	costs    Costs
	previous *partialPlan
	block    RowBlock
}

// SearchTradeoffs returns storage/download alternatives. It processes full
// blocks in descending capacity order, batching repetitions. Reordering disjoint
// full blocks does not change costs; a final padded block may use any configuration.
// Full-download suffixes are ordinary candidates. Completed feasible greedy and
// whole-download baselines bypass lossy reduction and retain explicit witnesses.
//
// At the same configuration stage and remaining row count, possible completions
// are identical, so dominance pruning is sound. With an upload budget compare
// storage/upload/download separately to preserve feasibility; otherwise compare
// storage/download. Download limits fit the existing objective dimensions.
// State and count sampling are bounded approximations, not cross-state dominance.
// No multiplicative error bound or global optimality is promised. ErrNoPlan returns
// nonnil search metadata and does not prove infeasibility after truncation.
func SearchTradeoffs(rows, recordBits uint64, catalogue *pirconfig.Catalogue, r Requirements, options TradeoffOptions) (*TradeoffResult, error) {
	if err := catalogue.Validate(); err != nil {
		return nil, err
	}
	if rows == 0 || recordBits == 0 {
		return nil, fmt.Errorf("planner: application dimensions must be nonempty")
	}
	if _, err := multiply(rows, recordBits); err != nil {
		return nil, err
	}
	if options.MaxFrontier == 0 {
		options.MaxFrontier = 32
	}
	if options.MaxCounts == 0 {
		options.MaxCounts = 8
	}
	if options.MaxBaselineBlocks == 0 {
		options.MaxBaselineBlocks = 256
	}
	if options.MaxStates == 0 {
		options.MaxStates = 256
	}
	if options.MaxFrontier < 2 || options.MaxStates < 1 || options.MaxCounts < 4 || options.MaxBaselineBlocks < 1 {
		return nil, fmt.Errorf("planner: require max_frontier >= 2, max_states >= 1, max_counts >= 4, max_baseline_blocks >= 1")
	}
	widths, err := logicalWidths(recordBits, catalogue, options.LogicalRecordBits)
	if err != nil {
		return nil, err
	}
	result := &TradeoffResult{}
	var completed []*partialPlan
	var layouts, protected []Alternative
	add := func(plan ApplicationPlan, baseline bool) {
		evaluation, err := EvaluateApplication(plan, catalogue, r)
		if err == nil {
			alternative := Alternative{Plan: plan, Evaluation: evaluation}
			if baseline {
				protected = append(protected, alternative)
			} else {
				layouts = append(layouts, alternative)
			}
		}
	}
	// Always include the identity matrix, even when indexed input is too narrow.
	add(ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: recordBits, LogicalPlan: Plan{FormatVersion: 1, Rows: rows, RecordBits: recordBits, RowBlocks: []RowBlock{{RowEnd: rows, DownloadAll: true}}}}, true)
	indexedAllowed := true
	for _, c := range catalogue.Configurations {
		if recordBits < uint64(c.Code.S) {
			indexedAllowed = false
		}
	}
	for _, width := range widths {
		mapping := ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: width}
		fragments, _ := mapping.FragmentCount()
		total, err := multiply(rows, fragments)
		if err != nil || !indexedAllowed {
			continue
		}
		// Protect useful greedy baselines from frontier and boundary truncation.
		if baseline, _, err := greedyWithBlockLimit(total, width, catalogue, r, false, fragments, 0, 1, options.MaxBaselineBlocks); err == nil {
			mapping.LogicalPlan = baseline
			add(mapping, true)
		} else if errors.Is(err, errBaselineLimit) {
			result.BaselineLimitReached = true
		}
		var choices []blockOption
		for i := range catalogue.Configurations {
			c := &catalogue.Configurations[i]
			if c.Capacity.Status == "unknown" || c.Security.ComputationalBits < r.MinSecurityBits || c.Security.Log2Samples < r.Log2Samples {
				continue
			}
			count := width / uint64(c.Code.S)
			if width%uint64(c.Code.S) != 0 {
				count++
			}
			cost, err := databaseCosts(c, count)
			if err == nil {
				cost, err = repeatTraffic(cost, fragments)
			}
			if err == nil {
				choices = append(choices, blockOption{id: c.ID, capacity: *c.Capacity.K, costs: cost})
			}
		}
		// Configuration order removes redundant permutations of disjoint row blocks.
		sort.Slice(choices, func(i, j int) bool {
			if choices[i].capacity != choices[j].capacity {
				return choices[i].capacity > choices[j].capacity
			}
			return choices[i].id < choices[j].id
		})
		states := map[uint64][]*partialPlan{total: {{}}}
		completed = nil
		finish := func(prefix *partialPlan) {
			if prefix.end == total {
				completed = append(completed, prefix)
				return
			}
			// Only full blocks are capacity-ordered. The single final padded block
			// may use any configuration; forcing its order would omit valid mixtures.
			for _, choice := range choices {
				if choice.capacity < total-prefix.end {
					continue
				}
				next, err := appendBatch(prefix, 1, total, choice)
				if err == nil && withinBudgets(next.costs, r) {
					completed = append(completed, next)
				}
			}
			if tail, ok := fullDownloadCosts(total-prefix.end, width, math.MaxUint64); ok {
				if cost, err := addModelCosts(prefix.costs, tail); err == nil && withinBudgets(cost, r) {
					completed = append(completed, &partialPlan{end: total, costs: cost, previous: prefix, block: RowBlock{RowStart: prefix.end, RowEnd: total, DownloadAll: true}})
				}
			}
		}
		for _, choice := range choices {
			nextStates := map[uint64][]*partialPlan{}
			for _, remaining := range stateKeys(states) {
				for _, prefix := range states[remaining] {
					result.StatesExplored++
					finish(prefix)
					counts, trimmed := blockCounts(remaining, choice.capacity, options.MaxCounts)
					result.CountChoicesTrimmed = result.CountChoicesTrimmed || trimmed
					for _, count := range counts {
						next, err := appendBatch(prefix, count, total, choice)
						if err != nil || !withinBudgets(next.costs, r) {
							continue
						}
						if next.end == total {
							finish(next)
							continue
						}
						rest := total - next.end
						kept, trimmed := boundFrontier(append(nextStates[rest], next), options.MaxFrontier, r.MaxUploadBits != 0)
						nextStates[rest] = kept
						result.FrontierTrimmed = result.FrontierTrimmed || trimmed
					}
				}
			}
			var trimmed bool
			states, trimmed = boundStates(nextStates, options.MaxStates)
			result.StateLimitReached = result.StateLimitReached || trimmed
			completed, trimmed = boundFrontier(completed, options.MaxFrontier, r.MaxUploadBits != 0)
			result.FrontierTrimmed = result.FrontierTrimmed || trimmed
		}
		for _, remaining := range stateKeys(states) {
			for _, prefix := range states[remaining] {
				finish(prefix)
			}
		}
		var trimmed bool
		completed, trimmed = boundFrontier(completed, options.MaxFrontier, r.MaxUploadBits != 0)
		result.FrontierTrimmed = result.FrontierTrimmed || trimmed
		for _, last := range completed {
			mapping.LogicalPlan = partialLayout(last, total, width)
			add(mapping, false)
		}
	}
	// Approximate exploration only. Protected baselines must never pass through
	// lossy reduction: seeding them before reduction would not preserve a witness.
	nodes := make([]*partialPlan, len(layouts))
	owners := make(map[*partialPlan]Alternative, len(layouts)+len(protected))
	for i, a := range layouts {
		nodes[i] = &partialPlan{costs: a.Evaluation.Costs}
		owners[nodes[i]] = a
	}
	kept, trimmed := boundFrontier(nodes, options.MaxFrontier, false)
	result.FrontierTrimmed = result.FrontierTrimmed || trimmed
	for _, a := range protected {
		node := &partialPlan{costs: a.Evaluation.Costs}
		kept = append(kept, node)
		owners[node] = a
	}
	// Exact pruning preserves an equal-or-better witness transitively for every
	// removed baseline. This merge deliberately has no output-size approximation.
	kept, _ = boundFrontier(kept, len(kept), false)
	for _, node := range kept {
		result.Alternatives = append(result.Alternatives, owners[node])
	}
	for _, baseline := range protected {
		witness := -1
		for i, a := range result.Alternatives {
			if dominates(a.Evaluation.Costs, baseline.Evaluation.Costs, false) {
				witness = i
				break
			}
		}
		if witness < 0 {
			return nil, fmt.Errorf("planner: protected baseline lost during exact pruning")
		}
		result.BaselineWitnesses = append(result.BaselineWitnesses, BaselineWitness{
			LogicalRecordBits: baseline.Plan.LogicalRecordBits, FullDownloadOnly: baseline.Evaluation.FullDownloadOnly,
			Costs: baseline.Evaluation.Costs, AlternativeIndex: witness,
		})
	}
	if len(result.Alternatives) == 0 {
		return result, ErrNoPlan
	}
	return result, nil
}

func partialLayout(last *partialPlan, rows, width uint64) Plan {
	var reversed []RowBlock
	for node := last; node.previous != nil; node = node.previous {
		reversed = append(reversed, node.block)
	}
	plan := Plan{FormatVersion: 1, Rows: rows, RecordBits: width}
	groups := map[string]uint64{}
	for i := len(reversed) - 1; i >= 0; i-- {
		block := reversed[i]
		if !block.DownloadAll {
			group, ok := groups[block.ConfigurationID]
			if !ok {
				group = uint64(len(groups))
				groups[block.ConfigurationID] = group
			}
			block.Group = group
		}
		plan.RowBlocks = append(plan.RowBlocks, block)
	}
	return plan
}

func dominates(a, b Costs, separate bool) bool {
	if a.StorageBits > b.StorageBits {
		return false
	}
	if separate {
		return a.UploadBits <= b.UploadBits && a.DownloadBits <= b.DownloadBits
	}
	return a.DownloadBits <= b.DownloadBits
}

// Retain the exact Pareto set first. If too large, keep minimum storage and
// minimum download plus the minimum-storage representative per geometric
// download bucket. Stable ordering makes ties reproducible.
func boundFrontier(nodes []*partialPlan, limit int, separate bool) ([]*partialPlan, bool) {
	sort.SliceStable(nodes, func(i, j int) bool {
		a, b := nodes[i].costs, nodes[j].costs
		if a.StorageBits != b.StorageBits {
			return a.StorageBits < b.StorageBits
		}
		if a.DownloadBits != b.DownloadBits {
			return a.DownloadBits < b.DownloadBits
		}
		if separate && a.UploadBits != b.UploadBits {
			return a.UploadBits < b.UploadBits
		}
		return false
	})
	kept := make([]*partialPlan, 0, len(nodes))
	for _, node := range nodes {
		dominated := false
		for _, other := range kept {
			if dominates(other.costs, node.costs, separate) {
				dominated = true
				break
			}
		}
		if !dominated {
			kept = append(kept, node)
		}
	}
	if len(kept) <= limit {
		return kept, false
	}
	low := kept[0]
	for _, node := range kept {
		if node.costs.DownloadBits < low.costs.DownloadBits {
			low = node
		}
	}
	selected := map[*partialPlan]bool{kept[0]: true, low: true}
	if limit > 2 {
		minLog := math.Log1p(float64(low.costs.DownloadBits))
		maxLog := minLog
		for _, node := range kept {
			maxLog = math.Max(maxLog, math.Log1p(float64(node.costs.DownloadBits)))
		}
		buckets := map[int]*partialPlan{}
		for _, node := range kept {
			bucket := 0
			if maxLog > minLog {
				bucket = int((math.Log1p(float64(node.costs.DownloadBits)) - minLog) / (maxLog - minLog) * float64(limit-2))
				if bucket >= limit-2 {
					bucket = limit - 3
				}
			}
			if buckets[bucket] == nil {
				buckets[bucket] = node
			}
		}
		for _, node := range buckets {
			selected[node] = true
		}
	}
	reduced := make([]*partialPlan, 0, limit)
	for _, node := range kept {
		if selected[node] {
			reduced = append(reduced, node)
		}
	}
	return reduced, true
}

// RepresentativeIndices selects display options in ascending storage order.
// It always includes every baseline witness and both frontier endpoints, then
// samples additional alternatives evenly by index. target is a soft size limit:
// witness protection can require more options. Zero means 3; negative is invalid.
// Alternatives and witness metadata are not modified. Use the same witnesses if
// implementing another display strategy; arbitrary sampling can hide guarantees.
func (result *TradeoffResult) RepresentativeIndices(target int) ([]int, error) {
	if result == nil || target < 0 {
		return nil, fmt.Errorf("planner: invalid representative selection")
	}
	if target == 0 {
		target = 3
	}
	n := len(result.Alternatives)
	if n == 0 {
		return nil, nil
	}
	if target > n {
		target = n
	}
	selected := map[int]bool{0: true, n - 1: true}
	for _, w := range result.BaselineWitnesses {
		if w.AlternativeIndex < 0 || w.AlternativeIndex >= n {
			return nil, fmt.Errorf("planner: baseline witness index out of range")
		}
		selected[w.AlternativeIndex] = true
	}
	if target > 1 {
		for i := 1; i < target-1 && len(selected) < target; i++ {
			// Floating-point positioning avoids integer overflow; all sampled indices
			// are checked by construction against the actual slice length.
			index := int(float64(i) * float64(n-1) / float64(target-1))
			selected[index] = true
		}
	}
	for i := 0; len(selected) < target; i++ {
		selected[i] = true
	}
	indices := make([]int, 0, len(selected))
	for i := range selected {
		indices = append(indices, i)
	}
	sort.Ints(indices)
	return indices, nil
}
