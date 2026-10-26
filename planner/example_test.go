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
