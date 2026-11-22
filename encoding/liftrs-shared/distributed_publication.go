package liftrs

import (
	"fmt"
	"io"
)

type rankStreamPlan struct {
	Operation, Domain, Count, TileWords, Limit, MaxLocalWords uint64
	Rank, Ranks                                               int
}
type rankValueLink[Word transposeWord] interface {
	SendOperand(operandPacket, []Word) error
	ReceiveOperand(operandPacket, []Word) error
}

// streamRankValues publishes canonical ordered words through rank0 without a
// complete receive array. Source shards are immutable and include zero padding.
// Each worker admits its input separately; rank0 adds only one bounded tile.
// Word selects file width and must match the section/header representation.
// On error collective callers must abort peers before attempting another phase.
func streamRankValues[Word transposeWord](p rankStreamPlan, input []Word, writer io.Writer, link rankValueLink[Word]) error {
	if p.Operation == 0 || p.Domain == 0 || p.Count == 0 || p.Count > p.Domain || p.TileWords == 0 || p.TileWords > 1<<20 || p.Ranks < 1 || p.Ranks > 64 || p.Rank < 0 || p.Rank >= p.Ranks || p.Limit < 2 || (p.Rank == 0 && writer == nil) || (p.Ranks > 1 && link == nil) {
		return fmt.Errorf("stream publication plan")
	}
	first, words := rankBand(p.Domain, p.Rank, p.Ranks)
	if words > p.MaxLocalWords || uint64(len(input)) != words {
		return fmt.Errorf("stream publication input admission")
	}
	var zero Word
	width := uint8(4)
	if _, ok := any(zero).(uint64); ok {
		width = 8
	}
	if width == 4 && p.Limit > 1<<32 {
		return fmt.Errorf("stream publication word range")
	}
	for j, v := range input {
		if uint64(v) >= p.Limit || (first+uint64(j) >= p.Count && v != 0) {
			return fmt.Errorf("stream publication coefficient/padding")
		}
	}
	var tile []Word
	if p.Rank == 0 {
		_, largest := rankBand(p.Domain, 0, p.Ranks)
		tile = make([]Word, min(p.TileWords, largest, p.Count))
	}
	write := func(values []Word) error {
		switch v := any(values).(type) {
		case []uint32:
			return writeScalingCodeValues(writer, width, p.Limit, v, nil)
		case []uint64:
			return writeScalingCodeValues(writer, width, p.Limit, nil, v)
		}
		return fmt.Errorf("stream publication width")
	}
	for source := 0; source < p.Ranks; source++ {
		start, count := rankBand(p.Domain, source, p.Ranks)
		count = min(count, p.Count-min(start, p.Count))
		if p.Rank != 0 && p.Rank != source {
			continue
		}
		for offset := uint64(0); offset < count; {
			take := min(p.TileWords, count-offset)
			packet := operandPacket{p.Operation, p.Domain, 0, start + offset, take, source, 0}
			if source == 0 {
				if err := write(input[offset : offset+take]); err != nil {
					return err
				}
			} else if p.Rank == source {
				if err := link.SendOperand(packet, input[offset:offset+take]); err != nil {
					return err
				}
			} else {
				values := tile[:take]
				if err := link.ReceiveOperand(packet, values); err != nil {
					return err
				}
				if err := write(values); err != nil {
					return err
				}
			}
			offset += take
		}
	}
	return nil
}
