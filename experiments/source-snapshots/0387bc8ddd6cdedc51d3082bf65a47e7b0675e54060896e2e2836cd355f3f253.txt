package planner

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// DPCostOptions controls a cost-indexed DP independently of SearchTradeoffs.
// Zero selects StorageCells=64, DownloadCells=64, MaxFrontier=32 and
// MaxBaselineBlocks=256. Each axis must have 1..512 cells. LogicalRecordBits
// follows ApplicationOptions. Window bits bound exploration only, not protected and one-code seeds: zero uses twice the storage of the least-download one-code
// or greedy seed, and geometrically spaced download windows from that seed
// through application size. Explicit window bits select a single window.
// Exact costs and Requirements.MaxStorageBits are always enforced. Upload budgets require a third constraint dimension and remain unsupported;
// storage and download budgets are enforced in the existing two dimensions.
type DPCostOptions struct {
	LogicalRecordBits  []uint64 `json:"logical_record_bits,omitempty"`
	StorageCells       int      `json:"storage_cells,omitempty"`
	DownloadCells      int      `json:"download_cells,omitempty"`
	StorageWindowBits  uint64   `json:"storage_window_bits,omitempty"`
	DownloadWindowBits uint64   `json:"download_window_bits,omitempty"`
	MaxFrontier        int      `json:"max_frontier,omitempty"`
	MaxBaselineBlocks  int      `json:"max_baseline_blocks,omitempty"`
}

// DPCostWindow records the numerical search window and rounding for one width.
// A path's full-block costs lose less than BinaryItems*UnitBits per axis through
// flooring; this is an accounting bound, not a global Pareto error guarantee.
// Full/padded final completions are costed exactly, without coordinate rounding.
// Transitions counts full-block grid attempts before exact budget checks; it
// excludes terminal padding/download attempts and independent baseline work.
type DPCostWindow struct {
	LogicalRecordBits uint64 `json:"logical_record_bits"`
	StorageBits       uint64 `json:"storage_bits"`
	DownloadBits      uint64 `json:"download_bits"`
	StorageUnitBits   uint64 `json:"storage_unit_bits"`
	DownloadUnitBits  uint64 `json:"download_unit_bits"`
	BinaryItems       int    `json:"binary_items"`
	Transitions       uint64 `json:"transitions"`
	Rounded           bool   `json:"rounded"`
}

// DPCostResult reuses protected baseline witnesses and display selection. State
// and count sampling flags are unused: the grid itself bounds DP state and all
// block counts are represented through exact binary multiplicity decomposition.
// FrontierTrimmed can report approximate completed-output selection. Windows and
// floor rounding limit search quality; neither global optimality nor an FPTAS is claimed.
type DPCostResult struct {
	TradeoffResult
	Windows []DPCostWindow `json:"windows"`
}

type dpItem struct {
	choice blockOption
	count  uint64
	cost   Costs
}

