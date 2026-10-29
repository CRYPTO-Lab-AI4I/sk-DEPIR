//go:build liftrs_cpp && cgo

package liftrs

import (
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestCPPOrderedAffine(t *testing.T) {
	polynomial, _ := BinaryPolynomial(4)
	_, f, _, _ := orbitFields(4, polynomial)
	for _, basis := range [][]uint32{{3, 5, 9}, {1, 16, 2, 32}, {7, 11}} {
		a, err := orderedTransform(f, basis)
		if err != nil {
			t.Fatal(err)
		}
		p, err := orderedCPP(a, basis)
		if err != nil {
			t.Fatal(err)
		}
		for _, offset := range []uint32{0, 1, 91, 255} {
			input := make([]uint32, 1<<len(basis))
			for i := range input {
				input[i] = uint32((31*i + 7) % int(f.q))
			}
			got := append([]uint32(nil), input...)
			if _, err := p.ApplyAt(got, offset, false); err != nil {
				t.Fatal(err)
			}
			for i, x := range got {
				point := offset
				for j, b := range basis {
					if i&(1<<j) != 0 {
						point ^= b
					}
				}
				var want uint32
				for k := len(input) - 1; k >= 0; k-- {
					want = f.mul(want, point) ^ input[k]
				}
				if x != want {
					t.Fatal("affine Horner mismatch", basis, offset, i)
				}
			}
			if _, err := p.ApplyAt(got, offset, true); err != nil || !reflect.DeepEqual(got, input) {
				t.Fatal("affine inverse", err)
			}
		}
		p.Close()
	}
	if _, err := orderedTransform(f, []uint32{1, 1}); err == nil {
		t.Fatal("dependent basis accepted")
	}
}
func TestCPPMatrixTransform(t *testing.T) {
	for s := 2; s <= 5; s++ {
		polynomial, _ := BinaryPolynomial(s)
		_, f, _, _ := orbitFields(s, polynomial)
		for dimension := 2; dimension <= 2*s; dimension++ {
			n := 1 << dimension
			plan, err := newMatrixCPP(f, dimension)
			if err != nil {
				t.Fatal(err)
			}
			input := make([]uint32, n)
			random := rand.New(rand.NewSource(int64(100 + n)))
			for i := range input {
				input[i] = uint32(random.Intn(int(f.q)))
			}
			want := append([]uint32(nil), input...)
			newSparseTransform(f).applySubspace(want, false)
			for _, workers := range []int{1, 2, 4} {
				got := append([]uint32(nil), input...)
				scratch := make([]uint32, n)
				if err := liftrscpp.ApplyMatrix(plan.rows, plan.columns, got, scratch, plan.split, workers, false); err != nil || !reflect.DeepEqual(got, want) {
					t.Fatal("matrix forward", s, dimension, workers, err)
				}
				if err := liftrscpp.ApplyMatrix(plan.rows, plan.columns, got, scratch, plan.split, workers, true); err != nil || !reflect.DeepEqual(got, input) {
					t.Fatal("matrix inverse", err)
				}
			}
			if err := liftrscpp.ApplyMatrix(plan.rows, plan.columns, input, input, plan.split, 2, false); err == nil {
				t.Fatal("alias accepted")
			}
			plan.close()
		}
	}
}
func TestCPPMatrixEncoding(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(1101))})
		if err != nil {
			t.Fatal(err)
		}
		e, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		p, err := e.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		for _, workers := range []int{1, 4} {
			native, err := p.WithCPPMatrix(workers)
			if err != nil {
				t.Fatal(err)
			}
			w := native.NewWorkspace()
			for repeat := 0; repeat < 2; repeat++ {
				input := make([]uint32, b.Dimension())
				for i := range input {
					input[i] = uint32((i*7 + repeat) % int(b.base.q))
				}
				want, err := e.Encode(input)
				if err != nil {
					t.Fatal(err)
				}
				got, err := native.Encode(input, w)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatal("complete matrix word", spec, workers, err)
				}
				if spec.FieldDegree <= 4 {
					oracle, err := Preprocess(spec)
					if err != nil {
						t.Fatal(err)
					}
					values := make([]uint32, oracle.Dimension())
					for i, pos := range oracle.SystematicPositions() {
						values[i] = got[pos]
					}
					dense, err := oracle.Encode(values)
					if err != nil || !reflect.DeepEqual(dense, got) {
						t.Fatal("dense oracle", err)
					}
				}
			}
			native.Close()
			if _, err := native.Encode(make([]uint32, p.k), w); err != ErrCPPClosed {
				t.Fatal("closed encoder", err)
			}
		}
	}
}

// Distinct workspaces share immutable native plans; concurrent writes remain
// disjoint. The Go race detector covers scheduling/ownership, not native loads.
func TestCPPMatrixConcurrent(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(1503))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	native, err := p.WithCPPMatrix(2)
	if err != nil {
		t.Fatal(err)
	}
	defer native.Close()
	input := make([]uint32, p.k)
	for i := range input {
		input[i] = uint32(i % 16)
	}
	want, err := e.Encode(input)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := native.Encode(input, native.NewWorkspace())
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Error("concurrent word", err)
			}
		}()
	}
	wg.Wait()
}
