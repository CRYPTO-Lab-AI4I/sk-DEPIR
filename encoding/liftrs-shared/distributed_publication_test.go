package liftrs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

type publicationMessage[Word transposeWord] struct {
	p      operandPacket
	values []Word
}
type publicationTestLink[Word transposeWord] struct {
	wire map[int]chan publicationMessage[Word]
}

func (l publicationTestLink[Word]) SendOperand(p operandPacket, v []Word) error {
	l.wire[p.Source] <- publicationMessage[Word]{p, append([]Word(nil), v...)}
	return nil
}
func (l publicationTestLink[Word]) ReceiveOperand(p operandPacket, v []Word) error {
	m := <-l.wire[p.Source]
	if m.p != p || len(v) != len(m.values) {
		return fmt.Errorf("publication descriptor")
	}
	copy(v, m.values)
	return nil
}

func publicationCases[Word transposeWord](t *testing.T, limit uint64, pattern func(uint64) Word) {
	for _, ranks := range []int{1, 2, 3, 5} {
		for _, count := range []uint64{1, 6, 7} {
			link := publicationTestLink[Word]{make(map[int]chan publicationMessage[Word])}
			for rank := 1; rank < ranks; rank++ {
				link.wire[rank] = make(chan publicationMessage[Word])
			}
			var actual bytes.Buffer
			var wait sync.WaitGroup
			errors := make([]error, ranks)
			for rank := 0; rank < ranks; rank++ {
				wait.Add(1)
				go func(rank int) {
					defer wait.Done()
					first, words := rankBand(7, rank, ranks)
					input := make([]Word, words)
					for j := range input {
						if k := first + uint64(j); k < count {
							input[j] = pattern(k)
						}
					}
					before := append([]Word{}, input...)
					errors[rank] = streamRankValues(rankStreamPlan{1, 7, count, 2, limit, 7, rank, ranks}, input, &actual, link)
					if !reflect.DeepEqual(append([]Word{}, input...), before) {
						errors[rank] = fmt.Errorf("source mutation")
					}
				}(rank)
			}
			wait.Wait()
			for _, err := range errors {
				if err != nil {
					t.Fatal(err)
				}
			}
			expected := make([]Word, count)
			for j := range expected {
				expected[j] = pattern(uint64(j))
			}
			var encoded bytes.Buffer
			if err := binary.Write(&encoded, binary.LittleEndian, expected); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual.Bytes(), encoded.Bytes()) {
				t.Fatal("canonical stream", ranks, count)
			}
		}
	}
}
func TestRankStreamCanonicalPublication(t *testing.T) {
	publicationCases[uint32](t, 1<<16, func(i uint64) uint32 { return uint32(i * 37) })
	publicationCases[uint64](t, 1<<40, func(i uint64) uint64 { return (1 << 35) + i*117 })
}

type failedPublicationWriter struct{}

func (failedPublicationWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("injected publication I/O")
}
func TestRankStreamPublicationErrors(t *testing.T) {
	p := rankStreamPlan{Operation: 1, Domain: 2, Count: 2, TileWords: 1, Limit: 16, MaxLocalWords: 2, Rank: 0, Ranks: 1}
	if err := streamRankValues[uint32](p, []uint32{1, 2}, failedPublicationWriter{}, nil); err == nil {
		t.Fatal("writer failure ignored")
	}
	p.Count = 1
	if err := streamRankValues[uint32](p, []uint32{1, 2}, &bytes.Buffer{}, nil); err == nil {
		t.Fatal("nonzero padding accepted")
	}
}

// Callback sources preserve wide high bits, exact local read order and canonical
// global ordering, including uneven/empty ownership. Each element is read once.
func publicationReadCases[Word transposeWord](t *testing.T, limit uint64, pattern func(uint64) Word) {
	t.Helper()
	for _, ranks := range []int{1, 2, 3, 5, 9} {
		link := publicationTestLink[Word]{make(map[int]chan publicationMessage[Word])}
		for rank := 1; rank < ranks; rank++ {
			link.wire[rank] = make(chan publicationMessage[Word])
		}
		var output bytes.Buffer
		var wait sync.WaitGroup
		errors := make([]error, ranks)
		for rank := 0; rank < ranks; rank++ {
			wait.Add(1)
			go func(rank int) {
				defer wait.Done()
				first, words := rankBand(7, rank, ranks)
				completed := uint64(0)
				read := func(offset uint64, values []Word) error {
					if offset != completed || len(values) == 0 || len(values) > 2 {
						return fmt.Errorf("source order/bound")
					}
					for j := range values {
						values[j] = pattern(first + offset + uint64(j))
					}
					completed += uint64(len(values))
					return nil
				}
				errors[rank] = streamRankReadValues(rankStreamPlan{1, 7, 7, 2, limit, words, rank, ranks}, read, &output, link)
				if completed != words {
					errors[rank] = fmt.Errorf("source read count")
				}
			}(rank)
		}
		wait.Wait()
		for _, err := range errors {
			if err != nil {
				t.Fatal(err)
			}
		}
		expected := []Word{}
		for j := uint64(0); j < 7; j++ {
			expected = append(expected, pattern(j))
		}
		var encoded bytes.Buffer
		if err := binary.Write(&encoded, binary.LittleEndian, expected); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(output.Bytes(), encoded.Bytes()) {
			t.Fatal("source publication", ranks)
		}
	}
}

func TestRankStreamReadCanonicalPublication(t *testing.T) {
	publicationReadCases[uint32](t, 1<<16, func(i uint64) uint32 { return uint32(i * 37) })
	publicationReadCases[uint64](t, 1<<40, func(i uint64) uint64 { return (1 << 35) + i*117 })
}

func TestRankStreamReadErrors(t *testing.T) {
	p := rankStreamPlan{1, 2, 2, 1, 16, 2, 0, 1}
	for _, invalid := range []rankStreamPlan{
		{1, 2, 1, 1, 16, 2, 0, 1}, // No implicit unchecked padding.
		{1, 2, 2, 1, 16, 1, 0, 1}, // Reject resident ownership admission first.
	} {
		called := false
		if err := streamRankReadValues[uint32](invalid, func(uint64, []uint32) error { called = true; return nil }, &bytes.Buffer{}, nil); err == nil || called {
			t.Fatal("invalid source admission")
		}
	}
	if err := streamRankReadValues[uint32](p, nil, &bytes.Buffer{}, nil); err == nil {
		t.Fatal("nil source")
	}
	if err := streamRankReadValues[uint32](p, func(uint64, []uint32) error { return fmt.Errorf("injected source read") }, &bytes.Buffer{}, nil); err == nil {
		t.Fatal("source failure")
	}
	if err := streamRankReadValues[uint32](p, func(_ uint64, v []uint32) error { v[0] = 16; return nil }, &bytes.Buffer{}, nil); err == nil {
		t.Fatal("source field range")
	}
	if err := streamRankReadValues[uint32](p, func(_ uint64, v []uint32) error { v[0] = 1; return nil }, failedPublicationWriter{}, nil); err == nil {
		t.Fatal("source writer failure")
	}
}