// SearchDPCosts maximizes covered logical rows at each rounded storage/download
// cell, after each 0/1 binary item. There is no sampling of counts or row positions.
// Items partition each configuration's full-block count bound into groups whose
// subset sums represent every count from zero through floor(rows/K).
//
// For equal cell coordinates and item stage retain maximum coverage; ties use
// smaller exact storage then download. With unit=1 labels are exact within
// the search windows and layout family; seeds may contribute outside those windows. With flooring, costs may differ inside a cell;
// retained plans and all reported costs are independently validated in exact bits.
// An additional upload budget requires another DP dimension and is rejected.
func SearchDPCosts(rows, recordBits uint64, catalogue *pirconfig.Catalogue, r Requirements, o DPCostOptions) (*DPCostResult, error) {
	if err := catalogue.Validate(); err != nil {
		return nil, err
	}
	if rows == 0 || recordBits == 0 {
		return nil, fmt.Errorf("planner: nonempty application dimensions required")
	}
	appBits, err := multiply(rows, recordBits)
	if err != nil {
		return nil, err
	}
	if r.MaxUploadBits != 0 {
		return nil, fmt.Errorf("planner: cost DP does not support an upload budget")
	}
	if o.StorageCells == 0 {
		o.StorageCells = 64
	}
	if o.DownloadCells == 0 {
		o.DownloadCells = 64
	}
	if o.MaxFrontier == 0 {
		o.MaxFrontier = 32
	}
	if o.MaxBaselineBlocks == 0 {
		o.MaxBaselineBlocks = 256
	}
	if o.StorageCells < 1 || o.StorageCells > 512 || o.DownloadCells < 1 || o.DownloadCells > 512 || o.MaxFrontier < 2 || o.MaxBaselineBlocks < 1 {
		return nil, fmt.Errorf("planner: invalid cost-grid limits")
	}
	widths, err := logicalWidths(recordBits, catalogue, o.LogicalRecordBits)
	if err != nil {
		return nil, err
	}
	result := &DPCostResult{}
	var protected, explored []Alternative
	add := func(p ApplicationPlan, baseline bool) {
		e, err := EvaluateApplication(p, catalogue, r)
		if err != nil {
			return
		}
		a := Alternative{Plan: p, Evaluation: e}
		if baseline {
			protected = append(protected, a)
		} else {
			explored = append(explored, a)
		}
	}
	add(ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: recordBits, LogicalPlan: Plan{FormatVersion: 1, Rows: rows, RecordBits: recordBits, RowBlocks: []RowBlock{{RowEnd: rows, DownloadAll: true}}}}, true)
	indexed := true
	for _, c := range catalogue.Configurations {
		if recordBits < uint64(c.Code.S) {
			indexed = false
		}
	}
	for _, width := range widths {
		if !indexed {
			continue
		}
		mapping := ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: width}
		f, _ := mapping.FragmentCount()
		total, err := multiply(rows, f)
		if err != nil {
			continue
		}
		var seed *partialPlan
		if p, e, err := greedyWithBlockLimit(total, width, catalogue, r, false, f, 0, 1, o.MaxBaselineBlocks); err == nil {
			mapping.LogicalPlan = p
			add(mapping, true)
			seed = &partialPlan{costs: e.Costs}
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
				cost, err = repeatTraffic(cost, f)
			}
			if err != nil {
				continue
			}
			choice := blockOption{id: c.ID, capacity: *c.Capacity.K, costs: cost}
			choices = append(choices, choice)
		}
		sort.Slice(choices, func(i, j int) bool { return choices[i].id < choices[j].id })
		for _, c := range choices {
			count := total / c.capacity
			if total%c.capacity != 0 {
				count++
			}
			p, err := appendBatch(&partialPlan{}, count, total, c)
			if err != nil {
				continue
			}
			mapping.LogicalPlan = coalesceRuns(partialLayout(p, total, width))
			add(mapping, false)
			if seed == nil || p.costs.DownloadBits < seed.costs.DownloadBits || (p.costs.DownloadBits == seed.costs.DownloadBits && p.costs.StorageBits < seed.costs.StorageBits) {
				seed = p
			}
		}
		if seed == nil {
			continue
		}
		sLimit := o.StorageWindowBits
		if sLimit == 0 {
			sLimit, err = multiply(seed.costs.StorageBits, 2)
			if err != nil {
				sLimit = math.MaxUint64
			}
		}
		for _, cLimit := range dpDownloadWindows(seed.costs.DownloadBits, appBits, o.DownloadWindowBits) {
			if r.MaxStorageBits != 0 && r.MaxStorageBits < sLimit {
				sLimit = r.MaxStorageBits
			}
			sUnit := ceilDivide(sLimit, uint64(o.StorageCells))
			cUnit := ceilDivide(cLimit, uint64(o.DownloadCells))
			sCells := int(sLimit / sUnit)
			cCells := int(cLimit / cUnit)
			stride := cCells + 1
			grid := make([]*partialPlan, (sCells+1)*stride)
			grid[0] = &partialPlan{}
			var items []dpItem
			for _, c := range choices {
				remaining := total / c.capacity
				for power := uint64(1); remaining > 0; {
					count := minWidth(power, remaining)
					cost, err := scaleModelCosts(c.costs, count)
					if err == nil {
						items = append(items, dpItem{c, count, cost})
					}
					remaining -= count
					if power > math.MaxUint64/2 {
						power = remaining
					} else {
						power *= 2
					}
				}
			}
			window := DPCostWindow{LogicalRecordBits: width, StorageBits: sLimit, DownloadBits: cLimit, StorageUnitBits: sUnit, DownloadUnitBits: cUnit, BinaryItems: len(items), Rounded: sUnit > 1 || cUnit > 1}
			var completed []*partialPlan
			keep := func(p *partialPlan) {
				if p.costs.StorageBits > sLimit || p.costs.DownloadBits > cLimit || !withinBudgets(p.costs, r) {
					return
				}
				completed = append(completed, p)
				if len(completed) > 2*o.MaxFrontier {
					var trimmed bool
					completed, trimmed = boundFrontier(completed, o.MaxFrontier, false)
					result.FrontierTrimmed = result.FrontierTrimmed || trimmed
				}
			}
			for _, item := range items {
				ds, dc := item.cost.StorageBits/sUnit, item.cost.DownloadBits/cUnit
				if ds > uint64(sCells) || dc > uint64(cCells) {
					continue
				}
				// Descending indices ensure every binary item is used at most once, including
				// zero-coordinate items. Nodes are immutable even when a grid slot changes.
				for index := len(grid) - 1; index >= 0; index-- {
					prefix := grid[index]
					if prefix == nil || prefix.end == total {
						continue
					}
					s, c := index/stride, index%stride
					if s+int(ds) > sCells || c+int(dc) > cCells {
						continue
					}
					if item.count > (total-prefix.end)/item.choice.capacity {
						// An overshooting binary group can finish with fewer blocks and one
						// padded block. Dropping it would invalidate maximum-coverage dominance.
						count := ceilDivide(total-prefix.end, item.choice.capacity)
						if p, err := appendBatch(prefix, count, total, item.choice); err == nil {
							keep(p)
						}
						continue
					}
					window.Transitions++
					costs, err := addModelCosts(prefix.costs, item.cost)
					if err != nil || costs.StorageBits > sLimit || costs.DownloadBits > cLimit || !withinBudgets(costs, r) {
						continue
					}
					target := (s+int(ds))*stride + c + int(dc)
					end := prefix.end + item.count*item.choice.capacity
					old := grid[target]
					if old != nil && (old.end > end || (old.end == end && !cheaper(costs, old.costs))) {
						continue
					}
					grid[target] = &partialPlan{end: end, costs: costs, previous: prefix, block: RowBlock{RowStart: prefix.end, RowEnd: end, ConfigurationID: item.choice.id, Repeat: item.count}}
				}
			}
			for _, prefix := range grid {
				if prefix == nil {
					continue
				}
				result.StatesExplored++
				if prefix.end == total {
					keep(prefix)
					continue
				}
				remaining := total - prefix.end
				if tail, ok := fullDownloadCosts(remaining, width, math.MaxUint64); ok {
					if cost, err := addModelCosts(prefix.costs, tail); err == nil {
						keep(&partialPlan{end: total, costs: cost, previous: prefix, block: RowBlock{RowStart: prefix.end, RowEnd: total, DownloadAll: true}})
					}
				}
				for _, c := range choices {
					if remaining <= c.capacity {
						if p, err := appendBatch(prefix, 1, total, c); err == nil {
							keep(p)
						}
					}
				}
			}
			completed, trimmed := boundFrontier(completed, o.MaxFrontier, false)
			result.FrontierTrimmed = result.FrontierTrimmed || trimmed
			for _, p := range completed {
				mapping.LogicalPlan = coalesceRuns(partialLayout(p, total, width))
				add(mapping, false)
			}
			result.Windows = append(result.Windows, window)
		}
	}
	// Protect baselines after all cost rounding and output approximation.
	nodes := make([]*partialPlan, len(explored))
	owners := map[*partialPlan]Alternative{}
	for i, a := range explored {
		nodes[i] = &partialPlan{costs: a.Evaluation.Costs}
		owners[nodes[i]] = a
	}
	kept, trimmed := boundFrontier(nodes, o.MaxFrontier, false)
	result.FrontierTrimmed = result.FrontierTrimmed || trimmed
	for _, a := range protected {
		node := &partialPlan{costs: a.Evaluation.Costs}
		kept = append(kept, node)
		owners[node] = a
	}
	kept, _ = boundFrontier(kept, len(kept), false)
	for _, node := range kept {
		result.Alternatives = append(result.Alternatives, owners[node])
	}
	for _, a := range protected {
		index := -1
		for i, b := range result.Alternatives {
			if dominates(b.Evaluation.Costs, a.Evaluation.Costs, false) {
				index = i
				break
			}
		}
		if index < 0 {
			return nil, fmt.Errorf("planner: cost DP lost baseline witness")
		}
		result.BaselineWitnesses = append(result.BaselineWitnesses, BaselineWitness{LogicalRecordBits: a.Plan.LogicalRecordBits, FullDownloadOnly: a.Evaluation.FullDownloadOnly, Costs: a.Evaluation.Costs, AlternativeIndex: index})
	}
	if len(result.Alternatives) == 0 {
		return result, ErrNoPlan
	}
	return result, nil
}

