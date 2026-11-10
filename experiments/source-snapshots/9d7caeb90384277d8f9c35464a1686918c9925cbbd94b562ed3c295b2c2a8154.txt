// This experiment evaluates greedy plans, not physical database encodings.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/planner"
)

type scenario struct {
	ApplicationBytes       uint64                  `json:"application_bytes"`
	RecordBytes            uint64                  `json:"record_bytes"`
	IndexedAlternativePlan planner.ApplicationPlan `json:"indexed_alternative_plan"`
	IndexedAlternative     *planner.Evaluation     `json:"indexed_alternative"`
	Plan                   planner.ApplicationPlan `json:"plan"`
	Evaluation             *planner.Evaluation     `json:"evaluation"`
}

type results struct {
	Version      int                        `json:"version"`
	Units        string                     `json:"units"`
	GoVersion    string                     `json:"go_version"`
	GOOS         string                     `json:"goos"`
	GOARCH       string                     `json:"goarch"`
	Options      planner.ApplicationOptions `json:"options"`
	Requirements planner.Requirements       `json:"requirements"`
	Scenarios    []scenario                 `json:"scenarios"`
}

// Minimum workload size for the default application cost grid; correctness fixtures
// and the planner API remain able to handle smaller inputs.
const minApplicationRecords uint64 = 5000

func main() {
	path := flag.String("catalogue", "experiments/application-costs/pir-configs.json", "fixed catalogue snapshot")
	flag.Parse()
	if err := run(*path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	c, err := pirconfig.Load(file)
	if err != nil {
		return err
	}
	options := planner.ApplicationOptions{AllowFullDownload: true, FullDownloadRatio: 1}
	requirements := planner.Requirements{MinSecurityBits: 100, Log2Samples: 50}
	output := results{Version: 1, Units: "Decimal GB=10^9 bytes, TB=10^12 bytes; KB=10^3 bytes, MB=10^6 bytes. Costs are logical bits; upload/download are per application-record retrieval.", GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Requirements: requirements, Options: options}
	for _, size := range []uint64{1_000_000_000, 20_000_000_000, 50_000_000_000, 100_000_000_000, 500_000_000_000, 1_000_000_000_000} {
		for _, record := range []uint64{1000, 10000, 1000000, 10000000} {
			if size/record < minApplicationRecords {
				continue
			}
			if size%record != 0 {
				return fmt.Errorf("application size is not divisible by record size")
			}
			plan, evaluation, err := planner.GreedyApplication(size/record, record*8, c, requirements, options)
			if err != nil {
				return fmt.Errorf("size=%d record=%d: %w", size, record, err)
			}
			if evaluation.Costs.ApplicationBits != size*8 {
				return fmt.Errorf("application size changed")
			}
			checked, err := planner.EvaluateApplication(plan, c, requirements)
			if err != nil || !reflect.DeepEqual(checked, evaluation) {
				return fmt.Errorf("re-evaluation disagrees: %v", err)
			}
			if !evaluation.FullDownloadOnly && evaluation.FullDownloadBits > evaluation.Costs.DownloadBits-evaluation.FullDownloadBits {
				return fmt.Errorf("default full-download ratio exceeded")
			}
			indexedOptions := options
			indexedOptions.AllowFullDownload = false
			indexedPlan, indexed, err := planner.GreedyApplication(size/record, record*8, c, requirements, indexedOptions)
			if err != nil {
				return err
			}
			_, original, err := planner.Greedy(size/record, record*8, c, requirements)
			if err != nil {
				return err
			}
			if evaluation.FullDownloadBits > original.Costs.DownloadBits {
				return fmt.Errorf("full pieces exceed original-width download")
			}
			fmt.Fprintf(os.Stderr, "Completed %d bytes / %d-byte records\n", size, record)
			output.Scenarios = append(output.Scenarios, scenario{IndexedAlternativePlan: indexedPlan, IndexedAlternative: indexed, ApplicationBytes: size, RecordBytes: record, Plan: plan, Evaluation: evaluation})
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
