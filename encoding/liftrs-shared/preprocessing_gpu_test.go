//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func gpuPreparationDevices(t *testing.T) []int {
	t.Helper()
	if os.Getenv("PIR_GPU_PREPARATION_LIBRARY") == "" {
		t.Skip("opt-in resident GPU preparation")
	}
	value := os.Getenv("PIR_GPU_PREPARATION_DEVICES")
	if value == "" {
		value = "0"
	}
	var devices []int
	for _, piece := range strings.Split(value, ",") {
		id, err := strconv.Atoi(piece)
		if err != nil {
			t.Fatal(err)
		}
		devices = append(devices, id)
	}
	return devices
}
func gpuPreparationCap(t *testing.T) uint64 {
	t.Helper()
	value := os.Getenv("PIR_GPU_PREPARATION_CAP")
	if value == "" {
		return 512 << 20
	}
	cap, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return cap
}
func TestGPUPreparationSmall(t *testing.T) {
	devices := gpuPreparationDevices(t)
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		t.Run(fmt.Sprintf("s%d", spec.FieldDegree), func(t *testing.T) {
			reference, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(170069))})
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
			options := preprocessingOptions(spec.FieldDegree, 4)
			options.Backend = "cuda"
			options.GPUPath = os.Getenv("PIR_GPU_PREPARATION_LIBRARY")
			options.GPUDevices = devices
			options.GPUBytes = gpuPreparationCap(t)
			session, err := NewCodePreprocessor(spec, options)
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			support, err := session.Discover(OrbitOptions{Random: rand.New(rand.NewSource(170069))})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(support.exponents, reference.support) || support.Parameters().Samples != reference.Samples() {
				t.Fatal("GPU discovery differs from Go")
			}
			var supportRaw bytes.Buffer
			supportManifest, err := support.Write(&supportRaw)
			if err != nil {
				t.Fatal(err)
			}
			var rootRaw bytes.Buffer
			var rootManifest RootRecord
			prepared, err := session.Prepare(support, PreparationReuse{SaveRoot: func(root *RootCheckpoint) error { var e error; rootManifest, e = root.Write(&rootRaw); return e }})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(prepared.code.p, encoder.p) || !reflect.DeepEqual(prepared.generator, encoder.g) {
				t.Fatal("GPU P/G oracle")
			}
			var actual bytes.Buffer
			if err = WriteScalingCode(&actual, prepared.CodeData()); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
				t.Fatal("GPU canonical bytes", err)
			}
			root, err := ReadPreparationRoot(bytes.NewReader(rootRaw.Bytes()), support, rootManifest, options.Limits)
			if err != nil {
				t.Fatal(err)
			}
			replay, err := session.Prepare(support, PreparationReuse{Root: root})
			if err != nil {
				t.Fatal(err)
			}
			var repeated bytes.Buffer
			if err = WriteScalingCode(&repeated, replay.CodeData()); err != nil || !bytes.Equal(repeated.Bytes(), expected.Bytes()) {
				t.Fatal("GPU checkpoint replay", err)
			}
			native, device := session.OwnedBytes()
			if native > 64<<20 || device == 0 || device > uint64(len(devices))*options.GPUBytes {
				t.Fatal("GPU/host accounting", native, device)
			}
			if directory := os.Getenv("PIR_GPU_PREPARATION_CACHE"); directory != "" {
				path := filepath.Join(directory, fmt.Sprintf("s%d", spec.FieldDegree))
				if err = os.MkdirAll(path, 0700); err != nil {
					t.Fatal(err)
				}
				files := map[string][]byte{"support.bin": supportRaw.Bytes(), "root.bin": rootRaw.Bytes(), "online.bin": actual.Bytes()}
				files["support.json"], err = json.MarshalIndent(supportManifest, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				files["root.json"], err = json.MarshalIndent(rootManifest, "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				var g bytes.Buffer
				if err = prepared.WriteGenerator(&g); err != nil {
					t.Fatal(err)
				}
				files["generator.bin"] = g.Bytes()
				for name, raw := range files {
					if err = os.WriteFile(filepath.Join(path, name), raw, 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			session.Close()
			session.Close()
			if native, device = session.OwnedBytes(); native != 0 || device != 0 {
				t.Fatal("GPU close ownership")
			}
			actual.Reset()
			if err = WriteScalingCode(&actual, prepared.CodeData()); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
				t.Fatal("Go output lifetime", err)
			}
		})
	}
}

