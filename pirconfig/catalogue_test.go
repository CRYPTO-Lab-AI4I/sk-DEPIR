package pirconfig

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
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

type rateObservation struct {
	ID         string  `json:"id"`
	Good       uint64  `json:"good"`
	Samples    uint64  `json:"samples"`
	LowerRate  float64 `json:"lower_rate"`
	UpperRate  float64 `json:"upper_rate"`
	Confidence float64 `json:"confidence"`
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
	raw, err := os.ReadFile("../experiments/liftrs-rates/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Estimates []rateObservation `json:"estimates"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	measurements := make(map[string]rateObservation)
	for _, result := range observed.Estimates {
		if result.Samples == 0 || result.Good > result.Samples {
			t.Fatal("invalid recorded sample count")
		}
		measurements[result.ID] = result
	}
	if len(measurements) != len(c.Configurations) {
		t.Fatal("recorded estimates do not cover the catalogue")
	}

	for i, entry := range c.Configurations {
		s, _ := strconv.Atoi(rows[i+1][0])
		degree, _ := strconv.Atoi(rows[i+1][2])
		if entry.Code.S != s || entry.Code.M != 2 || entry.Code.T != degree {
			t.Fatalf("target %d differs from documented security calculation", i)
		}
		if entry.Code.S >= 8 && entry.Code.S <= 12 {
			wantK, wantPolynomial := uint64(3121), uint32(0x11d)
			if entry.Code.S == 9 {
				wantK, wantPolynomial = 17200, 0x211
			}
			if entry.Code.S == 10 {
				wantK, wantPolynomial = 103556, 0x409
			}
			if entry.Code.S == 11 {
				wantK, wantPolynomial = 455434, 0x805
			}
			if entry.Code.S == 12 {
				wantK, wantPolynomial = 1965739, 0x1053
			}
			if entry.EncodingStatus != "supported" || entry.Encoder == nil || entry.Capacity.Status != "verified" || entry.Code.Polynomial == nil || *entry.Code.Polynomial != wantPolynomial || entry.Capacity.K == nil || *entry.Capacity.K != wantK {
				t.Fatal("missing bounded exact encoding evidence")
			}
		} else if entry.EncodingStatus != "candidate" || entry.Encoder != nil || entry.Capacity.Status != "estimated" {
			t.Fatal("unverified target promoted")
		}
		if entry.Capacity.K == nil || entry.Capacity.Rate == nil {
			t.Fatal("missing capacity")
		}
		if entry.Code.S >= 8 && entry.Code.S <= 12 {
			found := false
			for _, e := range entry.Capacity.Evidence {
				if e.Kind == "encoding" && (e.Results == "experiments/encoder-comparison/results.json" || e.Results == "experiments/tablefree-encoding/results.json" || e.Results == "experiments/encoder-scaling/results.json" || e.Results == "experiments/cyclic-scaling/results.json") {
					found = true
				}
			}
			if !found {
				t.Fatal("exact evidence absent")
			}
		}
		evidence := entry.Capacity.Evidence
		foundMC := false
		var paperPages []string
		for _, source := range evidence {
			if source.Kind == "paper" {
				paperPages = append(paperPages, source.Reference)
			}
			if source.Kind == "monte_carlo" {
				foundMC = true
				measured, ok := measurements[entry.ID]
				if !ok || source.Reference != "experiments/liftrs-rates/report.md" || source.Script != "experiments/liftrs-rates/run.py" || source.Results != "experiments/liftrs-rates/results.json" || source.MonteCarlo == nil {
					t.Fatal("empirical evidence missing or not reproducible")
				}
				wantK := (measured.Good*entry.Capacity.N + measured.Samples/2) / measured.Samples
				if (entry.Code.S > 12 && (*entry.Capacity.K != wantK || *entry.Capacity.Rate != float64(wantK)/float64(entry.Capacity.N))) || source.MonteCarlo.Samples != measured.Samples || source.MonteCarlo.LowerRate != measured.LowerRate || source.MonteCarlo.UpperRate != measured.UpperRate || source.MonteCarlo.Confidence != measured.Confidence {
					t.Fatal("catalogue does not match recorded rate experiment")
				}
			}
		}
		if !foundMC {
			t.Fatal("missing Monte Carlo evidence")
		}
		wantPages := []string{}
		if s == 12 || s == 15 || s == 18 || s == 20 {
			wantPages = append(wantPages, "paper.pdf#page=64")
		}
		if s == 15 || s == 18 {
			wantPages = append(wantPages, "paper.pdf#page=66")
		}
		if fmt.Sprint(paperPages) != fmt.Sprint(wantPages) {
			t.Fatal("paper evidence was lost")
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

func TestExactCapacityMatchesEncodingEvidence(t *testing.T) {
	raw, err := os.ReadFile("../experiments/encoder-comparison/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var saved struct {
		Scenarios []struct {
			Preprocessing struct {
				S, M, T int
				K, N    uint64
			}
			Methods []struct {
				Output string `json:"output_sha256"`
			}
		}
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	catalogue, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	entry := catalogue.Configurations[0]
	found := false
	for _, scenario := range saved.Scenarios {
		p := scenario.Preprocessing
		if p.S == entry.Code.S && p.M == entry.Code.M && p.T == entry.Code.T {
			found = true
			if p.K != *entry.Capacity.K || p.N != entry.Capacity.N || len(scenario.Methods) != 3 {
				t.Fatal("exact capacity evidence differs")
			}
			for _, m := range scenario.Methods {
				if m.Output == "" || m.Output != scenario.Methods[0].Output {
					t.Fatal("encoding evidence does not agree")
				}
			}
		}
	}
	if !found {
		t.Fatal("exact configuration evidence missing")
	}
}

func TestS9CapacityMatchesTableFreeEvidence(t *testing.T) {
	file, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	c, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../experiments/tablefree-encoding/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Scenarios []struct {
			Discovery struct {
				S, M, T int
				K, N    uint64
			}
			Methods []struct {
				Encoding struct {
					Output string `json:"output_sha256"`
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, caseResult := range result.Scenarios {
		d := caseResult.Discovery
		if d.S != 9 {
			continue
		}
		found = true
		e := c.Configurations[1]
		if d.K != *e.Capacity.K || d.N != e.Capacity.N || d.M != e.Code.M || d.T != e.Code.T {
			t.Fatal("s=9 evidence differs from catalogue")
		}
		if len(caseResult.Methods) != 2 || caseResult.Methods[0].Encoding.Output == "" || caseResult.Methods[0].Encoding.Output != caseResult.Methods[1].Encoding.Output {
			t.Fatal("s=9 encoder evidence disagrees")
		}
	}
	if !found {
		t.Fatal("missing s=9 evidence")
	}
}

// Check the catalogue's exact s=10 capacity against the separately recorded
// discovery and cross-encoder/curve evidence, retaining the old sampling record.
func TestS10CapacityMatchesScalingEvidence(t *testing.T) {
	file, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	c, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../experiments/encoder-scaling/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Discovery struct {
			S, M, T int
			K, N    uint64
		}
		Methods []struct {
			Encoding struct {
				Output string `json:"output_sha256"`
			}
		}
		Verification struct {
			K, N      uint64
			Equal     bool   `json:"paper_cyclic_equal"`
			Curves    int    `json:"curve_checks"`
			Positions uint64 `json:"systematic_positions_checked"`
		}
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	e := c.Configurations[2]
	d := result.Discovery
	if d.S != 10 || d.M != 2 || d.T != 6 || d.K != *e.Capacity.K || d.N != e.Capacity.N {
		t.Fatal("s=10 discovery/catalogue mismatch")
	}
	if len(result.Methods) != 2 || result.Methods[0].Encoding.Output == "" || result.Methods[0].Encoding.Output != result.Methods[1].Encoding.Output {
		t.Fatal("s=10 outputs disagree")
	}
	v := result.Verification
	if !v.Equal || v.Curves != 64 || v.K != d.K || v.N != d.N || v.Positions != d.K {
		t.Fatal("s=10 verification evidence incomplete")
	}
}

func TestLargerCapacitiesMatchScalingEvidence(t *testing.T) {
	file, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	catalogue, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../experiments/cyclic-scaling/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Cases []struct {
			S, T         int
			K, N         uint64
			Status       string
			Verification struct {
				K, N      uint64
				Positions uint64 `json:"systematic_positions_checked"`
				Curves    int    `json:"curve_checks"`
				Bytes     uint64 `json:"packed_bytes"`
			}
		}
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, c := range result.Cases {
		if c.Status != "measured" {
			continue
		}
		found++
		entry := catalogue.Configurations[c.S-8]
		if c.K != *entry.Capacity.K || c.N != entry.Capacity.N || c.T != entry.Code.T || c.Verification.K != c.K || c.Verification.N != c.N || c.Verification.Positions != c.K || c.Verification.Curves != 16 || c.Verification.Bytes != (c.N*uint64(c.S)+7)/8 {
			t.Fatal("scaling/catalogue evidence mismatch", c.S)
		}
	}
	if found != 2 {
		t.Fatal("missing s=11/12 measurements")
	}
}
