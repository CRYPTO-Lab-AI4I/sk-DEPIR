package planner_test

import (
	"fmt"
	"os"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/planner"
)

func ExampleEvaluate() {
	file, err := os.Open("../pirconfig/testdata/reference.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	catalogue, err := pirconfig.Load(file)
	if err != nil {
		panic(err)
	}
	id := catalogue.Configurations[0].ID
	plan := planner.Plan{FormatVersion: 1, Rows: 3, RecordBits: 4, Databases: []planner.Database{
		{RowEnd: 3, ColumnEnd: 2, ConfigurationID: id, Group: 0},
		{ColumnStart: 2, RowEnd: 3, ColumnEnd: 4, ConfigurationID: id, Group: 0},
	}}
	// The tiny fixture has no security claim. Real target plans request the
	// catalogue's conjectured profile with MinSecurityBits:100, Log2Samples:50.
	result, err := planner.Evaluate(plan, catalogue, planner.Requirements{})
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Costs.StorageBits, result.Costs.UploadBits, result.Costs.DownloadBits)
	fmt.Println(result.EffectiveStorageRate, result.Provisional)
	// Output:
	// 64 1048 524
	// 0.1875 false
}

func ExampleGreedy() {
	file, err := os.Open("../pirconfig/testdata/reference.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	catalogue, err := pirconfig.Load(file)
	if err != nil {
		panic(err)
	}
	// Correctness fixture only; use the target requirements for actual profiles.
	plan, result, err := planner.Greedy(7, 4, catalogue, planner.Requirements{})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(plan.RowBlocks), len(plan.Databases), result.DatabaseCount)
	fmt.Println(result.Costs.StorageBits, result.Costs.UploadBits, result.Costs.DownloadBits)
	// Output:
	// 2 0 4
	// 128 2096 1048
}

func ExampleGreedyApplication() {
	file, err := os.Open("../pirconfig/testdata/reference.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	catalogue, err := pirconfig.Load(file)
	if err != nil {
		panic(err)
	}
	// A tiny fixture: keep at most four residual bits without code expansion.
	plan, result, err := planner.GreedyApplication(6, 4, catalogue, planner.Requirements{}, planner.ApplicationOptions{
		LogicalRecordBits: []uint64{4}, AllowFullDownload: true,
		MaxFullDownloadBits: 4, // optional absolute cap in addition to default ratio 1
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(plan.LogicalRecordBits, result.FullDownloadBits)
	fmt.Println(result.Costs.StorageBits, result.Costs.UploadBits, result.Costs.DownloadBits)
	// Output:
	// 4 4
	// 68 1048 528
}

func ExampleSearchTradeoffs() {
	file, err := os.Open("../pirconfig/testdata/reference.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	catalogue, err := pirconfig.Load(file)
	if err != nil {
		panic(err)
	}
	result, err := planner.SearchTradeoffs(3, 4, catalogue, planner.Requirements{}, planner.TradeoffOptions{MaxFrontier: 16, MaxStates: 64})
	if err != nil {
		panic(err)
	}
	// This tiny correctness fixture communicates less by downloading everything.
	// Actual target inputs can return several storage/communication alternatives.
	fmt.Println(len(result.Alternatives), result.Alternatives[0].Evaluation.FullDownloadOnly)
	fmt.Printf("%.3f\n", result.Alternatives[0].Evaluation.CommunicationRate)
	indices, err := result.RepresentativeIndices(3)
	if err != nil {
		panic(err)
	}
	fmt.Println(indices)
	// Output:
	// 1 true
	// 0.333
	// [0]
}

func ExampleSearchDPCosts() {
	file, err := os.Open("../pirconfig/testdata/reference.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	catalogue, err := pirconfig.Load(file)
	if err != nil {
		panic(err)
	}
	result, err := planner.SearchDPCosts(3, 4, catalogue, planner.Requirements{}, planner.DPCostOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println(len(result.Alternatives), result.Alternatives[0].Evaluation.FullDownloadOnly)
	// Output:
	// 1 true
}
