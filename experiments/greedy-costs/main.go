// This experiment evaluates greedy plans, not physical database encodings.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/planner"
)

type scenario struct {
	ApplicationBytes uint64              `json:"application_bytes"`
	RecordBytes      uint64              `json:"record_bytes"`
	Plan             planner.Plan        `json:"plan"`
	Evaluation       *planner.Evaluation `json:"evaluation"`
}

type results struct {
	Version      int                  `json:"version"`
	Units        string               `json:"units"`
	GoVersion    string               `json:"go_version"`
	GOOS         string               `json:"goos"`
	GOARCH       string               `json:"goarch"`
	Requirements planner.Requirements `json:"requirements"`
	Scenarios    []scenario           `json:"scenarios"`
}

func main() {
	path := flag.String("catalogue", "experiments/greedy-costs/pir-configs.json", "fixed catalogue snapshot")
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
	requirements := planner.Requirements{MinSecurityBits: 100, Log2Samples: 50}
	output := results{Version: 1, Units: "Decimal GB=10^9 bytes, TB=10^12 bytes; KB=10^3 bytes, MB=10^6 bytes. Costs are logical bits; upload/download are per application-record retrieval.", GoVersion: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, Requirements: requirements}
	for _, size := range []uint64{1_000_000_000, 20_000_000_000, 50_000_000_000, 100_000_000_000, 500_000_000_000, 1_000_000_000_000} {
		for _, record := range []uint64{1000, 10000, 1000000, 10000000} {
			if size%record != 0 {
				return fmt.Errorf("application size is not divisible by record size")
			}
			plan, evaluation, err := planner.Greedy(size/record, record*8, c, requirements)
			if err != nil {
				return fmt.Errorf("size=%d record=%d: %w", size, record, err)
			}
			if evaluation.Costs.ApplicationBits != size*8 {
				return fmt.Errorf("application size changed")
			}
			output.Scenarios = append(output.Scenarios, scenario{ApplicationBytes: size, RecordBytes: record, Plan: plan, Evaluation: evaluation})
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
