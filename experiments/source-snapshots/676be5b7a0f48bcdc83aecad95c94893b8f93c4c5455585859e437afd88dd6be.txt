package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/appref"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

func fixtureCatalogue(t *testing.T, degrees ...int) *pirconfig.Catalogue {
	t.Helper()
	c := &pirconfig.Catalogue{Version: 1, Construction: "curve-lifted-rs-qminus1"}
	for _, s := range degrees {
		code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: s, Dimension: 2, CurveDegree: 2})
		if err != nil {
			t.Fatal(err)
		}
		k, n, p := uint64(code.Dimension()), uint64(code.Length()), code.Polynomial()
		rate := float64(k) / float64(n)
		evidence := []pirconfig.Evidence{{Kind: "reference_test", Reference: "planner/evaluate_test.go"}}
		c.Configurations = append(c.Configurations, pirconfig.Configuration{
			ID: fmt.Sprintf("fixture-s%d", s), EncodingStatus: "supported", Code: pirconfig.Code{S: s, M: 2, T: 2, Polynomial: &p},
			Capacity: pirconfig.Capacity{Status: "verified", K: &k, N: n, Rate: &rate, Evidence: evidence},
			Query:    pirconfig.Query{CurvePositions: 1<<uint(s) - 1, DummyPositions: 128, Length: 1<<uint(s) + 127, DummyRule: "fixed_count_uniform_positions"},
			Security: pirconfig.Security{Assumptions: "Synthetic correctness fixture, no security claim", Evidence: evidence},
			Encoder:  &pirconfig.Encoder{Implementation: "encoding/liftrs-go", Preprocessing: "dense systematic reference", Evidence: evidence},
		})
	}
	return c
}

func fixturePlan() Plan {
	return Plan{FormatVersion: 1, Rows: 3, RecordBits: 4, Databases: []Database{
		{RowStart: 0, ColumnStart: 0, RowEnd: 3, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 0, ColumnStart: 2, RowEnd: 3, ColumnEnd: 4, ConfigurationID: "fixture-s2", Group: 0},
	}}
}

func TestHandCalculatedCostsAndBudgets(t *testing.T) {
	c, p := fixtureCatalogue(t, 2), fixturePlan()
	got, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	// Two N=16,s=2 databases, each queried at 3+128 positions, each with
	// its own 4-bit address list even though query generation is shared.
	want := Costs{ApplicationBits: 12, StorageBits: 64, UploadBits: 1048, DownloadBits: 524, CommunicationBits: 1572}
	if got.Costs != want || got.EffectiveStorageRate != .1875 || got.Provisional || len(got.Configurations) != 1 {
		t.Fatalf("got %+v", got)
	}
	limits := Requirements{MaxStorageBits: 64, MaxUploadBits: 1048, MaxDownloadBits: 524}
	if _, err := Evaluate(p, c, limits); err != nil {
		t.Fatal("exact budget rejected", err)
	}
	for _, r := range []Requirements{{MaxStorageBits: 63}, {MaxUploadBits: 1047}, {MaxDownloadBits: 523}} {
		if _, err := Evaluate(p, c, r); err == nil {
			t.Fatal("over-budget layout accepted")
		}
	}
	p.Databases[1].Group = 1
	separate, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if separate.Costs != got.Costs {
		t.Fatal("query sharing incorrectly changes communication")
	}
}

func TestMixedFieldsAndSecurity(t *testing.T) {
	c, p := fixtureCatalogue(t, 2, 3), fixturePlan()
	p.Databases[1].ConfigurationID = "fixture-s3"
	p.Databases[1].Group = 1
	// Synthetic profile values test constraint comparison, not real security.
	c.Configurations[0].Security.ComputationalBits = 128
	c.Configurations[0].Security.Log2Samples = 60
	c.Configurations[1].Security.ComputationalBits = 96
	c.Configurations[1].Security.Log2Samples = 50
	got, err := Evaluate(p, c, Requirements{MinSecurityBits: 96, Log2Samples: 50})
	if err != nil {
		t.Fatal(err)
	}
	want := Costs{ApplicationBits: 12, StorageBits: 224, UploadBits: 1334, DownloadBits: 667, CommunicationBits: 2001}
	if got.Costs != want || got.SecurityBits != 96 || got.Log2Samples != 50 {
		t.Fatalf("got %+v", got)
	}
	for _, r := range []Requirements{{MinSecurityBits: 97}, {Log2Samples: 51}} {
		if _, err := Evaluate(p, c, r); err == nil {
			t.Fatal("insufficient profile accepted")
		}
	}
	p.Databases[1].Group = 0
	if _, err := Evaluate(p, c, Requirements{}); err == nil {
		t.Fatal("different configurations shared a group")
	}
}

