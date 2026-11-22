//go:build liftrs_cuda && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
)

// rankGPUExecutor retains one admitted three-array workspace for successive
// root/Newton/identity products and the final spectrum. Inputs/results are owned
// coefficient shards; no complete polynomial is gathered by this executor.
// A failed operation invalidates the session; collective callers must abort.
type rankGPUExecutor[Word transposeWord] struct {
	GPU                *cuda.RankStorage
	Plan               rankTransposePlan
	Capacity, Limit    uint64
	Transpose          transposeLink[Word]
	Windows            taylorLink[Word]
	Notice             func(string, uint64, uint64)
	Degrees            func(int) ([]uint64, error)
	UploadedInputWords uint64
}

func (e *rankGPUExecutor[Word]) load(slot int, values []uint32, count uint64) error {
	return e.loadInput(slot, values, count, false)
}

// A prefix input is a logical view;unused source values are never uploaded.
// The product zeroes both destination slots before either load. Dense sources
// retain their complete canonical-value and zero-padding validation.
func (e *rankGPUExecutor[Word]) loadInput(slot int, values []uint32, count uint64, prefix bool) error {
	p := e.Plan
	if uint64(len(values)) != p.InputWords || count == 0 || count > p.Rows*p.Columns {
		return fmt.Errorf("executor input shard/count")
	}
	extent := p.InputWords
	if prefix {
		var err error
		extent, err = rankPrefixInputWords(p.Rows*p.Columns, p.InputFirst*p.Columns, p.InputWords, count)
		if err != nil {
			return err
		}
	}
	tile := make([]Word, min(p.TileWords, extent))
	for first := uint64(0); first < extent; {
		take := min(uint64(len(tile)), extent-first)
		for j := uint64(0); j < take; j++ {
			v := values[first+j]
			k := p.InputFirst*p.Columns + first + j
			if uint64(v) >= e.Limit || (!prefix && k >= count && v != 0) {
				return fmt.Errorf("executor canonical input/padding")
			}
			tile[j] = Word(v)
		}
		if err := storageTransfer(e.GPU, slot, first, tile[:take], true); err != nil {
			return err
		}
		e.UploadedInputWords += take
		first += take
	}
	return nil
}

func (e *rankGPUExecutor[Word]) Product(left, right []uint32, ac, bc, keep uint64) ([]uint32, error) {
	return e.product(left, right, ac, bc, keep, false)
}

// ProductPrefixInputs treats each source as its selected prefix followed by zero.
// It borrows source slices unchanged and shares the dense product's transforms.
func (e *rankGPUExecutor[Word]) ProductPrefixInputs(left, right []uint32, ac, bc, keep uint64) ([]uint32, error) {
	return e.product(left, right, ac, bc, keep, true)
}

func (e *rankGPUExecutor[Word]) product(left, right []uint32, ac, bc, keep uint64, prefix bool) ([]uint32, error) {
	p := e.Plan
	n := p.Rows * p.Columns
	if ac == 0 || bc == 0 || ac > n || bc > n-ac+1 || keep == 0 || keep > ac+bc-1 {
		return nil, fmt.Errorf("executor nonaliasing product extent")
	}
	plans, err := rankProductPlans(p, e.Capacity)
	if err != nil {
		return nil, err
	}
	// Every successive operation has a disjoint communication identity.
	e.Plan.Operation = plans[2].Operation + (plans[1].Operation - plans[0].Operation)
	for slot := 0; slot < 2; slot++ {
		if err = e.GPU.Reset(slot, p.Rows, p.Columns, false); err != nil {
			return nil, err
		}
	}
	if err = e.loadInput(0, left, ac, prefix); err != nil {
		return nil, err
	}
	if err = e.loadInput(1, right, bc, prefix); err != nil {
		return nil, err
	}
	if err = rankResidentPolynomialProduct[Word](p, e.GPU, 0, 1, 2, e.Capacity, e.Transpose, e.Windows, e.degrees, e.Notice, true); err != nil {
		return nil, err
	}
	out := make([]uint32, p.InputWords)
	tile := make([]Word, min(p.TileWords, p.InputWords))
	for first := uint64(0); first < p.InputWords; {
		take := min(uint64(len(tile)), p.InputWords-first)
		if err = storageTransfer(e.GPU, 0, first, tile[:take], false); err != nil {
			return nil, err
		}
		for j := uint64(0); j < take; j++ {
			k := p.InputFirst*p.Columns + first + j
			v := uint64(tile[j])
			if v >= e.Limit || (k >= ac+bc-1 && v != 0) {
				return nil, fmt.Errorf("executor product range/padding")
			}
			if k < keep {
				out[first+j] = uint32(v)
			}
		}
		first += take
	}
	return out, nil
}

