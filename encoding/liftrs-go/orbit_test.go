package liftrs

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestOrbitTransforms(t *testing.T) {
	for s := 2; s <= 8; s++ {
		polys := map[int]uint32{2: 7, 3: 11, 4: 19, 5: 37, 6: 67, 7: 131, 8: 285}
		base, ext, _, _ := orbitFields(s, polys[s])
		for _, f := range []orbitField{base, ext} {
			a := newAdditiveTransform(f)
			r := rand.New(rand.NewSource(int64(s)))
			p := make([]uint32, f.q)
			for i := range p {
				p[i] = uint32(r.Intn(int(f.q)))
			}
			values := append([]uint32(nil), p...)
			a.apply(values, false)
			// Horner samples independently check ordering and forward correctness.
			for _, point := range []uint32{0, 1, f.q / 3, f.q - 1} {
				want := uint32(0)
				for j := len(p) - 1; j >= 0; j-- {
					want = f.mul(want, point) ^ p[j]
				}
				if values[point] != want {
					t.Fatalf("s=%d field=%d point=%d", s, f.q, point)
				}
			}
			a.apply(values, true)
			if !reflect.DeepEqual(values, p) {
				t.Fatalf("FFT round trip s=%d field=%d", s, f.q)
			}
		}
	}
}
func TestOrbitPolynomials(t *testing.T) {
	f, _, _, _ := orbitFields(4, 0x13)
	r := rand.New(rand.NewSource(43))
	for _, n := range []int{1, 23, 32, 33, 65, 130} {
		a, b := make([]uint32, n), make([]uint32, n+3)
		for i := range a {
			a[i] = uint32(r.Intn(16))
		}
		for i := range b {
			b[i] = uint32(r.Intn(16))
		}
		b[len(b)-1] = 1
		product := polyMul(f, a, b)
		want := make([]uint32, len(product))
		for i, x := range a {
			for j, y := range b {
				want[i+j] ^= f.mul(x, y)
			}
		}
		if !reflect.DeepEqual(product, want) {
			t.Fatal("Karatsuba mismatch", n)
		}
		q, rem := polyDiv(f, product, b)
		if len(rem) != 0 || !reflect.DeepEqual(polyTrim(a), q) {
			t.Fatal("division mismatch", n)
		}
	}
}
func TestOrbitEncodersAgainstDenseOracle(t *testing.T) {
	for _, spec := range []Spec{{2, 2, 1}, {2, 2, 2}, {3, 2, 1}, {3, 2, 2}, {3, 2, 5}, {4, 2, 1}, {4, 2, 2}, {4, 2, 5}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(99))})
		if err != nil {
			t.Fatal(spec, err)
		}
		dense, err := Preprocess(spec)
		if err != nil {
			t.Fatal(err)
		}
		if b.Dimension() != dense.Dimension() {
			t.Fatal("dimension mismatch", spec)
		}
		paper, err := b.PreparePaper()
		if err != nil {
			t.Fatal(spec, err)
		}
		cyclic, err := b.PrepareCyclic()
		if err != nil {
			t.Fatal(spec, err)
		}
		r := rand.New(rand.NewSource(101))
		for trial := 0; trial < b.Dimension()+8; trial++ {
			message := make([]uint32, b.Dimension())
			for i := range message {
				message[i] = uint32(r.Intn(int(b.base.q)))
			}
			if trial < b.Dimension() {
				for i := range message {
					message[i] = 0
				}
				message[trial] = 1
			}
			saved := append([]uint32(nil), message...)
			got, err := paper.Encode(message)
			if err != nil {
				t.Fatal(spec, err)
			}
			other, err := cyclic.Encode(message)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, other) {
				t.Fatal("paper/cyclic mismatch", spec)
			}
			for j, x := range message {
				pos, _ := b.SystematicPosition(j)
				if got[pos] != x {
					t.Fatalf("not systematic: %v index %d", spec, j)
				}
			}
			// The old oracle has a different information set. Restrict the new word
			// there and ask the independent dense encoder to reconstruct the entire word.
			oracleMessage := make([]uint32, dense.Dimension())
			for j, pos := range dense.SystematicPositions() {
				oracleMessage[j] = got[pos]
			}
			want, err := dense.Encode(oracleMessage)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatal("not in dense oracle code", spec)
			}
			if !reflect.DeepEqual(saved, message) {
				t.Fatal("message modified")
			}
		}
		artifact, _ := b.MarshalBinary()
		loaded, err := LoadOrbit(artifact)
		if err != nil {
			t.Fatal(err)
		}
		again, _ := loaded.MarshalBinary()
		if !bytes.Equal(artifact, again) {
			t.Fatal("artifact unstable")
		}
		if loaded.Samples() != 0 {
			t.Fatal("loading reran discovery")
		}
	}
}
func TestOrbitModerateAgreementAndReuse(t *testing.T) {
	b, err := DiscoverOrbit(Spec{5, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(321))})
	if err != nil {
		t.Fatal(err)
	}
	artifact, _ := b.MarshalBinary()
	b, err = LoadOrbit(artifact)
	if err != nil {
		t.Fatal(err)
	}
	p, err := b.PreparePaper()
	if err != nil {
		t.Fatal(err)
	}
	c, err := b.PrepareCyclic()
	if err != nil {
		t.Fatal(err)
	}
	message := make([]uint32, b.Dimension())
	for i := range message {
		message[i] = uint32(i) % b.base.q
	}
	want, err := p.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := c.Encode(message)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Error("concurrent reuse mismatch", err)
			}
		}()
	}
	wg.Wait()
}

