package planner

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// ErrNoPlan means the greedy heuristic found no feasible completion. It does
// not prove that no partition meets the constraints; search is not exhaustive.
var ErrNoPlan = errors.New("greedy planner found no feasible completion")

type blockOption struct {
	id       string
	capacity uint64
	costs    Costs
}

// Greedy constructs a compact row-block partition and evaluates it. Each step
// scores every eligible configuration plus a largest-fitting-capacity greedy
// completion. It chooses minimum total storage, breaking ties by total traffic,
// then greater next-block height and configuration ID for determinism.
//
// Tail costs batch repeated blocks without constructing rectangles. Estimated
// capacities are eligible; unknown capacities and insufficient security profiles
// are excluded. No encoding-time or memory measurement affects selection. Plans
// may be provisional, and neither row-aligned nor global optimality is promised.
// The result uses Plan.RowBlocks, not potentially millions of explicit databases.
func Greedy(rows, recordBits uint64, catalogue *pirconfig.Catalogue, requirements Requirements) (Plan, *Evaluation, error) {
	if err := catalogue.Validate(); err != nil {
		return Plan{}, nil, fmt.Errorf("planner: %w", err)
	}
	if rows == 0 || recordBits == 0 {
		return Plan{}, nil, fmt.Errorf("planner: application dimensions must be nonempty")
	}
	if _, err := multiply(rows, recordBits); err != nil {
		return Plan{}, nil, fmt.Errorf("planner: application bits: %w", err)
	}
	options := []blockOption{}
	for i := range catalogue.Configurations {
		c := &catalogue.Configurations[i]
		if recordBits < uint64(c.Code.S) {
			return Plan{}, nil, fmt.Errorf("planner: record_bits must be at least catalogue field width %d", c.Code.S)
		}
		if c.Capacity.Status == "unknown" || c.Security.ComputationalBits < requirements.MinSecurityBits || c.Security.Log2Samples < requirements.Log2Samples {
			continue
		}
		count := recordBits / uint64(c.Code.S)
		if recordBits%uint64(c.Code.S) != 0 {
			count++
		}
		costs, err := databaseCosts(c, count)
		if err != nil {
			continue
		} // An overflowing option need not exclude usable alternatives.
		options = append(options, blockOption{id: c.ID, capacity: *c.Capacity.K, costs: costs})
	}
	if len(options) == 0 {
		return Plan{}, nil, ErrNoPlan
	}
	// The baseline repeatedly fills the largest available capacity. Ties use the
	// cheaper block and then ID; input catalogue order never controls the result.
	sort.Slice(options, func(i, j int) bool {
		a, b := options[i], options[j]
		if a.capacity != b.capacity {
			return a.capacity > b.capacity
		}
		if cheaper(a.costs, b.costs) {
			return true
		}
		if cheaper(b.costs, a.costs) {
			return false
		}
		return a.id < b.id
	})
	plan := Plan{FormatVersion: 1, Rows: rows, RecordBits: recordBits}
	groups := map[string]uint64{}
	var prefix Costs
	remaining := rows
	for remaining > 0 {
		best := -1
		var bestScore Costs
		var bestHeight uint64
		for i, c := range options {
			height := c.capacity
			if height > remaining {
				height = remaining
			}
			tail, err := greedyTail(remaining-height, options)
			if err != nil {
				continue
			}
			score, err := addModelCosts(prefix, c.costs)
			if err == nil {
				score, err = addModelCosts(score, tail)
			}
			if err != nil || !withinBudgets(score, requirements) {
				continue
			}
			if best < 0 || cheaper(score, bestScore) || (!cheaper(bestScore, score) && (height > bestHeight || (height == bestHeight && c.id < options[best].id))) {
				best, bestScore, bestHeight = i, score, height
			}
		}
		if best < 0 {
			return Plan{}, nil, ErrNoPlan
		}
		chosen := options[best]
		group, exists := groups[chosen.id]
		if !exists {
			group = uint64(len(groups))
			groups[chosen.id] = group
		}
		start := rows - remaining
		plan.RowBlocks = append(plan.RowBlocks, RowBlock{RowStart: start, RowEnd: start + bestHeight, ConfigurationID: chosen.id, Group: group})
		var err error
		prefix, err = addModelCosts(prefix, chosen.costs)
		if err != nil {
			return Plan{}, nil, fmt.Errorf("planner: %w", err)
		}
		remaining -= bestHeight
	}
	evaluation, err := Evaluate(plan, catalogue, requirements)
	if err != nil {
		return Plan{}, nil, err
	}
	// Independently normalized evaluation must agree with the search accounting.
	if evaluation.Costs.StorageBits != prefix.StorageBits || evaluation.Costs.UploadBits != prefix.UploadBits || evaluation.Costs.DownloadBits != prefix.DownloadBits {
		return Plan{}, nil, fmt.Errorf("planner: greedy and evaluated costs disagree")
	}
	return plan, evaluation, nil
}

// options are sorted by descending capacity. Once a capacity is processed, the
// remainder is smaller, so each distinct size is visited at most once. The last
// partial block uses the cheapest covering configuration, not an arbitrary size.
func greedyTail(rows uint64, options []blockOption) (Costs, error) {
	var result Costs
	for _, c := range options {
		count := rows / c.capacity
		if count == 0 {
			continue
		}
		repeated, err := scaleModelCosts(c.costs, count)
		if err == nil {
			result, err = addModelCosts(result, repeated)
		}
		if err != nil {
			return Costs{}, err
		}
		rows %= c.capacity
	}
	if rows > 0 {
		best := -1
		for i, c := range options {
			if c.capacity >= rows && (best < 0 || cheaper(c.costs, options[best].costs)) {
				best = i
			}
		}
		if best < 0 {
			return Costs{}, ErrNoPlan
		}
		var err error
		result, err = addModelCosts(result, options[best].costs)
		if err != nil {
			return Costs{}, err
		}
	}
	return result, nil
}