// PrepareSpectrum leaves the transformed spectrum in GPU slot0. A caller may
// publish bounded reads before resetting/closing this workspace. This separates
// the transform from host ownership without adding transforms or transfers.
func (e *rankGPUExecutor[Word]) PrepareSpectrum(values []uint32, count uint64) error {
	p := e.Plan
	_, matrix, err := rankAdditivePlans(p, e.Capacity, false)
	if err != nil {
		return err
	}
	e.Plan.Operation = matrix.Operation + 2
	if err = e.GPU.Reset(0, p.Rows, p.Columns, false); err != nil {
		return err
	}
	if err = e.load(0, values, count); err != nil {
		return err
	}
	if err = rankResidentAdditiveTransform[Word](p, e.GPU, 0, 2, e.Capacity, false, e.Transpose, e.Windows, e.degrees, e.Notice, true); err != nil {
		return err
	}
	return nil
}

// Spectrum retains the original resident host result and allocation boundary.
func (e *rankGPUExecutor[Word]) Spectrum(values []uint32, count uint64) ([]Word, error) {
	if err := e.PrepareSpectrum(values, count); err != nil {
		return nil, err
	}
	return e.residentResult()
}

// residentResult keeps direct GPU reads into the final owned host allocation.
// Spectrum and discovery use the same width-selected transfer boundary.
func (e *rankGPUExecutor[Word]) residentResult() ([]Word, error) {
	p := e.Plan
	out := make([]Word, p.InputWords)
	for first := uint64(0); first < p.InputWords; {
		take := min(p.TileWords, p.InputWords-first)
		if err := storageTransfer(e.GPU, 0, first, out[first:first+take], false); err != nil {
			return nil, err
		}
		first += take
	}
	return out, nil
}

func (e *rankGPUExecutor[Word]) degrees(level int) ([]uint64, error) {
	if e.Degrees != nil {
		return e.Degrees(level)
	}
	return e.GPU.TaylorDegrees(level)
}

// PrepareDiscovery runs the same tensor Fq grid and extension interpolation
// on GPUs,retaining operation IDs and leaving the result in slot0. observe
// switches route identifiers between tensor and inverse stages.
func (e *rankGPUExecutor[Word]) PrepareDiscovery(values []uint32, observe func(bool, uint64)) error {
	p := e.Plan
	if p.Rows != p.Columns || p.Rows != e.Limit || p.Operation > ^uint64(0)-2 {
		return fmt.Errorf("discovery full base grid")
	}
	back, err := newRankTransposePlan(p.Operation+1, p.Columns, p.Rows, p.Rank, p.Ranks, max(p.InputWords, p.OutputWords), p.TileWords)
	if err != nil {
		return err
	}
	inverse := p
	inverse.Operation += 2
	_, _, err = rankAdditivePlans(inverse, e.Capacity, true)
	if err != nil {
		return err
	}
	// Inverse schedules consume the same complete operation span as forwards.
	plans, err := rankProductPlans(inverse, e.Capacity)
	if err != nil {
		return err
	}
	e.Plan.Operation = inverse.Operation + (plans[1].Operation - plans[0].Operation)
	if observe != nil {
		observe(false, p.Operation)
	}
	if err = e.GPU.Reset(0, p.Rows, p.Columns, false); err != nil {
		return err
	}
	if err = e.load(0, values, p.Rows*p.Columns); err != nil {
		return err
	}
	if err = transposeResident[Word](p, e.GPU, 0, 2, true, e.Transpose, e.Notice); err != nil {
		return err
	}
	if err = e.GPU.BaseVectors(2); err != nil {
		return err
	}
	if err = transposeResident[Word](back, e.GPU, 2, 0, true, e.Transpose, e.Notice); err != nil {
		return err
	}
	if err = e.GPU.BaseVectors(0); err != nil {
		return err
	}
	if observe != nil {
		observe(true, inverse.Operation)
	}
	if err = rankResidentAdditiveTransform[Word](inverse, e.GPU, 0, 2, e.Capacity, true, e.Transpose, e.Windows, e.degrees, e.Notice, true); err != nil {
		return err
	}
	return nil
}

