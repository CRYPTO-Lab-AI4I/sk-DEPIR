package liftrs

import "fmt"

// rankOperandPlan assigns coefficient intervals, independently of arithmetic or
// transport. Each operand initially has one owner; output is zero-padded shards.
// The two owners may be equal. This routes a merge's inputs, not its final result.
type rankOperandPlan struct {
	Operation, Domain, TileWords, Limit, MaxLocalWords uint64
	Counts                                             [2]uint64
	Owners                                             [2]int
	Rank, Ranks                                        int
}

type operandPacket struct {
	Operation, Domain, Operand, Offset, Words uint64
	Source, Target                            int
}

// Calls borrow slices until return. Receive checks the complete expected packet
// and fills exactly the supplied slice; Send preserves its source. No complete
// remote polynomial is retained.
type operandLink interface {
	SendOperand(operandPacket, []uint32) error
	ReceiveOperand(operandPacket, []uint32) error
}

// routePolynomialOperands preserves owned inputs and returns two independently
// owned coefficient shards. Traffic is deterministic: operand, target, tile.
// No global gather or padding traffic is needed. On error discard both outputs;
// collective callers must abort peers rather than proceed with incomplete data.
// MaxLocalWords admits each output separately; budget both outputs, owned input
// and transport/runtime memory together before invoking the collective.
func routePolynomialOperands(p rankOperandPlan, local [2][]uint32, link operandLink, notice func(int, uint64, uint64)) (out [2][]uint32, err error) {
	sources := [2]rootShardLayout{}
	for j := range sources {
		sources[j] = rootShardLayout{p.Owners[j], 1, p.Counts[j], p.Counts[j]}
	}
	return routeShardedOperands(rankShardOperandPlan{p.Operation, p.Domain, p.TileWords, p.Limit, p.MaxLocalWords, p.Rank, p.Ranks, sources}, local, link, notice)
}

// rankShardOperandPlan also admits already-distributed child polynomials. A
// one-rank source is the existing owner-to-shards case; both use the same codec,
// packet order and bounded interval intersections. Sources may overlap in ranks.
type rankShardOperandPlan struct {
	Operation, Domain, TileWords, Limit, MaxLocalWords uint64
	Rank, Ranks                                        int
	Sources                                            [2]rootShardLayout
}

func routeShardedOperands(p rankShardOperandPlan, local [2][]uint32, link operandLink, notice func(int, uint64, uint64)) (out [2][]uint32, err error) {
	if p.Operation == 0 || p.Domain == 0 || p.Domain > uint64(^uint(0)>>1) || p.Ranks < 1 || p.Ranks > 64 || p.Rank < 0 || p.Rank >= p.Ranks || p.TileWords == 0 || p.TileWords > 1<<20 || p.Limit < 2 || p.Limit > 1<<32 || (p.Ranks > 1 && link == nil) {
		return out, fmt.Errorf("operand routing plan")
	}
	for operand, source := range p.Sources {
		if source.First < 0 || source.Ranks < 1 || source.First > p.Ranks-source.Ranks || source.Domain == 0 || source.Domain > p.Domain || source.Count == 0 || source.Count > source.Domain {
			return out, fmt.Errorf("operand routing ownership/extent")
		}
		first, words := uint64(0), uint64(0)
		if p.Rank >= source.First && p.Rank < source.First+source.Ranks {
			first, words = rankBand(source.Domain, p.Rank-source.First, source.Ranks)
		}
		if uint64(len(local[operand])) != words {
			return out, fmt.Errorf("operand routing local extent")
		}
		for j, v := range local[operand] {
			if uint64(v) >= p.Limit || (first+uint64(j) >= source.Count && v != 0) {
				return out, fmt.Errorf("operand routing canonical input/padding")
			}
		}
	}
	_, words := rankBand(p.Domain, p.Rank, p.Ranks)
	if words > p.MaxLocalWords {
		return out, fmt.Errorf("operand routing local memory admission")
	}
	for operand, layout := range p.Sources {
		out[operand] = make([]uint32, words)
		// Enumerate only non-padding interval overlaps. The global order prevents
		// cyclic waits with synchronous links,including unequal/empty child shards.
		visits := func(visit func(int, int, uint64, uint64, uint64, uint64) error) error {
			for source := layout.First; source < layout.First+layout.Ranks; source++ {
				first, count := rankBand(layout.Domain, source-layout.First, layout.Ranks)
				end := min(first+count, layout.Count)
				for target := 0; target < p.Ranks; target++ {
					if p.Rank != source && p.Rank != target {
						continue
					}
					start, take := rankBand(p.Domain, target, p.Ranks)
					lo, hi := max(first, start), min(end, start+take)
					if lo < hi {
						if e := visit(source, target, first, start, lo, hi); e != nil {
							return e
						}
					}
				}
			}
			return nil
		}
		var done, total uint64
		visits(func(_, _ int, _, _, lo, hi uint64) error { total += hi - lo; return nil })
		if notice != nil {
			notice(operand, 0, total)
		}
		err = visits(func(source, target int, sourceFirst, targetFirst, lo, hi uint64) error {
			for offset := lo; offset < hi; {
				take := min(p.TileWords, hi-offset)
				packet := operandPacket{p.Operation, p.Domain, uint64(operand), offset, take, source, target}
				if source == target {
					copy(out[operand][offset-targetFirst:offset-targetFirst+take], local[operand][offset-sourceFirst:offset-sourceFirst+take])
				} else if p.Rank == source {
					if e := link.SendOperand(packet, local[operand][offset-sourceFirst:offset-sourceFirst+take]); e != nil {
						return e
					}
				} else {
					tile := out[operand][offset-targetFirst : offset-targetFirst+take]
					if e := link.ReceiveOperand(packet, tile); e != nil {
						return e
					}
					for _, v := range tile {
						if uint64(v) >= p.Limit {
							return fmt.Errorf("operand routing canonical receive")
						}
					}
				}
				offset += take
				done += take
				if notice != nil {
					notice(operand, done, total)
				}
			}
			return nil
		})
		if err != nil {
			return out, err
		}
	}
	return out, nil
}
