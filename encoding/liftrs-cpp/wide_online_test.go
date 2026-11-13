//go:build liftrs_cpp && cgo

package liftrscpp

import (
	field "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared/field"
	"testing"
)

func TestWideOnlineCounts(t *testing.T) {
	for _, s := range []int{16, 17, 18, 20} {
		polynomial, f := wideTestField(t, s)
		n := uint64(1) << uint(2*s)
		order := n - 1
		remaining := order
		var primes []uint64
		for d := uint64(2); d <= remaining/d; d++ {
			if remaining%d == 0 {
				primes = append(primes, d)
				for remaining%d == 0 {
					remaining /= d
				}
			}
		}
		if remaining > 1 {
			primes = append(primes, remaining)
		}
		var omega uint64
		for candidate := uint64(1) << uint(s); candidate < n; candidate++ {
			mask := uint64(1)<<uint(s) - 1
			element := field.QuadraticElement{A: uint32(candidate & mask), B: uint32(candidate >> uint(s))}
			primitive := true
			for _, p := range primes {
				if f.Pow(element, order/p) == (field.QuadraticElement{A: 1}) {
					primitive = false
					break
				}
			}
			if primitive {
				omega = candidate
				break
			}
		}
		e, err := NewWideOnline(s, polynomial, omega, n/4)
		if err != nil {
			t.Fatal(err)
		}
		layout := e.Layout()
		if layout.N != n || layout.K != n/4 || layout.Prefix != n/2 || layout.Arena != n || layout.WorkspaceBytes != 17*n || layout.OutputBackingBytes != 8*n || e.OwnedBytes() > 16<<20 {
			t.Fatal("wide count/allocation inventory", s, layout)
		}
		if _, _, err = e.EncodeBorrowed(nil, nil, nil, nil, nil, nil, nil, 1); err == nil {
			t.Fatal("missing huge buffers accepted")
		}
		expanded, err := NewWideOnline(s, polynomial, omega, n/4+1)
		if err != nil {
			t.Fatal(err)
		}
		bigger := expanded.Layout()
		if bigger.Prefix != n || bigger.Arena != 2*n || bigger.WorkspaceBytes != 25*n+4 {
			t.Fatal("expanded prefix count", s, bigger)
		}
		expanded.Close()
		e.Close()
		if e.OwnedBytes() != 0 {
			t.Fatal("metadata not released")
		}
		if _, err = NewWideOnline(s, polynomial, 1, 1); err == nil {
			t.Fatal("nonprimitive generator accepted")
		}
	}
}

func TestWideOnlineBorrowed(t *testing.T) {
	// The constant cyclic subcode has P=1+X and G=sum_{j=0}^{14}X^j.
	// G's field evaluations are1 at0/1 and0 elsewhere in GF16.
	e, err := NewWideOnline(2, 7, 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	layout := e.Layout()
	s32 := make([]uint32, 16)
	s32[0] = 1
	s32[1] = 1
	s64 := make([]uint64, 16)
	s64[0] = 1
	s64[1] = 1
	for _, wide := range []bool{false, true} {
		arena := make([]uint64, int(layout.Arena))
		scratch := make([]uint64, 16)
		saved := make([]uint32, 1)
		var narrow []uint32
		var spectrum []uint64
		if wide {
			spectrum = s64
		} else {
			narrow = s32
		}
		word, stats, err := e.EncodeBorrowed([]uint32{3}, []uint32{1, 1}, narrow, spectrum, arena, saved, scratch, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(word) != 16 || cap(word) != 16 || stats.Calls != 2 || stats.OutputBackingBytes != 128 {
			t.Fatal("constant/output shape")
		}
		for _, x := range word {
			if x != 3 {
				t.Fatal("constant code")
			}
		}
		if wide {
			s64[0] = 1 << 32
			if _, _, err = e.EncodeBorrowed([]uint32{3}, []uint32{1, 1}, nil, s64, arena, saved, scratch, 1); err == nil {
				t.Fatal("wide high bits truncated")
			}
			s64[0] = 1
		}
		if _, _, err = e.EncodeBorrowed([]uint32{4}, []uint32{1, 1}, narrow, spectrum, arena, saved, scratch, 1); err == nil {
			t.Fatal("base range")
		}
		if _, _, err = e.EncodeBorrowed([]uint32{3}, []uint32{1, 1}, narrow, spectrum, arena, saved, arena[:16], 1); err == nil {
			t.Fatal("alias accepted")
		}
	}
}
