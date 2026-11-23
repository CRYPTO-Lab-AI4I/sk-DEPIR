package liftrs

import (
	"errors"
	"reflect"
	"testing"
)

func TestRankReciprocalColdWarmModes(t *testing.T) {
	for _, prefix := range []bool{false, true} {
		for _, known := range []uint64{1, 4, 7} {
			root := make([]uint32, 16)
			root[0], root[1], root[14] = 1, 1, 9
			// The unused logical root tail must remain untouched in both modes.
			original := append([]uint32(nil), root...)
			initial := make([]uint32, 16)
			for i := uint64(0); i < known; i++ {
				initial[i] = 1
			}
			checkpoints := []uint64{}
			ops := rankReciprocalOps{
				Square: func(g []uint32, keep uint64) ([]uint32, error) {
					out := make([]uint32, 16)
					for j, v := range g {
						if uint64(2*j) < keep {
							out[2*j] = v
						}
					}
					return out, nil
				},
				Product: func(_ string, a, b []uint32, ac, bc, keep uint64, selected bool) ([]uint32, uint64, error) {
					if selected != prefix {
						t.Fatal("input mode")
					}
					if !prefix {
						for _, v := range a[ac:] {
							if v != 0 {
								t.Fatal("unmasked dense tail")
							}
						}
					}
					out := make([]uint32, 16)
					for i := uint64(0); i < ac; i++ {
						for j := uint64(0); j < bc; j++ {
							if i+j < keep {
								out[i+j] ^= a[i] & b[j]
							}
						}
					}
					return out, ac + bc, nil
				},
				Checkpoint: func(size uint64, g []uint32) ([]uint32, error) {
					checkpoints = append(checkpoints, size)
					return append([]uint32(nil), g...), nil
				},
			}
			g, stats, err := refineRankReciprocal(rankReciprocalPlan{16, 2, 7, known, 16, 0, 1, prefix}, root, initial, ops)
			if err != nil {
				t.Fatal(err)
			}
			want := make([]uint32, 16)
			for i := 0; i < 7; i++ {
				want[i] = 1
			}
			if !reflect.DeepEqual(g, want) || !reflect.DeepEqual(root, original) {
				t.Fatal("reciprocal/source mutation", known, prefix)
			}
			steps := map[uint64]int{1: 3, 4: 1, 7: 0}[known]
			if stats.Steps != steps || len(checkpoints) != steps {
				t.Fatal("schedule", stats, checkpoints)
			}
			clone := uint64(0)
			if !prefix && steps > 0 {
				clone = 16
			}
			if stats.RootCloneWordsPerStep != clone {
				t.Fatal("clone accounting")
			}
			if stats.UploadedWords != map[uint64]uint64{1: 17, 4: 9, 7: 0}[known] {
				t.Fatal("reported upload accounting", stats)
			}
		}
	}
}
func TestRankReciprocalRejectsFailures(t *testing.T) {
	p := rankReciprocalPlan{Domain: 16, RootCount: 2, Target: 7, Known: 1, MaxLocalWords: 16, Ranks: 1, PrefixInputs: true}
	failure := errors.New("backend failure")
	valid := rankReciprocalOps{Square: func([]uint32, uint64) ([]uint32, error) { return make([]uint32, 16), nil }, Product: func(string, []uint32, []uint32, uint64, uint64, uint64, bool) ([]uint32, uint64, error) {
		return make([]uint32, 16), 0, nil
	}}
	for _, stage := range []string{"square", "product", "checkpoint", "extent"} {
		ops := valid
		switch stage {
		case "square":
			ops.Square = func([]uint32, uint64) ([]uint32, error) { return nil, failure }
		case "product":
			ops.Product = func(string, []uint32, []uint32, uint64, uint64, uint64, bool) ([]uint32, uint64, error) {
				return nil, 0, failure
			}
		case "checkpoint":
			ops.Checkpoint = func(uint64, []uint32) ([]uint32, error) { return nil, failure }
		case "extent":
			ops.Checkpoint = func(uint64, []uint32) ([]uint32, error) { return make([]uint32, 15), nil }
		}
		out, _, err := refineRankReciprocal(p, make([]uint32, 16), make([]uint32, 16), ops)
		if err == nil || out != nil {
			t.Fatal("failed refinement published", stage)
		}
		if stage != "extent" && !errors.Is(err, failure) {
			t.Fatal("lost backend error", stage, err)
		}
	}
	for _, change := range []func(*rankReciprocalPlan){func(p *rankReciprocalPlan) { p.MaxLocalWords = 15 }, func(p *rankReciprocalPlan) { p.Known = 3 }, func(p *rankReciprocalPlan) { p.Target = 16 }, func(p *rankReciprocalPlan) { p.Domain = 15 }} {
		bad := p
		change(&bad)
		if _, _, err := refineRankReciprocal(bad, make([]uint32, 16), make([]uint32, 16), valid); err == nil {
			t.Fatal("bad plan accepted", bad)
		}
	}
}

func TestRankFullCodeIdentity(t *testing.T) {
	for rank := 0; rank < 4; rank++ {
		values := make([]uint32, 4)
		if rank == 0 {
			values[0] = 1
		}
		if rank == 3 {
			values[3] = 1
		}
		if err := verifyRankCodeIdentity(16, rank, 4, values); err != nil {
			t.Fatal(err)
		}
		values[1] ^= 1
		if err := verifyRankCodeIdentity(16, rank, 4, values); err == nil {
			t.Fatal("identity corruption accepted", rank)
		}
	}
	if err := verifyRankCodeIdentity(16, 0, 4, make([]uint32, 3)); err == nil {
		t.Fatal("short identity accepted")
	}
}