func ceilDivide(a, b uint64) uint64 {
	n := a / b
	if a%b != 0 {
		n++
	}
	return n
}

// Coalesce adjacent full runs of one code; shorter padded blocks stay separate.
func coalesceRuns(p Plan) Plan {
	var blocks []RowBlock
	for _, b := range p.RowBlocks {
		if len(blocks) > 0 && !b.DownloadAll {
			last := &blocks[len(blocks)-1]
			br := b.Repeat
			if br == 0 {
				br = 1
			}
			lr := last.Repeat
			if lr == 0 {
				lr = 1
			}
			if !last.DownloadAll && last.ConfigurationID == b.ConfigurationID && last.RowEnd == b.RowStart && (last.RowEnd-last.RowStart)/lr == (b.RowEnd-b.RowStart)/br {
				last.RowEnd = b.RowEnd
				last.Repeat = lr + br
				continue
			}
		}
		blocks = append(blocks, b)
	}
	p.RowBlocks = blocks
	return p
}

// Geometrically spaced download windows retain useful low-traffic resolution
// instead of using database size as the sole rounding scale. Include both ends.
func dpDownloadWindows(seed, application, explicit uint64) []uint64 {
	if explicit != 0 {
		return []uint64{explicit}
	}
	limit := minWidth(seed, application)
	windows := []uint64{limit}
	for limit < application {
		if limit > application/4 {
			limit = application
		} else {
			limit *= 4
		}
		windows = append(windows, limit)
	}
	return windows
}
