package liftrs

import "testing"

func TestRankRootActualCountAdmission(t *testing.T) {
	counts := []uint64{3442, 1507, 1541, 895}
	for rank := range counts {
		cap := uint64(2048)
		if rank < 2 {
			cap = 4096
		}
		stages, err := admitRankRootStages(rank, counts, 65536, cap)
		if err != nil || stages[len(stages)-1].Domain != 8192 {
			t.Fatal("actual rounded root domains", rank, err)
		}
		if _, err := admitRankRootStages(rank, counts, 65536, cap-1); err == nil {
			t.Fatal("rounded stage cap ignored", rank)
		}
	}
	wide := make([]uint64, 16)
	for i := range wide {
		wide[i] = 1
	}
	wide[0] = (1 << 33) + 1
	if _, err := admitRankRootStages(0, wide, 1<<36, 1<<32); err == nil {
		t.Fatal("wide leaf truncated")
	}
	if _, err := admitRankRootStages(0, wide, 1<<36, 1<<34); err != nil {
		t.Fatal("metadata-only wide admission", err)
	}
}
func TestRankRootIdentityAndGuards(t *testing.T) {
	p, err := prepareRankRoot(supportSorted, orbitField{q: 16}, 256, nil, nil, nil, nil)
	if err != nil || len(p) != 1 || p[0] != 1 {
		t.Fatal("empty root ownership", p, err)
	}
	if _, err := prepareRankRoot(supportSorted, orbitField{q: 16}, 256, nil, []int{0}, nil, nil); err == nil {
		t.Fatal("missing engine")
	}
	if _, err := prepareRankRoot(supportSorted, orbitField{q: 16}, 128, nil, nil, nil, nil); err == nil {
		t.Fatal("incorrect field domain")
	}
}
