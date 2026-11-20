package liftrs

import "fmt"

// rankVectorApply borrows interleaved complete vectors for one synchronous call.
// first/count are global vector coordinates; no pointer or tile may be retained.
type rankVectorApply[Word transposeWord] func(tile []Word, columns, inverse bool, first, count uint64) error

// rankVectorRows packs bounded batches of owned row-major complete vectors into
// the GPU's interleaved tile. It mutates only data, never builds a global grid.
func rankVectorRows[Word transposeWord](data []Word, first, vectors, length, capacity uint64, columns, inverse bool, apply rankVectorApply[Word], notice func(string, uint64, uint64)) error {
	if length == 0 || capacity < length || vectors > ^uint64(0)/length || uint64(len(data)) != vectors*length || apply == nil {
		return fmt.Errorf("rank vector extent/callback/capacity")
	}
	batch := min(vectors, capacity/length)
	if batch == 0 {
		return nil
	}
	var zero Word
	width := uint64(8)
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	if batch*length > uint64(^uint(0)>>1)/width {
		return fmt.Errorf("rank vector tile exceeds platform")
	}
	tile := make([]Word, int(batch*length))
	for offset := uint64(0); offset < vectors; {
		count := min(batch, vectors-offset)
		work := tile[:int(count*length)]
		for coefficient := uint64(0); coefficient < length; coefficient++ {
			for j := uint64(0); j < count; j++ {
				work[coefficient*count+j] = data[(offset+j)*length+coefficient]
			}
		}
		if err := apply(work, columns, inverse, first+offset, count); err != nil {
			return err
		}
		for coefficient := uint64(0); coefficient < length; coefficient++ {
			for j := uint64(0); j < count; j++ {
				data[(offset+j)*length+coefficient] = work[coefficient*count+j]
			}
		}
		offset += count
		if notice != nil {
			phase := "rank-rows-vectors"
			if columns {
				phase = "rank-columns-vectors"
			}
			notice(phase, offset, vectors)
		}
	}
	return nil
}

// rankMatrixVectors composes separable column/affine-row stages with two rank
// transposes, returning a row-owned matrix in the original layout. Inverse runs
// row inverse before column inverse. This is NOT an additive FFT: the Taylor
// basis conversion must precede/follow this sequence at a separate integration
// gate. All ranks use identical operations and admission; failure aborts peers.
// Input is consumed/mutated; callers needing it afterward must retain a copy.
// Peak host storage includes caller input plus two owned shards and bounded tiles,
// not an adopted full-size GPU-resident layout. The faster resident path stays.
func rankMatrixVectors[Word transposeWord](p rankTransposePlan, input []Word, capacity uint64, inverse bool, link transposeLink[Word], apply rankVectorApply[Word], notice func(string, uint64, uint64)) ([]Word, error) {
	checked, err := newRankTransposePlan(p.Operation, p.Rows, p.Columns, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if err != nil || checked != p || p.Operation == ^uint64(0) || uint64(len(input)) != p.InputWords || capacity < max(p.Rows, p.Columns) || apply == nil {
		return nil, fmt.Errorf("rank matrix-vector plan/capacity/callback")
	}
	back, err := newRankTransposePlan(p.Operation+1, p.Columns, p.Rows, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if err != nil {
		return nil, err
	}
	rows := func(data []Word) error {
		return rankVectorRows(data, p.InputFirst, p.InputRows, p.Columns, capacity, false, inverse, apply, notice)
	}

	current := input
	var columnData []Word
	err = rankSeparableSchedule(inverse,
		func() error { return rows(current) },
		func() error { var e error; columnData, e = transposeRank(p, current, link, notice); return e },
		func() error {
			return rankVectorRows(columnData, p.OutputFirst, p.OutputRows, p.Rows, capacity, true, inverse, apply, notice)
		},
		func() error { var e error; current, e = transposeRank(back, columnData, link, notice); return e })
	if err != nil {
		return nil, err
	}
	return current, nil
}

// rankSeparableSchedule is shared by host-tile and resident storage. It fixes
// forward/inverse stage ordering without imposing array ownership or layout.
func rankSeparableSchedule(inverse bool, rows, transpose, columns, back func() error) error {
	if inverse {
		if err := rows(); err != nil {
			return err
		}
	}
	for _, stage := range []func() error{transpose, columns, back} {
		if err := stage(); err != nil {
			return err
		}
	}
	if !inverse {
		return rows()
	}
	return nil
}
