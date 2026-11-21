package liftrs

import "fmt"

// Taylor metadata is public transform data, not protocol secrets. A unique
// Operation names one joined quotient/remainder pass across all ranks.
type rankTaylorPlan struct {
	Operation, Domain, Columns, Capacity uint64
	Rank, Ranks, Level                   int
	Quotient                             bool
	Degrees                              []uint64
}
type taylorTicket struct{ Block, First, Count uint64 }
type taylorPacket struct {
	Operation, Domain, Columns uint64
	Level                      int
	Quotient                   bool
	Target                     taylorTicket
	Slot, Source, Words        uint64
}
type taylorLink[Word transposeWord] interface {
	SendWindow(peer int, packet taylorPacket, values []Word) error
	ReceiveWindow(peer int, packet taylorPacket, values []Word) error
	Barrier()
}
type rankTaylorApply[Word transposeWord] func(snapshot []Word, level int, quotient bool, block, first, count uint64) error

func (p rankTaylorPlan) validate() (half, minimum, maximum, span uint64, err error) {
	if p.Operation == 0 || p.Domain < 4 || p.Domain&(p.Domain-1) != 0 || p.Columns < 2 || p.Columns&(p.Columns-1) != 0 || p.Columns >= p.Domain || p.Domain%p.Columns != 0 || p.Level < 0 || p.Level >= 63 || p.Ranks < 1 || p.Ranks > 64 || p.Rank < 0 || p.Rank >= p.Ranks || len(p.Degrees) > 21 || p.Capacity == 0 || p.Capacity > 1<<20 {
		err = fmt.Errorf("Taylor plan admission")
		return
	}
	if p.Columns > (p.Domain/2)>>uint(p.Level) {
		err = fmt.Errorf("Taylor level outside domain")
		return
	}
	half = p.Columns << uint(p.Level)
	minimum = half
	for j, d := range p.Degrees {
		if d == 0 || d > half/2 || (j > 0 && d <= p.Degrees[j-1]) {
			err = fmt.Errorf("Taylor degree independence/order")
			return
		}
		minimum = min(minimum, d)
		maximum = max(maximum, d)
	}
	span = p.Capacity / uint64(len(p.Degrees)+1)
	if span == 0 {
		err = fmt.Errorf("Taylor snapshot capacity")
		return
	}
	return
}

// A cursor holds one ticket per owner, never a domain-sized request list.
type taylorCursor struct {
	block, ownerFirst, ownerEnd, domain, half, begin, end, span, next uint64
	quotient                                                          bool
}

func newTaylorCursor(p rankTaylorPlan, rank int, half, minimum, maximum, span uint64) taylorCursor {
	rowFirst, rows := rankBand(p.Domain/p.Columns, rank, p.Ranks)
	first, end := rowFirst*p.Columns, (rowFirst+rows)*p.Columns
	begin, limit := minimum, half
	if p.Quotient {
		begin, limit = 0, maximum
	}
	return taylorCursor{block: (first / (2 * half)) * (2 * half), ownerFirst: first, ownerEnd: end, domain: p.Domain, half: half, begin: begin, end: limit, span: span, quotient: p.Quotient}
}
func (c *taylorCursor) take() (taylorTicket, bool) {
	for c.block < c.domain && c.block < c.ownerEnd {
		base := c.block
		if c.quotient {
			base += c.half
		}
		begin, end := max(base+c.begin, c.ownerFirst), min(base+c.end, c.ownerEnd)
		begin = max(begin, c.next)
		if begin < end {
			count := min(c.span, end-begin)
			c.next = begin + count
			return taylorTicket{c.block, begin - base, count}, true
		}
		c.block += 2 * c.half
		c.next = 0
	}
	return taylorTicket{}, false
}

// taylorSource intersects one term window with the supplying owner's row band.
// The returned destination offset is inside that source's packed snapshot slot.
func taylorSource(p rankTaylorPlan, t taylorTicket, slot int, owner int, half uint64) (source, count, offset uint64) {
	base := t.Block + half
	if p.Quotient {
		base += half
	}
	start := base + t.First - p.Degrees[slot]
	end := min(start+t.Count, t.Block+2*half)
	// Remainder terms with degree>j are inactive. Their unused prefix can
	// overlap another target: never read/send it concurrently with updates.
	skip := uint64(0)
	if !p.Quotient && p.Degrees[slot] > t.First {
		skip = min(t.Count, p.Degrees[slot]-t.First)
	}
	start += skip
	first, rows := rankBand(p.Domain/p.Columns, owner, p.Ranks)
	lo, hi := first*p.Columns, (first+rows)*p.Columns
	source = max(start, lo)
	end = min(end, hi)
	if source >= end {
		return 0, 0, 0
	}
	return source, end - source, uint64(slot+1)*t.Count + skip + source - start
}

