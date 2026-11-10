package pirconfig

import (
	"encoding/json"
	"os"
	"testing"
)

// Capacity promotion is tied to retained exact discovery and complete encoding
// verification, not merely to a new field-width guard or a Monte Carlo estimate.
func TestS14CapacityMatchesEncodingEvidence(t *testing.T) {
	var run struct {
		Status string
		Phases []struct {
			Phase      string
			K, N       uint64
			Systematic uint64 `json:"systematic_checked"`
			Curves     int    `json:"curves_checked"`
			Canonical  string `json:"canonical_sha256"`
			Match      bool   `json:"repeat_match"`
			Identity   bool   `json:"identity_verified"`
		}
	}
	raw, err := os.ReadFile("../experiments/parallel-s14/initial/results.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &run); err != nil {
		t.Fatal(err)
	}
	if run.Status != "passed" {
		t.Fatal("s14 run incomplete")
	}
	var k, n uint64
	checks := 0
	identity := false
	digest := ""
	for _, p := range run.Phases {
		if p.Phase == "discovery" {
			k, n = p.K, p.N
		}
		if p.Phase == "preparation" {
			identity = p.Identity
		}
		if p.Phase == "verify-0" || p.Phase == "verify-1" || p.Phase == "verify-2" {
			checks++
			if p.Systematic != 50976325 || p.Curves != 16 || !p.Match || len(p.Canonical) != 64 || (digest != "" && p.Canonical != digest) {
				t.Fatal("s14 output checks")
			}
			digest = p.Canonical
		}
	}
	if !identity || checks != 3 || k != 50976325 || n != 268435456 {
		t.Fatal("s14 evidence missing")
	}
	file, err := os.Open("configs_catalogue.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	c, err := Load(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range c.Configurations {
		if entry.Code.S == 14 {
			if entry.Capacity.K == nil || *entry.Capacity.K != k || entry.Capacity.N != n || entry.EncodingStatus != "supported" || entry.Capacity.Status != "verified" || entry.Code.Polynomial == nil || *entry.Code.Polynomial != 0x402b {
				t.Fatal("s14 catalogue differs")
			}
			return
		}
	}
	t.Fatal("s14 config absent")
}
