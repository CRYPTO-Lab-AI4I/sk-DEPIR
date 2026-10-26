package planner

import (
	"errors"
	"fmt"
	"sort"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// ErrNoPlan means the selected search found no feasible completion. It does
// not prove that no partition meets the constraints; search is not exhaustive.
var ErrNoPlan = errors.New("planner search found no feasible completion")

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
	return greedy(rows, recordBits, catalogue, requirements, true, 1, 0, 1)
}

// fullLimit=0 disables a full-download suffix; otherwise it caps its total bits.
func greedy(rows, recordBits uint64, catalogue *pirconfig.Catalogue, requirements Requirements, enforceWidth bool, queries, fullLimit uint64, fullRatio float64) (Plan, *Evaluation, error) {
	return greedyWithBlockLimit(rows, recordBits, catalogue, requirements, enforceWidth, queries, fullLimit, fullRatio, 0)
}

// errBaselineLimit reports bounded baseline construction, not infeasibility.
var errBaselineLimit = errors.New("planner: baseline block limit reached")

func greedyWithBlockLimit(rows, recordBits uint64, catalogue *pirconfig.Catalogue, requirements Requirements, enforceWidth bool, queries, fullLimit uint64, fullRatio float64, maxBlocks int) (Plan, *Evaluation, error) {
	if err := catalogue.Validate(); err != nil {
		return Plan{}, nil, fmt.Errorf("planner: %w", err)
	}
	if rows == 0 || recordBits == 0 {
		return Plan{}, nil, fmt.Errorf("planner: application dimensions must be nonempty")
	}
	if _, err := multiply(rows, recordBits); err != nil {
		return Plan{}, nil, fmt.Errorf("planner: application bits: %w", err)
	}
	maxWidth := 0
	for _, c := range catalogue.Configurations {
		if c.Code.S > maxWidth {
			maxWidth = c.Code.S
		}
	}
	tooNarrow := enforceWidth && recordBits < uint64(maxWidth)
	if tooNarrow && fullLimit == 0 {
		return Plan{}, nil, fmt.Errorf("planner: record_bits must be at least catalogue field width %d", maxWidth)
	}
	options := []blockOption{}
	for i := range catalogue.Configurations {
		c := &catalogue.Configurations[i]
		if c.Capacity.Status == "unknown" || c.Security.ComputationalBits < requirements.MinSecurityBits || c.Security.Log2Samples < requirements.Log2Samples {
			continue
		}
		count := recordBits / uint64(c.Code.S)
		if recordBits%uint64(c.Code.S) != 0 {
			count++
		}
		costs, err := databaseCosts(c, count)
		if err == nil {
			costs, err = repeatTraffic(costs, queries)
		}
		if err != nil {
			continue
		} // An overflowing option need not exclude usable alternatives.
		options = append(options, blockOption{id: c.ID, capacity: *c.Capacity.K, costs: costs})
	}
	if len(options) == 0 && fullLimit == 0 {
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
	// A pure full download is compared with the indexed-only baseline. Mixed
	// plans compare the suffix with the encoded prefix that is actually queried.
	var baselineDownload uint64
	if fullLimit != 0 {
		if baseline, err := greedyTail(rows, options); err == nil {
			baselineDownload = baseline.DownloadBits
		}
	}
	suffixLimit := func(prefix Costs) uint64 {
		reference := prefix.DownloadBits
		if reference == 0 {
			reference = baselineDownload
		}
		cap := float64(reference) * fullRatio
		limit := fullLimit
		if cap < float64(limit) {
			limit = uint64(cap)
		}
		return limit
	}
	plan := Plan{FormatVersion: 1, Rows: rows, RecordBits: recordBits}
	groups := map[string]uint64{}
	var prefix Costs
	remaining := rows
	for remaining > 0 {
		if maxBlocks > 0 && len(plan.RowBlocks) >= maxBlocks {
			return Plan{}, nil, errBaselineLimit
		}
		best := -1
		var bestScore Costs
		var bestHeight uint64
		bestID := ""
		if suffix, ok := fullDownloadCosts(remaining, recordBits, suffixLimit(prefix)); ok {
			score, err := addModelCosts(prefix, suffix)
			if err == nil && withinBudgets(score, requirements) {
				best = -2
				bestScore = score
				bestHeight = remaining
			}
		}
		for i, c := range options {
			if tooNarrow {
				continue
			}
			height := c.capacity
			if height > remaining {
				height = remaining
			}
			base, err := addModelCosts(prefix, c.costs)
			if err != nil {
				continue
			}
			score, ok := completionScore(base, remaining-height, recordBits, options, requirements, suffixLimit(base))
			if !ok {
				continue
			}

			if best == -1 || cheaper(score, bestScore) || (!cheaper(bestScore, score) && (height > bestHeight || (height == bestHeight && c.id < bestID))) {
				best, bestScore, bestHeight, bestID = i, score, height, c.id
			}
		}
		if best == -1 {
			return Plan{}, nil, ErrNoPlan
		}
		if best == -2 {
			plan.RowBlocks = append(plan.RowBlocks, RowBlock{DownloadAll: true, RowStart: rows - remaining, RowEnd: rows})
			suffix, _ := fullDownloadCosts(remaining, recordBits, suffixLimit(prefix))
			var err error
			prefix, err = addModelCosts(prefix, suffix)
			if err != nil {
				return Plan{}, nil, err
			}
			remaining = 0
			break
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
	evaluation, err := evaluate(plan, catalogue, requirements, enforceWidth, queries)
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

// Full pieces use no index-dependent upload: fixed range descriptors and database
// identifiers are transport overhead, excluded as in the ordinary logical model.
func fullDownloadCosts(rows, width, limit uint64) (Costs, bool) {
	if rows == 0 {
		return Costs{}, true
	}
	if limit == 0 {
		return Costs{}, false
	}
	size, err := multiply(rows, width)
	if err != nil || size > limit {
		return Costs{}, false
	}
	return Costs{StorageBits: size, DownloadBits: size, CommunicationBits: size}, true
}

// Compare complete continuations under the actual per-application budgets. A
// cheaper full-download tail must not hide a feasible indexed continuation when
// its download exceeds the budget.
func completionScore(prefix Costs, rows, width uint64, options []blockOption, r Requirements, fullLimit uint64) (Costs, bool) {
	var best Costs
	found := false
	if tail, err := greedyTail(rows, options); err == nil {
		if total, err := addModelCosts(prefix, tail); err == nil && withinBudgets(total, r) {
			best = total
			found = true
		}
	}
	if tail, ok := fullDownloadCosts(rows, width, fullLimit); ok {
		if total, err := addModelCosts(prefix, tail); err == nil && withinBudgets(total, r) && (!found || cheaper(total, best)) {
			best = total
			found = true
		}
	}
	return best, found
}
