package liftrs

import "fmt"

// rootShardLayout describes one normalized polynomial split among a contiguous
// subgroup. First is a rank in the containing merge group; Domain includes padding.
type rootShardLayout struct {
	First, Ranks  int
	Domain, Count uint64
}

// First is an original-rank offset; Sources.First is local to this group.
type rankRootMergeStage struct {
	Level, First, Ranks int
	Domain, Count       uint64
	Sources             [2]rootShardLayout
}

// rankRootMergeStages admits the complete path before data movement. Equal-sized
// power-of-two groups grow2,4,...; sibling groups execute concurrently. Canonical
// matrix splits need at least Ranks rows/columns,even for identity-only subtrees.
func rankRootMergeStages(rank int, counts []uint64, maxDomain uint64) ([]rankRootMergeStage, error) {
	ranks := len(counts)
	if ranks < 2 || ranks > 64 || ranks&(ranks-1) != 0 || rank < 0 || rank >= ranks || maxDomain < 4 || maxDomain&(maxDomain-1) != 0 || maxDomain > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("root shard tree admission")
	}
	nodes := make([]rootShardLayout, ranks)
	for i, count := range counts {
		if count == 0 || count > maxDomain {
			return nil, fmt.Errorf("root leaf extent")
		}
		nodes[i] = rootShardLayout{i, 1, count, count}
	}
	var stages []rankRootMergeStage
	for width, level := 2, 1; width <= ranks; width, level = width*2, level+1 {
		parents := make([]rootShardLayout, len(nodes)/2)
		for i := range parents {
			a, b := nodes[2*i], nodes[2*i+1]
			if a.Count > maxDomain-b.Count+1 {
				return nil, fmt.Errorf("root merged extent")
			}
			count := a.Count + b.Count - 1
			domain := uint64(4)
			for domain < count || domain < uint64(width*width) {
				if domain > maxDomain/2 {
					return nil, fmt.Errorf("root merge domain cap")
				}
				domain *= 2
			}
			parents[i] = rootShardLayout{a.First, width, domain, count}
			if rank >= a.First && rank < a.First+width {
				a.First, b.First = 0, width/2
				stages = append(stages, rankRootMergeStage{level, parents[i].First, width, domain, count, [2]rootShardLayout{a, b}})
			}
		}
		nodes = parents
	}
	return stages, nil
}

// reduceShardedRoots keeps only the current local polynomial shard. The executor
// owns communicator creation,redistribution and one distributed product per group;
// it borrows input immutably and must release each group's device/transport
// workspace before the next level.
// On error peers must abort; partial results must never be published as complete.
func reduceShardedRoots(rank int, counts []uint64, maxDomain, limit uint64, local []uint32,
	merge func(rankRootMergeStage, []uint32) ([]uint32, error)) ([]uint32, error) {
	stages, err := rankRootMergeStages(rank, counts, maxDomain)
	if err != nil {
		return nil, err
	}
	if limit < 2 || limit > 1<<32 || merge == nil || uint64(len(local)) != counts[rank] {
		return nil, fmt.Errorf("root shard input")
	}
	validate := func(values []uint32, first, count uint64) error {
		for j, v := range values {
			k := first + uint64(j)
			if uint64(v) >= limit || (k >= count && v != 0) || (k == 0 && v != 1) || (k == count-1 && v == 0) {
				return fmt.Errorf("root shard canonical degree/padding")
			}
		}
		return nil
	}
	if err = validate(local, 0, counts[rank]); err != nil {
		return nil, err
	}
	for _, stage := range stages {
		result, e := merge(stage, local)
		if e != nil {
			return nil, e
		}
		first, words := rankBand(stage.Domain, rank-stage.First, stage.Ranks)
		if uint64(len(result)) != words {
			return nil, fmt.Errorf("root merged shard extent")
		}
		if e = validate(result, first, stage.Count); e != nil {
			return nil, e
		}
		local = result
	}
	return local, nil
}
