package liftrs

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type transposeMail[Word transposeWord] struct {
	packet transposePacket
	values []Word
}
type transposeMailbox[Word transposeWord] struct {
	rank  int
	boxes map[[2]int]chan transposeMail[Word]
}

func (m transposeMailbox[Word]) SendTile(peer int, p transposePacket, values []Word) error {
	// Unbuffered channels deliberately model blocking rendezvous, not eager MPI.
	select {
	case m.boxes[[2]int{m.rank, peer}] <- transposeMail[Word]{p, append([]Word(nil), values...)}:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("test send rendezvous timeout")
	}
}
func (m transposeMailbox[Word]) ReceiveTile(peer int, p transposePacket, values []Word) error {
	select {
	case got := <-m.boxes[[2]int{peer, m.rank}]:
		if got.packet != p || len(got.values) != len(values) {
			return fmt.Errorf("test packet mismatch")
		}
		copy(values, got.values)
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("test receive rendezvous timeout")
	}
}

func checkRankTranspose[Word transposeWord](t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 4, 5, 7, 8, 16, 64} {
		for _, shape := range [][2]uint64{{1, 1}, {1, 3}, {3, 1}, {7, 11}, {16, 32}, {3, 17}} {
			rows, columns := shape[0], shape[1]
			full := make([]Word, rows*columns)
			for j := range full {
				full[j] = Word(uint64(j) * 0x9e3779b97f4a7c15)
			}
			// Independent serial transpose, with no shared packing/index helper.
			expected := make([]Word, len(full))
			for row := uint64(0); row < rows; row++ {
				for col := uint64(0); col < columns; col++ {
					expected[col*rows+row] = full[row*columns+col]
				}
			}
			boxes := map[[2]int]chan transposeMail[Word]{}
			for a := 0; a < ranks; a++ {
				for b := 0; b < ranks; b++ {
					if a != b {
						boxes[[2]int{a, b}] = make(chan transposeMail[Word])
					}
				}
			}
			results := make([][]Word, ranks)
			failures := make([]error, ranks)
			inputs := make([][]Word, ranks)
			var group sync.WaitGroup
			for rank := 0; rank < ranks; rank++ {
				p, err := newRankTransposePlan(11, rows, columns, rank, ranks, rows*columns, 7)
				if err != nil {
					t.Fatal(err)
				}
				inputs[rank] = append([]Word{}, full[p.InputFirst*columns:(p.InputFirst+p.InputRows)*columns]...)
				group.Add(1)
				go func(rank int, p rankTransposePlan) {
					defer group.Done()
					results[rank], failures[rank] = transposeRank(p, inputs[rank], transposeMailbox[Word]{rank, boxes}, nil)
				}(rank, p)
			}
			group.Wait()
			var joined []Word
			for rank := 0; rank < ranks; rank++ {
				if failures[rank] != nil {
					t.Fatal(ranks, shape, rank, failures[rank])
				}
				joined = append(joined, results[rank]...)
				first, count := rankBand(rows, rank, ranks)
				if !reflect.DeepEqual(inputs[rank], full[first*columns:(first+count)*columns]) {
					t.Fatal("input mutated")
				}
			}
			if !reflect.DeepEqual(joined, expected) {
				t.Fatal("transpose differs", ranks, shape)
			}
		}
	}
}
func TestRankTranspose32(t *testing.T) { checkRankTranspose[uint32](t) }
func TestRankTranspose64(t *testing.T) { checkRankTranspose[uint64](t) }

type failingTransposeLink struct{ err error }

func (f failingTransposeLink) SendTile(int, transposePacket, []uint32) error    { return f.err }
func (f failingTransposeLink) ReceiveTile(int, transposePacket, []uint32) error { return f.err }
func TestRankTransposeAdmissionAndFailure(t *testing.T) {
	for _, args := range [][7]uint64{{0, 4, 4, 0, 2, 8, 2}, {1, 0, 4, 0, 2, 8, 2}, {1, ^uint64(0), 2, 0, 2, ^uint64(0), 2}, {1, 4, 4, 2, 2, 8, 2}, {1, 4, 4, 0, 65, 8, 2}, {1, 4, 4, 0, 2, 7, 2}, {1, 4, 4, 0, 2, 8, 0}} {
		if _, err := newRankTransposePlan(args[0], args[1], args[2], int(args[3]), int(args[4]), args[5], args[6]); err == nil {
			t.Fatal("bad layout admitted", args)
		}
	}
	p, err := newRankTransposePlan(1, 4, 4, 0, 2, 8, 2)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected tile transport")
	if _, err = transposeRank(p, make([]uint32, 8), failingTransposeLink{injected}, nil); !errors.Is(err, injected) {
		t.Fatal("transport failure lost", err)
	}
	p.OutputWords++
	if _, err = transposeRank(p, make([]uint32, 8), failingTransposeLink{}, nil); err == nil {
		t.Fatal("modified plan accepted")
	}
	// Admission arithmetic must stay64-bit without allocating the projected grid.
	p, err = newRankTransposePlan(1, 1<<18, 1<<18, 15, 16, 1<<32, 16384)
	if err != nil || p.InputWords != 1<<32 || p.OutputWords != 1<<32 {
		t.Fatal("wide layout count", p, err)
	}
}

// Storage failures must stop the common schedule before reporting a complete
// target. A GPU failure has the same contract; no partial shard may be published.
type failedTransposeStorage struct {
	phase string
	calls []string
}

func (s *failedTransposeStorage) fail(phase string) error {
	s.calls = append(s.calls, phase)
	if s.phase == phase {
		return errors.New("storage " + phase)
	}
	return nil
}
func (s *failedTransposeStorage) Local(rankTransposePlan) error { return s.fail("local") }
func (s *failedTransposeStorage) Pack(rankTransposePlan, int, uint64, []uint32) error {
	return s.fail("pack")
}
func (s *failedTransposeStorage) Scatter(rankTransposePlan, int, uint64, []uint32) error {
	return s.fail("scatter")
}

type discardTransposeLink struct{}

func (discardTransposeLink) SendTile(int, transposePacket, []uint32) error    { return nil }
func (discardTransposeLink) ReceiveTile(int, transposePacket, []uint32) error { return nil }
func TestRankTransposeStorageFailure(t *testing.T) {
	for _, phase := range []string{"local", "pack", "scatter"} {
		p, err := newRankTransposePlan(1, 3, 5, 0, 2, 15, 2)
		if err != nil {
			t.Fatal(err)
		}
		s := &failedTransposeStorage{phase: phase}
		complete := false
		err = transposeStored[uint32](p, s, discardTransposeLink{}, func(phase string, _, _ uint64) { complete = complete || phase == "transpose-output" })
		if err == nil || err.Error() != "storage "+phase || complete || s.calls[len(s.calls)-1] != phase {
			t.Fatal(phase, s.calls, err, complete)
		}
	}
}
