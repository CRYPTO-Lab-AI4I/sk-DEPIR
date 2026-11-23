package liftrs

import "fmt"

// Named simultaneously retained arrays,not an RSS bound. The caller separately
// reserves native/runtime/GC/allocator slack; no forced collection is implied.
type rankHostInventory struct {
	Phase      string            `json:"phase"`
	Components map[string]uint64 `json:"components_bytes"`
	NamedBytes uint64            `json:"named_bytes"`
}

func hostInventory(phase string, components map[string]uint64) rankHostInventory {
	p := rankHostInventory{Phase: phase, Components: components}
	for _, bytes := range components {
		p.NamedBytes += bytes
	}
	return p
}
func hostTransformScratch(s int, n uint64) map[string]uint64 {
	width := uint64(4)
	if s > 16 {
		width = 8
	}
	rows, cols := uint64(1)<<uint(2*s-s), uint64(1)<<uint(s)
	// Two resident Taylor host windows, two transfer tiles,field log/power tables
	// and bounded codec/routing scratch. Native allocations are in caller reserve.
	capacity := min(n, 3*max(rows, cols))
	return map[string]uint64{"Taylor_windows": 2 * capacity * width, "transfer_tiles": 2 * 4096 * width, "field_tables": 8 * (uint64(1) << uint(s)), "codec_and_control": 4 * 65536}
}
func hostArrayPhase(phase string, s int, n, baseWords, indexWords, bitmapWords, extensionWords uint64) rankHostInventory {
	components := hostTransformScratch(s, n)
	width := uint64(4)
	if s > 16 {
		width = 8
	}
	components["base_coefficients"] = 4 * baseWords
	components["indices"] = 8 * indexWords
	components["bitmap"] = 8 * bitmapWords
	components["extension_result"] = width * extensionWords
	return hostInventory(phase, components)
}

// Actual root counts determine each rounded subgroup domain. Six coefficient
// shard equivalents conservatively cover builder/inputs/result/normalization;
// abandoned allocations and unmodeled native state belong to the explicit reserve.
func rootHostInventories(s, rank int, counts []uint64, n uint64, rootCapacity uint64) ([]rankHostInventory, error) {
	stages, err := rankRootMergeStages(rank, counts, n)
	if err != nil {
		return nil, err
	}
	out := []rankHostInventory{hostArrayPhase("local-roots", s, n, 6*counts[rank], rootCapacity, 0, 0)}
	out[0].Components["factor_batch"] = 16 * min(uint64(65536), counts[rank]-1)
	out[0].NamedBytes += out[0].Components["factor_batch"]
	for _, stage := range stages {
		words := stage.Domain / uint64(stage.Ranks)
		p := hostArrayPhase(fmt.Sprintf("root-group-%d", stage.Level), s, n, 6*words, rootCapacity, 0, 0)
		out = append(out, p)
	}
	_, words := rankBand(n, rank, len(counts))
	compact := stages[len(stages)-1].Domain / uint64(len(counts))
	out = append(out, hostArrayPhase("root-reshape", s, n, compact+2*words, rootCapacity, 0, 0))
	return out, nil
}

func (c DistributedPreprocessingConfig) admitHost(p rankHostInventory) error {
	if c.HostReserveBytes == 0 || c.HostReserveBytes >= c.RSSBytes || p.NamedBytes > c.RSSBytes-c.HostReserveBytes {
		return fmt.Errorf("host phase %s: named %d bytes + reserve %d bytes exceed RSS cap %d", p.Phase, p.NamedBytes, c.HostReserveBytes, c.RSSBytes)
	}
	return nil
}
