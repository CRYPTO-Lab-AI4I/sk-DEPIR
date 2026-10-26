// Package planner evaluates application-to-PIR partition plans and constructs
// greedy row-block partitions using documented configuration capacities. It never
// preprocesses codes, encodes databases, or generates queries. Estimated capacities support provisional
// planning; mathematical and backend validation remain separate prerequisites.
package planner

import (
	"fmt"
	"math/bits"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// Plan describes a version-1 application layout. Databases must be in ascending
// (row_start,column_start,row_end,column_end) order and cover the matrix exactly
// once. Alternatively, RowBlocks implicitly slice each full-width row block into
// field-sized columns. Supply exactly one representation. RecordBits must be at
// least the largest s in the supplied catalogue.
// This is public planning data, not a stored-database manifest or secret state.
type Plan struct {
	FormatVersion int        `json:"format_version"`
	Rows          uint64     `json:"rows"`
	RecordBits    uint64     `json:"record_bits"`
	Databases     []Database `json:"databases,omitempty"`
	RowBlocks     []RowBlock `json:"row_blocks,omitempty"`
}

// Database assigns a half-open application rectangle to one catalogue entry.
// Group numbers start at zero in first-occurrence order. All members of a group
// reference the same configuration ID, fixing its padded/code/query conventions.
// Their row ranges must be disjoint or have the same start.
type Database struct {
	RowStart        uint64 `json:"row_start"`
	ColumnStart     uint64 `json:"column_start"`
	RowEnd          uint64 `json:"row_end"`
	ColumnEnd       uint64 `json:"column_end"`
	ConfigurationID string `json:"configuration_id"`
	Group           uint64 `json:"group"`
}

// RowBlock represents adjacent vertical PIR databases covering all application
// columns for [RowStart,RowEnd). For field width s, column slice j is
// [j*s,min((j+1)*s,RecordBits)). Each slice remains a separate database with its
// own masks and full address upload. This is metadata compression, not query or
// storage compression. Row blocks must cover all rows in increasing order.
type RowBlock struct {
	RowStart        uint64 `json:"row_start"`
	RowEnd          uint64 `json:"row_end"`
	ConfigurationID string `json:"configuration_id"`
	Group           uint64 `json:"group"`
}

// Requirements imposes security-profile and logical-bit budgets. Zero cost
// limits mean unlimited; zero security requirements permit correctness fixtures.
// Security targets are compared with catalogue estimates, not independently proved.
type Requirements struct {
	MinSecurityBits uint32 `json:"min_security_bits"`
	Log2Samples     uint32 `json:"log2_samples"`
	MaxStorageBits  uint64 `json:"max_storage_bits,omitempty"`
	MaxUploadBits   uint64 `json:"max_upload_bits,omitempty"`
	MaxDownloadBits uint64 `json:"max_download_bits,omitempty"`
}

// Costs counts logical bits, with full-codeword storage, packed field elements,
// and minimal-width explicit addresses. Upload/download are per application-record
// retrieval and include every database, dummy, and repeated address. Byte rounding,
// metadata, storage request framing, and transport overheads are excluded.
type Costs struct {
	ApplicationBits   uint64 `json:"application_bits"`
	StorageBits       uint64 `json:"storage_bits"`
	UploadBits        uint64 `json:"upload_bits"`
	DownloadBits      uint64 `json:"download_bits"`
	CommunicationBits uint64 `json:"communication_bits"`
}

// ConfigurationUse retains the provenance behind each used capacity and security
// profile. Entries appear in first-use order; unrelated resource measurements are
// omitted because they do not influence feasibility or partition selection.
type ConfigurationUse struct {
	ID               string               `json:"id"`
	CapacityStatus   string               `json:"capacity_status"`
	EncodingStatus   string               `json:"encoding_status"`
	CapacityEvidence []pirconfig.Evidence `json:"capacity_evidence"`
	Security         pirconfig.Security   `json:"security"`
}

// Evaluation contains exact integer costs for the specified logical model and
// a rounded floating-point application/storage ratio. Provisional means at least
// one used capacity is estimated or its encoder is not declared supported.
// Even a nonprovisional result does not verify backend availability or security.
// Returned evidence is copied; changes do not alter the supplied catalogue.
type Evaluation struct {
	DatabaseCount        uint64             `json:"database_count"`
	Costs                Costs              `json:"costs"`
	EffectiveStorageRate float64            `json:"effective_storage_rate"`
	SecurityBits         uint32             `json:"security_bits"`
	Log2Samples          uint32             `json:"log2_samples"`
	Provisional          bool               `json:"provisional"`
	Configurations       []ConfigurationUse `json:"configurations"`
}

// Evaluate validates a supplied plan and calculates its cost. It accepts known
// estimated capacities, rejects unknown ones, and does not mutate either input.
// All products and totals must fit uint64. Coverage checks use O(D^2) time for
// D explicit rectangles or compact row blocks. Allocations hold metadata and copied evidence;
// no per-application-row or per-bit allocation is performed.
func Evaluate(plan Plan, catalogue *pirconfig.Catalogue, requirements Requirements) (*Evaluation, error) {
	if err := catalogue.Validate(); err != nil {
		return nil, fmt.Errorf("planner: %w", err)
	}
	if plan.FormatVersion != 1 || plan.Rows == 0 || plan.RecordBits == 0 || (len(plan.Databases) == 0 && len(plan.RowBlocks) == 0) {
		return nil, fmt.Errorf("planner: require format_version 1 and nonempty application dimensions and databases")
	}
	configurations := make(map[string]*pirconfig.Configuration, len(catalogue.Configurations))
	maxWidth := 0
	for i := range catalogue.Configurations {
		c := &catalogue.Configurations[i]
		configurations[c.ID] = c
		if c.Code.S > maxWidth {
			maxWidth = c.Code.S
		}
	}
	if plan.RecordBits < uint64(maxWidth) {
		return nil, fmt.Errorf("planner: record_bits must be at least catalogue field width %d", maxWidth)
	}
	regions, err := planRegions(plan, configurations)
	if err != nil {
		return nil, err
	}
	applicationBits, err := multiply(plan.Rows, plan.RecordBits)
	if err != nil {
		return nil, fmt.Errorf("planner: application bits: %w", err)
	}
	result := &Evaluation{Costs: Costs{ApplicationBits: applicationBits}}
	used := make(map[string]bool)
	groups := []string{}
	var covered uint64
	for i, region := range regions {
		p := region.bounds
		if p.RowStart >= p.RowEnd || p.RowEnd > plan.Rows || p.ColumnStart >= p.ColumnEnd || p.ColumnEnd > plan.RecordBits {
			return nil, fmt.Errorf("planner: database %d: rectangle is empty or out of bounds", i)
		}
		if i > 0 && !less(regions[i-1].bounds, p) {
			return nil, fmt.Errorf("planner: database %d: coordinates must be unique and in canonical order", i)
		}
		c := configurations[p.ConfigurationID]
		if c == nil {
			return nil, fmt.Errorf("planner: database %d: unknown configuration %q", i, p.ConfigurationID)
		}
		if c.Capacity.Status == "unknown" {
			return nil, fmt.Errorf("planner: configuration %s has unknown capacity", c.ID)
		}
		if p.RowEnd-p.RowStart > *c.Capacity.K || region.sliceWidth > uint64(c.Code.S) {
			return nil, fmt.Errorf("planner: database %d: rectangle exceeds configuration capacity or field width", i)
		}
		if c.Security.ComputationalBits < requirements.MinSecurityBits || c.Security.Log2Samples < requirements.Log2Samples {
			return nil, fmt.Errorf("planner: configuration %s does not meet the requested security profile", c.ID)
		}
		if p.Group > uint64(len(groups)) {
			return nil, fmt.Errorf("planner: query-sharing groups must be numbered in first-occurrence order")
		}
		if p.Group == uint64(len(groups)) {
			groups = append(groups, c.ID)
		}
		if groups[int(p.Group)] != c.ID {
			return nil, fmt.Errorf("planner: group %d requires the same catalogue configuration", p.Group)
		}
		for _, previous := range regions[:i] {
			other := previous.bounds
			rowOverlap := p.RowStart < other.RowEnd && other.RowStart < p.RowEnd
			if rowOverlap && p.ColumnStart < other.ColumnEnd && other.ColumnStart < p.ColumnEnd {
				return nil, fmt.Errorf("planner: database %d overlaps another rectangle", i)
			}
			if p.Group == other.Group && rowOverlap && p.RowStart != other.RowStart {
				return nil, fmt.Errorf("planner: group %d has incompatible row ranges", p.Group)
			}
		}
		area, err := multiply(p.RowEnd-p.RowStart, p.ColumnEnd-p.ColumnStart)
		if err != nil {
			return nil, fmt.Errorf("planner: rectangle area: %w", err)
		}
		if err := accumulate(&covered, area); err != nil {
			return nil, fmt.Errorf("planner: covered area: %w", err)
		}
		blockCosts, err := databaseCosts(c, region.count)
		if err == nil {
			result.Costs, err = addModelCosts(result.Costs, blockCosts)
		}
		if err != nil {
			return nil, fmt.Errorf("planner: %w", err)
		}
		if err := accumulate(&result.DatabaseCount, region.count); err != nil {
			return nil, fmt.Errorf("planner: database count: %w", err)
		}

		if !used[c.ID] {
			used[c.ID] = true
			security := c.Security
			security.Evidence = copyEvidence(security.Evidence)
			result.Configurations = append(result.Configurations, ConfigurationUse{ID: c.ID, CapacityStatus: c.Capacity.Status, EncodingStatus: c.EncodingStatus, CapacityEvidence: copyEvidence(c.Capacity.Evidence), Security: security})
			if len(result.Configurations) == 1 || c.Security.ComputationalBits < result.SecurityBits {
				result.SecurityBits = c.Security.ComputationalBits
			}
			if len(result.Configurations) == 1 || c.Security.Log2Samples < result.Log2Samples {
				result.Log2Samples = c.Security.Log2Samples
			}
		}
		if c.Capacity.Status != "verified" || c.EncodingStatus != "supported" {
			result.Provisional = true
		}
	}
	// Bounded, pairwise-disjoint rectangles cover exactly once iff their total
	// area equals the application's area. Area alone would not detect overlaps.
	if covered != applicationBits {
		return nil, fmt.Errorf("planner: partition leaves application entries uncovered")
	}

	for _, limit := range []struct {
		value, max uint64
		name       string
	}{
		{result.Costs.StorageBits, requirements.MaxStorageBits, "storage"},
		{result.Costs.UploadBits, requirements.MaxUploadBits, "upload"},
		{result.Costs.DownloadBits, requirements.MaxDownloadBits, "download"},
	} {
		if limit.max != 0 && limit.value > limit.max {
			return nil, fmt.Errorf("planner: %s cost %d bits exceeds limit %d", limit.name, limit.value, limit.max)
		}
	}
	result.EffectiveStorageRate = float64(applicationBits) / float64(result.Costs.StorageBits)
	return result, nil
}

func multiply(a, b uint64) (uint64, error) {
	hi, lo := bits.Mul64(a, b)
	if hi != 0 {
		return 0, fmt.Errorf("uint64 overflow")
	}
	return lo, nil
}

func accumulate(total *uint64, value uint64) error {
	sum, carry := bits.Add64(*total, value, 0)
	if carry != 0 {
		return fmt.Errorf("uint64 overflow")
	}
	*total = sum
	return nil
}

func less(a, b Database) bool {
	x := [4]uint64{a.RowStart, a.ColumnStart, a.RowEnd, a.ColumnEnd}
	y := [4]uint64{b.RowStart, b.ColumnStart, b.RowEnd, b.ColumnEnd}
	for i := range x {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	return false
}

func copyEvidence(evidence []pirconfig.Evidence) []pirconfig.Evidence {
	copied := append([]pirconfig.Evidence(nil), evidence...)
	for i := range copied {
		if copied[i].MonteCarlo != nil {
			mc := *copied[i].MonteCarlo
			copied[i].MonteCarlo = &mc
		}
	}
	return copied
}

// A region spans either one explicit rectangle or all column slices of a block.
// Geometry uses the complete span; sliceWidth validates each represented database.
type layoutRegion struct {
	bounds     Database
	count      uint64
	sliceWidth uint64
}

func planRegions(plan Plan, configurations map[string]*pirconfig.Configuration) ([]layoutRegion, error) {
	if len(plan.Databases) > 0 && len(plan.RowBlocks) > 0 {
		return nil, fmt.Errorf("planner: choose exactly one of databases and row_blocks")
	}
	regions := make([]layoutRegion, 0, len(plan.Databases)+len(plan.RowBlocks))
	for _, d := range plan.Databases {
		regions = append(regions, layoutRegion{bounds: d, count: 1, sliceWidth: d.ColumnEnd - d.ColumnStart})
	}
	for _, b := range plan.RowBlocks {
		c := configurations[b.ConfigurationID]
		if c == nil {
			return nil, fmt.Errorf("planner: unknown configuration %q", b.ConfigurationID)
		}
		width := uint64(c.Code.S)
		count := plan.RecordBits / width
		if plan.RecordBits%width != 0 {
			count++
		}
		regions = append(regions, layoutRegion{bounds: Database{RowStart: b.RowStart, RowEnd: b.RowEnd, ColumnEnd: plan.RecordBits, ConfigurationID: b.ConfigurationID, Group: b.Group}, count: count, sliceWidth: width})
	}
	return regions, nil
}
