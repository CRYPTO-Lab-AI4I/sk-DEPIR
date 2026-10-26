package pirconfig

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"rme/encoding/liftrs"
)

func fixture(t *testing.T) *Catalogue {
	t.Helper()
	f, err := os.Open("testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestReferenceEvidence(t *testing.T) {
	c := fixture(t)
	entry := c.Configurations[0]
	// Independently execute the cited backend in this test, never inside Load.
	code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: entry.Code.S, Dimension: entry.Code.M, CurveDegree: entry.Code.T})
	if err != nil {
		t.Fatal(err)
	}
	if uint64(code.Dimension()) != *entry.Capacity.K || uint64(code.Length()) != entry.Capacity.N || code.Polynomial() != *entry.Code.Polynomial {
		t.Fatal("fixture disagrees with reference encoder")
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bytes.NewReader(raw)); err != nil {
		t.Fatal(err)
	}
}

func TestTargetProfile(t *testing.T) {
	f, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := Load(f)
	if err != nil {
		t.Fatal(err)
	}
	fcsv, err := os.Open("../experiments/bounded-sample-parameters/results.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer fcsv.Close()
	rows, err := csv.NewReader(fcsv).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Configurations) != 13 || len(rows) != 14 {
		t.Fatal("missing target degrees")
	}
	for i, entry := range c.Configurations {
		s, _ := strconv.Atoi(rows[i+1][0])
		degree, _ := strconv.Atoi(rows[i+1][2])
		if entry.Code.S != s || entry.Code.M != 2 || entry.Code.T != degree {
			t.Fatalf("target %d differs from documented security calculation", i)
		}
		if entry.EncodingStatus != "candidate" || entry.Encoder != nil {
			t.Fatal("target implies unsupported encoding")
		}
		// Table 1 supplies s=12,20; Table 3 refines the matching s=15,18 rates.
		paperRates := map[int]float64{12: .117, 15: .2024, 18: .2367, 20: .259}
		if rate, ok := paperRates[s]; ok {
			wantK := uint64(math.Round(rate * float64(entry.Capacity.N)))
			if entry.Capacity.Status != "estimated" || entry.Capacity.K == nil || *entry.Capacity.K != wantK || entry.Capacity.Rate == nil || *entry.Capacity.Rate != float64(wantK)/float64(entry.Capacity.N) {
				t.Fatal("paper estimate missing or incorrectly transcribed")
			}
			evidence := entry.Capacity.Evidence
			wantCount := 1
			if s == 15 || s == 18 {
				wantCount = 2
			}
			if len(evidence) != wantCount || evidence[0].Kind != "paper" || evidence[0].Reference != "paper.pdf#page=64" || evidence[0].Method == "" {
				t.Fatal("paper estimate lacks source or rounding method")
			}
			if wantCount == 2 && (evidence[1].Kind != "paper" || evidence[1].Reference != "paper.pdf#page=66" || evidence[1].Method == "") {
				t.Fatal("more precise Table 3 evidence missing")
			}
		} else if entry.Capacity.Status != "unknown" || entry.Capacity.K != nil {
			t.Fatal("capacity asserted without a matching paper row")
		}
		if entry.Security.ComputationalBits != 100 || entry.Security.Log2Samples != 50 {
			t.Fatal("wrong initial security profile")
		}
	}
}

func TestRejectInconsistentConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Catalogue)
	}{
		{"version", func(c *Catalogue) { c.Version = 2 }},
		{"construction", func(c *Catalogue) { c.Construction = "rm" }},
		{"empty", func(c *Catalogue) { c.Configurations = nil }},
		{"duplicate id", func(c *Catalogue) { c.Configurations = append(c.Configurations, c.Configurations[0]) }},
		{"bad id", func(c *Catalogue) { c.Configurations[0].ID = "not an id" }},
		{"field bounds", func(c *Catalogue) { c.Configurations[0].Code.S = 64 }},
		{"dimension bounds", func(c *Catalogue) { c.Configurations[0].Code.M = 4 }},
		{"curve bounds", func(c *Catalogue) { c.Configurations[0].Code.T = 3 }},
		{"length", func(c *Catalogue) { c.Configurations[0].Capacity.N = 15 }},
		{"reducible polynomial", func(c *Catalogue) { *c.Configurations[0].Code.Polynomial = 5 }},
		{"wrong degree polynomial", func(c *Catalogue) { *c.Configurations[0].Code.Polynomial = 11 }},
		{"support status", func(c *Catalogue) { c.Configurations[0].EncodingStatus = "maybe" }},
		{"capacity status", func(c *Catalogue) { c.Configurations[0].Capacity.Status = "maybe" }},
		{"unknown with k", func(c *Catalogue) { c.Configurations[0].Capacity.Status = "unknown" }},
		{"zero k", func(c *Catalogue) { *c.Configurations[0].Capacity.K = 0 }},
		{"large k", func(c *Catalogue) { *c.Configurations[0].Capacity.K = 17 }},
		{"missing k", func(c *Catalogue) { c.Configurations[0].Capacity.K = nil }},
		{"rate mismatch", func(c *Catalogue) { *c.Configurations[0].Capacity.Rate = .5 }},
		{"negative rate", func(c *Catalogue) { *c.Configurations[0].Capacity.Rate = -.1 }},
		{"nonfinite rate", func(c *Catalogue) { *c.Configurations[0].Capacity.Rate = math.NaN() }},
		{"missing evidence", func(c *Catalogue) { c.Configurations[0].Capacity.Evidence = nil }},
		{"paper is not verified", func(c *Catalogue) { c.Configurations[0].Capacity.Evidence[0].Kind = "paper" }},
		{"curve count", func(c *Catalogue) { c.Configurations[0].Query.CurvePositions = 4 }},
		{"dummy rule", func(c *Catalogue) { c.Configurations[0].Query.DummyRule = "unknown" }},
		{"missing dummies in length", func(c *Catalogue) { c.Configurations[0].Query.Length = 3 }},
		{"query overflow", func(c *Catalogue) { c.Configurations[0].Query.DummyPositions = math.MaxUint64 }},
		{"security assumptions", func(c *Catalogue) { c.Configurations[0].Security.Assumptions = "" }},
		{"security evidence", func(c *Catalogue) { c.Configurations[0].Security.Evidence = nil }},
		{"missing polynomial", func(c *Catalogue) { c.Configurations[0].Code.Polynomial = nil }},
		{"missing encoder", func(c *Catalogue) { c.Configurations[0].Encoder = nil }},
		{"encoder implementation", func(c *Catalogue) { c.Configurations[0].Encoder.Implementation = "" }},
		{"encoder preprocessing", func(c *Catalogue) { c.Configurations[0].Encoder.Preprocessing = "" }},
		{"encoder evidence", func(c *Catalogue) { c.Configurations[0].Encoder.Evidence = nil }},
		{"bad evidence", func(c *Catalogue) { c.Configurations[0].Capacity.Evidence[0].Reference = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fixture(t)
			tt.change(c)
			if err := c.Validate(); err == nil {
				t.Fatal("accepted inconsistent catalogue")
			}
		})
	}
	var c *Catalogue
	if c.Validate() == nil {
		t.Fatal("accepted nil catalogue")
	}
}

func estimated(t *testing.T) *Catalogue {
	c := fixture(t)
	entry := &c.Configurations[0]
	entry.EncodingStatus = "candidate"
	entry.Capacity.Status = "estimated"
	entry.Encoder = nil
	entry.Code.Polynomial = nil
	entry.Capacity.Evidence = []Evidence{{Kind: "monte_carlo", Reference: "synthetic-test/report.md", Script: "synthetic-test/run.py", Results: "synthetic-test/results.json", Method: "Synthetic metadata fixture; no measurement or statistical claim.", MonteCarlo: &MonteCarlo{Samples: 1000, Confidence: .95, LowerRate: .28, UpperRate: .35}}}
	return c
}

