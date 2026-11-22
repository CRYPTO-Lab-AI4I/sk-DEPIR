package liftrs

import "fmt"

// rankSquarePlan maps coefficient i to2i, keeping only a requested prefix. The
// base-field square is injected, so ownership routing adds no field convention.
type rankSquarePlan struct {
	Operation, Domain, TileWords, Limit, Keep, MaxLocalWords uint64
	Rank, Ranks                                              int
}

// squarePolynomialShard uses the existing checked operand transport for sparse
// coefficient movement. Odd result coefficients stay zero. No full polynomial
// or all-to-all matrix is gathered; source is immutable and buffers are bounded.
func squarePolynomialShard(p rankSquarePlan, input []uint32, square func(uint32) uint32, link operandLink, notice func(uint64, uint64)) ([]uint32, error) {
	if p.Operation == 0 || p.Domain == 0 || p.Domain > uint64(^uint(0)>>1) || p.Keep == 0 || p.Keep > p.Domain || p.TileWords == 0 || p.TileWords > 1<<20 || p.Limit < 2 || p.Limit > 1<<32 || p.Ranks < 1 || p.Ranks > 64 || p.Rank < 0 || p.Rank >= p.Ranks || square == nil || (p.Ranks > 1 && link == nil) {
		return nil, fmt.Errorf("distributed square plan")
	}
	first, words := rankBand(p.Domain, p.Rank, p.Ranks)
	if words > p.MaxLocalWords || uint64(len(input)) != words {
		return nil, fmt.Errorf("distributed square shard admission")
	}
	for _, v := range input {
		if uint64(v) >= p.Limit {
			return nil, fmt.Errorf("distributed square canonical input")
		}
	}
	output := make([]uint32, words)
	tile := make([]uint32, min(p.TileWords, (p.Keep+1)/2))
	// Only i<ceil(Keep/2) contributes. Visit each source/target interval exactly
	// once; the same schedule computes the local progress total and execution.
	visit := func(fn func(int, int, uint64, uint64) error) error {
		for source := 0; source < p.Ranks; source++ {
			sf, sw := rankBand(p.Domain, source, p.Ranks)
			for target := 0; target < p.Ranks; target++ {
				if p.Rank != source && p.Rank != target {
					continue
				}
				tf, tw := rankBand(p.Domain, target, p.Ranks)
				start := max(sf, (tf+1)/2)
				end := min(sf+sw, (tf+tw+1)/2, (p.Keep+1)/2)
				if end > start {
					if err := fn(source, target, start, end); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	var done, total uint64
	_ = visit(func(_, _ int, start, end uint64) error { total += end - start; return nil })
	if notice != nil {
		notice(0, total)
	}
	err := visit(func(source, target int, start, end uint64) error {
		for offset := start; offset < end; {
			count := min(p.TileWords, end-offset)
			work := tile[:count]
			if p.Rank == source {
				for j := range work {
					work[j] = square(input[offset+uint64(j)-first])
				}
			}
			packet := operandPacket{p.Operation, p.Domain, 0, offset, count, source, target}
			if source != target {
				var err error
				if p.Rank == source {
					err = link.SendOperand(packet, work)
				} else {
					err = link.ReceiveOperand(packet, work)
				}
				if err != nil {
					return err
				}
			}
			if p.Rank == target {
				for j, v := range work {
					if uint64(v) >= p.Limit {
						return fmt.Errorf("distributed square canonical output")
					}
					output[2*(offset+uint64(j))-first] = v
				}
			}
			offset += count
			done += count
			if notice != nil {
				notice(done, total)
			}
		}
		return nil
	})
	return output, err
}
