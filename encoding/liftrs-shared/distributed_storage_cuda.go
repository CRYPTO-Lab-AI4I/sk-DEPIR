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
