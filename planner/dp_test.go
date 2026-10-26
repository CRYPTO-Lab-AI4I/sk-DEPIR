package planner

import (
	"math"
	"reflect"
	"testing"
)

func TestCostDPExactGridAgainstExhaustive(t *testing.T) {
	c := tradeoffCatalogue(t)
	for _, rows := range []uint64{1, 13, 15, 27, 39, 40, 41} {
		got, err := SearchDPCosts(rows, 6, c, Requirements{}, DPCostOptions{LogicalRecordBits: []uint64{3}, StorageCells: 512, CommunicationCells: 512, StorageWindowBits: 500, CommunicationWindowBits: 250, MaxFrontier: 1000})
		if err != nil {
			t.Fatal(err)
		}
		want := exhaustiveTradeoffs(t, rows, 6, c, Requirements{})
		pairs := map[costPair]bool{}
		for _, a := range got.Alternatives {
			costs := a.Evaluation.Costs
			pairs[costPair{costs.StorageBits, costs.CommunicationBits}] = true
			e, err := EvaluateApplication(a.Plan, c, Requirements{})
			if err != nil || !reflect.DeepEqual(e, a.Evaluation) {
				t.Fatal("cost DP reconstruction", err)
			}
		}
		if !reflect.DeepEqual(pairs, want) {
			t.Fatalf("rows %d: got %v want %v", rows, pairs, want)
		}
		for _, w := range got.Windows {
			if w.Rounded {
				t.Fatal("exact test used rounding")
			}
		}
	}
}

func TestCostDPRoundingAndWitnesses(t *testing.T) {
	c := tradeoffCatalogue(t)
	options := DPCostOptions{LogicalRecordBits: []uint64{3}, StorageCells: 8, CommunicationCells: 8, MaxFrontier: 2}
	result, err := SearchDPCosts(1000, 6, c, Requirements{}, options)
	if err != nil {
		t.Fatal(err)
	}
	again, err := SearchDPCosts(1000, 6, c, Requirements{}, options)
	if err != nil || !reflect.DeepEqual(result, again) {
		t.Fatal("DP not deterministic")
	}
	rounded := false
	for _, w := range result.Windows {
		rounded = rounded || w.Rounded
	}
	if !rounded {
		t.Fatal("rounding not exercised")
	}
	indices, err := result.RepresentativeIndices(3)
	if err != nil {
		t.Fatal(err)
	}
	shown := map[int]bool{}
	for _, i := range indices {
		shown[i] = true
	}
	for _, w := range result.BaselineWitnesses {
		if !shown[w.AlternativeIndex] || !dominates(result.Alternatives[w.AlternativeIndex].Evaluation.Costs, w.Costs, false) {
			t.Fatal("lost protected witness")
		}
	}
}

func TestCostDPValidation(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	for _, o := range []DPCostOptions{{StorageCells: -1}, {CommunicationCells: 513}, {MaxFrontier: 1}, {MaxBaselineBlocks: -1}, {LogicalRecordBits: []uint64{0}}} {
		if _, err := SearchDPCosts(3, 4, c, Requirements{}, o); err == nil {
			t.Fatal("invalid options")
		}
	}
	if _, err := SearchDPCosts(3, 4, c, Requirements{MaxDownloadBits: 1}, DPCostOptions{}); err == nil {
		t.Fatal("directional constraint silently accepted")
	}
	if _, err := SearchDPCosts(math.MaxUint64, 4, c, Requirements{}, DPCostOptions{}); err == nil {
		t.Fatal("overflow")
	}
	if _, err := SearchDPCosts(3, 4, nil, Requirements{}, DPCostOptions{}); err == nil {
		t.Fatal("nil catalogue")
	}
	result, err := SearchDPCosts(3, 4, c, Requirements{}, DPCostOptions{})
	if err != nil || len(result.Alternatives) != 1 || !result.Alternatives[0].Evaluation.FullDownloadOnly {
		t.Fatal("dominated encoded option")
	}
}