// rankTaylorStorage separates the shared ticket/wave schedule from owned-array
// access. Snapshot/Send borrow bounded ranges; Apply commits only the target interval.
// Host Send keeps its original borrowed slice; GPU Send uses a reusable tile.
type rankTaylorStorage[Word transposeWord] interface {
	Snapshot(p rankTaylorPlan, t taylorTicket, ownFirst, half uint64, target []Word) error
	Send(peer int, packet taylorPacket, offset, count uint64, link taylorLink[Word]) error
	Apply(p rankTaylorPlan, t taylorTicket, offset uint64, snapshot []Word) error
}
type hostTaylorStorage[Word transposeWord] struct {
	data  []Word
	apply rankTaylorApply[Word]
}

func (s hostTaylorStorage[Word]) Read(offset uint64, target []Word) error {
	copy(target, s.data[offset:offset+uint64(len(target))])
	return nil
}
func (s hostTaylorStorage[Word]) Send(peer int, p taylorPacket, offset, count uint64, link taylorLink[Word]) error {
	return link.SendWindow(peer, p, s.data[offset:offset+count])
}
func (s hostTaylorStorage[Word]) Apply(p rankTaylorPlan, t taylorTicket, offset uint64, snapshot []Word) error {
	if err := s.apply(snapshot, p.Level, p.Quotient, t.Block, t.First, t.Count); err != nil {
		return err
	}
	copy(s.data[offset:offset+t.Count], snapshot[:t.Count])
	return nil
}