func TestGPUPreparationWide(t *testing.T) {
	devices := gpuPreparationDevices(t)
	h, f, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := NewBinaryField(17, h.Polynomial)
	table := newPreparationProducts(base)
	makeEngine := func(cap uint64) (*cuda.Preparation, error) {
		return cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: 17, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: devices, MaxElements: h.N, DeviceBytes: cap})
	}
	engine, err := makeEngine(gpuPreparationCap(t))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	random := rand.New(rand.NewSource(170069))
	for _, n := range []int{4, 16, 128, 1024} {
		input := make([]uint64, n)
		for j := range input {
			input[j] = random.Uint64() & ((1 << 34) - 1)
		}
		input[0] = 1 << 33
		actual := append([]uint64(nil), input...)
		if err = engine.Apply64(actual, false); err != nil {
			t.Fatal(err)
		}
		// For the larger size sample32 Horner points plus round-trip the whole input.
		for point := 0; point < n; point += max(1, n/32) {
			var expected uint64
			for j := n - 1; j >= 0; j-- {
				expected = f.mul(expected, uint64(point)) ^ input[j]
			}
			if actual[point] != expected {
				t.Fatal("GPU wide Horner", n, point)
			}
		}
		if err = engine.Apply64(actual, true); err != nil || !reflect.DeepEqual(actual, input) {
			t.Fatal("GPU wide inverse", n, err)
		}
	}
	a, b := make([]uint32, 257), make([]uint32, 300)
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
	actual, err := engine.Product(a, b, len(expected), 1024)
	if err != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatal("GPU wide product", err)
	}
	truncated, err := engine.Product(a, b, 257, 1024)
	if err != nil || !reflect.DeepEqual(truncated, expected[:257]) {
		t.Fatal("GPU wide prefix", err)
	}
	var indices []int
	for j := 0; j < 1100; j++ {
		indices = append(indices, j)
	}
	indices = append(indices, 1<<32, (1<<32)+1, (1<<34)-2)
	roots := make([][4]uint32, len(indices))
	if err = engine.RootFactorsInto(indices, roots); err != nil {
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
			t.Fatal("GPU wide root", i, roots[j], expected)
		}
	}
	if _, err = engine.GoodMonomials(5, 64<<20); err == nil {
		t.Fatal("full s17 enumeration bypassed GPU cap")
	}
	if err = engine.Apply32(make([]uint32, 16), false); err == nil {
		t.Fatal("narrow wide-field buffer")
	}
	if err = engine.Apply64([]uint64{1, 2, 3}, false); err == nil {
		t.Fatal("invalid transform shape")
	}
	if err = engine.RootFactorsInto([]int{2, 1}, make([][4]uint32, 2)); err == nil {
		t.Fatal("unsorted roots")
	}
	bad := []uint64{1 << 34, 0, 0, 0}
	if err = engine.Apply64(bad, false); err == nil || bad[0] != 1<<34 {
		t.Fatal("input range validation", err)
	}
	before := engine.OwnedBytes()
	if err = engine.ReleaseProducts(); err != nil || engine.OwnedBytes() >= before {
		t.Fatal("dead products retained", err)
	}
	if _, err = makeEngine(1); err == nil {
		t.Fatal("metadata byte cap")
	}
	engine.Close()
	engine.Close()
	if engine.OwnedBytes() != 0 {
		t.Fatal("closed GPU allocation")
	}
	if err = engine.Apply64(make([]uint64, 16), false); err == nil {
		t.Fatal("closed GPU call")
	}
}

// Compare resident one/multiple-card tiles in the actual peer topology. The
// sparse Horner oracle is independent of the full forward/inverse pair.
func TestGPUPreparationTile(t *testing.T) {
	devices := gpuPreparationDevices(t)
	h, f, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := NewBinaryField(17, h.Polynomial)
	table := newPreparationProducts(base)
	const n = 1 << 22
	var records []map[string]any
	for _, ids := range [][]int{devices[:1], devices} {
		if len(records) > 0 && len(ids) == 1 {
			continue
		}
		engine, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: 17, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: ids, MaxElements: n, DeviceBytes: gpuPreparationCap(t)})
		if err != nil {
			t.Fatal(err)
		}
		defer engine.Close()
		input := make([]uint64, n)
		for j := 0; j < 16; j++ {
			input[j] = uint64(170069+j*71317) ^ (1 << 33)
		}
		original := append([]uint64(nil), input...)
		// Warm and verify sparse Horner first; timed data below is dense so
		// zero shortcuts do not make these component timings misleading.
		if err = engine.Apply64(input, false); err != nil {
			t.Fatal(err)
		}
		for point := 0; point < n; point += n / 64 {
			var expected uint64
			for j := 15; j >= 0; j-- {
				expected = f.mul(expected, uint64(point)) ^ original[j]
			}
			if input[point] != expected {
				t.Fatal("tile sparse Horner", len(ids), point)
			}
		}
		if err = engine.Apply64(input, true); err != nil || !reflect.DeepEqual(input, original) {
			t.Fatal("sparse tile inverse", err)
		}
		state := uint64(170069)
		var parity uint64
		for j := range input {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			input[j] = state & ((1 << 34) - 1)
			original[j] = input[j]
			parity ^= input[j]
		}
		began := time.Now()
		if err = engine.Apply64(input, false); err != nil {
			t.Fatal(err)
		}
		forward := time.Since(began)
		if input[0] != original[0] || input[1] != parity {
			t.Fatal("dense tile independent 0/1 evaluations")
		}
		// Hash the entire forward word, outside the timed transform boundary.
		digest := sha256.New()
		var element [8]byte
		for _, value := range input {
			binary.LittleEndian.PutUint64(element[:], value)
			digest.Write(element[:])
		}
		fingerprint := hex.EncodeToString(digest.Sum(nil))
		if len(records) > 0 && fingerprint != records[0]["forward_sha256"] {
			t.Fatal("peer tile differs from single GPU")
		}
		began = time.Now()
		if err = engine.Apply64(input, true); err != nil {
			t.Fatal(err)
		}
		inverse := time.Since(began)
		if !reflect.DeepEqual(input, original) {
			t.Fatal("complete tile inverse")
		}
		records = append(records, map[string]any{"s": 17, "values": n, "gpus": len(ids), "forward_seconds": forward.Seconds(), "inverse_seconds": inverse.Seconds(), "owned_device_bytes": engine.OwnedBytes(), "forward_sha256": fingerprint, "profiled": true, "input": "dense deterministic extension values"})
		engine.Close()
	}
	if directory := os.Getenv("PIR_GPU_PREPARATION_OUT"); directory != "" {
		raw, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(directory, "tiles.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
