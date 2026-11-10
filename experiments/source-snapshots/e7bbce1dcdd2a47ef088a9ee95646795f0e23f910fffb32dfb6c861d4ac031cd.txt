package liftrs

import (
	"math/rand"
	"reflect"
	"testing"
)

type testResidentProduct struct {
	field    orbitField
	fallback *additiveTransform
	calls    int
	bad      bool
}

func (e *testResidentProduct) apply(data []uint32, inverse bool) {
	e.fallback.applySubspace(data, inverse)
}
func (e *testResidentProduct) canConvolve(n int) bool { return n >= 64 }
func (e *testResidentProduct) convolve(a, b []uint32, limit, n int) []uint32 {
	e.calls++
	if n < len(a)+len(b)-1 || n > int(e.fallback.f.q) || limit > len(a)+len(b)-1 {
		panic("resident shape contract")
	}
	if e.bad {
		return nil
	}
	return polyMulLow(e.field, a, b, limit)
}

func TestResidentProductOverlapAndPrefixes(t *testing.T) {
	base, ext, _, _ := orbitFields(4, 0x13)
	fft := newSparseTransform(ext)
	executor := &testResidentProduct{field: base, fallback: newSparseTransform(ext)}
	fft.executor = executor
	workspace := &productWorkspace{}
	products := &polynomialProducts{transform: fft, cutoff: 0, workspace: workspace}
	rng := rand.New(rand.NewSource(1701))
	a, b := make([]uint32, 300), make([]uint32, 257)
	for i := range a {
		a[i] = uint32(rng.Intn(16))
	}
	for i := range b {
		b[i] = uint32(rng.Intn(16))
	}
	saveA, saveB := append([]uint32(nil), a...), append([]uint32(nil), b...)
	for _, limit := range []int{1, 17, 64, 127, 128, 129, 255, 300, 556} {
		got := products.multiply(a, b, limit)
		want := polyMulLow(base, a, b, limit)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("resident overlap/prefix", limit)
		}
		if workspace.busy {
			t.Fatal("resident workspace lease retained")
		}
	}
	if executor.calls == 0 || !reflect.DeepEqual(a, saveA) || !reflect.DeepEqual(b, saveB) {
		t.Fatal("resident input/call invariant")
	}
	executor.bad = true
	func() {
		defer func() {
			if recover() == nil {
				t.Error("accepted bad resident output length")
			}
		}()
		products.multiply(a, b, 300)
	}()
	if workspace.busy {
		t.Fatal("workspace lease retained after resident panic")
	}
}
