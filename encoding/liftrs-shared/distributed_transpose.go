package liftrs

import "fmt"

// transposeWord is a canonical storage slot, not a new field representation.
// Moving these words performs no arithmetic and preserves all bits.
type transposeWord interface{ uint32 | uint64 }

type transposePacket struct {
	Operation, Rows, Columns, Round, Offset, Words uint64
}

// A link borrows the tile only until the blocking call returns. Receive must
// check the complete expected descriptor and fill exactly len(tile) words.
type transposeLink[Word transposeWord] interface {
	SendTile(peer int, packet transposePacket, tile []Word) error
	ReceiveTile(peer int, packet transposePacket, tile []Word) error
}

type rankTransposePlan struct {
	Operation, Rows, Columns                         uint64
	Rank, Ranks                                      int
	InputFirst, InputRows, OutputFirst, OutputRows   uint64
	InputWords, OutputWords, TileWords, ScratchWords uint64
}

// rankBand divides complete rows, assigning the remainder to the first owners.
// The quotient/remainder form avoids size*rank overflow for large virtual grids.
func rankBand(size uint64, rank, ranks int) (first, count uint64) {
	quotient, remainder := size/uint64(ranks), size%uint64(ranks)
	first = quotient*uint64(rank) + min(uint64(rank), remainder)
	count = quotient
	if uint64(rank) < remainder {
		count++
	}
	return
}

// The input is row-major Rows x Columns, split by original rows. The output is
// row-major Columns x Rows, split by original columns. No global array is built.
// maxLocalWords admits each principal array separately; sum live arrays plus
// runtime/transport memory when admitting a real job. This is an experimental
// packing reference/control plane, not yet a resident distributed GPU transform.
func newRankTransposePlan(operation, rows, columns uint64, rank, ranks int, maxLocalWords, tileWords uint64) (rankTransposePlan, error) {
	p := rankTransposePlan{Operation: operation, Rows: rows, Columns: columns, Rank: rank, Ranks: ranks}
	if operation == 0 || rows == 0 || columns == 0 || rows > ^uint64(0)/columns || ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks || tileWords == 0 || tileWords > 1<<20 {
		return p, fmt.Errorf("transpose operation/shape/rank/tile admission")
	}
	p.InputFirst, p.InputRows = rankBand(rows, rank, ranks)
	p.OutputFirst, p.OutputRows = rankBand(columns, rank, ranks)
	p.InputWords, p.OutputWords = p.InputRows*columns, p.OutputRows*rows
	if p.InputWords > maxLocalWords || p.OutputWords > maxLocalWords {
		return p, fmt.Errorf("transpose local shard not admitted")
	}
	_, largestRows := rankBand(rows, 0, ranks)
	_, largestColumns := rankBand(columns, 0, ranks)
	p.TileWords = tileWords
	p.ScratchWords = min(tileWords, max(p.InputRows*largestColumns, p.OutputRows*largestRows))
	return p, nil
}

// rankTransposeStorage separates owned-array access from the communication
// schedule. Pack/Scatter borrow only a bounded tile; Local preserves the source.
// On error the destination is incomplete and the caller must discard/reset it.
type rankTransposeStorage[Word transposeWord] interface {
	Local(rankTransposePlan) error
	Pack(rankTransposePlan, int, uint64, []Word) error
	Scatter(rankTransposePlan, int, uint64, []Word) error
}

type hostTransposeStorage[Word transposeWord] struct{ input, output []Word }

func (h hostTransposeStorage[Word]) Local(p rankTransposePlan) error {
	for column := uint64(0); column < p.OutputRows; column++ {
		for row := uint64(0); row < p.InputRows; row++ {
			h.output[column*p.Rows+p.InputFirst+row] = h.input[row*p.Columns+p.OutputFirst+column]
		}
	}
	return nil
}
func (h hostTransposeStorage[Word]) Pack(p rankTransposePlan, peer int, offset uint64, tile []Word) error {
	first, _ := rankBand(p.Columns, peer, p.Ranks)
	for j := range tile {
		index := offset + uint64(j)
		tile[j] = h.input[(index%p.InputRows)*p.Columns+first+index/p.InputRows]
	}
	return nil
}
func (h hostTransposeStorage[Word]) Scatter(p rankTransposePlan, peer int, offset uint64, tile []Word) error {
	first, rows := rankBand(p.Rows, peer, p.Ranks)
	for j := range tile {
		index := offset + uint64(j)
		h.output[(index/rows)*p.Rows+first+index%rows] = tile[j]
	}
	return nil
}

