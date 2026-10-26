// This experiment evaluates a rounded cost-grid DP, not encoded database usage.
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
	ReferenceConfigID          string                `json:"reference_config_id"`
	ReferenceCapacityBits      uint64                `json:"reference_capacity_bits"`
	ReferenceStorageOverhead   float64               `json:"reference_storage_overhead"`
	ApplicationBytes           uint64                `json:"application_bytes"`
	RecordBytes                uint64                `json:"record_bytes"`
	OriginalGreedy             planner.Costs         `json:"original_greedy"`
	RepresentativeIndices      []int                 `json:"representative_indices"`
	OriginalGreedyWitnessIndex int                   `json:"original_greedy_witness_index"`
	Result                     *planner.DPCostResult `json:"result"`
}

func run(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		return err
	}
	requirements := planner.Requirements{MinSecurityBits: 100, Log2Samples: 50}
	options := planner.DPCostOptions{MaxFrontier: 32, StorageCells: 128, CommunicationCells: 128, MaxBaselineBlocks: 256}
	var scenarios []scenario
	for _, size := range []uint64{1000000000, 20000000000, 50000000000, 100000000000, 500000000000, 1000000000000} {
		for _, record := range []uint64{1000, 10000, 1000000, 10000000} {
			if size/record < minApplicationRecords {
				continue
			}
			result, err := planner.SearchDPCosts(size/record, record*8, c, requirements, options)
			if err != nil {
				return fmt.Errorf("size %d record %d: %w", size, record, err)
			}
			for i, a := range result.Alternatives {
				checked, err := planner.EvaluateApplication(a.Plan, c, requirements)
				if err != nil || !reflect.DeepEqual(a.Evaluation, checked) {
					return fmt.Errorf("evaluation mismatch: %v", err)
				}
				costs := checked.Costs
				if costs.ApplicationBits != size*8 {
					return fmt.Errorf("input size changed")
				}
				if !checked.FullDownloadOnly && costs.StorageBits >= size*8 && costs.CommunicationBits >= size*8 {
					return fmt.Errorf("whole-download dominated plan returned")
				}
				for j, b := range result.Alternatives {
					other := b.Evaluation.Costs
					if i != j && other.StorageBits <= costs.StorageBits && other.CommunicationBits <= costs.CommunicationBits {
						return fmt.Errorf("dominated or duplicate alternative")
					}
				}
			}
			_, original, err := planner.Greedy(size/record, record*8, c, requirements)
			if err != nil {
				return err
			}
			representatives, err := result.RepresentativeIndices(3)
			if err != nil {
				return err
			}
			displayed := map[int]bool{}
			for _, index := range representatives {
				displayed[index] = true
			}
			for _, w := range result.BaselineWitnesses {
				cost := result.Alternatives[w.AlternativeIndex].Evaluation.Costs
				if !displayed[w.AlternativeIndex] || cost.StorageBits > w.Costs.StorageBits || cost.CommunicationBits > w.Costs.CommunicationBits {
					return fmt.Errorf("baseline/display guarantee failed")
				}
			}
			witness := -1
			for _, index := range representatives {
				cost := result.Alternatives[index].Evaluation.Costs
				if cost.StorageBits <= original.Costs.StorageBits && cost.CommunicationBits <= original.Costs.CommunicationBits {
					witness = index
					break
				}
			}
			if witness < 0 {
				return fmt.Errorf("original greedy/display regression")
			}
			reference := c.Configurations[0]
			difference := func(entry pirconfig.Configuration) uint64 {
				capacity := *entry.Capacity.K * uint64(entry.Code.S)
				if capacity > size*8 {
					return capacity - size*8
				}
				return size*8 - capacity
			}
			for _, entry := range c.Configurations {
				if difference(entry) < difference(reference) || (difference(entry) == difference(reference) && entry.ID < reference.ID) {
					reference = entry
				}
			}
			fmt.Fprintf(os.Stderr, "DP completed %d bytes / %d-byte records\n", size, record)
			scenarios = append(scenarios, scenario{ReferenceConfigID: reference.ID, ReferenceCapacityBits: *reference.Capacity.K * uint64(reference.Code.S), ReferenceStorageOverhead: float64(reference.Capacity.N) / float64(*reference.Capacity.K), ApplicationBytes: size, RecordBytes: record, OriginalGreedy: original.Costs, RepresentativeIndices: representatives, OriginalGreedyWitnessIndex: witness, Result: result})
		}
	}
	output := struct {
		Version      int                   `json:"version"`
		Units        string                `json:"units"`
		GoVersion    string                `json:"go_version"`
		GOOS         string                `json:"goos"`
		GOARCH       string                `json:"goarch"`
		Requirements planner.Requirements  `json:"requirements"`
		Options      planner.DPCostOptions `json:"options"`
		Scenarios    []scenario            `json:"scenarios"`
	}{1, "Decimal bytes; communication rate = original record bits / (upload + download per original lookup).", runtime.Version(), runtime.GOOS, runtime.GOARCH, requirements, options, scenarios}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// Minimum workload size for the default application cost grid; correctness fixtures
// and the planner API remain able to handle smaller inputs.
const minApplicationRecords uint64 = 5000

func main() {
	path := flag.String("catalogue", "experiments/dp-costs/pir-configs.json", "fixed catalogue snapshot")
	flag.Parse()
	if err := run(*path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
