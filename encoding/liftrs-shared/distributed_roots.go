package liftrs

import "fmt"

// rootWorkerEngine is borrowed for local normalized root construction. The
// caller owns close and backend/field compatibility; no test or oracle is needed.
type rootWorkerEngine interface {
	Product([]uint32, []uint32, int, int) ([]uint32, error)
	RootFactorsInto([]int, [][4]uint32) error
	OwnedBytes() uint64
	Close() error
}

func rootFieldWithEngine(base orbitField, domain uint64, engine rootWorkerEngine) orbitField {
	base.products = &polynomialProducts{domain: domain, cutoff: 4096, convolver: preparationConvolver{engine}}
	return base
}

// prepareRankRoot shares the existing factor/product algorithm for one owned
// exponent subset. Domain is the complete quadratic-field grid, not a smaller
// temporary FFT extent. Empty ownership is the identity polynomial [1].
func prepareRankRoot(mode supportMembership, base orbitField, domain uint64, engine rootWorkerEngine, subset []int, checkpoint func(string), notice preparationNotice) (p []uint32, err error) {
	if base.q < 2 || base.q > 1<<20 || domain != base.q*base.q || domain > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("local root field/domain admission")
	}
	if len(subset) == 0 {
		return []uint32{1}, nil
	}
	if engine == nil {
		return nil, fmt.Errorf("local root engine required")
	}
	if checkpoint == nil {
		checkpoint = func(string) {}
	}
	defer recoverPreparation(&err)
	return prepareCyclicRootWith(mode, rootFieldWithEngine(base, domain, engine), subset, int(domain), base.q, engine.RootFactorsInto, nil, nil, checkpoint, notice)
}

// admitRankRootStages uses actual leaf lengths and rounded subgroup domains.
// The cap covers each base-coefficient input/output shard, not aggregate RSS,
// device tables/workspace, support storage, GC or complete job admission.
func admitRankRootStages(rank int, counts []uint64, maxDomain, maxLocalWords uint64) ([]rankRootMergeStage, error) {
	stages, err := rankRootMergeStages(rank, counts, maxDomain)
	if err != nil {
		return nil, err
	}
	if maxLocalWords == 0 || counts[rank] > maxLocalWords {
		return nil, fmt.Errorf("root leaf host-shard admission")
	}
	for _, stage := range stages {
		if stage.Domain/uint64(stage.Ranks) > maxLocalWords {
			return nil, fmt.Errorf("root stage host-shard admission at level %d", stage.Level)
		}
	}
	return stages, nil
}
