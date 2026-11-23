//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"math/rand"
	"os"
	"testing"
)

// Bounded experiment policy stays here;runtime has no deterministic RNG default.
func partitionedDiscovery[Word transposeWord](t *testing.T, w *mpi.World, spec Spec, emit func(map[string]any), archiveSink func([]uint64) error) discoveryResult {
	t.Helper()
	n := uint64(1) << uint(2*spec.FieldDegree)
	result, err := discoverRankSupport[Word](w, spec, rankDiscoveryOptions{Library: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), ResultMode: os.Getenv("PIR_PIPELINE_DISCOVERY_RESULT"), DeviceBytes: 64 << 20, MaxLocalWords: 8 << 20, MaxGoodWords: n, MaxRootWords: n, MaxArchiveWords: n, Random: rand.New(rand.NewSource(109))}, emit, archiveSink)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// Bound the actual assembled shapes and validate the new orbit ownership before
// an allocation; no CUDA/MPI initialization and no new calibration workload.
func TestPartitionedDiscoveryShapes(t *testing.T) {
	for _, s := range []int{4, 6} {
		support, _ := pilotReference(t, s)
		n := uint64(1) << uint(2*s)
		q := uint64(1) << uint(s)
		count := 0
		owners := [2][]int{}
		for rank := 0; rank < 2; rank++ {
			first, words := rankBand(n, rank, 2)
			mask, err := cpp.GoodMonomialRange(s, 2, first, int(words), 1)
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range mask {
				count += int(v)
			}
		}
		if count != len(support.exponents) {
			t.Fatal("local monomial count")
		}
		for _, index := range support.exponents {
			owner := discoveryRootOwner(uint64(index), q, n-1, 2)
			owners[owner] = append(owners[owner], index)
		}
		for _, part := range owners {
			if len(part) == 0 {
				t.Fatal("empty pilot orbit owner")
			}
			if err := validateExponentSupportWith(supportSorted, part, int(n), q, nil); err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("s=%d N=%d K=%d samples=%d owned roots=%d/%d", s, n, count, support.parameters.Samples, len(owners[0]), len(owners[1]))
	}
}