func TestReferenceStorageAndReads(t *testing.T) {
	p := fixturePlan()
	c := fixtureCatalogue(t, 2)
	result, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	reference := appref.Plan{Rows: p.Rows, RecordBits: p.RecordBits}
	for _, d := range p.Databases {
		reference.Databases = append(reference.Databases, appref.Partition{Coordinates: appref.Coordinates{RowStart: d.RowStart, ColumnStart: d.ColumnStart, RowEnd: d.RowEnd, ColumnEnd: d.ColumnEnd}, Code: liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2}, Group: d.Group})
	}
	records := [][]byte{{6}, {11}, {13}}
	client, stored, err := appref.Encode(reference, records)
	if err != nil {
		t.Fatal(err)
	}
	var storedBits uint64
	for _, d := range stored {
		storedBits += uint64(len(d.Data)) * 8
	}
	if storedBits != result.Costs.StorageBits {
		t.Fatal("actual packed storage disagrees with model")
	}
	for i, want := range records {
		var calls uint64
		got, err := client.Retrieve(context.Background(), uint64(i), func(_ context.Context, id appref.Coordinates, dst []byte, offset int64) (int, error) {
			calls++
			for _, d := range stored {
				if d.Coordinates == id {
					return bytes.NewReader(d.Data).ReadAt(dst, offset)
				}
			}
			return 0, fmt.Errorf("unexpected database")
		})
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("retrieval %d: %v, %v", i, got, err)
		}
		// The adapter transfers aligned bytes; model D counts the two useful field
		// bits per response. Count logical addresses, not actual backend wire bytes.
		if calls != 262 || calls*2 != result.Costs.DownloadBits || calls*4 != result.Costs.UploadBits {
			t.Fatal("read count disagrees with query model")
		}
	}
}

func TestGeneralPartitionAndGroupRows(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	p := Plan{FormatVersion: 1, Rows: 4, RecordBits: 4, Databases: []Database{
		{RowStart: 0, ColumnStart: 0, RowEnd: 4, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 0, ColumnStart: 2, RowEnd: 2, ColumnEnd: 4, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 2, ColumnStart: 2, RowEnd: 4, ColumnEnd: 4, ConfigurationID: "fixture-s2", Group: 1},
	}}
	if _, err := Evaluate(p, c, Requirements{}); err != nil {
		t.Fatal("same-start/different-end rows rejected", err)
	}
	disjoint := Plan{FormatVersion: 1, Rows: 6, RecordBits: 2, Databases: []Database{
		{RowStart: 0, ColumnStart: 0, RowEnd: 3, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 3, ColumnStart: 0, RowEnd: 6, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 0},
	}}
	if _, err := Evaluate(disjoint, c, Requirements{}); err != nil {
		t.Fatal("disjoint row ranges rejected", err)
	}
	// A valid general partition whose first and third rectangles cannot share
	// queries because their overlapping rows would have different local indices.
	incompatible := Plan{FormatVersion: 1, Rows: 4, RecordBits: 4, Databases: []Database{
		{RowStart: 0, ColumnStart: 0, RowEnd: 3, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 0, ColumnStart: 2, RowEnd: 1, ColumnEnd: 4, ConfigurationID: "fixture-s2", Group: 1},
		{RowStart: 1, ColumnStart: 2, RowEnd: 4, ColumnEnd: 4, ConfigurationID: "fixture-s2", Group: 0},
		{RowStart: 3, ColumnStart: 0, RowEnd: 4, ColumnEnd: 2, ConfigurationID: "fixture-s2", Group: 1},
	}}
	if _, err := Evaluate(incompatible, c, Requirements{}); err == nil || !strings.Contains(err.Error(), "incompatible row") {
		t.Fatal("incompatible shared indices accepted", err)
	}
	incompatible.Databases[2].Group = 2
	if _, err := Evaluate(incompatible, c, Requirements{}); err != nil {
		t.Fatal("valid nonaligned partition rejected", err)
	}
}

