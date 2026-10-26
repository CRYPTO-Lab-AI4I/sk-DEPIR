package planner

import (
	"errors"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// High synthetic capacities and zero dummies produce distinct storage/traffic
// choices on tiny grids. They test optimization only, not real code dimensions.
func tradeoffCatalogue(t *testing.T) *pirconfig.Catalogue {
	c := searchCatalogue(t, 14, 40)
	for i := range c.Configurations {
		c.Configurations[i].Query.DummyPositions = 0
		c.Configurations[i].Query.Length = c.Configurations[i].Query.CurvePositions
	}
	return c
}

type costPair struct{ storage, communication uint64 }

// Exhaustive oracle builds every allowed block sequence and evaluates full plans.
// It uses neither partial-plan pruning nor the search's cost accumulation.
func exhaustiveTradeoffs(t *testing.T, rows, bits uint64, c *pirconfig.Catalogue, r Requirements) map[costPair]bool {
	t.Helper()
	var values []costPair
	for _, width := range []uint64{3, bits} {
		fragments := (bits + width - 1) / width
		total := rows * fragments
		var visit func(uint64, []RowBlock)
		visit = func(start uint64, blocks []RowBlock) {
			if start < total {
				visit(total, append(append([]RowBlock(nil), blocks...), RowBlock{RowStart: start, RowEnd: total, DownloadAll: true}))
				for _, entry := range c.Configurations {
					end := start + *entry.Capacity.K
					if end > total {
						end = total
					}
					visit(end, append(append([]RowBlock(nil), blocks...), RowBlock{RowStart: start, RowEnd: end, ConfigurationID: entry.ID}))
				}
				return
			}
			groups := map[string]uint64{}
			for i := range blocks {
				b := &blocks[i]
				if b.DownloadAll {
					continue
				}
				g, ok := groups[b.ConfigurationID]
				if !ok {
					g = uint64(len(groups))
					groups[b.ConfigurationID] = g
				}
				b.Group = g
			}
			e, err := EvaluateApplication(ApplicationPlan{Rows: rows, RecordBits: bits, LogicalRecordBits: width, LogicalPlan: Plan{FormatVersion: 1, Rows: total, RecordBits: width, RowBlocks: blocks}}, c, r)
			if err == nil {
				values = append(values, costPair{e.Costs.StorageBits, e.Costs.CommunicationBits})
			}
		}
		visit(0, nil)
	}
	result := map[costPair]bool{}
	for _, a := range values {
		dominated := false
		for _, b := range values {
			if b.storage <= a.storage && b.communication <= a.communication && b != a {
				dominated = true
				break
			}
		}
		if !dominated {
			result[a] = true
		}
	}
	return result
}

func TestTradeoffsMatchExhaustiveRestrictedSearch(t *testing.T) {
	c := tradeoffCatalogue(t)
	for _, rows := range []uint64{1, 13, 14, 15, 27, 39, 40, 41} {
		for _, r := range []Requirements{{}, {MaxUploadBits: 200}, {MaxDownloadBits: 100}, {MaxStorageBits: 240}} {
			got, err := SearchTradeoffs(rows, 6, c, r, TradeoffOptions{LogicalRecordBits: []uint64{3}, MaxFrontier: 1000, MaxStates: 1000})
			want := exhaustiveTradeoffs(t, rows, 6, c, r)
			if len(want) == 0 {
				if !errors.Is(err, ErrNoPlan) {
					t.Fatalf("expected no plan: %v", err)
				}
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.FrontierTrimmed || got.StateLimitReached || got.CountChoicesTrimmed || got.BaselineLimitReached {
				t.Fatal("unexpected approximation")
			}
			pairs := map[costPair]bool{}
			for _, a := range got.Alternatives {
				pairs[costPair{a.Evaluation.Costs.StorageBits, a.Evaluation.Costs.CommunicationBits}] = true
				e, err := EvaluateApplication(a.Plan, c, r)
				if err != nil || !reflect.DeepEqual(e, a.Evaluation) {
					t.Fatal("result failed independent evaluation", err)
				}
				if e.CommunicationRate != float64(6)/float64(e.Costs.CommunicationBits) {
					t.Fatal("inverted communication rate")
				}
			}
			if !reflect.DeepEqual(pairs, want) {
				t.Fatalf("rows=%d req=%+v got=%v want=%v", rows, r, pairs, want)
			}
		}
	}
}

func TestTradeoffBoundsAndDeterminism(t *testing.T) {
	c := tradeoffCatalogue(t)
	options := TradeoffOptions{LogicalRecordBits: []uint64{3}, MaxFrontier: 2, MaxStates: 1}
	got, err := SearchTradeoffs(40, 6, c, Requirements{}, options)
	if err != nil {
		t.Fatal(err)
	}
	if !got.StateLimitReached || len(got.Alternatives) > 2+len(got.BaselineWitnesses) || got.StatesExplored > 4 {
		t.Fatal("search limits not enforced", got)
	}
	c.Configurations[0], c.Configurations[1] = c.Configurations[1], c.Configurations[0]
	again, err := SearchTradeoffs(40, 6, c, Requirements{}, options)
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatal("catalogue order changed result")
	}
	// All points on this synthetic curve are nondominated. Reduction must retain
	// both endpoints and at most the requested count, deterministically.
	var nodes []*partialPlan
	for i := uint64(1); i <= 100; i++ {
		nodes = append(nodes, &partialPlan{costs: Costs{StorageBits: i, CommunicationBits: 101 - i}})
	}
	bounded, trimmed := boundFrontier(nodes, 8, false)
	if !trimmed || len(bounded) > 8 || bounded[0].costs.StorageBits != 1 || bounded[len(bounded)-1].costs.CommunicationBits != 1 {
		t.Fatal("endpoint loss")
	}
	if other, _ := boundFrontier(nodes, 8, false); !reflect.DeepEqual(bounded, other) {
		t.Fatal("bucket choice changed")
	}
}

func TestTradeoffValidationAndFullBaseline(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	got, err := SearchTradeoffs(3, 4, c, Requirements{}, TradeoffOptions{})
	if err != nil || len(got.Alternatives) != 1 || !got.Alternatives[0].Evaluation.FullDownloadOnly {
		t.Fatal("dominated encoded plan survived", err)
	}
	if got.Alternatives[0].Evaluation.CommunicationRate != 1.0/3 {
		t.Fatal("wrong full-download rate")
	}
	for _, o := range []TradeoffOptions{{MaxFrontier: 1}, {MaxFrontier: -1}, {MaxStates: -1}, {MaxCounts: 3}, {MaxBaselineBlocks: -1}, {LogicalRecordBits: []uint64{0}}} {
		if _, err := SearchTradeoffs(3, 4, c, Requirements{}, o); err == nil {
			t.Fatal("invalid options accepted")
		}
	}
	if _, err := SearchTradeoffs(math.MaxUint64, 4, c, Requirements{}, TradeoffOptions{}); err == nil {
		t.Fatal("overflow accepted")
	}
	if _, err := SearchTradeoffs(3, 4, nil, Requirements{}, TradeoffOptions{}); err == nil {
		t.Fatal("nil catalogue accepted")
	}
	// No eligible indexed profile and full read exceeds the download budget.
	got, err = SearchTradeoffs(3, 4, c, Requirements{MinSecurityBits: 100, MaxDownloadBits: 1}, TradeoffOptions{})
	if !errors.Is(err, ErrNoPlan) || got == nil {
		t.Fatal("missing infeasibility/search metadata")
	}
}

func TestTradeoffsTargetCoordinates(t *testing.T) {
	f, err := os.Open("../pirconfig/configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	result, err := SearchTradeoffs(100000, 80000000, c, Requirements{MinSecurityBits: 100, Log2Samples: 50}, TradeoffOptions{MaxFrontier: 8, MaxStates: 16})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Alternatives) > 8+len(result.BaselineWitnesses) {
		t.Fatal("output bound exceeded")
	}
	for _, a := range result.Alternatives {
		costs := a.Evaluation.Costs
		if costs.StorageBits >= 8000000000000 && costs.CommunicationBits >= 8000000000000 && !a.Evaluation.FullDownloadOnly {
			t.Fatal("dominated target plan survived")
		}
	}
}

func TestTradeoffBaselineWorkBound(t *testing.T) {
	// A huge row count and tiny capacity must not build millions of greedy blocks.
	c := fixtureCatalogue(t, 2)
	result, err := SearchTradeoffs(1000000000, 4, c, Requirements{}, TradeoffOptions{MaxStates: 1, MaxBaselineBlocks: 1, MaxFrontier: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !result.BaselineLimitReached || !result.StateLimitReached {
		t.Fatal("baseline truncation not reported")
	}
}

func TestTradeoffApproximationKeepsTinyEndpoints(t *testing.T) {
	c := tradeoffCatalogue(t)
	exact, err := SearchTradeoffs(40, 6, c, Requirements{}, TradeoffOptions{LogicalRecordBits: []uint64{3}, MaxFrontier: 1000, MaxStates: 1000})
	if err != nil {
		t.Fatal(err)
	}
	bounded, err := SearchTradeoffs(40, 6, c, Requirements{}, TradeoffOptions{LogicalRecordBits: []uint64{3}, MaxFrontier: 2, MaxStates: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if !bounded.FrontierTrimmed || len(bounded.Alternatives) > 2+len(bounded.BaselineWitnesses) {
		t.Fatal("approximation did not run")
	}
	if bounded.Alternatives[0].Evaluation.Costs.StorageBits != exact.Alternatives[0].Evaluation.Costs.StorageBits || bounded.Alternatives[len(bounded.Alternatives)-1].Evaluation.Costs.CommunicationBits != exact.Alternatives[len(exact.Alternatives)-1].Evaluation.Costs.CommunicationBits {
		t.Fatal("tiny endpoint comparison failed")
	}
}

func TestTradeoffProtectedBaselinesAndDisplay(t *testing.T) {
	c := tradeoffCatalogue(t)
	result, err := SearchTradeoffs(40, 6, c, Requirements{}, TradeoffOptions{LogicalRecordBits: []uint64{3}, MaxFrontier: 2, MaxStates: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.BaselineWitnesses) != 3 {
		t.Fatal("missing full or per-width baseline")
	}
	if len(result.Alternatives) <= 2 {
		t.Fatal("test did not exercise output beyond exploratory cap")
	}
	indices, err := result.RepresentativeIndices(2)
	if err != nil {
		t.Fatal(err)
	}
	displayed := map[int]bool{}
	for _, i := range indices {
		displayed[i] = true
	}
	for _, w := range result.BaselineWitnesses {
		if !dominates(result.Alternatives[w.AlternativeIndex].Evaluation.Costs, w.Costs, false) {
			t.Fatal("invalid baseline witness")
		}
		if !displayed[w.AlternativeIndex] {
			t.Fatal("display selection lost baseline witness")
		}
	}
	again, err := result.RepresentativeIndices(2)
	if err != nil || !reflect.DeepEqual(indices, again) {
		t.Fatal("display not deterministic")
	}
	if _, err := result.RepresentativeIndices(-1); err == nil {
		t.Fatal("negative display target accepted")
	}
	bad := *result
	bad.BaselineWitnesses = []BaselineWitness{{AlternativeIndex: len(result.Alternatives)}}
	if _, err := bad.RepresentativeIndices(2); err == nil {
		t.Fatal("invalid witness accepted")
	}
}

func TestTradeoffsDoNotRegressOriginalGreedy(t *testing.T) {
	f, err := os.Open("../pirconfig/configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	r := Requirements{MinSecurityBits: 100, Log2Samples: 50}
	// Original 1 GB / 1 KB row lost its greedy witness under old lossy merging.
	_, baseline, err := Greedy(1000000, 8000, c, r)
	if err != nil {
		t.Fatal(err)
	}
	result, err := SearchTradeoffs(1000000, 8000, c, r, TradeoffOptions{MaxFrontier: 2, MaxStates: 256})
	if err != nil {
		t.Fatal(err)
	}
	indices, err := result.RepresentativeIndices(3)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range indices {
		if dominates(result.Alternatives[i].Evaluation.Costs, baseline.Costs, false) {
			found = true
		}
	}
	if !found {
		t.Fatal("original greedy witness lost from results/display")
	}
}
