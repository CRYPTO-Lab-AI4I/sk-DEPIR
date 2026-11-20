package liftrs

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

// A reversible, position-dependent permutation makes row/column order matter.
// It tests orchestration independently of finite-field/GPU arithmetic.
func vectorReference[Word transposeWord](v []Word, columns, inverse bool, first uint64) {
	mask := func(k int) Word {
		role := uint64(0x79)
		if columns {
			role = 0x163
		}
		return Word((first+1)*role + uint64(k)*0x9e3779b9)
	}
	original := append([]Word(nil), v...)
	for k := range v {
		if inverse {
			v[k] = original[(k+len(v)-1)%len(v)] ^ mask((k+len(v)-1)%len(v))
		} else {
			v[k] = original[(k+1)%len(v)] ^ mask(k)
		}
	}
}
func testRankMatrixVectors[Word transposeWord](t *testing.T) {
	for _, shape := range [][2]uint64{{2, 3}, {7, 11}, {16, 8}} {
		for _, ranks := range []int{1, 2, 3, 8, 16} {
			rows, cols := shape[0], shape[1]
			full := make([]Word, rows*cols)
			for j := range full {
				full[j] = Word(uint64(j) * 0x9e3779b97f4a7c15)
			}
			expected := append([]Word(nil), full...)
			for c := uint64(0); c < cols; c++ {
				v := make([]Word, rows)
				for r := range v {
					v[r] = expected[uint64(r)*cols+c]
				}
				vectorReference(v, true, false, c)
				for r := range v {
					expected[uint64(r)*cols+c] = v[r]
				}
			}
			for r := uint64(0); r < rows; r++ {
				vectorReference(expected[r*cols:(r+1)*cols], false, false, r)
			}
			shards := make([][]Word, ranks)
			for rank := range shards {
				first, count := rankBand(rows, rank, ranks)
				shards[rank] = append([]Word(nil), full[first*cols:(first+count)*cols]...)
			}
			for pass := 0; pass < 2; pass++ {
				boxes := map[[2]int]chan transposeMail[Word]{}
				for a := 0; a < ranks; a++ {
					for b := 0; b < ranks; b++ {
						if a != b {
							boxes[[2]int{a, b}] = make(chan transposeMail[Word])
						}
					}
				}
				outputs := make([][]Word, ranks)
				failures := make([]error, ranks)
				var group sync.WaitGroup
				for rank := 0; rank < ranks; rank++ {
					group.Add(1)
					go func(rank int) {
						defer group.Done()
						p, err := newRankTransposePlan(uint64(1+2*pass), rows, cols, rank, ranks, rows*cols, 7)
						if err != nil {
							failures[rank] = err
							return
						}
						apply := func(tile []Word, columns, inverse bool, first, count uint64) error {
							length := uint64(len(tile)) / count
							for j := uint64(0); j < count; j++ {
								v := make([]Word, length)
								for k := range v {
									v[k] = tile[uint64(k)*count+j]
								}
								vectorReference(v, columns, inverse, first+j)
								for k := range v {
									tile[uint64(k)*count+j] = v[k]
								}
							}
							return nil
						}
						outputs[rank], failures[rank] = rankMatrixVectors(p, shards[rank], 2*max(rows, cols), pass == 1, transposeMailbox[Word]{rank, boxes}, apply, nil)
					}(rank)
				}
				group.Wait()
				for rank := range outputs {
					if failures[rank] != nil {
						t.Fatal(shape, ranks, pass, rank, failures[rank])
					}
					first, count := rankBand(rows, rank, ranks)
					want := expected
					if pass == 1 {
						want = full
					}
					if !reflect.DeepEqual(outputs[rank], want[first*cols:(first+count)*cols]) {
						t.Fatal("matrix-vector ordering/ownership", shape, ranks, pass, rank)
					}
				}
				shards = outputs
			}
		}
	}
}
func TestRankMatrixVectors(t *testing.T) {
	testRankMatrixVectors[uint32](t)
	testRankMatrixVectors[uint64](t)
}
func TestRankVectorAdmission(t *testing.T) {
	p, _ := newRankTransposePlan(1, 4, 4, 0, 1, 16, 7)
	input := make([]uint32, 16)
	fail := errors.New("vector backend")
	apply := func([]uint32, bool, bool, uint64, uint64) error { return fail }
	if _, err := rankMatrixVectors(p, input, 4, false, nil, apply, nil); !errors.Is(err, fail) {
		t.Fatal("backend error", err)
	}
	if _, err := rankMatrixVectors(p, input, 3, false, nil, apply, nil); err == nil {
		t.Fatal("undersized tile")
	}
	if _, err := rankMatrixVectors(p, input, 4, false, nil, nil, nil); err == nil {
		t.Fatal("nil callback")
	}
	p.Operation = ^uint64(0)
	if _, err := rankMatrixVectors(p, input, 4, false, nil, apply, nil); err == nil {
		t.Fatal("operation wrap")
	}
}
