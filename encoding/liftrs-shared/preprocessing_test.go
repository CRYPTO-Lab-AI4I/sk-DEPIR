//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"encoding/json"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func preprocessingOptions(s, workers int) PreprocessingOptions {
	return PreprocessingOptions{Backend: "cpp", Workers: workers, ProductCutoff: 24, Limits: PreprocessingLimits{MaxElements: 1 << uint(2*s), MaxHostArrayBytes: 64 << 20, NativeBytes: 64 << 20}}
}

func TestPreprocessingNative(t *testing.T) {
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		t.Run(fmt.Sprintf("s%d", spec.FieldDegree), func(t *testing.T) {
			reference, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(170066))})
			if err != nil {
				t.Fatal(err)
			}
			encoder, err := reference.PrepareCyclicFFT()
			if err != nil {
				t.Fatal(err)
			}
			legacy, err := encoder.ScalingPlan()
			if err != nil {
				t.Fatal(err)
			}
			var expected bytes.Buffer
			if err = WriteScalingOnline(&expected, legacy); err != nil {
				t.Fatal(err)
			}
			for _, workers := range []int{1, 4} {
				session, err := NewCodePreprocessor(spec, preprocessingOptions(spec.FieldDegree, workers))
				if err != nil {
					t.Fatal(err)
				}
				defer session.Close()
				support, err := session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(170066))})
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(support.exponents, reference.support) || support.parameters.Samples != reference.Samples() {
					t.Fatal("discovery changed")
				}
				var supportBytes bytes.Buffer
				manifest, err := support.Write(&supportBytes)
				if err != nil {
					t.Fatal(err)
				}
				limits := preprocessingOptions(spec.FieldDegree, workers).Limits
				support, err = ReadCodeSupport(bytes.NewReader(supportBytes.Bytes()), manifest, limits)
				if err != nil {
					t.Fatal(err)
				}
				var rootBytes bytes.Buffer
				var rootRecord RootRecord
				prepared, err := session.Prepare(support, PreparationReuse{SaveRoot: func(root *RootCheckpoint) error { var e error; rootRecord, e = root.Write(&rootBytes); return e }})
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(prepared.code.p, encoder.p) || !reflect.DeepEqual(prepared.generator, encoder.g) {
					t.Fatal("root/generator oracle")
				}
				var actual bytes.Buffer
				if err = WriteScalingCode(&actual, prepared.CodeData()); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
					t.Fatal("canonical code bytes", err)
				}
				checkpoint, err := ReadPreparationRoot(bytes.NewReader(rootBytes.Bytes()), support, rootRecord, limits)
				if err != nil {
					t.Fatal(err)
				}
				replay, err := session.Prepare(support, PreparationReuse{Root: checkpoint})
				if err != nil {
					t.Fatal(err)
				}
				actual.Reset()
				if err = WriteScalingCode(&actual, replay.CodeData()); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
					t.Fatal("restart changed code", err)
				}
				if directory := os.Getenv("PIR_PREPROCESSING_CACHE"); directory != "" && workers == 4 {
					target := filepath.Join(directory, fmt.Sprintf("s%d", spec.FieldDegree))
					if err = os.MkdirAll(target, 0700); err != nil {
						t.Fatal(err)
					}
					files := map[string][]byte{"support.bin": supportBytes.Bytes(), "root.bin": rootBytes.Bytes(), "online.bin": actual.Bytes()}
					files["support.json"], err = json.MarshalIndent(manifest, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					files["root.json"], err = json.MarshalIndent(rootRecord, "", "  ")
					if err != nil {
						t.Fatal(err)
					}
					var generator bytes.Buffer
					if err = prepared.WriteGenerator(&generator); err != nil {
						t.Fatal(err)
					}
					files["generator.bin"] = generator.Bytes()
					for name, raw := range files {
						if err = os.WriteFile(filepath.Join(target, name), raw, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				session.Close()
				session.Close()
				if native, device := session.OwnedBytes(); native != 0 || device != 0 {
					t.Fatal("closed bytes")
				}
				if _, err = session.Discover(OrbitOptions{}); err == nil {
					t.Fatal("closed discovery")
				}
				if _, err = session.Prepare(support, PreparationReuse{}); err == nil {
					t.Fatal("closed preparation")
				}
				// Returned code must survive releasing all executor-owned memory.
				actual.Reset()
				if err = WriteScalingCode(&actual, prepared.CodeData()); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
					t.Fatal("ownership", err)
				}
			}
		})
	}
}

