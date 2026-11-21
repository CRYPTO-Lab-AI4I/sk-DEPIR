package liftrs

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type taylorMail[Word transposeWord] struct {
	packet taylorPacket
	data   []Word
}
type taylorBarrier struct {
	mu                         sync.Mutex
	wake                       *sync.Cond
	arrived, generation, ranks int
}

func newTaylorBarrier(ranks int) *taylorBarrier {
	b := &taylorBarrier{ranks: ranks}
	b.wake = sync.NewCond(&b.mu)
	return b
}
func (b *taylorBarrier) wait() {
	b.mu.Lock()
	defer b.mu.Unlock()
	g := b.generation
	b.arrived++
	if b.arrived == b.ranks {
		b.arrived = 0
		b.generation++
		b.wake.Broadcast()
		return
	}
	for g == b.generation {
		b.wake.Wait()
	}
}

type taylorMailbox[Word transposeWord] struct {
	rank    int
	boxes   map[[2]int]chan taylorMail[Word]
	barrier *taylorBarrier
}

func (m taylorMailbox[Word]) Barrier() { m.barrier.wait() }
func (m taylorMailbox[Word]) SendWindow(peer int, p taylorPacket, data []Word) error {
	select {
	case m.boxes[[2]int{m.rank, peer}] <- taylorMail[Word]{p, append([]Word(nil), data...)}:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("Taylor send timeout")
	}
}
func (m taylorMailbox[Word]) ReceiveWindow(peer int, p taylorPacket, data []Word) error {
	select {
	case got := <-m.boxes[[2]int{peer, m.rank}]:
		if got.packet != p || len(got.data) != len(data) {
			return fmt.Errorf("Taylor descriptor")
		}
		copy(data, got.data)
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("Taylor receive timeout")
	}
}
func checkRankTaylor[Word transposeWord](t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 5, 8, 16, 32} {
		for _, capacity := range []uint64{6, 13, 64} {
			const n, cols = 128, 8
			input := make([]Word, n)
			for j := range input {
				input[j] = Word(uint64(j) * 0x9e3779b97f4a7c15)
			}
			expected := append([]Word(nil), input...)
			// Whole-domain serial arithmetic uses global offsets, not packed snapshots.
			for level := 3; level >= 0; level-- {
				half := uint64(cols) << uint(level)
				degrees := []uint64{1 << uint(level), 4 << uint(level)}
				for block := uint64(0); block < n; block += 2 * half {
					for j := uint64(0); j < half; j++ {
						for _, d := range degrees {
							if d > j {
								expected[block+half+j] ^= expected[block+2*half+j-d]
							}
						}
					}
					for j := uint64(0); j < half; j++ {
						for _, d := range degrees {
							if d <= j {
								expected[block+j] ^= expected[block+half+j-d]
							}
						}
					}
				}
			}
			shards := make([][]Word, ranks)
			for rank := range shards {
				first, count := rankBand(n/cols, rank, ranks)
				shards[rank] = append([]Word(nil), input[first*cols:(first+count)*cols]...)
			}
			boxes := map[[2]int]chan taylorMail[Word]{}
			for a := 0; a < ranks; a++ {
				for b := 0; b < ranks; b++ {
					if a != b {
						boxes[[2]int{a, b}] = make(chan taylorMail[Word])
					}
				}
			}
			barrier := newTaylorBarrier(ranks)
			for pass := 0; pass < 2; pass++ {
				failures := make([]error, ranks)
				var group sync.WaitGroup
				for rank := 0; rank < ranks; rank++ {
					group.Add(1)
					go func(rank int) {
						defer group.Done()
						base := rankTaylorPlan{Operation: uint64(1 + pass*8), Domain: n, Columns: cols, Capacity: capacity, Rank: rank, Ranks: ranks}
						degrees := func(level int) ([]uint64, error) { return []uint64{1 << uint(level), 4 << uint(level)}, nil }
						apply := func(snapshot []Word, level int, q bool, block, first, count uint64) error {
							terms, _ := degrees(level)
							for j := uint64(0); j < count; j++ {
								for slot, d := range terms {
									if (q && d > first+j) || (!q && d <= first+j) {
										snapshot[j] ^= snapshot[uint64(slot+1)*count+j]
									}
								}
							}
							return nil
						}
						failures[rank] = rankTaylorConvert(base, shards[rank], pass == 1, degrees, taylorMailbox[Word]{rank, boxes, barrier}, apply, nil)
					}(rank)
				}
				group.Wait()
				want := expected
				if pass == 1 {
					want = input
				}
				for rank := range shards {
					if failures[rank] != nil {
						t.Fatal(ranks, capacity, pass, rank, failures[rank])
					}
					first, count := rankBand(n/cols, rank, ranks)
					if count > 0 && !reflect.DeepEqual(shards[rank], want[first*cols:(first+count)*cols]) {
						t.Fatal("Taylor ownership/window/barrier", ranks, capacity, pass, rank)
					}
				}
			}
		}
	}
}
func TestRankTaylorConversion(t *testing.T) { checkRankTaylor[uint32](t); checkRankTaylor[uint64](t) }
func TestRankTaylorAdmission(t *testing.T) {
	p := rankTaylorPlan{Operation: 1, Domain: 1 << 36, Columns: 1 << 18, Capacity: 64, Rank: 8, Ranks: 16, Level: 17, Quotient: true, Degrees: []uint64{1 << 17}}
	half, mi, ma, span, err := p.validate()
	if err != nil {
		t.Fatal(err)
	}
	cursor := newTaylorCursor(p, 8, half, mi, ma, span)
	ticket, ok := cursor.take()
	if !ok || ticket.Block+half+ticket.First <= 1<<32 {
		t.Fatal("wide coordinates")
	}
	p.Degrees = []uint64{half}
	if _, _, _, _, err = p.validate(); err == nil {
		t.Fatal("dependent source range")
	}
	p = rankTaylorPlan{Operation: 1, Domain: 16, Columns: 4, Capacity: 8, Rank: 0, Ranks: 1, Degrees: []uint64{1}}
	boom := errors.New("Taylor backend")
	link := taylorMailbox[uint32]{0, nil, newTaylorBarrier(1)}
	err = rankTaylorPass(p, make([]uint32, 16), make([]uint32, 8), link, func([]uint32, int, bool, uint64, uint64, uint64) error { return boom }, nil)
	if !errors.Is(err, boom) {
		t.Fatal("backend propagation", err)
	}
}

