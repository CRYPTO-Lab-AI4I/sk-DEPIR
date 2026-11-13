//go:build liftrs_cpp && cgo

package liftrscpp

import (
	field "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func wideTestField(t *testing.T, s int) (uint32, field.QuadraticField) {
	t.Helper()
	p, err := field.BinaryPolynomial(s)
	if err != nil {
		t.Fatal(err)
	}
	base, err := field.NewBinaryField(s, p)
	if err != nil {
		t.Fatal(err)
	}
	tower, err := field.NewQuadraticField(base)
	if err != nil {
		t.Fatal(err)
	}
	return p, tower
}
func wideTestMul(f field.QuadraticField, a, b uint64) uint64 {
	s := uint(f.Base().Degree())
	mask := uint64(1)<<s - 1
	x := f.Mul(field.QuadraticElement{A: uint32(a & mask), B: uint32(a >> s)}, field.QuadraticElement{A: uint32(b & mask), B: uint32(b >> s)})
	return uint64(x.A) | uint64(x.B)<<s
}
func wideTestEvaluate(f field.QuadraticField, input []uint64, x uint64) uint64 {
	var y uint64
	for j := len(input) - 1; j >= 0; j-- {
		y = wideTestMul(f, y, x) ^ input[j]
	}
	return y
}

func TestWidePlan(t *testing.T) {
	random := rand.New(rand.NewSource(170053))
	for _, s := range []int{8, 16, 17, 18, 20} {
		poly, f := wideTestField(t, s)
		q := uint64(1) << uint(2*s)
		basis := []uint64{1 << uint(2*s-1), 1, 2, 4, 8, 16}
		saved := append([]uint64(nil), basis...)
		p, err := NewWidePlan(s, poly, basis)
		if err != nil {
			t.Fatal(err)
		}
		basis[0] = 0 // Native metadata must own a copy.
		if p.OwnedBytes() == 0 {
			t.Fatal("missing metadata accounting")
		}
		input := make([]uint64, 64)
		for i := range input {
			input[i] = random.Uint64() & (q - 1)
		}
		input[0] = q - 1
		start := uint64(1) << uint(2*s-2)
		want := make([]uint64, len(input))
		for i := range want {
			x := start
			for j, b := range saved {
				if i&(1<<uint(j)) != 0 {
					x ^= b
				}
			}
			want[i] = wideTestEvaluate(f, input, x)
		}
		got := append([]uint64(nil), input...)
		if _, err = p.ApplyAt(got, start, false); err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("affine oracle", s, err)
		}
		if _, err = p.ApplyAt(got, start, true); err != nil || !reflect.DeepEqual(got, input) {
			t.Fatal("affine inverse", s, err)
		}
		bad := append([]uint64(nil), input...)
		bad[3] = q
		before := append([]uint64(nil), bad...)
		if _, err = p.ApplyAt(bad, 0, false); err == nil || !reflect.DeepEqual(bad, before) {
			t.Fatal("range validation mutated input")
		}
		if _, err = p.ApplyAt(got[:3], 0, false); err == nil {
			t.Fatal("bad shape accepted")
		}
		if _, err = p.ApplyAt(got, q, false); err == nil {
			t.Fatal("bad offset accepted")
		}
		p.Close()
		p.Close()
		if p.OwnedBytes() != 0 {
			t.Fatal("closed accounting")
		}
		if _, err = p.ApplyAt(got, 0, false); err == nil {
			t.Fatal("closed handle accepted")
		}
		if _, err = NewWidePlan(s, poly, []uint64{1, 1}); err == nil {
			t.Fatal("dependent basis accepted")
		}
	}
	if _, err := NewWidePlan(17, 0, []uint64{1}); err == nil {
		t.Fatal("bad field accepted")
	}
}