func TestPreprocessingGoodMonomials(t *testing.T) {
	for s := 2; s <= 8; s++ {
		h, _, err := canonicalScalingCodeHeader(s, 1, 1, 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		engine, err := cpp.NewPreparation(s, h.Polynomial, h.Omega, 4, h.N, 64<<20)
		if err != nil {
			t.Fatal(err)
		}
		for _, curve := range []int{1, min(5, (1<<uint(s))-2)} {
			got, err := engine.GoodMonomials(curve, 64<<20)
			if err != nil {
				t.Fatal(err)
			}
			var expected []int
			for index := 0; index < int(h.N); index++ {
				yes, err := GoodMonomial(Spec{s, 2, curve}, []uint32{uint32(index) & ((1 << uint(s)) - 1), uint32(index) >> uint(s)})
				if err != nil {
					t.Fatal(err)
				}
				if yes {
					expected = append(expected, index)
				}
			}
			if !reflect.DeepEqual(got, expected) {
				t.Fatal("weight-class enumeration", s, curve)
			}
		}
		engine.Close()
	}
	// This criterion witness uses coordinates above 2^32 without full discovery.
	for _, first := range []uint64{0, 1 << 32, (1 << 34) - 257} {
		got, err := cpp.GoodMonomialRange(17, 5, first, 257, 4)
		if err != nil {
			t.Fatal(err)
		}
		for j, value := range got {
			i := first + uint64(j)
			yes, err := GoodMonomial(Spec{17, 2, 5}, []uint32{uint32(i & ((1 << 17) - 1)), uint32(i >> 17)})
			if err != nil || yes != (value == 1) {
				t.Fatal("wide criterion", i, err)
			}
		}
	}
}

func TestPreprocessingLimits(t *testing.T) {
	options := preprocessingOptions(4, 1)
	bad := options
	bad.Backend = ""
	if _, err := NewCodePreprocessor(Spec{4, 2, 2}, bad); err == nil {
		t.Fatal("implicit backend")
	}
	bad = options
	bad.Limits.MaxElements = 128
	if _, err := NewCodePreprocessor(Spec{4, 2, 2}, bad); err == nil {
		t.Fatal("count cap")
	}
	bad = options
	bad.Limits.NativeBytes = 1
	if _, err := NewCodePreprocessor(Spec{4, 2, 2}, bad); err == nil {
		t.Fatal("native metadata cap")
	}
	session, err := NewCodePreprocessor(Spec{4, 2, 2}, options)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if _, err = session.Discover(OrbitOptions{MaxSamples: -1}); err == nil {
		t.Fatal("negative samples")
	}
	if _, err = session.Discover(OrbitOptions{Random: bytes.NewReader(nil)}); err == nil {
		t.Fatal("random EOF")
	}
	if _, err = session.Discover(OrbitOptions{Random: zeroPreparationReader{}, MaxSamples: 1}); err == nil {
		t.Fatal("incomplete discovery")
	}
	support, err := session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(17))})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = session.Prepare(support, PreparationReuse{SaveRoot: func(*RootCheckpoint) error { return nil }}); err == nil {
		t.Fatal("unbound checkpoint")
	}
	// Metadata-only s17 admission. Native discovery must reject its bitset/table
	// before attempting the enormous Go grid; numeric caps authorize no allocation.
	wide := PreprocessingOptions{Backend: "cpp", Workers: 1, Limits: PreprocessingLimits{MaxElements: 1 << 34, MaxHostArrayBytes: 128 << 30, NativeBytes: 1 << 20}}
	large, err := NewCodePreprocessor(Spec{17, 2, 5}, wide)
	if err != nil {
		t.Fatal(err)
	}
	defer large.Close()
	if _, err = large.Discover(OrbitOptions{}); err == nil {
		t.Fatal("s17 bitset cap")
	}
	if native, _ := large.OwnedBytes(); native > 1<<20 {
		t.Fatal("cap leaked huge allocation")
	}
	if _, err = large.Prepare(nil, PreparationReuse{}); err == nil {
		t.Fatal("missing support")
	}
}

