package pirconfig

import (
	"encoding/json"
	"os"
	"testing"
)

// Promotion requires completed database encoding evidence, not only the exact
// discovery dimension. The in-memory benchmark does not establish cache loading
// or application-wrapper support.
func TestS13CapacityMatchesEncodingEvidence(t *testing.T) {
	raw, err := os.ReadFile("../experiments/cyclic-s13/results.json")
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Status     string
		Returncode int
		Code       struct {
			S, M, T int
			K, N    uint64
		}
		Verification struct {
			K, N                uint64
			CurveChecks         int    `json:"curve_checks"`
			SystematicPositions uint64 `json:"systematic_positions_checked"`
			PackedBytes         uint64 `json:"packed_bytes"`
			Canonical           string `json:"canonical_sha256"`
			Packed              string `json:"packed_sha256"`
		}
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.Returncode != 0 {
		t.Fatal("s=13 encoding evidence is incomplete")
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
	for _, entry := range catalogue.Configurations {
		if entry.Code.S != 13 {
			continue
		}
		c, v := result.Code, result.Verification
		if c.S != 13 || c.M != 2 || c.T != 5 || c.K != *entry.Capacity.K || c.N != entry.Capacity.N || v.K != c.K || v.N != c.N {
			t.Fatal("s=13 catalogue differs from measured code")
		}
		if v.SystematicPositions != c.K || v.CurveChecks != 16 || v.PackedBytes != (c.N*13+7)/8 || len(v.Canonical) != 64 || len(v.Packed) != 64 {
			t.Fatal("s=13 output verification is incomplete")
		}
		if entry.EncodingStatus != "supported" || entry.Capacity.Status != "verified" || entry.Code.Polynomial == nil || *entry.Code.Polynomial != 0x201b {
			t.Fatal("s=13 standalone support metadata missing")
		}
		return
	}
	t.Fatal("s=13 entry missing")
}
