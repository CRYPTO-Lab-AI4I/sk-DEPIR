//go:build liftrs_cuda && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
)

// residentTransposeStorage moves bounded host tiles directly to/from GPU shards.
// The schedule, packet validation and proofs remain in transposeStored/link.
type residentTransposeStorage[Word transposeWord] struct {
	gpu            *cuda.RankStorage
	source, target int
}

func (s residentTransposeStorage[Word]) Local(p rankTransposePlan) error {
	return s.gpu.Local(s.source, s.target)
}
func (s residentTransposeStorage[Word]) Pack(p rankTransposePlan, peer int, offset uint64, tile []Word) error {
	switch v := any(tile).(type) {
	case []uint32:
		return s.gpu.Tile32(s.source, peer, offset, v, true)
	case []uint64:
		return s.gpu.Tile64(s.source, peer, offset, v, true)
	}
	return fmt.Errorf("GPU storage slot width")
}
func (s residentTransposeStorage[Word]) Scatter(p rankTransposePlan, peer int, offset uint64, tile []Word) error {
	switch v := any(tile).(type) {
	case []uint32:
		return s.gpu.Tile32(s.target, peer, offset, v, false)
	case []uint64:
		return s.gpu.Tile64(s.target, peer, offset, v, false)
	}
	return fmt.Errorf("GPU storage slot width")
}

// transposeResident replaces only target. Interleaved selects its physical view;
// source and the third slot survive. On failure target is incomplete and must be
// reset before use. No complete host input/output shard is allocated here.
func transposeResident[Word transposeWord](p rankTransposePlan, gpu *cuda.RankStorage, source, target int, interleaved bool, link transposeLink[Word], notice func(string, uint64, uint64)) error {
	if err := validateTranspose(p, link); err != nil {
		return err
	}
	if !gpu.AdmitsTranspose(p.Rows, p.Columns, p.Rank, p.Ranks, p.TileWords) || source < 0 || source >= 3 || target < 0 || target >= 3 || source == target {
		return fmt.Errorf("GPU storage transpose ownership/slot mismatch")
	}
	if err := gpu.Reset(target, p.Columns, p.Rows, interleaved); err != nil {
		return err
	}
	return transposeStored(p, residentTransposeStorage[Word]{gpu, source, target}, link, notice)
}

