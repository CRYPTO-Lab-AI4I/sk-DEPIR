package planner

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// Synthetic capacities isolate search choices; they are not measured code dimensions.
func searchCatalogue(t *testing.T, smallK, largeK uint64) *pirconfig.Catalogue {
	c := fixtureCatalogue(t, 2, 3)
	for i, k := range []uint64{smallK, largeK} {
		entry := &c.Configurations[i]
		entry.EncodingStatus = "candidate"
		entry.Capacity.Status = "estimated"
		value := k
		rate := float64(k) / float64(entry.Capacity.N)
		entry.Capacity.K = &value
		entry.Capacity.Rate = &rate
	}
	return c
}

func TestGreedyImprovementAndOversizedChoice(t *testing.T) {
	c := searchCatalogue(t, 4, 12)
	p, e, err := Greedy(12, 6, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	// Largest-fitting alone chooses the large block (384 bits). Lookahead chooses
	// three small blocks (288 bits), the optimum for this synthetic model.
	if e.Costs.StorageBits != 288 || len(p.RowBlocks) != 3 || e.DatabaseCount != 9 {
		t.Fatalf("unexpected improvement: %+v %+v", p, e)
	}
	if !e.Provisional {
		t.Fatal("synthetic estimates treated as executable")
	}
	if _, _, err := Greedy(12, 6, c, Requirements{MaxStorageBits: 288}); err != nil {
		t.Fatal("feasible budget rejected", err)
	}
	if _, _, err := Greedy(12, 6, c, Requirements{MaxStorageBits: 287}); !errors.Is(err, ErrNoPlan) {
		t.Fatal("impossible budget accepted", err)
	}
	c = searchCatalogue(t, 2, 12)
	p, e, err = Greedy(11, 6, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.RowBlocks) != 1 || p.RowBlocks[0].ConfigurationID != "fixture-s3" || p.RowBlocks[0].RowEnd != 11 || e.Costs.StorageBits != 384 {
		t.Fatal("beneficial oversized block was missed")
	}
}

func expandSmall(t *testing.T, p Plan, c *pirconfig.Catalogue) Plan {
	t.Helper()
	explicit := Plan{FormatVersion: p.FormatVersion, Rows: p.Rows, RecordBits: p.RecordBits}
	widths := map[string]uint64{}
	for _, entry := range c.Configurations {
		widths[entry.ID] = uint64(entry.Code.S)
	}
	for _, b := range p.RowBlocks {
		s := widths[b.ConfigurationID]
		repeat := b.Repeat
		if repeat == 0 {
			repeat = 1
		}
		height := (b.RowEnd - b.RowStart) / repeat
		for index := uint64(0); index < repeat; index++ {
			row := b.RowStart + index*height
			for col := uint64(0); col < p.RecordBits; col += s {
				end := col + s
				if end > p.RecordBits {
					end = p.RecordBits
				}
				explicit.Databases = append(explicit.Databases, Database{RowStart: row, RowEnd: row + height, ColumnStart: col, ColumnEnd: end, ConfigurationID: b.ConfigurationID, Group: b.Group})
			}
		}
	}

	return explicit
}

func TestCompactAndExplicitEvaluationAgree(t *testing.T) {
	c := fixtureCatalogue(t, 2, 3, 4)
	rng := rand.New(rand.NewSource(83))
	for i := 0; i < 30; i++ {
		rows, width := uint64(1+rng.Intn(30)), uint64(4+rng.Intn(40))
		p, got, err := Greedy(rows, width, c, Requirements{})
		if err != nil {
			t.Fatal(err)
		}
		explicit := expandSmall(t, p, c)
		want, err := Evaluate(explicit, c, Requirements{})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("implicit/explicit mismatch for %d x %d", rows, width)
		}
		if len(p.Databases) != 0 {
			t.Fatal("greedy materialized column slices")
		}
	}
	p, _, err := Greedy(7, 7, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	p.Databases = expandSmall(t, p, c).Databases
	if _, err := Evaluate(p, c, Requirements{}); err == nil {
		t.Fatal("mixed representations accepted")
	}
	p.Databases = nil
	p.RowBlocks[0].RowEnd = 0
	if _, err := Evaluate(p, c, Requirements{}); err == nil {
		t.Fatal("empty compact block accepted")
	}
	p.RowBlocks[0].ConfigurationID = "missing"
	if _, err := Evaluate(p, c, Requirements{}); err == nil {
		t.Fatal("unknown compact config accepted")
	}
}

func TestGreedyBaselineAndOptimalLowerBound(t *testing.T) {
	c := searchCatalogue(t, 3, 11)
	width := uint64(6)
	// Independent, small integer recurrence for the optimum within homogeneous
	// row-block plans. No claim is made that one-step lookahead always attains it.
	type item struct{ k, s, u, d uint64 }
	var items []item
	for _, entry := range c.Configurations {
		columns := (width + uint64(entry.Code.S) - 1) / uint64(entry.Code.S)
		items = append(items, item{*entry.Capacity.K, columns * entry.Capacity.N * uint64(entry.Code.S), columns * entry.Query.Length * uint64(entry.Code.S*entry.Code.M), columns * entry.Query.Length * uint64(entry.Code.S)})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].k > items[j].k })
	type pair struct{ storage, traffic uint64 }
	lessPair := func(a, b pair) bool {
		return a.storage < b.storage || (a.storage == b.storage && a.traffic < b.traffic)
	}
	dp := make([]pair, 81)
	for rows := 1; rows <= 80; rows++ {
		dp[rows] = pair{math.MaxUint64, math.MaxUint64}
		for _, x := range items {
			rest := 0
			if uint64(rows) > x.k {
				rest = rows - int(x.k)
			}
			candidate := pair{x.s + dp[rest].storage, x.d + dp[rest].traffic}
			if lessPair(candidate, dp[rows]) {
				dp[rows] = candidate
			}
		}
		var baseline pair
		remaining := uint64(rows)
		for remaining > 0 {
			selected := -1
			for i, x := range items {
				if x.k <= remaining {
					selected = i
					break
				}
			}
			if selected < 0 {
				selected = 0
				for i, x := range items {
					if lessPair(pair{x.s, x.d}, pair{items[selected].s, items[selected].d}) {
						selected = i
					}
				}
			}
			x := items[selected]
			baseline.storage += x.s
			baseline.traffic += x.d
			if x.k >= remaining {
				remaining = 0
			} else {
				remaining -= x.k
			}
		}
		_, e, err := Greedy(uint64(rows), width, c, Requirements{})
		if err != nil {
			t.Fatal(err)
		}
		got := pair{e.Costs.StorageBits, e.Costs.DownloadBits}
		if lessPair(got, dp[rows]) || lessPair(baseline, got) {
			t.Fatalf("rows=%d optimum=%v greedy=%v baseline=%v", rows, dp[rows], got, baseline)
		}
	}
	// Tail scoring batches even a huge number of repeated blocks.
	costs, err := databaseCosts(&c.Configurations[0], 2)
	if err != nil {
		t.Fatal(err)
	}
	tail, err := greedyTail(3_000_000_000_000, []blockOption{{id: "small", capacity: 3, costs: costs}})
	if err != nil || tail.StorageBits != costs.StorageBits*1_000_000_000_000 {
		t.Fatal("batched tail is incorrect", err)
	}
}

