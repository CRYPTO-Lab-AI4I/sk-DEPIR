package planner

import (
	"reflect"
	"testing"
)

func TestRepeatedBlockMatchesExpandedCosts(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	p := Plan{FormatVersion: 1, Rows: 15, RecordBits: 4, RowBlocks: []RowBlock{{RowEnd: 15, ConfigurationID: c.Configurations[0].ID, Repeat: 3}}}
	compact, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := Evaluate(expandSmall(t, p, c), c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(compact, expanded) || compact.DatabaseCount != 6 {
		t.Fatal("batch differs from physical expansion")
	}
	p.Rows = 5000000000
	p.RowBlocks[0].RowEnd = p.Rows
	p.RowBlocks[0].Repeat = 1000000000
	e, err := Evaluate(p, c, Requirements{})
	if err != nil || e.DatabaseCount != 2000000000 {
		t.Fatal("large compact batch", err)
	}
	for _, b := range []RowBlock{{RowEnd: 15, Repeat: 2, ConfigurationID: c.Configurations[0].ID}, {RowEnd: 15, Repeat: 20, ConfigurationID: c.Configurations[0].ID}, {RowEnd: 15, Repeat: 2, DownloadAll: true}} {
		p.Rows = 15
		p.RowBlocks = []RowBlock{b}
		if _, err := Evaluate(p, c, Requirements{}); err == nil {
			t.Fatal("invalid repetition accepted")
		}
	}
}

func TestCountSamplingAndBatchedPadding(t *testing.T) {
	values, trimmed := blockCounts(11, 4, 8)
	if trimmed || !reflect.DeepEqual(values, []uint64{0, 1, 2, 3}) {
		t.Fatal(values, trimmed)
	}
	values, trimmed = blockCounts(1000000, 3, 8)
	if !trimmed || len(values) > 8 {
		t.Fatal("unbounded count search")
	}
	wanted := map[uint64]bool{0: false, 1: false, 333333: false, 333334: false}
	for _, value := range values {
		if _, ok := wanted[value]; ok {
			wanted[value] = true
		}
	}
	for _, found := range wanted {
		if !found {
			t.Fatal("count endpoints missing")
		}
	}
	prefix := &partialPlan{}
	last, err := appendBatch(prefix, 3, 11, blockOption{id: "test", capacity: 4, costs: Costs{StorageBits: 10, UploadBits: 2, DownloadBits: 3, CommunicationBits: 5}})
	if err != nil {
		t.Fatal(err)
	}
	p := partialLayout(last, 11, 4)
	if len(p.RowBlocks) != 2 || p.RowBlocks[0].Repeat != 2 || p.RowBlocks[0].RowEnd != 8 || p.RowBlocks[1].RowStart != 8 || last.costs.StorageBits != 30 {
		t.Fatal("padded batch accounting", p)
	}
}

func TestCapacityOrderedSearch(t *testing.T) {
	c := tradeoffCatalogue(t)
	result, err := SearchTradeoffs(1000, 6, c, Requirements{}, TradeoffOptions{LogicalRecordBits: []uint64{6}, MaxStates: 8, MaxCounts: 4})
	if err != nil {
		t.Fatal(err)
	}
	if !result.CountChoicesTrimmed {
		t.Fatal("count approximation not reported")
	}
	for _, a := range result.Alternatives {
		// Protected greedy baselines need not follow the search's canonical order.
		// Independently reconstructed costs must still agree for every returned plan.
		e, err := EvaluateApplication(a.Plan, c, Requirements{})
		if err != nil || !reflect.DeepEqual(e, a.Evaluation) {
			t.Fatal("returned batched plan invalid", err)
		}
	}
}
