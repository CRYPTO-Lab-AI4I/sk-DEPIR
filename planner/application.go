package planner

import (
	"fmt"
	"math"
	"sort"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

// ApplicationPlan splits every application record into fixed-width logical
// records and stacks them in record-major order. The last fragment is zero-padded.
// Original record j, fragment i maps to logical row j*FragmentCount()+i. LogicalPlan
// partitions those rows; it does not directly address original application rows.
type ApplicationPlan struct {
	Rows              uint64 `json:"rows"`
	RecordBits        uint64 `json:"record_bits"`
	LogicalRecordBits uint64 `json:"logical_record_bits"`
	LogicalPlan       Plan   `json:"logical_plan"`
}

// ApplicationOptions controls width search and optional identity-storage pieces.
// Empty LogicalRecordBits searches the original width, all field widths, and
// powers of two from 8 bits up to the original width. Explicit widths must be
// positive and at most RecordBits; the original width is always included.
// Full-download pieces are opt-in. Their aggregate download is bounded by
// FullDownloadRatio times indexed download in the selected plan (default 1).
// For a full-only plan, use the indexed-only largest-fitting baseline for that width.
// Additionally, cap against the lowest indexed download among no-full greedy
// plans across all searched widths, including the original width.
// MaxFullDownloadBits optionally adds an absolute cap; ordinary budgets also apply.
type ApplicationOptions struct {
	LogicalRecordBits   []uint64 `json:"logical_record_bits,omitempty"`
	AllowFullDownload   bool     `json:"allow_full_download"`
	FullDownloadRatio   float64  `json:"full_download_ratio,omitempty"`
	MaxFullDownloadBits uint64   `json:"max_full_download_bits,omitempty"`
}

// FragmentCount returns the fixed number of logical records per application
// record. That many indexed queries are sent to every used PIR query-sharing
// group per application retrieval; full-download pieces are fetched once instead.
func (p ApplicationPlan) FragmentCount() (uint64, error) {
	if p.Rows == 0 || p.RecordBits == 0 || p.LogicalRecordBits == 0 || p.LogicalRecordBits > p.RecordBits {
		return 0, fmt.Errorf("planner: invalid application or logical record dimensions")
	}
	n := p.RecordBits / p.LogicalRecordBits
	if p.RecordBits%p.LogicalRecordBits != 0 {
		n++
	}
	return n, nil
}

// Fragment maps a source record/fragment to its logical row and half-open source
// bit range. Padding beyond the returned bit range is zero. No data is copied.
func (p ApplicationPlan) Fragment(record, fragment uint64) (logicalRow, startBit, endBit uint64, err error) {
	n, err := p.FragmentCount()
	if err != nil {
		return 0, 0, 0, err
	}
	total, err := multiply(p.Rows, n)
	if err != nil {
		return 0, 0, 0, err
	}
	if record >= p.Rows || fragment >= n {
		return 0, 0, 0, fmt.Errorf("planner: fragment index out of range")
	}
	row := record*n + fragment
	if row >= total {
		return 0, 0, 0, fmt.Errorf("planner: fragment row overflow")
	}
	start := fragment * p.LogicalRecordBits
	end := start + minWidth(p.LogicalRecordBits, p.RecordBits-start)
	return row, start, end, nil
}

// EvaluateApplication checks the split/stack mapping and evaluates traffic for
// one original application-record retrieval. Its ApplicationBits and storage rate
// use the original size; per-fragment padding remains included in stored data.
// Requirements.Log2Samples remains a protocol-query budget, not an application
// retrieval budget: each used PIR group consumes FragmentCount samples per lookup.
func EvaluateApplication(plan ApplicationPlan, catalogue *pirconfig.Catalogue, r Requirements) (*Evaluation, error) {
	fragments, err := plan.FragmentCount()
	if err != nil {
		return nil, err
	}
	rows, err := multiply(plan.Rows, fragments)
	if err != nil {
		return nil, fmt.Errorf("planner: logical rows: %w", err)
	}
	if plan.LogicalPlan.Rows != rows || plan.LogicalPlan.RecordBits != plan.LogicalRecordBits {
		return nil, fmt.Errorf("planner: logical plan dimensions do not match application mapping")
	}
	result, err := evaluate(plan.LogicalPlan, catalogue, r, false, fragments)
	if err != nil {
		return nil, err
	}
	if !result.FullDownloadOnly {
		for _, c := range catalogue.Configurations {
			if plan.RecordBits < uint64(c.Code.S) {
				return nil, fmt.Errorf("planner: original record_bits must be at least catalogue field width %d", c.Code.S)
			}
		}
	}
	applicationBits, err := multiply(plan.Rows, plan.RecordBits)
	if err != nil {
		return nil, fmt.Errorf("planner: application bits: %w", err)
	}
	result.CommunicationRate = float64(plan.RecordBits) / float64(result.Costs.CommunicationBits)
	result.Costs.ApplicationBits = applicationBits
	result.EffectiveStorageRate = float64(applicationBits) / float64(result.Costs.StorageBits)
	return result, nil
}

// GreedyApplication searches logical widths around the existing greedy planner.
// Selection minimizes storage then communication, using greater logical width for
// deterministic exact-cost ties. Indexed traffic is charged for every fragment;
// full-download suffixes are charged once per application retrieval. The original
// Greedy API is unchanged and continues to construct unsliced indexed plans.
func GreedyApplication(rows, recordBits uint64, catalogue *pirconfig.Catalogue, r Requirements, options ApplicationOptions) (ApplicationPlan, *Evaluation, error) {
	if err := catalogue.Validate(); err != nil {
		return ApplicationPlan{}, nil, err
	}
	if rows == 0 || recordBits == 0 {
		return ApplicationPlan{}, nil, fmt.Errorf("planner: application dimensions must be nonempty")
	}
	if _, err := multiply(rows, recordBits); err != nil {
		return ApplicationPlan{}, nil, err
	}
	originalTooNarrow := false
	for _, c := range catalogue.Configurations {
		if recordBits < uint64(c.Code.S) {
			originalTooNarrow = true
		}
	}
	if originalTooNarrow && !options.AllowFullDownload {
		return ApplicationPlan{}, nil, fmt.Errorf("planner: original records are narrower than the catalogue field widths")
	}
	ratio := options.FullDownloadRatio
	if ratio == 0 {
		ratio = 1
	}
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) || ratio < 0 {
		return ApplicationPlan{}, nil, fmt.Errorf("planner: full_download_ratio must be finite and positive (zero selects default 1)")
	}
	widths, err := logicalWidths(recordBits, catalogue, options.LogicalRecordBits)
	if err != nil {
		return ApplicationPlan{}, nil, err
	}
	limit := uint64(0)
	if options.AllowFullDownload {
		limit = options.MaxFullDownloadBits
		if limit == 0 {
			limit = math.MaxUint64
		}
	}
	// Establish one reference before considering identity storage. A width must
	// not inflate its own indexed traffic to qualify a large whole download.
	// Use security-qualified baselines without cost budgets: even a baseline that
	// exceeds storage limits remains a valid communication reference.
	if options.AllowFullDownload {
		referenceRequirements := r
		referenceRequirements.MaxStorageBits = 0
		referenceRequirements.MaxUploadBits = 0
		referenceRequirements.MaxDownloadBits = 0
		var reference uint64
		found := false
		for _, width := range widths {
			mapping := ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: width}
			fragments, _ := mapping.FragmentCount()
			logicalRows, err := multiply(rows, fragments)
			if err != nil {
				continue
			}
			_, baseline, err := greedy(logicalRows, width, catalogue, referenceRequirements, false, fragments, 0, 1)
			if err != nil {
				continue
			}
			if !found || baseline.Costs.DownloadBits < reference {
				reference = baseline.Costs.DownloadBits
				found = true
			}
		}
		if !found {
			limit = 0
		} else if cap := float64(reference) * ratio; cap < float64(limit) {
			limit = uint64(cap)
		}
	}
	var best ApplicationPlan
	var bestEvaluation *Evaluation
	for _, width := range widths {
		candidate := ApplicationPlan{Rows: rows, RecordBits: recordBits, LogicalRecordBits: width}
		fragments, _ := candidate.FragmentCount()
		logicalRows, err := multiply(rows, fragments)
		if err != nil {
			continue
		}
		layout, _, err := greedy(logicalRows, width, catalogue, r, originalTooNarrow, fragments, limit, ratio)
		if err != nil {
			continue
		}
		candidate.LogicalPlan = layout
		evaluation, err := EvaluateApplication(candidate, catalogue, r)
		if err != nil {
			continue
		}
		if bestEvaluation == nil || cheaper(evaluation.Costs, bestEvaluation.Costs) || (!cheaper(bestEvaluation.Costs, evaluation.Costs) && width > best.LogicalRecordBits) {
			best = candidate
			bestEvaluation = evaluation
		}
	}
	if bestEvaluation == nil {
		return ApplicationPlan{}, nil, ErrNoPlan
	}
	return best, bestEvaluation, nil
}

func logicalWidths(recordBits uint64, c *pirconfig.Catalogue, requested []uint64) ([]uint64, error) {
	seen := map[uint64]bool{recordBits: true}
	if len(requested) > 0 {
		for _, width := range requested {
			if width == 0 || width > recordBits {
				return nil, fmt.Errorf("planner: logical width must be in [1,record_bits]")
			}
			seen[width] = true
		}
	} else {
		for _, entry := range c.Configurations {
			width := uint64(entry.Code.S)
			if width <= recordBits {
				seen[width] = true
			}
		}
		for width := uint64(8); width <= recordBits; {
			seen[width] = true
			if width > math.MaxUint64/2 {
				break
			}
			width *= 2
		}
	}
	widths := make([]uint64, 0, len(seen))
	for width := range seen {
		widths = append(widths, width)
	}
	sort.Slice(widths, func(i, j int) bool { return widths[i] < widths[j] })
	return widths, nil
}
