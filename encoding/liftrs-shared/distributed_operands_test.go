package liftrs

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

type operandMessage struct {
	packet operandPacket
	values []uint32
}
type operandTestLink struct {
	wire map[[2]int]chan operandMessage
}

func (l operandTestLink) SendOperand(p operandPacket, v []uint32) error {
	l.wire[[2]int{p.Source, p.Target}] <- operandMessage{p, append([]uint32(nil), v...)}
	return nil
}
func (l operandTestLink) ReceiveOperand(p operandPacket, v []uint32) error {
	m := <-l.wire[[2]int{p.Source, p.Target}]
	if m.packet != p || len(m.values) != len(v) {
		return fmt.Errorf("packet mismatch")
	}
	copy(v, m.values)
	return nil
}

func TestRankOperandRouting(t *testing.T) {
	for _, ranks := range []int{1, 2, 3, 5} {
		for _, domain := range []uint64{2, 7, 19} {
			for _, owners := range [][2]int{{0, ranks - 1}, {ranks - 1, ranks - 1}} {
				counts := [2]uint64{domain - 1, domain}
				inputs := [2][]uint32{make([]uint32, counts[0]), make([]uint32, counts[1])}
				for op := range inputs {
					for j := range inputs[op] {
						inputs[op][j] = uint32(100*op + j + 1)
					}
				}
				before := [2][]uint32{append([]uint32(nil), inputs[0]...), append([]uint32(nil), inputs[1]...)}
				link := operandTestLink{make(map[[2]int]chan operandMessage)}
				for a := 0; a < ranks; a++ {
					for b := 0; b < ranks; b++ {
						if a != b {
							link.wire[[2]int{a, b}] = make(chan operandMessage)
						}
					}
				}
				results := make([][2][]uint32, ranks)
				errors := make([]error, ranks)
				var wait sync.WaitGroup
				for rank := 0; rank < ranks; rank++ {
					wait.Add(1)
					go func(rank int) {
						defer wait.Done()
						var own [2][]uint32
						for op, owner := range owners {
							if owner == rank {
								own[op] = inputs[op]
							}
						}
						results[rank], errors[rank] = routePolynomialOperands(rankOperandPlan{Operation: 1, Domain: domain, TileWords: 2, Limit: 256, MaxLocalWords: domain, Counts: counts, Owners: owners, Rank: rank, Ranks: ranks}, own, link, nil)
					}(rank)
				}
				wait.Wait()
				for rank, e := range errors {
					if e != nil {
						t.Fatal(e)
					}
					first, words := rankBand(domain, rank, ranks)
					for op := range inputs {
						want := make([]uint32, words)
						for j := range want {
							if k := first + uint64(j); k < counts[op] {
								want[j] = inputs[op][k]
							}
						}
						if !reflect.DeepEqual(results[rank][op], want) {
							t.Fatal("owned coefficients", ranks, domain, owners, rank, op)
						}
					}
				}
				if !reflect.DeepEqual(inputs, before) {
					t.Fatal("input mutated")
				}
			}
		}
	}
}

type failingOperandLink struct{}

func (failingOperandLink) SendOperand(operandPacket, []uint32) error {
	return fmt.Errorf("injected send")
}
func (failingOperandLink) ReceiveOperand(operandPacket, []uint32) error {
	return fmt.Errorf("injected receive")
}

func TestRankOperandAdmissionAndFailure(t *testing.T) {
	p := rankOperandPlan{Operation: 1, Domain: 4, TileWords: 2, Limit: 16, MaxLocalWords: 4, Counts: [2]uint64{2, 3}, Owners: [2]int{0, 0}, Rank: 0, Ranks: 1}
	own := [2][]uint32{{1, 2}, {1, 2, 3}}
	for _, bad := range []func(*rankOperandPlan){func(p *rankOperandPlan) { p.Operation = 0 }, func(p *rankOperandPlan) { p.Counts[0] = 5 }, func(p *rankOperandPlan) { p.Owners[0] = 1 }, func(p *rankOperandPlan) { p.TileWords = 0 }, func(p *rankOperandPlan) { p.Limit = 2 }, func(p *rankOperandPlan) { p.MaxLocalWords = 3 }} {
		q := p
		bad(&q)
		if _, e := routePolynomialOperands(q, own, nil, nil); e == nil {
			t.Fatal("bad plan/input accepted", q)
		}
	}
	p.Ranks = 2
	p.Counts = [2]uint64{4, 4}
	own = [2][]uint32{{1, 2, 3, 4}, {1, 2, 3, 4}}
	if _, e := routePolynomialOperands(p, own, failingOperandLink{}, nil); e == nil {
		t.Fatal("send failure ignored")
	}
	p.Rank = 1
	own = [2][]uint32{}
	if _, e := routePolynomialOperands(p, own, failingOperandLink{}, nil); e == nil {
		t.Fatal("receive failure ignored")
	}
}

type badValueOperandLink struct{}

func (badValueOperandLink) SendOperand(operandPacket, []uint32) error {
	return fmt.Errorf("unexpected send")
}
func (badValueOperandLink) ReceiveOperand(_ operandPacket, v []uint32) error {
	for j := range v {
		v[j] = 16
	}
	return nil
}
func TestRankOperandRejectsNoncanonicalReceive(t *testing.T) {
	p := rankOperandPlan{Operation: 1, Domain: 4, TileWords: 2, Limit: 16, MaxLocalWords: 2, Counts: [2]uint64{4, 4}, Owners: [2]int{0, 0}, Rank: 1, Ranks: 2}
	if _, err := routePolynomialOperands(p, [2][]uint32{}, badValueOperandLink{}, nil); err == nil {
		t.Fatal("noncanonical receive accepted")
	}
}