type zeroPreparationReader struct{}

func (zeroPreparationReader) Read(b []byte) (int, error) { clear(b); return len(b), nil }

func ExampleNewCodePreprocessor() {
	options := PreprocessingOptions{Backend: "cpp", Workers: 1, Limits: PreprocessingLimits{MaxElements: 256, MaxHostArrayBytes: 1 << 20, NativeBytes: 8 << 20}}
	p, err := NewCodePreprocessor(Spec{4, 2, 2}, options)
	if err != nil {
		panic(err)
	}
	defer p.Close()
	support, err := p.Discover(OrbitOptions{})
	if err != nil {
		panic(err)
	}
	// Persist support + its returned manifest before asking to save root checkpoints.
	if _, err = support.Write(io.Discard); err != nil {
		panic(err)
	}
	prepared, err := p.Prepare(support, PreparationReuse{})
	if err != nil {
		panic(err)
	}
	fmt.Println(prepared.CodeData().Header().N)
	// Output: 256
}

func TestPreprocessingWideKernels(t *testing.T) {
	h, f, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := cpp.NewPreparation(17, h.Polynomial, h.Omega, 4, 256, 8<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	input := make([]uint64, 128)
	random := rand.New(rand.NewSource(170066))
	for j := range input {
		input[j] = random.Uint64() & ((1 << 34) - 1)
	}
	input[0] = 1 << 33
	transformed := append([]uint64(nil), input...)
	if err = engine.Apply64(transformed, false); err != nil {
		t.Fatal(err)
	}
	for point, value := range transformed {
		var expected uint64
		for j := len(input) - 1; j >= 0; j-- {
			expected = f.mul(expected, uint64(point)) ^ input[j]
		}
		if value != expected {
			t.Fatal("wide Horner", point)
		}
	}
	if err = engine.Apply64(transformed, true); err != nil || !reflect.DeepEqual(input, transformed) {
		t.Fatal("wide inverse", err)
	}
	a, b := make([]uint32, 65), make([]uint32, 70)
	for j := range a {
		a[j] = random.Uint32() & ((1 << 17) - 1)
	}
	for j := range b {
		b[j] = random.Uint32() & ((1 << 17) - 1)
	}
	expected := make([]uint32, len(a)+len(b)-1)
	for i, x := range a {
		for j, y := range b {
			expected[i+j] ^= uint32(f.mul(uint64(x), uint64(y)))
		}
	}
	actual, err := engine.Product(a, b, len(expected), 256)
	if err != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatal("wide product", err)
	}
	indices := []int{0, 1, 2, 3, 1 << 32, (1 << 32) + 1, (1 << 34) - 2}
	roots, err := engine.RootFactors(indices)
	if err != nil {
		t.Fatal(err)
	}
	for j, i := range indices {
		partner := uint64(i) * (1 << 17) % ((1 << 34) - 1)
		var expected [4]uint32
		if partner >= uint64(i) {
			root := f.pow(h.Omega, ((1<<34)-1)-uint64(i))
			if partner == uint64(i) {
				expected = [4]uint32{uint32(root), 1, 0, 1}
			} else {
				conjugate := f.pow(root, 1<<17)
				expected = [4]uint32{uint32(f.mul(root, conjugate)), uint32(root ^ conjugate), 1, 2}
			}
		}
		if roots[j] != expected {
			t.Fatal("wide root", i, roots[j], expected)
		}
	}
	if err = engine.Apply32(make([]uint32, 128), false); err == nil {
		t.Fatal("narrow buffer accepted")
	}
	before := engine.OwnedBytes()
	engine.ReleaseProducts()
	if engine.OwnedBytes() >= before {
		t.Fatal("dead operands retained")
	}
	if _, err = engine.GoodMonomials(5, 1<<20); err == nil {
		t.Fatal("partial-domain discovery accepted")
	}
}