// Discovery preserves the resident host result and original direct read targets.
func (e *rankGPUExecutor[Word]) Discovery(values []uint32, observe func(bool, uint64)) ([]Word, error) {
	if err := e.PrepareDiscovery(values, observe); err != nil {
		return nil, err
	}
	return e.residentResult()
}

// DiscoveryTiles reads each local result word once into one bounded tile.
// consume must process it before return and must not reset the GPU workspace.
// Callback/transfer errors invalidate the collective session just like Discovery.
func (e *rankGPUExecutor[Word]) DiscoveryTiles(values []uint32, observe func(bool, uint64), consume func(uint64, []Word) error) error {
	if consume == nil {
		return fmt.Errorf("discovery tile consumer required")
	}
	if err := e.PrepareDiscovery(values, observe); err != nil {
		return err
	}
	p := e.Plan
	tile := make([]Word, min(p.TileWords, p.InputWords))
	if e.Notice != nil {
		e.Notice("discovery-result-consume", 0, p.InputWords)
	}
	for first := uint64(0); first < p.InputWords; {
		take := min(uint64(len(tile)), p.InputWords-first)
		values := tile[:take]
		if err := storageTransfer(e.GPU, 0, first, values, false); err != nil {
			return err
		}
		if err := consume(first, values); err != nil {
			return err
		}
		first += take
		if e.Notice != nil {
			e.Notice("discovery-result-consume", first, p.InputWords)
		}
	}
	return nil
}

// MergeRoots consumes only this rank's child shard and a validated stage plan.
// All source intervals and field/padding bounds are checked before multiplication.
// reduceShardedRoots additionally enforces normalization/degree after each stage.
// Communicator lifetime and abort-on-peer-failure remain the caller's responsibility.
func (e *rankGPUExecutor[Word]) MergeRoots(stage rankRootMergeStage, input []uint32, link operandLink) ([]uint32, error) {
	if e == nil || e.GPU == nil {
		return nil, fmt.Errorf("root merge executor unavailable")
	}
	p := e.Plan
	n := p.Rows * p.Columns
	ac, bc := stage.Sources[0].Count, stage.Sources[1].Count
	if stage.Ranks != p.Ranks || stage.Domain != n || ac == 0 || bc == 0 || ac > n || bc > n-ac+1 || stage.Count != ac+bc-1 {
		return nil, fmt.Errorf("root merge executor shape")
	}
	var local [2][]uint32
	for j, source := range stage.Sources {
		if p.Rank >= source.First && p.Rank < source.First+source.Ranks {
			local[j] = input
		}
	}
	pair, err := routeShardedOperands(rankShardOperandPlan{p.Operation, n, p.TileWords, e.Limit, p.InputWords, p.Rank, p.Ranks, stage.Sources}, local, link, nil)
	if err != nil {
		return nil, err
	}
	return e.Product(pair[0], pair[1], ac, bc, stage.Count)
}
