package liftrs

import "fmt"

// rootTreeStep gives a unique owner for each subtree. Senders leave the tree;
// receivers merge exactly once per existing peer. Odd sizes carry an unmatched
// subtree unchanged. No padding ranks, duplicated final merge or global array
// gather is required. This schedule is independent of MPI, GPU or field degree.
type rootTreeStep struct {
	Peer, Stride int
	Send         bool
}

func rootTreeSteps(rank, ranks int) ([]rootTreeStep, error) {
	if ranks < 1 || ranks > 64 || rank < 0 || rank >= ranks {
		return nil, fmt.Errorf("root tree rank bounds")
	}
	var steps []rootTreeStep
	for stride := 1; stride < ranks; stride *= 2 {
		if rank%(2*stride) == stride {
			steps = append(steps, rootTreeStep{rank - stride, stride, true})
			break
		}
		if rank+stride < ranks {
			steps = append(steps, rootTreeStep{rank + stride, stride, false})
		}
	}
	return steps, nil
}

type rootTreeTransport interface {
	SendRoot(peer int, values []uint32, maxWords int) error
	ReceiveRoot(peer, maxWords int) ([]uint32, error)
}

// reduceRootProduct transfers one owned normalized polynomial per subtree and
// releases sender ownership (nil result). Transport is synchronous. The receiver
// works with left/right/result host polynomials while multiplying; caller
// references and GC can extend lifetimes. Full-size
// product/FFT workspace must still be admitted to its executor. This is NOT a
// distributed transform or a claim that the last product fits one GPU at s18.
func reduceRootProduct(rank, ranks, maxWords int, limit uint64, local []uint32, transport rootTreeTransport,
	multiply func([]uint32, []uint32) ([]uint32, error), notice func(rootTreeStep, string, []uint32)) ([]uint32, error) {
	steps, err := rootTreeSteps(rank, ranks)
	if err != nil {
		return nil, err
	}
	canonical := func(values []uint32) bool {
		if len(values) == 0 || len(values) > maxWords || values[0] != 1 {
			return false
		}
		for _, v := range values {
			if uint64(v) >= limit {
				return false
			}
		}
		return true
	}
	if maxWords < 1 || limit < 2 || limit > 1<<32 || transport == nil || multiply == nil || !canonical(local) {
		return nil, fmt.Errorf("root tree product bounds/normalization")
	}
	for _, step := range steps {
		if notice != nil {
			notice(step, "started", local)
		}
		if step.Send {
			if err := transport.SendRoot(step.Peer, local, maxWords); err != nil {
				return nil, err
			}
			if notice != nil {
				notice(step, "completed", local)
			}
			return nil, nil
		}
		peer, err := transport.ReceiveRoot(step.Peer, maxWords)
		if err != nil {
			return nil, err
		}
		if !canonical(peer) || len(local) > maxWords-len(peer)+1 {
			return nil, fmt.Errorf("root tree peer/product extent")
		}
		expected := len(local) + len(peer) - 1
		result, err := multiply(local, peer)
		if err != nil {
			return nil, err
		}
		if len(result) != expected || !canonical(result) {
			return nil, fmt.Errorf("root tree merged product")
		}
		local = result
		if notice != nil {
			notice(step, "completed", local)
		}
	}
	return local, nil
}