// readTaylorSnapshot is the shared host-staging adapter. The resident-local
// adapter skips these local reads; target/source addresses still come from the
// same plan and kernel, with remote windows unchanged.
func readTaylorSnapshot[Word transposeWord](p rankTaylorPlan, t taylorTicket, ownFirst, half uint64, snapshot []Word, read func(uint64, []Word) error) error {
	target := t.Block + t.First
	if p.Quotient {
		target += half
	}
	if err := read(target-ownFirst, snapshot[:t.Count]); err != nil {
		return err
	}
	for slot := range p.Degrees {
		source, count, offset := taylorSource(p, t, slot, p.Rank, half)
		if count > 0 {
			if err := read(source-ownFirst, snapshot[offset:offset+count]); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s hostTaylorStorage[Word]) Snapshot(p rankTaylorPlan, t taylorTicket, ownFirst, half uint64, snapshot []Word) error {
	return readTaylorSnapshot(p, t, ownFirst, half, snapshot, s.Read)
}

// rankTaylorPass is the host reference adapter. It retains the original borrowed
// send slices and callback mutation semantics; both adapters share the scheduler.
func rankTaylorPass[Word transposeWord](p rankTaylorPlan, data, workspace []Word, link taylorLink[Word], apply rankTaylorApply[Word], notice func(string, uint64, uint64)) error {
	if _, _, _, _, err := p.validate(); err != nil {
		return err
	}
	_, rows := rankBand(p.Domain/p.Columns, p.Rank, p.Ranks)
	if uint64(len(data)) != rows*p.Columns || apply == nil {
		return fmt.Errorf("Taylor shard/callback")
	}
	return rankTaylorStoredPass(p, workspace, hostTaylorStorage[Word]{data, apply}, link, notice)
}

// rankTaylorStoredPass gathers bounded snapshots and updates owned targets in
// parallel waves. All ranks derive the same ticket/round order, including empty
// owners. Sources never overlap a target in the pass; barriers prevent unfinished
// quotient/remainder writes from feeding the next stage. Storage does not alter
// headers, source trimming, term order, wave order or synchronization.
func rankTaylorStoredPass[Word transposeWord](p rankTaylorPlan, workspace []Word, storage rankTaylorStorage[Word], link taylorLink[Word], notice func(string, uint64, uint64)) error {
	half, minimum, maximum, span, err := p.validate()
	if err != nil {
		return err
	}
	first, _ := rankBand(p.Domain/p.Columns, p.Rank, p.Ranks)
	ownFirst := first * p.Columns
	if uint64(len(workspace)) < p.Capacity || storage == nil || link == nil {
		return fmt.Errorf("Taylor shard/link/callback")
	}
	var zero Word
	width := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	if p.Capacity > uint64(^uint(0)>>1)/width {
		return fmt.Errorf("Taylor snapshot exceeds platform")
	}
	link.Barrier()
	if len(p.Degrees) == 0 {
		link.Barrier()
		return nil
	}
	cursors := make([]taylorCursor, p.Ranks)
	for rank := range cursors {
		cursors[rank] = newTaylorCursor(p, rank, half, minimum, maximum, span)
	}
	packed := workspace[:p.Capacity]
	var completed uint64
	totalCursor := cursors[p.Rank]
	var total uint64
	for {
		t, ok := totalCursor.take()
		if !ok {
			break
		}
		total += t.Count
	}
	if notice != nil {
		notice("rank-Taylor-targets", 0, total)
	}
	tickets := make([]taylorTicket, p.Ranks)
	active := make([]bool, p.Ranks)
	for {
		anyActive := false
		for rank := range cursors {
			tickets[rank], active[rank] = cursors[rank].take()
			anyActive = anyActive || active[rank]
		}
		if !anyActive {
			break
		}
		own := tickets[p.Rank]
		snapshot := packed[:0]
		if active[p.Rank] {
			snapshot = packed[:uint64(len(p.Degrees)+1)*own.Count]
			clear(snapshot)
			if err := storage.Snapshot(p, own, ownFirst, half, snapshot); err != nil {
				return err
			}
		}
		send := func(peer int) error {
			if !active[peer] {
				return nil
			}
			ticket := tickets[peer]
			for slot := range p.Degrees {
				source, count, _ := taylorSource(p, ticket, slot, p.Rank, half)
				if count == 0 {
					continue
				}
				packet := taylorPacket{p.Operation, p.Domain, p.Columns, p.Level, p.Quotient, ticket, uint64(slot), source, count}
				if err := storage.Send(peer, packet, source-ownFirst, count, link); err != nil {
					return err
				}
			}
			return nil
		}
		receive := func(peer int) error {
			if !active[p.Rank] {
				return nil
			}
			for slot := range p.Degrees {
				source, count, offset := taylorSource(p, own, slot, peer, half)
				if count == 0 {
					continue
				}
				packet := taylorPacket{p.Operation, p.Domain, p.Columns, p.Level, p.Quotient, own, uint64(slot), source, count}
				if err := link.ReceiveWindow(peer, packet, snapshot[offset:offset+count]); err != nil {
					return err
				}
			}
			return nil
		}
		for round := 1; round < p.Ranks; round++ {
			dest, source := (p.Rank+round)%p.Ranks, (p.Rank-round+p.Ranks)%p.Ranks
			if p.Rank < dest {
				if err = send(dest); err == nil {
					err = receive(source)
				}
			} else {
				if err = receive(source); err == nil {
					err = send(dest)
				}
			}
			if err != nil {
				return err
			}
		}
		if active[p.Rank] {
			target := own.Block + own.First
			if p.Quotient {
				target += half
			}
			if err = storage.Apply(p, own, target-ownFirst, snapshot); err != nil {
				return err
			}
			completed += own.Count
			if notice != nil {
				notice("rank-Taylor-targets", completed, total)
			}
		}
	}
	link.Barrier()
	return nil
}

// rankTaylorConvert keeps quotient/remainder and level barriers explicit. Degree
// arrays are immutable canonical metadata; arithmetic stays in the callback.
func rankTaylorConvert[Word transposeWord](base rankTaylorPlan, data []Word, inverse bool, degrees func(int) ([]uint64, error), link taylorLink[Word], apply rankTaylorApply[Word], notice func(string, uint64, uint64)) error {
	shape := base
	shape.Level = 0
	shape.Degrees = nil
	if _, _, _, _, err := shape.validate(); err != nil {
		return err
	}
	workspace := make([]Word, base.Capacity)
	return rankTaylorConversion(base, inverse, degrees, func(p rankTaylorPlan, passNotice func(string, uint64, uint64)) error {
		return rankTaylorPass(p, data, workspace, link, apply, passNotice)
	}, notice)
}

// rankTaylorConversion shares degree/level/pass ordering for all storage adapters.
func rankTaylorConversion(base rankTaylorPlan, inverse bool, degrees func(int) ([]uint64, error), execute func(rankTaylorPlan, func(string, uint64, uint64)) error, notice func(string, uint64, uint64)) error {
	shape := base
	shape.Level = 0
	shape.Degrees = nil
	if _, _, _, _, err := shape.validate(); err != nil {
		return err
	}
	levels := 0
	for rows := base.Domain / base.Columns; rows > 1; rows >>= 1 {
		levels++
	}
	if degrees == nil || levels == 0 || base.Operation > ^uint64(0)-uint64(2*levels) {
		return fmt.Errorf("Taylor conversion metadata/operation")
	}
	for step := 0; step < levels; step++ {
		level := levels - 1 - step
		if inverse {
			level = step
		}
		terms, err := degrees(level)
		if err != nil {
			return err
		}
		for pass := 0; pass < 2; pass++ {
			p := base
			p.Level = level
			p.Degrees = terms
			p.Operation = base.Operation + uint64(2*step+pass)
			p.Quotient = (pass == 0) != inverse
			passNotice := notice
			if notice != nil {
				kind := "remainder"
				if p.Quotient {
					kind = "quotient"
				}
				phase := fmt.Sprintf("rank-Taylor-level-%d-%s", level, kind)
				passNotice = func(_ string, done, total uint64) { notice(phase, done, total) }
			}
			if err = execute(p, passNotice); err != nil {
				return err
			}
		}
	}
	return nil
}

// rankAdditiveTransform adds Taylor conversion around the validated separable
// matrix stages. Forward: conversion then columns/rows; inverse reverses both.
// It consumes input and returns an owning row shard. This host-shard control
// layout validates arithmetic/communication, not full-size GPU-resident support.
func rankAdditiveTransform[Word transposeWord](p rankTransposePlan, input []Word, capacity uint64, inverse bool, transpose transposeLink[Word], windows taylorLink[Word], degrees func(int) ([]uint64, error), vectors rankVectorApply[Word], taylor rankTaylorApply[Word], notice func(string, uint64, uint64)) ([]Word, error) {
	base, matrix, err := rankAdditivePlans(p, capacity, inverse)
	if err != nil {
		return nil, err
	}
	if uint64(len(input)) != p.InputWords || windows == nil || degrees == nil || vectors == nil || taylor == nil || (p.Ranks > 1 && transpose == nil) {
		return nil, fmt.Errorf("additive rank input/backend")
	}
	current := input
	err = rankAdditiveSchedule(inverse,
		func() error { return rankTaylorConvert(base, current, inverse, degrees, windows, taylor, notice) },
		func() error {
			var e error
			current, e = rankMatrixVectors(matrix, current, capacity, inverse, transpose, vectors, notice)
			return e
		})
	if err != nil {
		return nil, err
	}
	return current, nil
}

// rankAdditivePlans fixes canonical shape, capacity and disjoint operation spans
// for both storage backends; storage/transport validation belongs to the caller.
func rankAdditivePlans(p rankTransposePlan, capacity uint64, inverse bool) (base rankTaylorPlan, matrix rankTransposePlan, err error) {
	fail := func(message string) (rankTaylorPlan, rankTransposePlan, error) {
		return base, matrix, fmt.Errorf("additive %s", message)
	}
	if p.Columns < 2 || p.Columns&(p.Columns-1) != 0 || p.Rows < 2 || p.Rows&(p.Rows-1) != 0 || p.Rows > ^uint64(0)/p.Columns {
		return fail("domain shape")
	}
	checked, e := newRankTransposePlan(p.Operation, p.Rows, p.Columns, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if e != nil || checked != p {
		return fail("rank plan")
	}
	if (p.Rows != p.Columns && p.Rows/p.Columns != 2) || p.Rows < p.Columns || capacity < max(p.Rows, p.Columns) || capacity > 1<<20 {
		return fail("split/capacity")
	}
	levels := 0
	for rows := p.Rows; rows > 1; rows >>= 1 {
		levels++
	}
	if p.Operation > ^uint64(0)-uint64(2*levels+2) {
		return fail("operation wrap")
	}
	base = rankTaylorPlan{Operation: p.Operation, Domain: p.Rows * p.Columns, Columns: p.Columns, Capacity: capacity, Rank: p.Rank, Ranks: p.Ranks}
	matrix = p
	if inverse {
		base.Operation += 2
	} else {
		matrix.Operation += uint64(2 * levels)
	}
	return base, matrix, nil
}

// rankAdditiveSchedule shares the complete FFT stage order, without imposing a
// host array or resident view: Taylor before vectors forward, after them inverse.
func rankAdditiveSchedule(inverse bool, taylor, vectors func() error) error {
	if !inverse {
		if err := taylor(); err != nil {
			return err
		}
	}
	if err := vectors(); err != nil {
		return err
	}
	if inverse {
		return taylor()
	}
	return nil
}
