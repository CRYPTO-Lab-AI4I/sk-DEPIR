package liftrs

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

// In characteristic2,(1+X)^d has coefficient1 exactly when k is a bit-subset
// of d. This oracle is independent of polynomial convolution/tree scheduling.
func rootParityShard(first, words, count uint64) []uint32 {
	out := make([]uint32, words)
	degree := count - 1
	for j := range out {
		k := first + uint64(j)
		if k < count && k&degree == k {
			out[j] = 1
		}
	}
	return out
}

func TestShardedRootTree(t *testing.T) {
	for _, ranks := range []int{2, 4, 8} {
		counts := make([]uint64, ranks)
		inputs := make([][]uint32, ranks)
		for rank := range counts {
			counts[rank] = uint64(1 + (rank*7)%9)
			inputs[rank] = rootParityShard(0, counts[rank], counts[rank])
		}
		links := map[[2]int]operandTestLink{}
		for width, level := 2, 1; width <= ranks; width, level = width*2, level+1 {
			for first := 0; first < ranks; first += width {
				link := operandTestLink{map[[2]int]chan operandMessage{}}
				for a := 0; a < width; a++ {
					for b := 0; b < width; b++ {
						if a != b {
							link.wire[[2]int{a, b}] = make(chan operandMessage)
						}
					}
				}
				links[[2]int{level, first}] = link
			}
		}
		results := make([][]uint32, ranks)
		failures := make([]error, ranks)
		var wait sync.WaitGroup
		for rank := 0; rank < ranks; rank++ {
			wait.Add(1)
			go func(rank int) {
				defer wait.Done()
				before := append([]uint32(nil), inputs[rank]...)
				results[rank], failures[rank] = reduceShardedRoots(rank, counts, 4096, 2, inputs[rank], func(stage rankRootMergeStage, local []uint32) ([]uint32, error) {
					own := rank - stage.First
					var values [2][]uint32
					for j, source := range stage.Sources {
						if own >= source.First && own < source.First+source.Ranks {
							values[j] = local
						}
					}
					_, words := rankBand(stage.Domain, own, stage.Ranks)
					pair, err := routeShardedOperands(rankShardOperandPlan{uint64(stage.Level), stage.Domain, 2, 2, words, own, stage.Ranks, stage.Sources}, values, links[[2]int{stage.Level, stage.First}], nil)
					if err != nil {
						return nil, err
					}
					first, _ := rankBand(stage.Domain, own, stage.Ranks)
					for j, source := range stage.Sources {
						if !reflect.DeepEqual(pair[j], rootParityShard(first, words, source.Count)) {
							t.Errorf("child shard mismatch ranks=%d rank=%d level=%d", ranks, rank, stage.Level)
						}
					}
					return rootParityShard(first, words, stage.Count), nil
				})
				if !reflect.DeepEqual(before, inputs[rank]) {
					t.Error("source mutated")
				}
			}(rank)
		}
		wait.Wait()
		total := uint64(1)
		for _, count := range counts {
			total += count - 1
		}
		for rank, err := range failures {
			if err != nil {
				t.Fatal(err)
			}
			stages, _ := rankRootMergeStages(rank, counts, 4096)
			last := stages[len(stages)-1]
			first, words := rankBand(last.Domain, rank, ranks)
			if !reflect.DeepEqual(results[rank], rootParityShard(first, words, total)) {
				t.Fatal("final root shard")
			}
		}
	}
}

func TestShardedRootAdmission(t *testing.T) {
	for _, counts := range [][]uint64{nil, {1}, {1, 1, 1}, {0, 1}, {^uint64(0), 1}, {17, 17}} {
		if _, err := rankRootMergeStages(0, counts, 32); err == nil {
			t.Fatal("invalid tree admitted", counts)
		}
	}
	for _, rank := range []int{-1, 2} {
		if _, err := rankRootMergeStages(rank, []uint64{1, 1}, 16); err == nil {
			t.Fatal("invalid rank")
		}
	}
	for _, max := range []uint64{0, 3, 8} {
		if _, err := rankRootMergeStages(0, []uint64{1, 1, 1, 1}, max); err == nil {
			t.Fatal("canonical split cap")
		}
	}
	failure := errors.New("backend")
	_, err := reduceShardedRoots(0, []uint64{2, 3}, 16, 2, []uint32{1, 1}, func(rankRootMergeStage, []uint32) ([]uint32, error) { return nil, failure })
	if !errors.Is(err, failure) {
		t.Fatal("backend error lost")
	}
	_, err = reduceShardedRoots(0, []uint64{2, 3}, 16, 2, []uint32{1, 1}, func(stage rankRootMergeStage, _ []uint32) ([]uint32, error) {
		return make([]uint32, stage.Domain/2), nil
	})
	if err == nil {
		t.Fatal("unnormalized result")
	}
	// A distributed child with padding must reject nonzero/out-of-field input
	// before any packet or output allocation,including memory admission failures.
	p := rankShardOperandPlan{1, 16, 2, 2, 8, 0, 2, [2]rootShardLayout{{0, 1, 4, 2}, {1, 1, 4, 2}}}
	for _, input := range [][]uint32{{1, 1, 1, 0}, {1, 2, 0, 0}} {
		if _, err := routeShardedOperands(p, [2][]uint32{input, nil}, operandTestLink{}, nil); err == nil {
			t.Fatal("bad padding/range")
		}
	}
	p.MaxLocalWords = 1
	if _, err := routeShardedOperands(p, [2][]uint32{{1, 1, 0, 0}, nil}, operandTestLink{}, nil); err == nil {
		t.Fatal("memory cap")
	}
}