// Independent small-field oracle: explicit subspace polynomial, direct Horner
// evaluation and Lagrange interpolation. No native/GPU transform metadata or
// recurrence is used here; this checks the complete stage ordering before HPC.
func checkRankAdditive[Word transposeWord](t *testing.T) {
	_, f, err := canonicalScalingCodeHeader(4, 3, 1, 256)
	if err != nil {
		t.Fatal(err)
	}
	multiplyLinear := func(poly []uint64, root uint64) []uint64 {
		out := make([]uint64, len(poly)+1)
		for k, c := range poly {
			out[k] ^= f.mul(c, root)
			out[k+1] ^= c
		}
		return out
	}
	evaluate := func(poly []uint64, x uint64) uint64 {
		value := uint64(0)
		for k := len(poly) - 1; k >= 0; k-- {
			value = f.mul(value, x) ^ poly[k]
		}
		return value
	}
	for _, dim := range []int{4, 5, 6} {
		n := uint64(1) << uint(dim)
		cols := uint64(1) << uint(dim/2)
		rows := n / cols
		subspace := []uint64{1}
		for root := uint64(0); root < cols; root++ {
			subspace = multiplyLinear(subspace, root)
		}
		gamma := make([]uint64, dim-dim/2)
		for bit := range gamma {
			gamma[bit] = evaluate(subspace, uint64(1)<<uint(dim/2+bit))
		}
		degrees := func(level int) ([]uint64, error) {
			var out []uint64
			for j, c := range subspace {
				if j > 0 && uint64(j) < cols && c != 0 {
					out = append(out, uint64(j)<<uint(level))
				}
			}
			return out, nil
		}
		vectors := func(tile []Word, columns, inverse bool, first, count uint64) error {
			length := uint64(len(tile)) / count
			for j := uint64(0); j < count; j++ {
				values := make([]uint64, length)
				points := make([]uint64, length)
				for k := uint64(0); k < length; k++ {
					values[k] = uint64(tile[k*count+j])
					if columns {
						for bit, b := range gamma {
							if (k>>uint(bit))&1 != 0 {
								points[k] ^= b
							}
						}
					} else {
						points[k] = (first+j)*cols ^ k
					}
				}
				out := make([]uint64, length)
				if !inverse {
					for k, x := range points {
						out[k] = evaluate(values, x)
					}
				} else {
					for k, x := range points {
						poly := []uint64{1}
						denominator := uint64(1)
						for l, y := range points {
							if k != l {
								poly = multiplyLinear(poly, y)
								denominator = f.mul(denominator, x^y)
							}
						}
						scale := f.mul(values[k], f.pow(denominator, 254))
						for l, c := range poly {
							out[l] ^= f.mul(scale, c)
						}
					}
				}
				for k, v := range out {
					tile[uint64(k)*count+j] = Word(v)
				}
			}
			return nil
		}
		update := func(snapshot []Word, level int, q bool, block, first, count uint64) error {
			terms, _ := degrees(level)
			for j := uint64(0); j < count; j++ {
				for slot, d := range terms {
					if (q && d > first+j) || (!q && d <= first+j) {
						coefficient := f.pow(subspace[d>>uint(level)], uint64(1)<<uint(level))
						snapshot[j] ^= Word(f.mul(coefficient, uint64(snapshot[uint64(slot+1)*count+j])))
					}
				}
			}
			return nil
		}
		input := make([]Word, n)
		coefficients := make([]uint64, n)
		expected := make([]Word, n)
		for j := range input {
			coefficients[j] = uint64(j*79+1) & 255
			input[j] = Word(coefficients[j])
		}
		for point := range expected {
			expected[point] = Word(evaluate(coefficients, uint64(point)))
		}
		for _, ranks := range []int{1, 2, 3, 8} {
			shards := make([][]Word, ranks)
			for rank := range shards {
				first, count := rankBand(rows, rank, ranks)
				shards[rank] = append([]Word(nil), input[first*cols:(first+count)*cols]...)
			}
			for pass := 0; pass < 2; pass++ {
				transposes := map[[2]int]chan transposeMail[Word]{}
				windows := map[[2]int]chan taylorMail[Word]{}
				for a := 0; a < ranks; a++ {
					for b := 0; b < ranks; b++ {
						if a != b {
							transposes[[2]int{a, b}] = make(chan transposeMail[Word])
							windows[[2]int{a, b}] = make(chan taylorMail[Word])
						}
					}
				}
				barrier := newTaylorBarrier(ranks)
				outputs := make([][]Word, ranks)
				failures := make([]error, ranks)
				var group sync.WaitGroup
				for rank := range shards {
					group.Add(1)
					go func(rank int) {
						defer group.Done()
						p, e := newRankTransposePlan(uint64(1+pass*32), rows, cols, rank, ranks, n, 3)
						if e != nil {
							failures[rank] = e
							return
						}
						outputs[rank], failures[rank] = rankAdditiveTransform(p, shards[rank], 3*max(rows, cols), pass == 1, transposeMailbox[Word]{rank, transposes}, taylorMailbox[Word]{rank, windows, barrier}, degrees, vectors, update, nil)
					}(rank)
				}
				group.Wait()
				want := expected
				if pass == 1 {
					want = input
				}
				for rank := range outputs {
					if failures[rank] != nil {
						t.Fatal(dim, ranks, pass, rank, failures[rank])
					}
					first, count := rankBand(rows, rank, ranks)
					if count > 0 && !reflect.DeepEqual(outputs[rank], want[first*cols:(first+count)*cols]) {
						t.Fatal("complete additive/Horner", dim, ranks, pass, rank)
					}
				}
				shards = outputs
			}
			// Compose the same reference FFT callbacks into a full polynomial
			// product; schoolbook coefficients are independent of the schedule.
			a, bpoly := make([]Word, n), make([]Word, n)
			wantProduct := make([]Word, n)
			ac, bc := n/4+1, 3*n/4
			for j := uint64(0); j < ac; j++ {
				a[j] = Word((j*3 + 1) & 15)
			}
			for j := uint64(0); j < bc; j++ {
				bpoly[j] = Word((j*7 + 1) & 15)
			}
			for i := uint64(0); i < ac; i++ {
				for j := uint64(0); j < bc; j++ {
					wantProduct[i+j] ^= Word(f.mul(uint64(a[i]), uint64(bpoly[j])))
				}
			}
			transposes := map[[2]int]chan transposeMail[Word]{}
			windows := map[[2]int]chan taylorMail[Word]{}
			for x := 0; x < ranks; x++ {
				for y := 0; y < ranks; y++ {
					if x != y {
						transposes[[2]int{x, y}] = make(chan transposeMail[Word])
						windows[[2]int{x, y}] = make(chan taylorMail[Word])
					}
				}
			}
			barrier := newTaylorBarrier(ranks)
			outputs := make([][]Word, ranks)
			failures := make([]error, ranks)
			var group sync.WaitGroup
			for rank := 0; rank < ranks; rank++ {
				group.Add(1)
				go func(rank int) {
					defer group.Done()
					p, e := newRankTransposePlan(100, rows, cols, rank, ranks, n, 3)
					if e != nil {
						failures[rank] = e
						return
					}
					first, count := rankBand(rows, rank, ranks)
					left := append([]Word(nil), a[first*cols:(first+count)*cols]...)
					right := append([]Word(nil), bpoly[first*cols:(first+count)*cols]...)
					backend := rankProductBackend[Word]{Transpose: transposeMailbox[Word]{rank, transposes}, Windows: taylorMailbox[Word]{rank, windows, barrier}, Degrees: degrees, Vectors: vectors, Taylor: update, Multiply: func(left, right []Word) error {
						for j := range left {
							left[j] = Word(f.mul(uint64(left[j]), uint64(right[j])))
						}
						return nil
					}}
					outputs[rank], failures[rank] = rankPolynomialProduct(p, left, right, 3*max(rows, cols), backend, nil)
				}(rank)
			}
			group.Wait()
			for rank := range outputs {
				if failures[rank] != nil {
					t.Fatal("product backend", dim, ranks, rank, failures[rank])
				}
				first, count := rankBand(rows, rank, ranks)
				if count > 0 && !reflect.DeepEqual(outputs[rank], wantProduct[first*cols:(first+count)*cols]) {
					t.Fatal("distributed product/schoolbook", dim, ranks, rank)
				}
			}
		}
	}
}
func TestRankAdditiveTransform(t *testing.T) {
	checkRankAdditive[uint32](t)
	checkRankAdditive[uint64](t)
}

