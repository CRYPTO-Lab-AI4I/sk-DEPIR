package liftrs

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

// Runtime acceleration tables are reconstructed from the field/code, not added
// to version-1 artifacts. Rebuilding must preserve the historical serialized data.
func TestConstantTablesPreserveArtifacts(t *testing.T) {
	for _, item := range []struct{ s, t int }{{4, 2}, {8, 8}, {9, 7}} {
		root := "../../experiments/encoder-comparison/artifacts"
		if item.s == 9 {
			root = "../../experiments/tablefree-encoding/artifacts"
		}
		stem := fmt.Sprintf("%s/s%02d-m2-t%d", root, item.s, item.t)
		raw, err := os.ReadFile(stem + "-basis.json")
		if err != nil {
			t.Fatal(err)
		}
		b, err := LoadOrbit(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, method := range []string{"paper", "cyclic-fft"} {
			var serialized []byte
			if method == "paper" {
				e, err := b.PreparePaper()
				if err != nil {
					t.Fatal(err)
				}
				serialized, err = e.MarshalBinary()
				if err != nil {
					t.Fatal(err)
				}
			} else {
				e, err := b.PrepareCyclicFFT()
				if err != nil {
					t.Fatal(err)
				}
				serialized, err = e.MarshalBinary()
				if err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(stem + "-" + method + ".json")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(serialized, want) {
				t.Fatalf("s=%d %s artifact changed", item.s, method)
			}
		}
	}
}
func TestConstantTableSizes(t *testing.T) {
	for s := 9; s <= 13; s++ {
		p, _ := BinaryPolynomial(s)
		base, ext, _, _ := orbitFields(s, p)
		baseBytes := newAdditiveTransform(base).constantTableBytes
		extBytes := newAdditiveTransform(ext).constantTableBytes
		t.Logf("s=%d base_table_bytes=%d extension_table_bytes=%d", s, baseBytes, extBytes)
	}
}
