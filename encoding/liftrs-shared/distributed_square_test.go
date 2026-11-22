package liftrs

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestRankSquareAndReciprocalSchedule(t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 5} {
		for _, domain := range []uint64{2, 7, 32} {
			for _, keep := range []uint64{1, domain - 1, domain} {
				link := operandTestLink{make(map[[2]int]chan operandMessage)}
				for a := 0; a < ranks; a++ {
					for b := 0; b < ranks; b++ {
						if a != b {
							link.wire[[2]int{a, b}] = make(chan operandMessage)
						}
					}
				}
				outputs := make([][]uint32, ranks)
				errors := make([]error, ranks)
				var wait sync.WaitGroup
				for rank := 0; rank < ranks; rank++ {
					wait.Add(1)
					go func(rank int) {
						defer wait.Done()
						first, words := rankBand(domain, rank, ranks)
						input := make([]uint32, words)
						for j := range input {
							input[j] = uint32(first + uint64(j) + 1)
						}
						before := append([]uint32{}, input...)
						outputs[rank], errors[rank] = squarePolynomialShard(rankSquarePlan{1, domain, 2, 1 << 16, keep, domain, rank, ranks}, input, func(v uint32) uint32 { return v * v }, link, nil)
						if !reflect.DeepEqual(append([]uint32{}, input...), before) {
							errors[rank] = fmt.Errorf("source mutation")
						}
					}(rank)
				}
				wait.Wait()
				for rank, e := range errors {
					if e != nil {
						t.Fatal(e)
					}
					first, _ := rankBand(domain, rank, ranks)
					for j, v := range outputs[rank] {
						k := first + uint64(j)
						var want uint32
						if k < keep && k%2 == 0 {
							want = uint32(k/2 + 1)
							want *= want
						}
						if v != want {
							t.Fatal("square coordinate", ranks, domain, keep, rank, j)
						}
					}
				}
			}
		}
	}
	var sizes []int
	if err := reciprocalSchedule(1, 13, func(n int) error { sizes = append(sizes, n); return nil }); err != nil || !reflect.DeepEqual(sizes, []int{2, 4, 8, 13}) {
		t.Fatal(sizes, err)
	}
	stop := fmt.Errorf("stop")
	if err := reciprocalSchedule(4, 13, func(int) error { return stop }); err != stop {
		t.Fatal("refinement error lost")
	}
}
