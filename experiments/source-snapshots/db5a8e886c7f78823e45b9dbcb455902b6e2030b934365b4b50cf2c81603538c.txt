//go:build liftrs_cpp && cgo

package liftrs

import (
	"errors"
	liftrscpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"reflect"
	"sync"
	"testing"
)

func TestCPPFullCodewordsAndLifetime(t *testing.T) {
	for _, spec := range []Spec{{2, 2, 1}, {3, 2, 2}, {4, 2, 2}, {6, 2, 4}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(1201))})
		if err != nil {
			t.Fatal(err)
		}
		source, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		// Force the same FFT threshold on both paths to cover truncated-prefix FFTs
		// even on small fixtures; default thresholds are preserved in ordinary use.
		b.base.products.cutoff = 0
		native, err := source.WithCPPTransforms()
		if err != nil {
			t.Fatal(err)
		}
		if native.Basis() != b || native.PlanInfo().DescriptionBytes == 0 || native.PlanInfo().NativeOwnedBytes == 0 {
			t.Fatal("native plan ownership metadata")
		}
		random := rand.New(rand.NewSource(1202))
		messages := make([][]uint32, 4)
		for trial := range messages {
			message := make([]uint32, b.Dimension())
			for j := range message {
				message[j] = uint32(random.Intn(int(b.base.q)))
			}
			messages[trial] = message
			saved := append([]uint32(nil), message...)
			want, err := source.Encode(message)
			if err != nil {
				t.Fatal(err)
			}
			got, stats, err := native.EncodeWithStats(message)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("native codeword mismatch", spec, err)
			}
			if spec.FieldDegree <= 4 {
				oracle, err := Preprocess(spec)
				if err != nil {
					t.Fatal(err)
				}
				oracleMessage := make([]uint32, oracle.Dimension())
				for j, pos := range oracle.SystematicPositions() {
					oracleMessage[j] = got[pos]
				}
				independent, err := oracle.Encode(oracleMessage)
				if err != nil || !reflect.DeepEqual(got, independent) {
					t.Fatal("native output differs from dense oracle", err)
				}
			}
			if !reflect.DeepEqual(message, saved) || stats.Calls < 2 || stats.MessageCopyBytes != 0 {
				t.Fatal("buffer/call accounting mismatch")
			}
		}
		var group sync.WaitGroup
		for _, message := range messages {
			group.Add(1)
			go func(message []uint32) {
				defer group.Done()
				want, err := source.Encode(message)
				if err != nil {
					t.Error(err)
					return
				}
				got, err := native.Encode(message)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Error("concurrent native encoding mismatch", err)
				}
			}(message)
		}
		group.Wait()
		if _, err := native.Encode(nil); err == nil {
			t.Fatal("invalid message accepted")
		}
		if err := native.Close(); err != nil {
			t.Fatal(err)
		}
		if err := native.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := native.Encode(messages[0]); !errors.Is(err, ErrCPPClosed) {
			t.Fatal("encoding after Close", err)
		}
		if _, err := source.Encode(messages[0]); err != nil {
			t.Fatal("native Close affected Go source", err)
		}
	}
}

func TestCPPPolynomialStrategies(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(1203))})
	if err != nil {
		t.Fatal(err)
	}
	source, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	native, err := source.WithCPPTransforms()
	if err != nil {
		t.Fatal(err)
	}
	defer native.Close()
	var stats CPPEncodeStats
	engine := source.withTransformExecutor(&meteredCPP{native.plan, &stats})
	f, g := b.base, engine.basis.base
	f.products = &polynomialProducts{b.extensionFFT, 0}
	g.products = &polynomialProducts{engine.basis.extensionFFT, 0}
	random := rand.New(rand.NewSource(1204))
	for _, shape := range [][3]int{{17, 13, 9}, {33, 32, 20}, {101, 27, 79}, {300, 270, 310}, {511, 300, 650}} {
		a, z := make([]uint32, shape[0]), make([]uint32, shape[1])
		for j := range a {
			a[j] = uint32(random.Intn(16))
		}
		for j := range z {
			z[j] = uint32(random.Intn(16))
		}
		got := polyMulLow(g, a, z, shape[2])
		want := polyMulLow(f, a, z, shape[2])
		if !reflect.DeepEqual(got, want) {
			t.Fatal("FFT/truncation/overlap mismatch", shape)
		}
		// Suppress FFTs in both variants to independently exercise pruned Karatsuba
		// and unequal-operand blocking with the very same dispatch choices.
		f.products.cutoff = 1 << 30
		g.products.cutoff = 1 << 30
		if !reflect.DeepEqual(polyMulLow(g, a, z, shape[2]), polyMulLow(f, a, z, shape[2])) {
			t.Fatal("small-product strategy mismatch", shape)
		}
		f.products.cutoff = 0
		g.products.cutoff = 0
	}
	if stats.Calls == 0 {
		t.Fatal("polynomial FFT calls bypassed native backend")
	}
}

func TestCPPPlanValidation(t *testing.T) {
	base, extension, _, _ := orbitFields(4, 0x13)
	for _, field := range []orbitField{base, extension} {
		a := newAdditiveTransform(field)
		words := cppDescription(a)
		for _, bad := range [][]uint32{nil, {2}, words[:len(words)-1], append(append([]uint32(nil), words...), 0)} {
			if plan, err := liftrscpp.NewPlan(bad); err == nil {
				plan.Close()
				t.Fatal("malformed plan accepted")
			}
		}
		plan, err := liftrscpp.NewPlan(words)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := plan.Apply(make([]uint32, 3), false); err == nil {
			t.Fatal("invalid transform size accepted")
		}
		data := make([]uint32, field.q)
		for j := range data {
			data[j] = uint32(j) % field.q
		}
		want := append([]uint32(nil), data...)
		a.apply(want, false)
		if _, err := plan.Apply(data, false); err != nil || !reflect.DeepEqual(data, want) {
			t.Fatal("native forward transform", err)
		}
		if _, err := plan.Apply(data, true); err != nil {
			t.Fatal(err)
		}
		for j, x := range data {
			if x != uint32(j)%field.q {
				t.Fatal("native inverse")
			}
		}
		plan.Close()
		if _, err := plan.Apply(data, false); err == nil {
			t.Fatal("closed native plan accepted")
		}
	}
}
