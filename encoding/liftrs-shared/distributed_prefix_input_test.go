package liftrs

import "testing"

func TestRankPrefixInputView(t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 9} {
		for _, count := range []uint64{1, 5, 7} {
			total := uint64(0)
			for rank := 0; rank < ranks; rank++ {
				first, words := rankBand(7, rank, ranks)
				active, err := rankPrefixInputWords(7, first, words, count)
				if err != nil {
					t.Fatal(err)
				}
				// Reconstructing a zeroed destination from the view must match an explicit
				// masked copy,including partial and empty ownership. Source is immutable.
				source, actual := make([]uint32, words), make([]uint32, words)
				for j := range source {
					source[j] = uint32(first + uint64(j) + 1)
				}
				copy(actual, source[:active])
				for j := range actual {
					want := uint32(0)
					if first+uint64(j) < count {
						want = source[j]
					}
					if actual[j] != want || source[j] != uint32(first+uint64(j)+1) {
						t.Fatal("prefix view/mutation")
					}
				}
				total += active
			}
			if total != count {
				t.Fatal("prefix partition", total, count)
			}
		}
	}
}

func TestRankPrefixInputWideBounds(t *testing.T) {
	n := uint64(1) << 36
	active, err := rankPrefixInputWords(n, 1<<34, 1<<32, (1<<34)+17)
	if err != nil || active != 17 {
		t.Fatal(active, err)
	}
	for _, p := range [][4]uint64{{0, 0, 0, 1}, {7, 0, 7, 0}, {7, 0, 7, 8}, {7, 8, 0, 1}, {7, 6, 2, 7}} {
		if _, err := rankPrefixInputWords(p[0], p[1], p[2], p[3]); err == nil {
			t.Fatal("bad prefix input", p)
		}
	}
	for _, value := range []string{"", "dense", "prefix"} {
		mode, err := newtonInputMode(value)
		if err != nil || (value == "" && mode != "dense") || (value != "" && mode != value) {
			t.Fatal(mode, err)
		}
	}
	if _, err := newtonInputMode("unknown"); err == nil {
		t.Fatal("bad input mode")
	}
}