func TestEstimatedCapacity(t *testing.T) {
	c := estimated(t)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	changes := []func(*Evidence){
		func(e *Evidence) { e.Script = "" }, func(e *Evidence) { e.Results = "" }, func(e *Evidence) { e.Method = "" },
		func(e *Evidence) { e.MonteCarlo = nil }, func(e *Evidence) { e.MonteCarlo.Samples = 0 },
		func(e *Evidence) { e.MonteCarlo.Confidence = 1 }, func(e *Evidence) { e.MonteCarlo.Confidence = math.NaN() },
		func(e *Evidence) { e.MonteCarlo.LowerRate = .4 }, func(e *Evidence) { e.MonteCarlo.UpperRate = .3 },
		func(e *Evidence) { e.MonteCarlo.LowerRate = -.1 }, func(e *Evidence) { e.Kind = "paper" },
	}
	for i, change := range changes {
		c := estimated(t)
		change(&c.Configurations[0].Capacity.Evidence[0])
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted invalid estimate %d", i)
		}
	}
	c = estimated(t)
	c.Configurations[0].EncodingStatus = "supported"
	if c.Validate() == nil {
		t.Fatal("empirical capacity permits encoding")
	}
}

func TestMeasurements(t *testing.T) {
	measurement := Measurement{Stage: "code_preprocessing", Quantity: "peak_bytes", Value: 1024, Evidence: Evidence{Kind: "measurement", Reference: "synthetic-test/report.md", Script: "synthetic-test/run.sh", Results: "synthetic-test/results.json"}}
	c := fixture(t)
	c.Configurations[0].Measurements = []Measurement{measurement}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Measurement){func(m *Measurement) { m.Stage = "query" }, func(m *Measurement) { m.Quantity = "speed" }, func(m *Measurement) { m.Value = -1 }, func(m *Measurement) { m.Value = math.Inf(1) }, func(m *Measurement) { m.Evidence.Kind = "paper" }} {
		m := measurement
		change(&m)
		c.Configurations[0].Measurements[0] = m
		if c.Validate() == nil {
			t.Fatal("accepted invalid measurement")
		}
	}
}

func TestStrictJSON(t *testing.T) {
	raw, err := os.ReadFile("testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []string{"", "null", "[]", "{}", string(raw) + " {}", string(raw) + " garbage",
		strings.Replace(string(raw), `"version": 1`, `"version": 1, "version": 1`, 1),
		strings.Replace(string(raw), `"s": 2`, `"s": 2, "s": 3`, 1),
		strings.Replace(string(raw), `"version": 1`, `"Version": 1`, 1),
		strings.Replace(string(raw), `"version": 1`, `"version": 1, "extra": true`, 1),
		strings.Replace(string(raw), `"dummy_positions": 128,`, "", 1),
		strings.Replace(string(raw), `"computational_bits": 0,`, "", 1),
		strings.Replace(string(raw), `"version": 1`, `"version": null`, 1),
		strings.Replace(string(raw), `"polynomial": 7`, `"polynomial": null`, 1),
		strings.Replace(string(raw), `"s": 2`, `"s": 2.5`, 1),
		strings.Replace(string(raw), `"s": 2`, `"s": "2"`, 1),
		strings.Replace(string(raw), `"k": 5`, `"k": -5`, 1),
		strings.Repeat("[", 66) + strings.Repeat("]", 66),
	}
	for i, raw := range tests {
		if _, err := Load(strings.NewReader(raw)); err == nil {
			t.Fatalf("accepted malformed JSON %d", i)
		}
	}
	if _, err := Load(nil); err == nil {
		t.Fatal("accepted nil reader")
	}
	sentinel := errors.New("read failed")
	if _, err := Load(failingReader{sentinel}); !errors.Is(err, sentinel) {
		t.Fatal("reader error not preserved")
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

var _ io.Reader = failingReader{}

func TestPolynomialValidation(t *testing.T) {
	for _, test := range []struct {
		p     uint32
		s     int
		valid bool
	}{{7, 2, true}, {11, 3, true}, {19, 4, true}, {0x11b, 8, true}, {0x1100b, 16, true}, {5, 2, false}, {21, 4, false}, {0x101, 8, false}, {0, 4, false}, {1, 4, false}} {
		if got := irreducible(test.p, test.s); got != test.valid {
			t.Fatalf("polynomial %x, degree %d: got %v", test.p, test.s, got)
		}
	}
}

func FuzzLoad(f *testing.F) {
	for _, path := range []string{"configs_catalogue.json", "testdata/reference.json"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		c, err := Load(bytes.NewReader(raw))
		if err == nil {
			if err := c.Validate(); err != nil {
				t.Fatal(err)
			}
		}
	})
}
