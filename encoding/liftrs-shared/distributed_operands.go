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
	if p.Operation == 0 || p.Domain == 0 || p.Domain > uint64(^uint(0)>>1) || p.Ranks < 1 || p.Ranks > 64 || p.Rank < 0 || p.Rank >= p.Ranks || p.TileWords == 0 || p.TileWords > 1<<20 || p.Limit < 2 || p.Limit > 1<<32 || (p.Ranks > 1 && link == nil) {
		return out, fmt.Errorf("operand routing plan")
	}
	for operand := range local {
		if p.Counts[operand] == 0 || p.Counts[operand] > p.Domain || p.Owners[operand] < 0 || p.Owners[operand] >= p.Ranks {
			return out, fmt.Errorf("operand routing ownership/extent")
		}
		expected := uint64(0)
		if p.Owners[operand] == p.Rank {
			expected = p.Counts[operand]
		}
		if uint64(len(local[operand])) != expected {
			return out, fmt.Errorf("operand routing local extent")
		}
		for _, v := range local[operand] {
			if uint64(v) >= p.Limit {
				return out, fmt.Errorf("operand routing canonical input")
			}
		}
	}
	first, words := rankBand(p.Domain, p.Rank, p.Ranks)
	if words > p.MaxLocalWords {
		return out, fmt.Errorf("operand routing local memory admission")
	}
	for operand := range out {
		out[operand] = make([]uint32, words)
		total := p.Counts[operand]
		if p.Rank != p.Owners[operand] {
			total = min(words, total-min(first, total))
		}
		var done uint64
		if notice != nil {
			notice(operand, 0, total)
		}
		for target := 0; target < p.Ranks; target++ {
			start, count := rankBand(p.Domain, target, p.Ranks)
			count = min(count, p.Counts[operand]-min(start, p.Counts[operand]))
			if p.Rank != p.Owners[operand] && p.Rank != target {
				continue
			}
			for offset := uint64(0); offset < count; {
				take := min(p.TileWords, count-offset)
				packet := operandPacket{p.Operation, p.Domain, uint64(operand), start + offset, take, p.Owners[operand], target}
				if target == p.Owners[operand] {
					copy(out[operand][offset:offset+take], local[operand][start+offset:start+offset+take])
				} else if p.Rank == p.Owners[operand] {
					err = link.SendOperand(packet, local[operand][start+offset:start+offset+take])
				} else {
					tile := out[operand][offset : offset+take]
					err = link.ReceiveOperand(packet, tile)
					if err == nil {
						for _, v := range tile {
							if uint64(v) >= p.Limit {
								err = fmt.Errorf("operand routing canonical receive")
								break
							}
						}
					}
				}
				if err != nil {
					return out, err
				}
				offset += take
				done += take
				if notice != nil {
					notice(operand, done, total)
				}
			}
		}
	}
	return out, nil
}