func validateTranspose[Word transposeWord](p rankTransposePlan, link transposeLink[Word]) error {
	admitted, err := newRankTransposePlan(p.Operation, p.Rows, p.Columns, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if err != nil || admitted != p || (p.Ranks > 1 && link == nil) {
		return fmt.Errorf("transpose plan/link mismatch")
	}
	var zero Word
	wordBytes := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		wordBytes = 4
	}
	maxWords := uint64(^uint(0)>>1) / wordBytes
	if p.InputWords > maxWords || p.OutputWords > maxWords || p.ScratchWords > maxWords {
		return fmt.Errorf("transpose byte extent exceeds platform")
	}
	return nil
}

// transposeRank is the host correctness adapter, retaining one input/output shard
// and a bounded transport tile. Storage adapters share transposeStored verbatim.
func transposeRank[Word transposeWord](p rankTransposePlan, input []Word, link transposeLink[Word], notice func(string, uint64, uint64)) ([]Word, error) {
	if err := validateTranspose(p, link); err != nil {
		return nil, err
	}
	if uint64(len(input)) != p.InputWords {
		return nil, fmt.Errorf("transpose input mismatch")
	}
	out := make([]Word, int(p.OutputWords))
	if err := transposeStored(p, hostTransposeStorage[Word]{input, out}, link, notice); err != nil {
		return nil, err
	}
	return out, nil
}

// transposeStored keeps exactly one host tile. Pairwise rounds send to rank+round
// and receive from rank-round, send first iff rank<destination, avoiding blocking
// cycles. Empty bands exchange nothing. Headers, proofs and notice boundaries are
// independent of storage; GPU packing does not change wire order or arithmetic.
func transposeStored[Word transposeWord](p rankTransposePlan, storage rankTransposeStorage[Word], link transposeLink[Word], notice func(string, uint64, uint64)) error {
	if err := validateTranspose(p, link); err != nil {
		return err
	}
	if storage == nil {
		return fmt.Errorf("transpose storage missing")
	}
	tile := make([]Word, int(p.ScratchWords))
	localWords := p.InputRows * p.OutputRows
	if err := storage.Local(p); err != nil {
		return err
	}
	sent, received := uint64(0), uint64(0)
	emit := func(phase string, done, total uint64) {
		if notice != nil {
			notice(phase, done, total)
		}
	}
	emit("transpose-local", localWords, localWords)
	send := func(peer int, round uint64) error {
		_, columns := rankBand(p.Columns, peer, p.Ranks)
		words := p.InputRows * columns
		for offset := uint64(0); offset < words; {
			take := min(p.TileWords, words-offset)
			if err := storage.Pack(p, peer, offset, tile[:int(take)]); err != nil {
				return err
			}
			packet := transposePacket{p.Operation, p.Rows, p.Columns, round, offset, take}
			if err := link.SendTile(peer, packet, tile[:int(take)]); err != nil {
				return err
			}
			sent += take
			emit("transpose-send", sent, p.InputWords-localWords)
			offset += take
		}
		return nil
	}
	receive := func(peer int, round uint64) error {
		_, rows := rankBand(p.Rows, peer, p.Ranks)
		words := rows * p.OutputRows
		for offset := uint64(0); offset < words; {
			take := min(p.TileWords, words-offset)
			packet := transposePacket{p.Operation, p.Rows, p.Columns, round, offset, take}
			if err := link.ReceiveTile(peer, packet, tile[:int(take)]); err != nil {
				return err
			}
			if err := storage.Scatter(p, peer, offset, tile[:int(take)]); err != nil {
				return err
			}
			received += take
			emit("transpose-receive", received, p.OutputWords-localWords)
			offset += take
		}
		return nil
	}
	var err error
	for round := 1; round < p.Ranks; round++ {
		destination, source := (p.Rank+round)%p.Ranks, (p.Rank-round+p.Ranks)%p.Ranks
		if p.Rank < destination {
			if err = send(destination, uint64(round)); err == nil {
				err = receive(source, uint64(round))
			}
		} else {
			if err = receive(source, uint64(round)); err == nil {
				err = send(destination, uint64(round))
			}
		}
		if err != nil {
			return err
		}
	}
	if sent != p.InputWords-localWords || received != p.OutputWords-localWords {
		return fmt.Errorf("transpose ownership coverage")
	}
	emit("transpose-output", p.OutputWords, p.OutputWords)
	return nil
}
