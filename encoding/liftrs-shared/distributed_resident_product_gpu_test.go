//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"reflect"
	"testing"
)

// Direct extension-field checks must cover high components even when a small
// polynomial's evaluation domain lies inside the base field. Tail zeros and a
// third-slot sentinel also detect writes outside the requested operand views.
func residentPointwiseCheck[Word transposeWord](t *testing.T, gpu *cuda.RankStorage, field wideOrbitField, s, dim int, rows, cols, words, capacity uint64, emit func(map[string]any)) {
	t.Helper()
	count := min(words, capacity, 17)
	left, right, want, third := make([]Word, words), make([]Word, words), make([]Word, words), make([]Word, words)
	mask := (uint64(1) << uint(2*s)) - 1
	for j := uint64(0); j < count; j++ {
		left[j] = Word((uint64(j+1)*0x9e3779b9 ^ (uint64(1) << uint(2*s-1))) & mask)
		right[j] = Word((uint64(j+3)*0x16379 ^ (uint64(1) << uint(s))) & mask)
		want[j] = Word(field.mul(uint64(left[j]), uint64(right[j])))
	}
	for j := range third {
		third[j] = ^Word(j)
	}
	storageIO(t, gpu, 0, left, true)
	storageIO(t, gpu, 1, right, true)
	storageIO(t, gpu, 2, third, true)
	if err := gpu.Pointwise(0, 0); err == nil {
		t.Fatal("resident pointwise alias accepted")
	}
	if err := gpu.Reset(1, rows, cols, true); err != nil {
		t.Fatal(err)
	}
	if err := gpu.Pointwise(0, 1); err == nil {
		t.Fatal("pointwise mismatched layout accepted")
	}
	if err := gpu.Reset(1, rows, cols, false); err != nil {
		t.Fatal(err)
	}
	invalid := append([]Word(nil), right...)
	invalid[0] = Word(uint64(1) << uint(2*s))
	storageIO(t, gpu, 1, invalid, true)
	if err := gpu.Pointwise(0, 1); err == nil {
		t.Fatal("pointwise invalid field accepted")
	}
	saved := make([]Word, words)
	storageIO(t, gpu, 0, saved, false)
	if !reflect.DeepEqual(saved, left) {
		t.Fatal("rejected pointwise mutated left")
	}
	storageIO(t, gpu, 1, saved, false)
	if !reflect.DeepEqual(saved, invalid) {
		t.Fatal("rejected pointwise mutated right")
	}
	storageIO(t, gpu, 1, right, true)
	if err := gpu.Pointwise(0, 1); err != nil {
		t.Fatal(err)
	}
	got := make([]Word, words)
	storageIO(t, gpu, 0, got, false)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("resident pointwise/scalar extension field")
	}
	storageIO(t, gpu, 1, got, false)
	if !reflect.DeepEqual(got, right) {
		t.Fatal("pointwise changed right")
	}
	storageIO(t, gpu, 2, got, false)
	if !reflect.DeepEqual(got, third) {
		t.Fatal("pointwise changed scratch")
	}
	emit(map[string]any{"kind": "pointwise-check", "s": s, "dimension": dim, "words": count, "checked_words": words, "matches": true, "resident": true, "alias_rejected": true, "layout_rejected": true, "invalid_rejected": true, "third_preserved": true})
}
