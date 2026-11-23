//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"sort"
	"testing"
)

// rootFixture shares deterministic canonical-field inputs across scheduling
// experiments. Synthetic subsets are Frobenius-closed, not genuine PIR codes.
// No full domain grid is allocated; s>=17 cases deliberately include wide indices.
type rootFixture struct {
	header     ScalingCodeHeader
	extension  wideOrbitField
	support    []int
	n, maximum int
	base       orbitField
}

func makeRootFixture(s, count int, genuine bool) (*rootFixture, error) {
	fixture, err := makeRootArithmetic(s)
	if err != nil {
		return nil, err
	}
	n := fixture.n
	var support []int
	if genuine {
		basis, err := DiscoverOrbit(Spec{s, 2, 5}, OrbitOptions{Random: rand.New(rand.NewSource(109))})
		if err != nil {
			return nil, err
		}
		support = append([]int(nil), basis.support...)
	} else {
		if count < 2 || count%2 != 0 || uint64(count) >= uint64(n)-(1<<uint(s)) {
			return nil, fmt.Errorf("bounded root fixture count")
		}
		rng := rand.New(rand.NewSource(int64(109 + s + count)))
		found := map[int]bool{}
		for len(found) < count {
			exponent := int(rng.Uint64() % uint64(n-1))
			partner := int(uint64(exponent) * (1 << uint(s)) % uint64(n-1))
			if exponent == partner {
				continue
			}
			found[exponent], found[partner] = true, true
		}
		for exponent := range found {
			support = append(support, exponent)
		}
		sort.Ints(support)
	}
	if s >= 17 && support[len(support)-1] <= 1<<32 {
		return nil, fmt.Errorf("wide root indices absent")
	}
	maximum := 4
	for maximum < len(support)+1 {
		maximum *= 2
	}
	fixture.support, fixture.maximum = support, maximum
	return fixture, nil
}

// Field-only setup permits non-coordinator workers to avoid a full support or
// full CPU oracle. The same canonical arithmetic is used by both fixtures.
func makeRootArithmetic(s int) (*rootFixture, error) {
	n := 1 << uint(2*s)
	h, extension, err := canonicalScalingCodeHeader(s, 5, 1, uint64(n))
	if err != nil {
		return nil, err
	}
	base, err := NewBinaryField(s, h.Polynomial)
	if err != nil {
		return nil, err
	}
	return &rootFixture{header: h, extension: extension, n: n, maximum: 4,
		base: orbitField{q: 1 << uint(s), base: base, lookup: newBoundedPreparationProducts(base, 18)}}, nil
}

func (f *rootFixture) withEngine(engine rootWorkerEngine) orbitField {
	return rootFieldWithEngine(f.base, uint64(f.n), engine)
}

func (f *rootFixture) root(mode supportMembership, engine rootWorkerEngine, subset []int, checkpoint func(string), notice preparationNotice) ([]uint32, error) {
	return prepareRankRoot(mode, f.base, uint64(f.n), engine, subset, checkpoint, notice)
}

// These complete normalized-polynomial digests were independently retained by
// job77213 before the shared fixture extraction. They guard seed, field,
// support ordering and root normalization without requiring a GPU or MPI run.
func TestRootFixtureReferenceDigests(t *testing.T) {
	cases := []struct {
		s, count int
		genuine  bool
		digest   string
	}{
		{8, 0, true, "346dd306a890fb403c787cf1a7edfb7adb8e4038f2844d00e44c32ce3e8c776a"},
		{12, 1 << 16, false, "6d11674d02c933e8e4d15f966084d4626e00aff0802db88ead6535dfdd3e3d98"},
		{17, 1 << 16, false, "e0e7070a60c00c836dce6ce806a35200ede8149454b8c6743d49c11245301649"},
		{17, 1 << 18, false, "f167d29b7e5212b1e63b4215bc0ef34427a562e8e569b9d4c0cc74f1a52fc223"},
	}
	for _, item := range cases {
		fixture, err := makeRootFixture(item.s, item.count, item.genuine)
		if err != nil {
			t.Fatal(err)
		}
		engine, err := cpp.NewPreparation(item.s, fixture.header.Polynomial, fixture.header.Omega, 1, uint64(fixture.maximum), 64<<20)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := fixture.root(supportSorted, engine, fixture.support, func(string) {}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err = engine.Close(); err != nil {
			t.Fatal(err)
		}
		h := sha256.New()
		if err = binary.Write(h, binary.LittleEndian, actual); err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", h.Sum(nil)) != item.digest {
			t.Fatal("historical fixture changed", item.s, item.count)
		}
	}
}