func TestGreedyDeterminismFiltersAndTrafficTie(t *testing.T) {
	c := searchCatalogue(t, 4, 12)
	before, _ := json.Marshal(c)
	p, e, err := Greedy(31, 7, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	c.Configurations[0], c.Configurations[1] = c.Configurations[1], c.Configurations[0]
	reversed, re, err := Greedy(31, 7, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p, reversed) || !reflect.DeepEqual(e, re) {
		t.Fatal("catalogue order changes result")
	}
	c.Configurations[0], c.Configurations[1] = c.Configurations[1], c.Configurations[0]
	after, _ := json.Marshal(c)
	if string(before) != string(after) {
		t.Fatal("catalogue mutated")
	}
	c.Configurations[0].Security.ComputationalBits = 10
	c.Configurations[1].Security.ComputationalBits = 100
	c.Configurations[1].Security.Log2Samples = 50
	p, _, err = Greedy(12, 6, c, Requirements{MinSecurityBits: 100, Log2Samples: 50})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range p.RowBlocks {
		if b.ConfigurationID != "fixture-s3" {
			t.Fatal("ineligible security profile selected")
		}
	}
	if _, _, err := Greedy(12, 6, c, Requirements{MinSecurityBits: 101}); !errors.Is(err, ErrNoPlan) {
		t.Fatal("insufficient security accepted", err)
	}
	c = fixtureCatalogue(t, 2)
	second := c.Configurations[0]
	second.ID = "low-traffic"
	second.Query.DummyPositions = 0
	second.Query.Length = 3
	c.Configurations = append(c.Configurations, second)
	p, _, err = Greedy(3, 4, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if p.RowBlocks[0].ConfigurationID != "low-traffic" {
		t.Fatal("storage tie did not use total traffic")
	}
	c.Configurations[0].EncodingStatus = "candidate"
	c.Configurations[0].Capacity.Status = "unknown"
	c.Configurations[0].Capacity.K = nil
	c.Configurations[0].Capacity.Rate = nil
	c.Configurations[0].Capacity.Evidence = nil
	if _, _, err := Greedy(3, 4, c, Requirements{}); err != nil {
		t.Fatal("unused unknown capacity excluded usable option", err)
	}
}

func TestWideRecordsRemainCompact(t *testing.T) {
	f, err := os.Open("../pirconfig/configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	p, e, err := Greedy(100, 80_000_000, c, Requirements{MinSecurityBits: 100, Log2Samples: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.RowBlocks) != 1 || len(p.Databases) != 0 || e.DatabaseCount != 10_000_000 {
		t.Fatal("wide-record plan is not compact")
	}
	if e.Costs.StorageBits != 5_242_880_000_000 || e.Costs.UploadBits != 61_280_000_000 || e.Costs.DownloadBits != 30_640_000_000 {
		t.Fatalf("incorrect wide-record costs: %+v", e.Costs)
	}
}

func TestGreedyErrorsAndConcurrency(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	if _, _, err := Greedy(0, 4, c, Requirements{}); err == nil {
		t.Fatal("empty application accepted")
	}
	if _, _, err := Greedy(1, 1, c, Requirements{}); err == nil {
		t.Fatal("too-small records accepted")
	}
	if _, _, err := Greedy(math.MaxUint64, 4, c, Requirements{}); err == nil {
		t.Fatal("application overflow accepted")
	}
	if _, _, err := Greedy(1, 4, nil, Requirements{}); err == nil {
		t.Fatal("nil catalogue accepted")
	}
	if _, _, err := Greedy(1, 4, c, Requirements{MaxUploadBits: 1}); !errors.Is(err, ErrNoPlan) {
		t.Fatal("upload budget ignored", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := Greedy(31, 7, c, Requirements{}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
