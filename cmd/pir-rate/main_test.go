package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-go"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/pirconfig"
)

func TestEstimatorOnExactFixture(t *testing.T) {
	args := []string{"-catalogue", "../../pirconfig/testdata/reference.json", "-samples", "10000", "-seed", "73", "-confidence", "0.99"}
	var first, second bytes.Buffer
	if err := run(args, &first); err != nil {
		t.Fatal(err)
	}
	if err := run(args, &second); err != nil {
		t.Fatal(err)
	}
	var a, b results
	if err := json.Unmarshal(first.Bytes(), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second.Bytes(), &b); err != nil {
		t.Fatal(err)
	}
	if len(a.Estimates) != 1 {
		t.Fatal("wrong count")
	}
	e, f := a.Estimates[0], b.Estimates[0]
	if e.Good != f.Good || e.Rate != f.Rate || e.Seed != f.Seed || e.LowerRate != f.LowerRate {
		t.Fatal("seeded counts/intervals not reproducible")
	}
	if e.LowerRate > 5.0/16 || e.UpperRate < 5.0/16 {
		t.Fatal("exact fixture dimension not in fixed-seed interval")
	}
	margin := math.Sqrt(math.Log(200) / (2 * 10000))
	if math.Abs(e.LowerRate-(e.Rate-margin)) > 1e-14 || math.Abs(e.UpperRate-(e.Rate+margin)) > 1e-14 {
		t.Fatal("wrong confidence interval")
	}
	if e.Samples != 10000 || e.Seed != 2275 || e.Confidence != .99 || len(a.CatalogueSHA256) != 64 || a.GoVersion == "" {
		t.Fatalf("wrong provenance: %+v", e)
	}
	// Evaluate the same sampled tuples through actual systematic code membership,
	// obtained from the dense reference's independently tested monomial set.
	code, err := liftrs.Preprocess(liftrs.Spec{FieldDegree: 2, Dimension: 2, CurveDegree: 2})
	if err != nil {
		t.Fatal(err)
	}
	goodSet := map[string]bool{}
	for _, m := range code.Monomials() {
		goodSet[fmt.Sprint(m)] = true
	}
	rng := rand.New(rand.NewSource(e.Seed))
	var want uint64
	for i := uint64(0); i < e.Samples; i++ {
		m := []uint32{rng.Uint32() & 3, rng.Uint32() & 3}
		if goodSet[fmt.Sprint(m)] {
			want++
		}
	}
	if e.Good != want {
		t.Fatalf("sample classification differs from dense code basis: %d vs %d", e.Good, want)
	}
}

func TestEstimatorInputErrors(t *testing.T) {
	for _, args := range [][]string{{"-samples", "0"}, {"-confidence", "1"}, {"-confidence", "NaN"}, {"unexpected"}, {"-catalogue", "missing"}, {"-unknown"}} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if run([]string{"-catalogue", path}, &bytes.Buffer{}) == nil {
		t.Fatal("accepted malformed catalogue")
	}
}

func TestFamilyConfidence(t *testing.T) {
	raw, err := os.ReadFile("../../pirconfig/testdata/reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var c pirconfig.Catalogue
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	second := c.Configurations[0]
	second.ID = "second-fixture"
	c.Configurations = append(c.Configurations, second)
	raw, err = json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "configs.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"-catalogue", path, "-samples", "1000"}, &out); err != nil {
		t.Fatal(err)
	}
	var r results
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	margin := math.Sqrt(math.Log(400) / 2000)
	for _, e := range r.Estimates {
		if e.Confidence != .995 || math.Abs(e.LowerRate-math.Max(0, e.Rate-margin)) > 1e-14 {
			t.Fatal("family error was not split across entries")
		}
	}
}