// rankResidentVectors consumes data in one slot and reuses another as scratch.
// Forward accepts row-major or interleaved input; inverse requires interleaved
// input from the forward pass. Output is interleaved in the original slot, with
// logical row-major chunked I/O unchanged. Third slot is preserved. This is the
// separable stage only: a complete additive FFT also needs Taylor conversion.
func rankResidentVectors[Word transposeWord](p rankTransposePlan, gpu *cuda.RankStorage, data, scratch int, inverse bool, link transposeLink[Word], notice func(string, uint64, uint64)) error {
	if err := validateTranspose(p, link); err != nil {
		return err
	}
	if p.Operation == ^uint64(0) || !gpu.AdmitsTranspose(p.Rows, p.Columns, p.Rank, p.Ranks, p.TileWords) || data < 0 || data >= 3 || scratch < 0 || scratch >= 3 || data == scratch {
		return fmt.Errorf("resident vectors plan/slots")
	}
	back, err := newRankTransposePlan(p.Operation+1, p.Columns, p.Rows, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if err != nil {
		return err
	}
	return rankSeparableSchedule(inverse,
		func() error { return gpu.Vectors(data, false, inverse) },
		func() error { return transposeResident[Word](p, gpu, data, scratch, true, link, notice) },
		func() error { return gpu.Vectors(scratch, true, inverse) },
		func() error { return transposeResident[Word](back, gpu, scratch, data, true, link, notice) })
}

// residentTaylorStorage retains only a bounded send tile; conversion separately
// owns the packed snapshot. Logical I/O supports either resident physical view.
type residentTaylorStorage[Word transposeWord] struct {
	gpu  *cuda.RankStorage
	slot int
	tile uint64
	send []Word
}

func (s residentTaylorStorage[Word]) Read(offset uint64, target []Word) error {
	for done := 0; done < len(target); {
		take := min(int(s.tile), len(target)-done)
		work := target[done : done+take]
		var err error
		switch v := any(work).(type) {
		case []uint32:
			err = s.gpu.Transfer32(s.slot, offset+uint64(done), v, false)
		case []uint64:
			err = s.gpu.Transfer64(s.slot, offset+uint64(done), v, false)
		}
		if err != nil {
			return err
		}
		done += take
	}
	return nil
}
func (s residentTaylorStorage[Word]) Send(peer int, p taylorPacket, offset, count uint64, link taylorLink[Word]) error {
	if count > uint64(len(s.send)) {
		return fmt.Errorf("resident Taylor send capacity")
	}
	work := s.send[:count]
	if err := s.Read(offset, work); err != nil {
		return err
	}
	return link.SendWindow(peer, p, work)
}
func (s residentTaylorStorage[Word]) Apply(p rankTaylorPlan, t taylorTicket, offset uint64, snapshot []Word) error {
	tile := cuda.TaylorTile{Level: p.Level, Quotient: p.Quotient, Block: t.Block, First: t.First, Count: t.Count}
	switch v := any(snapshot).(type) {
	case []uint32:
		return s.gpu.Taylor32(s.slot, v, p.Capacity, tile)
	case []uint64:
		return s.gpu.Taylor64(s.slot, v, p.Capacity, tile)
	}
	return fmt.Errorf("resident Taylor slot width")
}

// rankResidentTaylor uses the unchanged Taylor ticket/source/round scheduler.
// Two bounded host buffers and one capped device window replace a host shard.
// Sources/third slot are preserved; failure leaves an incomplete target slot.
func rankResidentTaylor[Word transposeWord](base rankTaylorPlan, gpu *cuda.RankStorage, slot int, tileWords uint64, inverse bool, link taylorLink[Word], degrees func(int) ([]uint64, error), notice func(string, uint64, uint64)) error {
	if _, _, _, _, err := base.validate(); err != nil {
		return err
	}
	if slot < 0 || slot >= 3 || link == nil || !gpu.AdmitsTranspose(base.Domain/base.Columns, base.Columns, base.Rank, base.Ranks, tileWords) {
		return fmt.Errorf("resident Taylor ownership/slot")
	}
	workspace := make([]Word, base.Capacity)
	storage := residentTaylorStorage[Word]{gpu: gpu, slot: slot, tile: tileWords, send: make([]Word, base.Capacity)}
	return rankTaylorConversion(base, inverse, degrees, func(p rankTaylorPlan, notice func(string, uint64, uint64)) error {
		return rankTaylorStoredPass(p, workspace, storage, link, notice)
	}, notice)
}

// rankResidentAdditiveTransform is the resident adapter to the same complete FFT
// plans/order used by the host reference. Input is consumed in data, scratch is
// reused, and the third slot survives. Result physical layout stays interleaved.
func rankResidentAdditiveTransform[Word transposeWord](p rankTransposePlan, gpu *cuda.RankStorage, data, scratch int, capacity uint64, inverse bool, transpose transposeLink[Word], windows taylorLink[Word], degrees func(int) ([]uint64, error), notice func(string, uint64, uint64)) error {
	base, matrix, err := rankAdditivePlans(p, capacity, inverse)
	if err != nil {
		return err
	}
	if data < 0 || data >= 3 || scratch < 0 || scratch >= 3 || data == scratch || degrees == nil || windows == nil || (p.Ranks > 1 && transpose == nil) || !gpu.AdmitsTranspose(p.Rows, p.Columns, p.Rank, p.Ranks, p.TileWords) {
		return fmt.Errorf("resident additive ownership/backend")
	}
	return rankAdditiveSchedule(inverse,
		func() error {
			return rankResidentTaylor[Word](base, gpu, data, p.TileWords, inverse, windows, degrees, notice)
		},
		func() error { return rankResidentVectors[Word](matrix, gpu, data, scratch, inverse, transpose, notice) })
}

// rankResidentPolynomialProduct consumes distinct left/right coefficient slots.
// Right retains its spectrum after multiplication; only left and scratch change
// during inverse. Three disjoint arrays suffice for all transforms and pointwise
// work. Caller admits a domain exceeding the summed degrees to avoid aliasing.
func rankResidentPolynomialProduct[Word transposeWord](p rankTransposePlan, gpu *cuda.RankStorage, left, right, scratch int, capacity uint64, transpose transposeLink[Word], windows taylorLink[Word], degrees func(int) ([]uint64, error), notice func(string, uint64, uint64)) error {
	plans, err := rankProductPlans(p, capacity)
	if err != nil {
		return err
	}
	if left < 0 || left >= 3 || right < 0 || right >= 3 || scratch < 0 || scratch >= 3 || left == right || left == scratch || right == scratch || degrees == nil || windows == nil || (p.Ranks > 1 && transpose == nil) || !gpu.AdmitsTranspose(p.Rows, p.Columns, p.Rank, p.Ranks, p.TileWords) {
		return fmt.Errorf("resident product ownership/backend/alias")
	}
	slots := [2]int{left, right}
	return rankProductSchedule(plans, p.InputWords,
		func(plan rankTransposePlan, operand int, inverse bool) error {
			return rankResidentAdditiveTransform[Word](plan, gpu, slots[operand], scratch, capacity, inverse, transpose, windows, degrees, notice)
		},
		func() error { return gpu.Pointwise(left, right) }, notice)
}
