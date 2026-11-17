package main

import "testing"

func TestSpectrumTileSelection(t *testing.T) {
	base := []string{"--s", "4", "--t", "3", "--cache", "unused", "--backend", "cuda", "--gpu-library", "unused.so", "--gpu-bytes", "1048576", "--max-array-bytes", "1048576", "--native-bytes", "1048576", "--max-code-bytes", "1048576"}
	c, err := parse(base)
	if err != nil || c.spectrumTile != 0 {
		t.Fatal("resident default", err)
	}
	for _, tile := range []string{"8", "24", "1048576", "67108864"} {
		args := append(append([]string{}, base...), "--spectrum-tile-bytes", tile)
		if _, err = parse(args); err != nil {
			t.Fatal(tile, err)
		}
	}
	for _, tail := range [][]string{{"--spectrum-tile-bytes", "7"}, {"--spectrum-tile-bytes", "67108865"}, {"--spectrum-tile-bytes", "24", "--backend", "cpp"}, {"--spectrum-tile-bytes", "24", "--dimension-only"}} {
		args := append(append([]string{}, base...), tail...)
		if _, err = parse(args); err == nil {
			t.Fatal("invalid tile/mode accepted", tail)
		}
	}
}