type orbitFailedReader struct{}

func (orbitFailedReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func TestOrbitValidation(t *testing.T) {
	for _, spec := range []Spec{{1, 2, 1}, {14, 2, 1}, {4, 3, 1}, {4, 2, 0}, {4, 2, 15}} {
		if _, err := DiscoverOrbit(spec, OrbitOptions{}); err == nil {
			t.Fatal("invalid spec accepted", spec)
		}
	}
	if _, err := DiscoverOrbit(Spec{2, 2, 1}, OrbitOptions{Random: bytes.NewReader(make([]byte, 100)), MaxSamples: 1}); err == nil {
		t.Fatal("incomplete discovery accepted")
	}
	if _, err := DiscoverOrbit(Spec{2, 2, 1}, OrbitOptions{Random: orbitFailedReader{}}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal("reader error lost", err)
	}
	b, err := DiscoverOrbit(Spec{2, 2, 1}, OrbitOptions{Random: rand.New(rand.NewSource(9))})
	if err != nil {
		t.Fatal(err)
	}
	p, _ := b.PreparePaper()
	c, _ := b.PrepareCyclic()
	for _, message := range [][]uint32{nil, make([]uint32, b.Dimension()+1), {4}} {
		if _, err := p.Encode(message); err == nil {
			t.Fatal("invalid paper message")
		}
		if _, err := c.Encode(message); err == nil {
			t.Fatal("invalid cyclic message")
		}
	}
	if _, err := b.SystematicPosition(-1); err == nil {
		t.Fatal("negative index")
	}
	for _, data := range [][]byte{nil, []byte(`{"version":2}`), []byte(`{"version":1,"spec":{"FieldDegree":2,"Dimension":2,"CurveDegree":1},"support":[1]}`)} {
		if _, err := LoadOrbit(data); err == nil {
			t.Fatal("invalid artifact")
		}
	}
}

func TestPreparedOrbitArtifactReuse(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(72))})
	if err != nil {
		t.Fatal(err)
	}
	p, _ := b.PreparePaper()
	c, _ := b.PrepareCyclic()
	message := make([]uint32, b.Dimension())
	for i := range message {
		message[i] = uint32(i) % b.base.q
	}
	want, _ := p.Encode(message)
	pa, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	ca, err := c.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	pp, err := LoadPaperEncoder(pa)
	if err != nil {
		t.Fatal(err)
	}
	cc, err := LoadCyclicEncoder(ca)
	if err != nil {
		t.Fatal(err)
	}
	for _, encode := range []func([]uint32) ([]uint32, error){pp.Encode, cc.Encode} {
		got, err := encode(message)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("prepared artifact output differs", err)
		}
	}
	if _, err := LoadPaperEncoder(ca); err == nil {
		t.Fatal("wrong method accepted")
	}
	if _, err := LoadCyclicEncoder(pa); err == nil {
		t.Fatal("wrong method accepted")
	}
}

func TestCyclicCachedFFT(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 1}, {4, 2, 2}, {5, 2, 3}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(481))})
		if err != nil {
			t.Fatal(err)
		}
		p, err := b.PreparePaper()
		if err != nil {
			t.Fatal(err)
		}
		c, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := c.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		c, err = LoadCyclicEncoder(raw)
		if err != nil {
			t.Fatal(err)
		}
		r := rand.New(rand.NewSource(68))
		for trial := 0; trial < 4; trial++ {
			msg := make([]uint32, b.Dimension())
			for i := range msg {
				msg[i] = uint32(r.Intn(int(b.base.q)))
			}
			want, err := p.Encode(msg)
			if err != nil {
				t.Fatal(err)
			}
			got, err := c.Encode(msg)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("cached FFT mismatch", spec, err)
			}
		}
	}
}
