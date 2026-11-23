package liftrs

import "fmt"

// rankReciprocalPlan describes base-field coefficient ownership, not GPU slots.
// MaxLocalWords admits each shard; it is not a bound on total process memory.
type rankReciprocalPlan struct {
	Domain, RootCount, Target, Known, MaxLocalWords uint64
	Rank, Ranks                                     int
	PrefixInputs                                    bool
}

// Counters describe completed products, not partial failed transfers or durable
// checkpoint commits.
type rankReciprocalStats struct {
	Steps                                int
	RootCloneWordsPerStep, UploadedWords uint64
}
type rankReciprocalOps struct {
	// Both kernels borrow their source slices and retain their existing field,
	// ownership and padding checks. Product reports actual backend upload words.
	Square  func([]uint32, uint64) ([]uint32, error)
	Product func(string, []uint32, []uint32, uint64, uint64, uint64, bool) ([]uint32, uint64, error)
	// Checkpoint may replace values with a verified reload. The caller owns
	// durability, coordinated commit and hash verification; nil omits persistence.
	Checkpoint func(uint64, []uint32) ([]uint32, error)
}

// refineRankReciprocal shares cold/resumed Newton orchestration without native
// oracle data or a testing framework. In characteristic two, G_new=P*G^2 doubles
// the verified inverse prefix. Callers supply normalized P(0)=1 and a verified
// initial inverse prefix. They must still verify the full final code
// identity; a successful refinement alone does not establish a prepared code.
// Dense mode clones/masks the root as before. Prefix mode borrows its logical
// prefix unchanged and omits the unused tail, using the same nonaliasing product.
func refineRankReciprocal(p rankReciprocalPlan, root, initial []uint32, ops rankReciprocalOps) ([]uint32, rankReciprocalStats, error) {
	var stats rankReciprocalStats
	if p.Domain == 0 || p.Domain&(p.Domain-1) != 0 || p.Domain > uint64(^uint(0)>>1) || p.Ranks < 1 || p.Ranks > 64 || p.Ranks&(p.Ranks-1) != 0 || p.Domain < uint64(p.Ranks) || p.Rank < 0 || p.Rank >= p.Ranks || p.RootCount == 0 || p.RootCount > p.Domain || p.Target == 0 || p.Target > p.Domain-p.RootCount+1 || p.Known == 0 || p.Known > p.Target || (p.Known != p.Target && p.Known&(p.Known-1) != 0) || ops.Square == nil || ops.Product == nil {
		return nil, stats, fmt.Errorf("reciprocal plan admission")
	}
	first, words := rankBand(p.Domain, p.Rank, p.Ranks)
	if words > p.MaxLocalWords || uint64(len(root)) != words || uint64(len(initial)) != words {
		return nil, stats, fmt.Errorf("reciprocal shard admission")
	}
	g, known := initial, p.Known
	err := reciprocalSchedule(int(known), int(p.Target), func(size int) error {
		keep := uint64(size)
		squared, err := ops.Square(g, keep)
		if err != nil {
			return err
		}
		if uint64(len(squared)) != words {
			return fmt.Errorf("reciprocal square extent")
		}
		ac := min(p.RootCount, keep)
		a := root
		if !p.PrefixInputs {
			a = append([]uint32(nil), root...)
			stats.RootCloneWordsPerStep = uint64(len(a))
			for j := range a {
				if first+uint64(j) >= ac {
					a[j] = 0
				}
			}
		}
		next, uploaded, err := ops.Product(fmt.Sprintf("newton-%d", size), a, squared, ac, min(2*known-1, keep), keep, p.PrefixInputs)
		if err != nil {
			return err
		}
		if uint64(len(next)) != words {
			return fmt.Errorf("reciprocal product extent")
		}
		stats.UploadedWords += uploaded
		stats.Steps++
		g = next
		if ops.Checkpoint != nil {
			g, err = ops.Checkpoint(keep, g)
			if err != nil {
				return err
			}
			if uint64(len(g)) != words {
				return fmt.Errorf("reciprocal checkpoint extent")
			}
		}
		known = keep
		return nil
	})
	if err != nil {
		return nil, stats, err
	}
	return g, stats, nil
}

// verifyRankCodeIdentity checks this owned shard of P*G=1+X^(N-1). This is the
// full code identity, not merely a truncated inverse check. It needs no oracle;
// discovery, source provenance and complete code admission remain caller duties.
func verifyRankCodeIdentity(domain uint64, rank, ranks int, values []uint32) error {
	if domain < 2 || domain&(domain-1) != 0 || ranks < 1 || ranks > 64 || ranks&(ranks-1) != 0 || rank < 0 || rank >= ranks || domain < uint64(ranks) {
		return fmt.Errorf("code identity ownership")
	}
	first, words := rankBand(domain, rank, ranks)
	if uint64(len(values)) != words {
		return fmt.Errorf("code identity shard extent")
	}
	for j, value := range values {
		index := first + uint64(j)
		want := uint32(0)
		if index == 0 || index == domain-1 {
			want = 1
		}
		if value != want {
			return fmt.Errorf("full P*G identity at coefficient %d", index)
		}
	}
	return nil
}