func TestRankAdditiveAdmission(t *testing.T) {
	p, _ := newRankTransposePlan(1, 4, 4, 0, 1, 16, 3)
	input := make([]uint32, 16)
	input[0] = 17
	if _, err := rankAdditiveTransform[uint32](p, input, 3, false, nil, nil, nil, nil, nil, nil); err == nil || input[0] != 17 {
		t.Fatal("invalid capacity mutated input", err)
	}
	base := rankTaylorPlan{Operation: 1, Domain: 16, Capacity: 8, Rank: 0, Ranks: 1}
	if err := rankTaylorConvert[uint32](base, input, false, nil, nil, nil, nil); err == nil {
		t.Fatal("zero columns")
	}
	p.Operation = ^uint64(0)
	if _, err := rankAdditiveTransform[uint32](p, input, 16, false, nil, nil, nil, nil, nil, nil); err == nil {
		t.Fatal("operation wrap")
	}
}

func TestRankProductAdmission(t *testing.T) {
	p, _ := newRankTransposePlan(1, 4, 4, 0, 1, 16, 3)
	a := make([]uint32, 32)
	backend := rankProductBackend[uint32]{Multiply: func([]uint32, []uint32) error { return nil }}
	if _, err := rankPolynomialProduct(p, a[:16], a[8:24], 16, backend, nil); err == nil {
		t.Fatal("overlap")
	}
	if _, err := rankPolynomialProduct(p, a[:16], a[16:31], 16, backend, nil); err == nil {
		t.Fatal("unequal extents")
	}
	p.Operation = ^uint64(0)
	if _, err := rankPolynomialProduct(p, a[:16], a[16:], 16, backend, nil); err == nil {
		t.Fatal("product operation wrap")
	}
}

