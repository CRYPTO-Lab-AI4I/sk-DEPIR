//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"encoding/json"
	gpumatrix "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/gpu-matrix"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type gpuProductTestFallback struct{ fft *additiveTransform }

func (e *gpuProductTestFallback) apply(data []uint32, inverse bool) {
	e.fft.applySubspace(data, inverse)
}

func TestGPUResidentProducts(t *testing.T) {
	out := os.Getenv("PIR_GPU_PRODUCT_TEST_OUT")
	if out == "" {
		t.Skip("opt-in resident product checks")
	}
	base, ext, _, _ := orbitFields(4, 0x13)
	table := newPreparationProducts(base.base)
	engine, err := gpumatrix.New(os.Getenv("PIR_GPU_LIBRARY"), 4, 0x13, ext.extension.Delta(), table.log, table.powers)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	for dimension := 2; dimension <= 8; dimension++ {
		words, _, _, err := gpuMatrixDescription(ext, dimension)
		if err != nil {
			t.Fatal(err)
		}
		if err = engine.AddPlan(words); err != nil {
			t.Fatal(err)
		}
	}
	executor := &gpuPreparationExecutor{gpu: engine, fallback: &gpuProductTestFallback{fft: newSparseTransform(ext)}, resident: true, minimum: 4}
	fft := newSparseTransform(ext)
	fft.executor = executor
	rng := rand.New(rand.NewSource(1702))
	checked := 0
	for _, field := range []orbitField{base, ext} {
		products := &polynomialProducts{transform: fft, cutoff: 0, workspace: &productWorkspace{}}
		for _, shape := range [][2]int{{300, 257}, {128, 129}, {23, 24}, {3, 7}, {1, 1}} {
			a, b := make([]uint32, shape[0]), make([]uint32, shape[1])
			for i := range a {
				a[i] = uint32(rng.Intn(int(field.q)))
			}
			for i := range b {
				b[i] = uint32(rng.Intn(int(field.q)))
			}
			saveA, saveB := append([]uint32(nil), a...), append([]uint32(nil), b...)
			for _, limit := range []int{1, 2, 5, 17, 31, 64, 127, 255, len(a) + len(b) - 1} {
				limit = minInt(limit, len(a)+len(b)-1)
				got, want := products.multiply(a, b, limit), polyMulLow(field, a, b, limit)
				if !reflect.DeepEqual(got, want) {
					t.Fatal("resident full/prefix/overlap", field.q, shape, limit)
				}
				checked++
			}
			if !reflect.DeepEqual(a, saveA) || !reflect.DeepEqual(b, saveB) {
				t.Fatal("resident mutated inputs")
			}
			clear(a)
			got := products.multiply(a, b, len(a)+len(b)-1)
			for _, v := range got {
				if v != 0 {
					t.Fatal("stale resident padding")
				}
			}
			checked++
		}
	}
	for _, bad := range []struct {
		a, b     []uint32
		limit, n int
	}{{nil, []uint32{1}, 1, 4}, {[]uint32{1}, []uint32{1}, 2, 4}, {make([]uint32, 3), make([]uint32, 3), 1, 4}, {[]uint32{1}, []uint32{1}, 1, 1024}} {
		if _, _, err := engine.Product(bad.a, bad.b, bad.limit, bad.n); err == nil {
			t.Fatal("accepted invalid resident shape/prefix")
		}
	}
	engine.Close()
	if _, _, err := engine.Product([]uint32{1}, []uint32{1}, 1, 4); err == nil {
		t.Fatal("closed resident engine")
	}
	if engine.OwnedBytes() != 0 {
		t.Fatal("resident resources retained")
	}
	if old := os.Getenv("PIR_OLD_GPU_LIBRARY"); old != "" {
		previous, err := gpumatrix.New(old, 4, 0x13, ext.extension.Delta(), table.log, table.powers)
		if err != nil {
			t.Fatal(err)
		}
		defer previous.Close()
		words, _, _, err := gpuMatrixDescription(ext, 2)
		if err != nil {
			t.Fatal(err)
		}
		if err = previous.AddPlan(words); err != nil {
			t.Fatal(err)
		}
		input := []uint32{1, 2, 3, 4}
		if _, err = previous.Apply(input, false); err != nil {
			t.Fatal(err)
		}
		if _, _, err = previous.Product([]uint32{1}, []uint32{1}, 1, 4); err == nil {
			t.Fatal("old library unexpectedly exposes resident ABI")
		}
	}
	raw, _ := json.MarshalIndent(map[string]any{"cases_checked": checked, "product_calls": executor.products, "overlap_add": true, "input_immutability": true, "zero_padding": true, "invalid_inputs": true, "closed_engine": true, "old_transform_library": os.Getenv("PIR_OLD_GPU_LIBRARY") != ""}, "", "  ")
	if err = os.WriteFile(filepath.Join(out, "product-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
