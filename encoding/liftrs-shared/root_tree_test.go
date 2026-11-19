package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

type rootMail struct {
	rank  int
	boxes map[[2]int]chan []uint32
}

func (m rootMail) SendRoot(peer int, p []uint32, max int) error {
	select {
	case m.boxes[[2]int{m.rank, peer}] <- append([]uint32(nil), p...):
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("test tree send timeout")
	}
}
func (m rootMail) ReceiveRoot(peer, max int) ([]uint32, error) {
	select {
	case p := <-m.boxes[[2]int{peer, m.rank}]:
		return p, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("test tree receive timeout")
	}
}
func binaryRootProduct(a, b []uint32) ([]uint32, error) {
	out := make([]uint32, len(a)+len(b)-1)
	for i, x := range a {
		for j, y := range b {
			out[i+j] ^= x & y
		}
	}
	return out, nil
}

func TestRootTreeOwnership(t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 4, 5, 7, 8, 16, 64} {
		boxes := map[[2]int]chan []uint32{}
		inputs := make([][]uint32, ranks)
		expected := []uint32{1}
		words := 1
		for rank := 0; rank < ranks; rank++ {
			inputs[rank] = []uint32{1}
			if rank%3 != 0 {
				inputs[rank] = make([]uint32, rank+2)
				inputs[rank][0] = 1
				inputs[rank][rank+1] = 1
			}
			expected, _ = binaryRootProduct(expected, inputs[rank])
			words += len(inputs[rank]) - 1
			steps, err := rootTreeSteps(rank, ranks)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range steps {
				if step.Send {
					boxes[[2]int{rank, step.Peer}] = make(chan []uint32, 1)
				}
			}
		}
		results := make([][]uint32, ranks)
		failures := make([]error, ranks)
		var group sync.WaitGroup
		var mu sync.Mutex
		merges := 0
		for rank := 0; rank < ranks; rank++ {
			group.Add(1)
			go func(rank int) {
				defer group.Done()
				results[rank], failures[rank] = reduceRootProduct(rank, ranks, words, 2, inputs[rank], rootMail{rank, boxes},
					func(a, b []uint32) ([]uint32, error) {
						mu.Lock()
						merges++
						mu.Unlock()
						return binaryRootProduct(a, b)
					}, nil)
			}(rank)
		}
		group.Wait()
		for rank, err := range failures {
			if err != nil {
				t.Fatal(ranks, rank, err)
			}
		}
		if merges != ranks-1 || !reflect.DeepEqual(results[0], expected) {
			t.Fatal("tree coverage/final", ranks, merges)
		}
		for rank := 1; rank < ranks; rank++ {
			if results[rank] != nil {
				t.Fatal("sender retained ownership", rank)
			}
		}
	}
}

type badRootLink struct {
	value []uint32
	err   error
}

func (b badRootLink) SendRoot(int, []uint32, int) error      { return b.err }
func (b badRootLink) ReceiveRoot(int, int) ([]uint32, error) { return b.value, b.err }
func TestRootTreeAdmissionAndFailure(t *testing.T) {
	for _, pair := range [][2]int{{-1, 4}, {4, 4}, {0, 0}, {0, 65}} {
		if _, err := rootTreeSteps(pair[0], pair[1]); err == nil {
			t.Fatal("rank accepted", pair)
		}
	}
	if strconv.IntSize == 64 {
		wide := uint64(1) << 34
		result, err := reduceRootProduct(0, 1, int(wide), 2, []uint32{1}, badRootLink{}, binaryRootProduct, nil)
		if err != nil || !reflect.DeepEqual(result, []uint32{1}) {
			t.Fatal("64bit admitted extent rejected", err)
		}
	}
	injected := errors.New("injected transport")
	for _, link := range []badRootLink{{nil, injected}, {[]uint32{0, 1}, nil}, {[]uint32{1, 2}, nil}, {[]uint32{1, 1, 1}, nil}} {
		if _, err := reduceRootProduct(0, 2, 2, 2, []uint32{1, 1}, link, binaryRootProduct, nil); err == nil {
			t.Fatal("bad peer accepted")
		}
	}
	if _, err := reduceRootProduct(1, 2, 2, 2, []uint32{1}, badRootLink{err: injected}, binaryRootProduct, nil); !errors.Is(err, injected) {
		t.Fatal("send error lost", err)
	}
	if _, err := reduceRootProduct(0, 2, 3, 2, []uint32{1, 1}, badRootLink{value: []uint32{1, 1}},
		func([]uint32, []uint32) ([]uint32, error) { return []uint32{1}, nil }, nil); err == nil {
		t.Fatal("wrong merge degree accepted")
	}
}

func TestRootRoutingPackets(t *testing.T) {
	support := make([]int, 255)
	for j := range support {
		support[j] = j
	}
	before := append([]int(nil), support...)
	for _, workers := range []int{1, 2, 3, 4, 7, 16, 64} {
		for _, capacity := range []int{2, 3, 17, 16384} {
			parts := make([][]int, workers)
			err := routeRootSupport(supportSorted, support, 256, 16, workers, capacity, func(owner int, packet []uint64) error {
				if len(packet) == 0 || len(packet) > capacity {
					t.Fatal("packet bounds")
				}
				for _, index := range packet {
					parts[owner] = append(parts[owner], int(index))
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			seen := map[int]bool{}
			low, high := 255, 0
			for _, part := range parts {
				sort.Ints(part)
				if len(part) > 0 {
					if err := validateExponentSupportWith(supportSorted, part, 256, 16); err != nil {
						t.Fatal(err)
					}
				}
				for _, x := range part {
					if seen[x] {
						t.Fatal("duplicate owner", x)
					}
					seen[x] = true
				}
				if len(part) < low {
					low = len(part)
				}
				if len(part) > high {
					high = len(part)
				}
			}
			if len(seen) != 255 || high-low > 2 || !reflect.DeepEqual(support, before) {
				t.Fatal("routing coverage/balance/source")
			}
			if workers == 4 {
				// Golden per-owner uint64 index hashes from the pre-refactor
				// greedy router in2018219; independent of the new wrapper.
				golden := []string{"5a586ad0288fa12af83ea4176293115da604935cb911a0cefe376915f8112c1d", "d080f3bb7563ffe33a8aae0f6fcac2b7ac2c752bd1594f3e31f7419f4f23280b", "5659ffd6144d369d4616753002ef32bcb45b0a83978d2ece6b06b322e693f377", "8e7455f160f770721f2c3ea7403f9d7105720689f8ba030df000532eb363f3a7"}
				for rank, part := range parts {
					h := sha256.New()
					for _, x := range part {
						if err := binary.Write(h, binary.LittleEndian, uint64(x)); err != nil {
							t.Fatal(err)
						}
					}
					if fmt.Sprintf("%x", h.Sum(nil)) != golden[rank] {
						t.Fatal("owner order changed", rank)
					}
				}
			}
		}
	}
	injected := errors.New("send failure")
	if err := routeRootSupport(supportSorted, support, 256, 16, 4, 2, func(int, []uint64) error { return injected }); !errors.Is(err, injected) {
		t.Fatal("routing error lost")
	}
	if err := routeRootSupport(supportSorted, []int{0, 1}, 256, 16, 4, 2, func(int, []uint64) error { t.Fatal("invalid support emitted"); return nil }); err == nil {
		t.Fatal("nonclosed support accepted")
	}
}
