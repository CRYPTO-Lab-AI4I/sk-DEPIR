package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"testing"
)

// A full 32-bit coordinate domain is exercised with small buffers only.
func TestOrbitCountBoundary(t *testing.T) {
	polynomial, _ := BinaryPolynomial(16)
	_, f, _, omega := orbitFields(16, polynomial)
	if f.q != uint64(1)<<32 || f.pow(omega, int(f.q-1)) != 1 {
		t.Fatal("32-bit coordinate count or primitive generator")
	}
	r := rand.New(rand.NewSource(3201))
	for i := 0; i < 100; i++ {
		x, y := r.Uint32(), r.Uint32()
		want := quadraticOracleMul(*f.extension, QuadraticElement{A: x & 65535, B: x >> 16}, QuadraticElement{A: y & 65535, B: y >> 16})
		if f.mul(x, y) != want.A|want.B<<16 {
			t.Fatal("full-width product differs from independent oracle")
		}
	}
	a := newSparseTransform(f)
	values := make([]uint32, 16)
	for i := range values {
		values[i] = r.Uint32()
	}
	input := append([]uint32(nil), values...)
	a.applySubspace(values, false)
	for i, got := range values {
		var want uint32
		for j := len(input) - 1; j >= 0; j-- {
			want = f.mul(want, uint32(i)) ^ input[j]
		}
		if got != want {
			t.Fatal("small subspace Horner check")
		}
	}
	a.applySubspace(values, true)
	if !reflect.DeepEqual(values, input) {
		t.Fatal("small subspace inverse")
	}
}

func TestScalingOnlineCountWidths(t *testing.T) {
	for _, s := range []int{15, 16} {
		var header bytes.Buffer
		n := uint64(1) << uint(2*s)
		if err := writeScalingOnlineHeader(&header, s, 5, n/4, n); err != nil {
			t.Fatal(err)
		}
		length := 25
		if s == 16 {
			length = 33
			if binary.LittleEndian.Uint64(header.Bytes()[25:]) != n {
				t.Fatal("2^32 count lost")
			}
		}
		if header.Len() != length {
			t.Fatal("unnecessary header width", s, header.Len())
		}
		// A real wide header must be rejected against the caller's small cap
		// before allocation, even with a correct digest of the truncated file.
		h := sha256.Sum256(header.Bytes())
		if _, err := LoadScalingOnline(bytes.NewReader(header.Bytes()), hex.EncodeToString(h[:]), 256); err == nil {
			t.Fatal("large allocation cap ignored")
		}
	}
	for _, counts := range [][2]uint64{{1, 0}, {1, ^uint64(0)}, {1 << 63, 1 << 32}, {1, (1 << 32) - 1}} {
		if validScalingOnlineCounts(16, 5, counts[0], counts[1], int(^uint(0)>>1)) {
			t.Fatal("invalid/overflowed count accepted", counts)
		}
	}
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(3202))})
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
	var old, current bytes.Buffer
	if err := writeScalingOnline(&old, p); err != nil {
		t.Fatal(err)
	}
	if err := WriteScalingOnline(&current, p); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(old.Bytes(), current.Bytes()) {
		t.Fatal("existing file bytes changed")
	}
	// Exercise wide-header decoding with tiny real constants and a full round trip.
	var wide bytes.Buffer
	wide.WriteString("LRSprobe2")
	binary.Write(&wide, binary.LittleEndian, [2]uint32{4, 2})
	binary.Write(&wide, binary.LittleEndian, [2]uint64{uint64(p.k), uint64(p.Length())})
	wide.Write(old.Bytes()[25:])
	digest := sha256.Sum256(wide.Bytes())
	loaded, err := LoadScalingOnline(bytes.NewReader(wide.Bytes()), hex.EncodeToString(digest[:]), 256)
	if err != nil || !reflect.DeepEqual(loaded.p, p.p) || !reflect.DeepEqual(loaded.spectrum, p.spectrum) {
		t.Fatal("wide file round trip", err)
	}
	if err := WriteScalingOnline(&current, nil); err == nil {
		t.Fatal("nil plan accepted")
	}
	if err := WriteScalingOnline(widthErrorWriter{}, p); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal("writer error lost", err)
	}
}

type widthErrorWriter struct{}

func (widthErrorWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func ExampleWriteScalingOnline() {
	basis, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(3203))})
	if err != nil {
		panic(err)
	}
	prepared, err := basis.PrepareCyclicFFT()
	if err != nil {
		panic(err)
	}
	plan, err := prepared.ScalingPlan()
	if err != nil {
		panic(err)
	}
	var data bytes.Buffer
	if err := WriteScalingOnline(&data, plan); err != nil {
		panic(err)
	}
	digest := sha256.Sum256(data.Bytes())
	loaded, err := LoadScalingOnline(bytes.NewReader(data.Bytes()), hex.EncodeToString(digest[:]), 256)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data.Bytes()[:9]), loaded.Length())
	// Output: LRSprobe1 256
}