func TestInvalidPlans(t *testing.T) {
	tests := []struct {
		name, want string
		change     func(*Plan)
	}{
		{"version", "format_version", func(p *Plan) { p.FormatVersion = 2 }},
		{"empty rows", "nonempty", func(p *Plan) { p.Rows = 0 }},
		{"empty bits", "nonempty", func(p *Plan) { p.RecordBits = 0 }},
		{"empty databases", "nonempty", func(p *Plan) { p.Databases = nil }},
		{"record too small", "record_bits", func(p *Plan) { p.RecordBits = 1 }},
		{"empty row range", "rectangle", func(p *Plan) { p.Databases[0].RowEnd = 0 }},
		{"row bounds", "rectangle", func(p *Plan) { p.Databases[0].RowEnd = 4 }},
		{"column bounds", "rectangle", func(p *Plan) { p.Databases[0].ColumnEnd = 5 }},
		{"empty columns", "rectangle", func(p *Plan) { p.Databases[0].ColumnEnd = 0 }},
		{"order", "canonical", func(p *Plan) { p.Databases[0], p.Databases[1] = p.Databases[1], p.Databases[0] }},
		{"duplicate", "canonical", func(p *Plan) { p.Databases[1] = p.Databases[0] }},
		{"overlap with equal total area", "overlap", func(p *Plan) { p.Databases[1].ColumnStart = 1; p.Databases[1].ColumnEnd = 3 }},
		{"gap", "uncovered", func(p *Plan) { p.Databases[1].ColumnStart = 3 }},
		{"too many rows", "capacity", func(p *Plan) { p.Rows = 6; p.Databases[0].RowEnd = 6; p.Databases[1].RowEnd = 6 }},
		{"field overflow", "capacity", func(p *Plan) { p.Databases[0].ColumnEnd = 3 }},
		{"unknown configuration", "unknown configuration", func(p *Plan) { p.Databases[0].ConfigurationID = "missing" }},
		{"skipped group", "first-occurrence", func(p *Plan) { p.Databases[0].Group = 1 }},
		{"huge group", "first-occurrence", func(p *Plan) { p.Databases[0].Group = math.MaxUint64 }},
		{"application overflow", "application bits", func(p *Plan) { p.Rows = math.MaxUint64 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := fixturePlan()
			tt.change(&p)
			_, err := Evaluate(p, fixtureCatalogue(t, 2), Requirements{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("wanted %q, got %v", tt.want, err)
			}
		})
	}
	if _, err := Evaluate(fixturePlan(), nil, Requirements{}); err == nil {
		t.Fatal("nil catalogue accepted")
	}
	c := fixtureCatalogue(t, 2)
	c.Version = 0
	if _, err := Evaluate(fixturePlan(), c, Requirements{}); err == nil {
		t.Fatal("invalid catalogue accepted")
	}
	c = fixtureCatalogue(t, 2, 4)
	p := fixturePlan()
	p.RecordBits = 3
	p.Databases[1].ColumnEnd = 3
	if _, err := Evaluate(p, c, Requirements{}); err == nil || !strings.Contains(err.Error(), "record_bits") {
		t.Fatal("largest supplied field width ignored")
	}
}

func TestCapacityStatusAndEvidenceOwnership(t *testing.T) {
	c := fixtureCatalogue(t, 2)
	entry := &c.Configurations[0]
	entry.EncodingStatus = "candidate"
	entry.Capacity.Status = "estimated"
	entry.Capacity.Evidence = []pirconfig.Evidence{{Kind: "monte_carlo", Reference: "synthetic-test/report.md", Script: "synthetic-test/run.py", Results: "synthetic-test/results.json", Method: "Synthetic metadata, no measurement claim", MonteCarlo: &pirconfig.MonteCarlo{Samples: 1000, Confidence: .95, LowerRate: .2, UpperRate: .4}}}
	p := fixturePlan()
	before, _ := json.Marshal(c)
	beforePlan, _ := json.Marshal(p)
	got, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Provisional || got.Configurations[0].CapacityStatus != "estimated" {
		t.Fatal("estimated plan not marked provisional")
	}
	got.Configurations[0].CapacityEvidence[0].MonteCarlo.LowerRate = 0
	got.Configurations[0].CapacityEvidence[0].Reference = "changed"
	got.Configurations[0].Security.Evidence[0].Reference = "changed"
	after, _ := json.Marshal(c)
	afterPlan, _ := json.Marshal(p)
	if !bytes.Equal(before, after) || !bytes.Equal(beforePlan, afterPlan) {
		t.Fatal("input mutated or returned evidence aliases it")
	}
	entry.Capacity.Status = "unknown"
	entry.Capacity.K = nil
	entry.Capacity.Rate = nil
	entry.Capacity.Evidence = nil
	if _, err := Evaluate(p, c, Requirements{}); err == nil || !strings.Contains(err.Error(), "unknown capacity") {
		t.Fatal("unknown capacity accepted", err)
	}
	c = fixtureCatalogue(t, 2)
	c.Configurations[0].EncodingStatus = "candidate"
	got, err = Evaluate(p, c, Requirements{})
	if err != nil || !got.Provisional {
		t.Fatal("unavailable encoder not marked provisional", err)
	}
}

func TestEncodingResourceMeasurementsAreInformational(t *testing.T) {
	c, p := fixtureCatalogue(t, 2), fixturePlan()
	before, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	c.Configurations[0].Measurements = []pirconfig.Measurement{{Stage: "database_encoding", Quantity: "elapsed_seconds", Value: 1e300, Evidence: pirconfig.Evidence{Kind: "measurement", Reference: "synthetic-test/report.md", Script: "synthetic-test/run.py", Results: "synthetic-test/results.json"}}}
	after, err := Evaluate(p, c, Requirements{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("encoding costs influenced evaluation")
	}
}

func TestTerabyteScaleCoordinates(t *testing.T) {
	f, err := os.Open("../pirconfig/configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := pirconfig.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	entry := c.Configurations[len(c.Configurations)-1]
	k := *entry.Capacity.K
	p := Plan{FormatVersion: 1, Rows: 2 * k, RecordBits: 20, Databases: []Database{
		{RowStart: 0, ColumnStart: 0, RowEnd: k, ColumnEnd: 20, ConfigurationID: entry.ID, Group: 0},
		{RowStart: k, ColumnStart: 0, RowEnd: 2 * k, ColumnEnd: 20, ConfigurationID: entry.ID, Group: 0},
	}}
	got, err := Evaluate(p, c, Requirements{MinSecurityBits: 100, Log2Samples: 50})
	if err != nil {
		t.Fatal(err)
	}
	if !got.Provisional || got.Costs.ApplicationBits != 2*k*20 || got.Costs.StorageBits != 2*(uint64(1)<<40)*20 || got.Costs.UploadBits != 2*entry.Query.Length*40 || got.Costs.DownloadBits != 2*entry.Query.Length*20 {
		t.Fatalf("wrong large-coordinate model: %+v", got)
	}
	if got.Costs.ApplicationBits < 8_000_000_000_000 {
		t.Fatal("fixture did not reach terabyte scale")
	}
}

func TestOverflowAndConcurrentEvaluation(t *testing.T) {
	if v, err := multiply(math.MaxUint64, 1); err != nil || v != math.MaxUint64 {
		t.Fatal("valid boundary rejected")
	}
	if _, err := multiply(math.MaxUint64, 2); err == nil {
		t.Fatal("multiplication wrapped")
	}
	total := uint64(math.MaxUint64)
	if err := accumulate(&total, 1); err == nil || total != math.MaxUint64 {
		t.Fatal("sum wrapped or changed on error")
	}
	c := fixtureCatalogue(t, 2)
	entry := &c.Configurations[0]
	entry.Query.Length = math.MaxUint64
	entry.Query.DummyPositions = math.MaxUint64 - 3
	if _, err := Evaluate(fixturePlan(), c, Requirements{}); err == nil || !strings.Contains(err.Error(), "upload bits") {
		t.Fatal("upload overflow accepted", err)
	}
	c = fixtureCatalogue(t, 2)
	entry = &c.Configurations[0]
	entry.Query.Length = math.MaxUint64/6 + 1
	entry.Query.DummyPositions = entry.Query.Length - 3
	one := Plan{FormatVersion: 1, Rows: 1, RecordBits: 2, Databases: []Database{{RowEnd: 1, ColumnEnd: 2, ConfigurationID: entry.ID}}}
	if _, err := Evaluate(one, c, Requirements{}); err == nil || !strings.Contains(err.Error(), "communication bits") {
		t.Fatal("communication overflow accepted", err)
	}
	c = fixtureCatalogue(t, 2)
	entry = &c.Configurations[0]
	entry.Code = pirconfig.Code{S: 20, M: 3, T: 5}
	entry.EncodingStatus = "candidate"
	entry.Capacity.Status = "estimated"
	entry.Capacity.N = uint64(1) << 60
	k := uint64(1)
	rate := 1 / float64(entry.Capacity.N)
	entry.Capacity.K = &k
	entry.Capacity.Rate = &rate
	entry.Query = pirconfig.Query{CurvePositions: (1 << 20) - 1, DummyPositions: 128, Length: (1 << 20) + 127, DummyRule: "fixed_count_uniform_positions"}
	one.RecordBits = 20
	one.Databases[0].ColumnEnd = 20
	if _, err := Evaluate(one, c, Requirements{}); err == nil || !strings.Contains(err.Error(), "storage bits") {
		t.Fatal("storage overflow accepted", err)
	}
	c = fixtureCatalogue(t, 2)
	p := fixturePlan()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Evaluate(p, c, Requirements{}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
