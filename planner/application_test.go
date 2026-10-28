package planner

import (
	"bytes"
	"context"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/appref"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
)

func TestSplitStackStorageAndMapping(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	a, ea, err := GreedyApplication(8, 4, c, Requirements{}, ApplicationOptions{LogicalRecordBits: []uint64{2}})
	if err != nil {
		t.Fatal(err)
	}
	b, eb, err := GreedyApplication(1, 32, c, Requirements{}, ApplicationOptions{LogicalRecordBits: []uint64{2}})
	if err != nil {
		t.Fatal(err)
	}
	_, old, err := Greedy(1, 32, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if ea.Costs.ApplicationBits != 32 || eb.Costs.ApplicationBits != 32 || ea.Costs.StorageBits != eb.Costs.StorageBits || eb.Costs.StorageBits >= old.Costs.StorageBits {
		t.Fatal("slicing did not remove artificial storage penalty")
	}
	if a.RecordBits != 4 || b.RecordBits != 32 {
		t.Fatal("original dimensions were lost")
	}
	p := ApplicationPlan{Rows: 2, RecordBits: 9, LogicalRecordBits: 4}
	count, err := p.FragmentCount()
	if err != nil || count != 3 {
		t.Fatal("wrong fragment count")
	}
	for j := uint64(0); j < 2; j++ {
		for i := uint64(0); i < 3; i++ {
			row, start, end, err := p.Fragment(j, i)
			wantEnd := (i + 1) * 4
			if wantEnd > 9 {
				wantEnd = 9
			}
			if err != nil || row != j*3+i || start != i*4 || end != wantEnd {
				t.Fatalf("mapping %d,%d: %d,%d,%d,%v", j, i, row, start, end, err)
			}
		}
	}
	if _, _, _, err := p.Fragment(2, 0); err == nil {
		t.Fatal("out-of-range source record accepted")
	}
	if _, _, _, err := p.Fragment(0, 3); err == nil {
		t.Fatal("out-of-range fragment accepted")
	}
}

func TestSplitStackActualReferenceRecovery(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	p, e, err := GreedyApplication(2, 9, c, Requirements{}, ApplicationOptions{LogicalRecordBits: []uint64{4}})
	if err != nil {
		t.Fatal(err)
	}
	if p.LogicalRecordBits != 4 {
		t.Fatalf("unexpected width %d", p.LogicalRecordBits)
	}
	sources := [][]byte{{0xa5, 1}, {0x36, 0}}
	fragments, _ := p.FragmentCount()
	logical := make([][]byte, p.LogicalPlan.Rows)
	for j := uint64(0); j < p.Rows; j++ {
		for i := uint64(0); i < fragments; i++ {
			row, start, end, err := p.Fragment(j, i)
			if err != nil {
				t.Fatal(err)
			}
			logical[row] = []byte{0}
			for bit := start; bit < end; bit++ {
				logical[row][0] |= ((sources[j][bit/8] >> uint(bit%8)) & 1) << uint(bit-start)
			}
		}
	}
	expanded := expandSmall(t, p.LogicalPlan, c)
	reference := appref.Plan{Rows: expanded.Rows, RecordBits: expanded.RecordBits}
	for _, d := range expanded.Databases {
		reference.Databases = append(reference.Databases, appref.Partition{Coordinates: appref.Coordinates{RowStart: d.RowStart, RowEnd: d.RowEnd, ColumnStart: d.ColumnStart, ColumnEnd: d.ColumnEnd}, Code: liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2}, Group: d.Group})
	}
	client, stored, err := appref.Encode(reference, logical)
	if err != nil {
		t.Fatal(err)
	}
	for j := uint64(0); j < p.Rows; j++ {
		restored := []byte{0, 0}
		var calls uint64
		for i := uint64(0); i < fragments; i++ {
			row, start, end, _ := p.Fragment(j, i)
			piece, err := client.Retrieve(context.Background(), row, func(_ context.Context, id appref.Coordinates, dst []byte, offset int64) (int, error) {
				calls++
				for _, d := range stored {
					if d.Coordinates == id {
						return bytes.NewReader(d.Data).ReadAt(dst, offset)
					}
				}
				t.Fatal("unknown reference database")
				return 0, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			for bit := start; bit < end; bit++ {
				restored[bit/8] |= ((piece[0] >> uint(bit-start)) & 1) << uint(bit%8)
			}
		}
		if !bytes.Equal(restored, sources[j]) || calls*2 != e.Costs.DownloadBits || calls*4 != e.Costs.UploadBits {
			t.Fatal("application recovery or per-fragment accounting disagrees")
		}
	}
	if e.Costs.ApplicationBits != 18 {
		t.Fatal("padding counted as application input")
	}
}

func TestFullDownloadEvaluationAndFragmentReuse(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	full := Plan{FormatVersion: 1, Rows: 3, RecordBits: 4, RowBlocks: []RowBlock{{DownloadAll: true, RowEnd: 3}}}
	e, err := Evaluate(full, c, Requirements{MinSecurityBits: 100, Log2Samples: 50})
	if err != nil {
		t.Fatal(err)
	}
	if !e.FullDownloadOnly || e.Provisional || e.DatabaseCount != 1 || e.FullDownloadBits != 12 || e.Costs.StorageBits != 12 || e.Costs.UploadBits != 0 || e.Costs.DownloadBits != 12 {
		t.Fatalf("wrong identity costs %+v", e)
	}
	mixed := ApplicationPlan{Rows: 1, RecordBits: 12, LogicalRecordBits: 4, LogicalPlan: Plan{FormatVersion: 1, Rows: 3, RecordBits: 4, Databases: []Database{
		{RowEnd: 3, ColumnEnd: 2, ConfigurationID: "fixture-s2"},
		{DownloadAll: true, ColumnStart: 2, RowEnd: 3, ColumnEnd: 4},
	}}}
	e, err = EvaluateApplication(mixed, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	// Three indexed lookups, but the same six-bit full piece is fetched just once.
	if e.Costs.StorageBits != 38 || e.Costs.UploadBits != 1572 || e.Costs.DownloadBits != 792 || e.FullDownloadBits != 6 || e.FullDownloadOnly {
		t.Fatalf("full piece repeated or PIR traffic undercounted: %+v", e)
	}
	if _, err := EvaluateApplication(mixed, c, Requirements{MaxDownloadBits: 791}); err == nil {
		t.Fatal("application-level budget ignored")
	}
	for _, change := range []func(*Plan){func(p *Plan) { p.RowBlocks[0].Group = 1 }, func(p *Plan) { p.RowBlocks[0].ConfigurationID = "fixture-s2" }} {
		bad := full
		bad.RowBlocks = append([]RowBlock(nil), full.RowBlocks...)
		change(&bad)
		if _, err := Evaluate(bad, c, Requirements{}); err == nil {
			t.Fatal("full piece with indexed metadata accepted")
		}
	}
}

func TestFullDownloadSuffixAndBudgets(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	options := ApplicationOptions{LogicalRecordBits: []uint64{4}, AllowFullDownload: true, MaxFullDownloadBits: 4}
	p, e, err := GreedyApplication(6, 4, c, Requirements{}, options)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.LogicalPlan.RowBlocks) != 2 || !p.LogicalPlan.RowBlocks[1].DownloadAll || e.Costs.StorageBits != 68 || e.Costs.DownloadBits != 528 || e.FullDownloadBits != 4 {
		t.Fatalf("small residual was not downloaded: %+v %+v", p, e)
	}
	options.MaxFullDownloadBits = 0
	p, e, err = GreedyApplication(6, 4, c, Requirements{}, options)
	if err != nil {
		t.Fatal(err)
	}
	if !e.FullDownloadOnly || e.Costs.StorageBits != 24 || e.Costs.DownloadBits != 24 || e.Costs.UploadBits != 0 {
		t.Fatal("unrestricted identity-storage plan not selected")
	}
	// Remove fixture dummies so a larger indexed database can communicate less
	// than a full download. The ordinary download budget must constrain choices.
	c.Configurations[0].Query.DummyPositions = 0
	c.Configurations[0].Query.Length = 3
	options.LogicalRecordBits = []uint64{4}
	_, e, err = GreedyApplication(1000, 4, c, Requirements{MaxDownloadBits: 3000}, options)
	if err != nil {
		t.Fatal(err)
	}
	if e.FullDownloadOnly || e.Costs.DownloadBits > 3000 || e.FullDownloadBits == 0 {
		t.Fatal("download budget was not enforced while using a suffix")
	}
	if _, _, err := GreedyApplication(1000, 4, c, Requirements{MaxDownloadBits: 1}, options); !errors.Is(err, ErrNoPlan) {
		t.Fatal("impossible budget accepted", err)
	}
}

func TestApplicationValidationAndBaselinePreservation(t *testing.T) {
	c := fixtureCatalogue(t, 2, 3)
	original, originalEval, err := Greedy(7, 12, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	p, e, err := GreedyApplication(7, 12, c, Requirements{}, ApplicationOptions{LogicalRecordBits: []uint64{12}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.LogicalPlan, original) || !reflect.DeepEqual(e, originalEval) {
		t.Fatal("unsliced application changed existing behavior")
	}
	for _, widths := range [][]uint64{{0}, {13}} {
		if _, _, err := GreedyApplication(7, 12, c, Requirements{}, ApplicationOptions{LogicalRecordBits: widths}); err == nil {
			t.Fatal("invalid width accepted")
		}
	}
	malformed := p
	malformed.LogicalPlan.Rows++
	if _, err := EvaluateApplication(malformed, c, Requirements{}); err == nil {
		t.Fatal("inconsistent logical rows accepted")
	}
	malformed = p
	malformed.LogicalPlan.RecordBits++
	if _, err := EvaluateApplication(malformed, c, Requirements{}); err == nil {
		t.Fatal("inconsistent logical width accepted")
	}
	malformed = p
	malformed.LogicalRecordBits = 0
	if _, err := EvaluateApplication(malformed, c, Requirements{}); err == nil {
		t.Fatal("zero logical width accepted")
	}
	if _, _, err := GreedyApplication(1, 12, nil, Requirements{}, ApplicationOptions{}); err == nil {
		t.Fatal("nil catalogue accepted")
	}
	if _, _, err := GreedyApplication(math.MaxUint64, 12, c, Requirements{}, ApplicationOptions{}); err == nil {
		t.Fatal("application overflow accepted")
	}
	if _, _, err := GreedyApplication(1, 1, c, Requirements{}, ApplicationOptions{}); err == nil {
		t.Fatal("undersized original record accepted")
	}
	_, e, err = GreedyApplication(1, 1, c, Requirements{}, ApplicationOptions{AllowFullDownload: true})
	if err != nil || !e.FullDownloadOnly {
		t.Fatal("tiny full-download database rejected", err)
	}
}

// The ratio bounds full downloads against indexed traffic, not database size.
func TestFullDownloadRatio(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	c.Configurations[0].Query.DummyPositions = 0
	c.Configurations[0].Query.Length = 3
	for _, ratio := range []float64{0.25, 1, 2} {
		_, e, err := GreedyApplication(1000, 4, c, Requirements{}, ApplicationOptions{AllowFullDownload: true, FullDownloadRatio: ratio, LogicalRecordBits: []uint64{4}})
		if err != nil {
			t.Fatal(err)
		}
		indexed := e.Costs.DownloadBits - e.FullDownloadBits
		if e.FullDownloadOnly {
			_, baseline, err := Greedy(1000, 4, c, Requirements{})
			if err != nil {
				t.Fatal(err)
			}
			indexed = baseline.Costs.DownloadBits
		}
		if float64(e.FullDownloadBits) > ratio*float64(indexed) {
			t.Fatalf("ratio %v violated: %+v", ratio, e)
		}
	}
	for _, ratio := range []float64{-1, math.NaN(), math.Inf(1)} {
		if _, _, err := GreedyApplication(10, 4, c, Requirements{}, ApplicationOptions{FullDownloadRatio: ratio}); err == nil {
			t.Fatal("invalid ratio accepted")
		}
	}
}

// Changing logical width must not inflate the baseline used to admit full pieces.
func TestFullDownloadUsesSharedWidthBaseline(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	c.Configurations[0].Query.DummyPositions = 0
	c.Configurations[0].Query.Length = 3
	const rows, bits = uint64(20), uint64(40)
	_, original, err := Greedy(rows, bits, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	// Capture both fractional and expanded limits, with and without budgets.
	for _, ratio := range []float64{0.25, 1, 2} {
		for _, r := range []Requirements{{}, {MaxStorageBits: 2000000}} {
			_, e, err := GreedyApplication(rows, bits, c, r, ApplicationOptions{AllowFullDownload: true, FullDownloadRatio: ratio, LogicalRecordBits: []uint64{2, 4}})
			if err != nil {
				t.Fatal(err)
			}
			if float64(e.FullDownloadBits) > ratio*float64(original.Costs.DownloadBits) {
				t.Fatalf("width inflated reference: %+v", e)
			}
			if ratio == 1 && e.FullDownloadOnly {
				t.Fatal("inflated baseline admitted full download")
			}
		}
	}
}