// Storage errors must propagate through the common scheduler before publishing a
// completed target or crossing the successful end barrier.
type failedTaylorStorage struct{ phase string }

func (s failedTaylorStorage) Snapshot(rankTaylorPlan, taylorTicket, uint64, uint64, []uint32) error {
	return s.Read(0, nil)
}
func (s failedTaylorStorage) Read(uint64, []uint32) error {
	if s.phase == "read" {
		return fmt.Errorf("storage read")
	}
	return nil
}
func (s failedTaylorStorage) Send(int, taylorPacket, uint64, uint64, taylorLink[uint32]) error {
	if s.phase == "send" {
		return fmt.Errorf("storage send")
	}
	return nil
}
func (s failedTaylorStorage) Apply(rankTaylorPlan, taylorTicket, uint64, []uint32) error {
	return fmt.Errorf("storage apply")
}

type discardTaylorLink struct{ barriers int }

func (l *discardTaylorLink) SendWindow(int, taylorPacket, []uint32) error    { return nil }
func (l *discardTaylorLink) ReceiveWindow(int, taylorPacket, []uint32) error { return nil }
func (l *discardTaylorLink) Barrier()                                        { l.barriers++ }
func TestRankTaylorStorageErrors(t *testing.T) {
	for _, phase := range []string{"read", "apply", "send"} {
		p := rankTaylorPlan{Operation: 1, Domain: 16, Columns: 4, Capacity: 12, Rank: 0, Ranks: 1, Level: 1, Quotient: true, Degrees: []uint64{1, 2}}
		if phase == "send" {
			p.Rank = 1
			p.Ranks = 2
			p.Quotient = false
		}
		link := &discardTaylorLink{}
		err := rankTaylorStoredPass(p, make([]uint32, 12), failedTaylorStorage{phase}, link, nil)
		if err == nil || err.Error() != "storage "+phase || link.barriers != 1 {
			t.Fatal(phase, err, link.barriers)
		}
	}
}

func TestRankProductScheduleFailure(t *testing.T) {
	p, err := newRankTransposePlan(1, 4, 4, 0, 1, 16, 3)
	if err != nil {
		t.Fatal(err)
	}
	plans, err := rankProductPlans(p, 12)
	if err != nil {
		t.Fatal(err)
	}
	for _, fail := range []string{"right", "multiply"} {
		calls := 0
		inverse := false
		failure := errors.New("test backend failure")
		err = rankProductSchedule(plans, 16, func(_ rankTransposePlan, operand int, inv bool) error {
			calls++
			inverse = inverse || inv
			if operand == 1 && fail == "right" {
				return failure
			}
			return nil
		}, func() error { calls++; return failure }, nil)
		want := 2
		if fail == "multiply" {
			want = 3
		}
		if !errors.Is(err, failure) || calls != want || inverse {
			t.Fatal(fail, err, calls, inverse)
		}
	}
}