func TestWideMatrix(t *testing.T) {
	random := rand.New(rand.NewSource(170054))
	for _, s := range []int{8, 16, 17, 18, 20} {
		poly, f := wideTestField(t, s)
		q := uint64(1) << uint(2*s)
		for dim := 2; dim <= 8; dim++ {
			n := 1 << uint(dim)
			input := make([]uint64, n)
			for i := range input {
				input[i] = random.Uint64() & (q - 1)
			}
			input[0] = q - 1
			want := make([]uint64, n)
			for i := range want {
				want[i] = wideTestEvaluate(f, input, uint64(i))
			}
			for split := 1; split < dim; split++ {
				m, err := NewWideMatrix(s, poly, dim, split)
				if err != nil {
					t.Fatal(err)
				}
				owned := m.OwnedBytes()
				if owned == 0 {
					t.Fatal("missing matrix metadata")
				}
				for _, workers := range []int{1, 4} {
					data := append([]uint64(nil), input...)
					scratch := make([]uint64, n)
					if _, err = m.Apply(data, scratch, workers, false); err != nil || !reflect.DeepEqual(data, want) {
						t.Fatal("matrix oracle", s, dim, split, workers, err)
					}
					if _, err = m.Apply(data, scratch, workers, true); err != nil || !reflect.DeepEqual(data, input) {
						t.Fatal("matrix inverse", s, dim, split, workers, err)
					}
					if _, err = m.Apply(data, data, workers, false); err == nil {
						t.Fatal("alias accepted")
					}
					overlap := make([]uint64, n+1)
					if _, err = m.Apply(overlap[:n], overlap[1:], workers, false); err == nil {
						t.Fatal("partial overlap accepted")
					}
					data[0] = q
					before := append([]uint64(nil), data...)
					if _, err = m.Apply(data, scratch, workers, false); err == nil || !reflect.DeepEqual(data, before) {
						t.Fatal("matrix range guard")
					}
					data[0] = input[0]
					if _, err = m.Apply(data, scratch, 0, false); err == nil {
						t.Fatal("zero workers accepted")
					}
					if _, err = m.Apply(data, scratch, workers, false); err != nil || !reflect.DeepEqual(data, want) {
						t.Fatal("matrix error recovery", err)
					}
				}
				if m.OwnedBytes() != owned {
					t.Fatal("mutable plan")
				}
				m.Close()
				m.Close()
				if m.OwnedBytes() != 0 {
					t.Fatal("matrix close accounting")
				}
				if _, err = m.Apply(input, make([]uint64, n), 1, false); err == nil {
					t.Fatal("closed matrix accepted")
				}
			}
		}
	}
}

func TestWideMatrixParallelTaylor(t *testing.T) {
	poly, _ := wideTestField(t, 17)
	random := rand.New(rand.NewSource(170055))
	for _, dim := range []int{15, 16} {
		n := 1 << uint(dim)
		basis := make([]uint64, dim)
		for j := range basis {
			basis[j] = 1 << uint(j)
		}
		recursive, err := NewWidePlan(17, poly, basis)
		if err != nil {
			t.Fatal(err)
		}
		input := make([]uint64, n)
		for i := range input {
			input[i] = random.Uint64() & ((1 << 34) - 1)
		}
		want := append([]uint64(nil), input...)
		if _, err = recursive.ApplyAt(want, 0, false); err != nil {
			t.Fatal(err)
		}
		for _, split := range []int{1, 8, dim - 1} {
			m, err := NewWideMatrix(17, poly, dim, split)
			if err != nil {
				t.Fatal(err)
			}
			data := append([]uint64(nil), input...)
			scratch := make([]uint64, n)
			if _, err = m.Apply(data, scratch, 4, false); err != nil || !reflect.DeepEqual(data, want) {
				t.Fatal("parallel Taylor", dim, split, err)
			}
			if _, err = m.Apply(data, scratch, 4, true); err != nil || !reflect.DeepEqual(data, input) {
				t.Fatal("parallel inverse", dim, split, err)
			}
			m.Close()
		}
		recursive.Close()
	}
	// Full34-level plan metadata must preserve counts without allocating2^34 data.
	m, err := NewWideMatrix(17, poly, 34, 17)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if m.n != 1<<34 {
		t.Fatal("count truncation")
	}
	if _, err = m.Apply(make([]uint64, 8), make([]uint64, 8), 1, false); err == nil {
		t.Fatal("short full-shape buffer accepted")
	}
}

func TestWideCloseConcurrent(t *testing.T) {
	poly, _ := wideTestField(t, 17)
	m, err := NewWideMatrix(17, poly, 6, 3)
	if err != nil {
		t.Fatal(err)
	}
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			data := make([]uint64, 64)
			data[0] = 1 << 33
			scratch := make([]uint64, 64)
			if _, err := m.Apply(data, scratch, 2, false); err == nil {
				for _, v := range data {
					if v != 1<<33 {
						t.Error("concurrent output")
					}
				}
			} else if !strings.Contains(err.Error(), "closed") {
				t.Error(err)
			}
		}()
	}
	m.Close()
	group.Wait()
	m.Close()
	var nilPlan *WidePlan
	nilPlan.Close()
	if nilPlan.OwnedBytes() != 0 {
		t.Fatal("nil plan accounting")
	}
}
