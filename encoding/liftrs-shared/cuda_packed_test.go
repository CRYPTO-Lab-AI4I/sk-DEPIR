//go:build liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type failingPackedWriter struct{ short bool }

func (w failingPackedWriter) Write(p []byte) (int, error) {
	if w.short {
		return 0, nil
	}
	return len(p) / 2, io.ErrClosedPipe
}

func packedReference(word []uint32, s int) []byte {
	var out []byte
	var acc uint64
	var used uint
	for _, x := range word {
		acc |= uint64(x) << used
		used += uint(s)
		for used >= 8 {
			out = append(out, byte(acc))
			acc >>= 8
			used -= 8
		}
	}
	if used != 0 {
		out = append(out, byte(acc))
	}
	return out
}

func TestCUDAPackedOutput(t *testing.T) {
	out := os.Getenv("PIR_CUDA_PACKED_TEST_OUT")
	if out == "" {
		t.Skip("opt-in GPU packed sink")
	}
	var records []map[string]any
	for _, spec := range []Spec{{2, 2, 1}, {3, 2, 2}, {4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		b, err := DiscoverOrbit(spec, OrbitOptions{Random: rand.New(rand.NewSource(3801))})
		if err != nil {
			t.Fatal(err)
		}
		code, err := b.PrepareCyclicFFT()
		if err != nil {
			t.Fatal(err)
		}
		p, err := code.ScalingPlan()
		if err != nil {
			t.Fatal(err)
		}
		if cache := os.Getenv("PIR_CUDA_PACKED_TEST_CACHE"); cache != "" {
			if err = saveCUDAOnlineFixture(p, filepath.Join(cache, fmt.Sprintf("s%d-code.bin", spec.FieldDegree))); err != nil {
				t.Fatal(err)
			}
		}
		modes := []bool{false}
		_, _, n, l, _ := cudaShape(p)
		if l <= n/2 {
			modes = append(modes, true)
		}
		for _, streamed := range modes {
			backend, err := newCUDAOnlineBackendSchedule(p, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, streamed, false, streamed)
			if err != nil {
				t.Fatal(err)
			}
			gpu := &ScalingCUDAEncoder{backend: backend, k: p.k, n: n}
			defer gpu.Close()
			owned := gpu.OwnedBytes()
			input := make([]uint32, p.k)
			for i := range input {
				input[i] = uint32((17*i + 3) % (1 << uint(spec.FieldDegree)))
			}
			original := append([]uint32(nil), input...)
			buffer := 4096
			if spec.FieldDegree <= 3 {
				buffer = 1
			} else if spec.FieldDegree == 4 {
				buffer = 7
			} else if spec.FieldDegree == 6 {
				buffer = 257
			}
			for pass := 0; pass < 3; pass++ {
				if pass == 1 {
					clear(input)
				} else if pass == 2 {
					copy(input, original)
				}
				word, err := p.Encode(input, p.NewWorkspace())
				if err != nil {
					t.Fatal(err)
				}
				want := packedReference(word, spec.FieldDegree)
				savedInput := append([]uint32(nil), input...)
				var sink bytes.Buffer
				st, err := gpu.EncodePackedTo(input, &sink, buffer)
				hash := sha256.Sum256(want)
				if err != nil || !reflect.DeepEqual(input, savedInput) || !bytes.Equal(sink.Bytes(), want) || st.PackedBytes != uint64(len(want)) || st.PackedSHA256 != hex.EncodeToString(hash[:]) || st.HostBufferBytes != minInt(buffer, len(want)) || gpu.OwnedBytes() != owned {
					first := -1
					for i := range want {
						if i >= sink.Len() || sink.Bytes()[i] != want[i] {
							first = i
							break
						}
					}
					t.Fatal("packed oracle/digest/allocation", spec, streamed, "pass", pass, "first_byte", first, "bytes", st.PackedBytes, "buffer", st.HostBufferBytes, "owned", gpu.OwnedBytes(), owned, err)
				}
			}
			if !reflect.DeepEqual(input, original) {
				t.Fatal("packed input mutation")
			}
			for _, sink := range []io.Writer{failingPackedWriter{}, failingPackedWriter{short: true}} {
				if st, err := gpu.EncodePackedTo(input, sink, buffer); err == nil || st.PackedSHA256 != "" {
					t.Fatal("packed sink failure")
				}
			}
			for _, size := range []int{0, 65 << 20} {
				if _, err := gpu.EncodePackedTo(input, io.Discard, size); err == nil {
					t.Fatal("packed buffer bound")
				}
			}
			bad := append([]uint32(nil), input...)
			bad[0] = 1 << uint(spec.FieldDegree)
			if _, err := gpu.EncodePackedTo(bad, io.Discard, buffer); err == nil {
				t.Fatal("packed invalid input")
			}
			var recovered bytes.Buffer
			if _, err := gpu.EncodePackedTo(input, &recovered, buffer); err != nil {
				t.Fatal("packed error recovery", err)
			}
			// A normal owning encode after sink/error paths must still work.
			word, err := gpu.Encode(input)
			if err != nil || !bytes.Equal(recovered.Bytes(), packedReference(word, spec.FieldDegree)) {
				t.Fatal("mixed owning/packed recovery", err)
			}
			if old := os.Getenv("PIR_PACKED_OLD_LIBRARY"); old != "" {
				options := CUDAOptions{LibraryPath: old, DeviceMemoryLimit: 512 << 20}
				oldBackend, err := newCUDAOnlineBackendSchedule(p, options, streamed, false, streamed)
				if err != nil {
					t.Fatal(err)
				}
				previous := &ScalingCUDAEncoder{backend: oldBackend, k: p.k, n: n}
				_, err = previous.EncodePackedTo(input, io.Discard, buffer)
				previous.Close()
				if err == nil || !strings.Contains(err.Error(), "packed output ABI unavailable") {
					t.Fatal("old packed ABI", err)
				}
			}
			gpu.Close()
			if gpu.OwnedBytes() != 0 {
				t.Fatal("packed close leak")
			}
			if _, err := gpu.EncodePackedTo(input, io.Discard, buffer); !errors.Is(err, ErrCUDAClosed) {
				t.Fatal("packed closed", err)
			}
			records = append(records, map[string]any{"s": spec.FieldDegree, "streamed": streamed, "matches": 3, "writer_errors": true, "mixed_owning": true, "closed": true, "owned_bytes": owned, "buffer_bytes": buffer})
		}
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "packed-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}

// N=2^16 spans16 coefficient chunks. For a nonzero constant word, each
// 4096-value chunk has parity0 except the final4095-value chunk. Thread0's
// original partial sum is therefore0, while the joined reduction must be nonzero.
func TestCUDAPackedParity(t *testing.T) {
	out := os.Getenv("PIR_CUDA_PACKED_TEST_OUT")
	if out == "" {
		t.Skip("opt-in packed multi-chunk parity regression")
	}
	const s = 8
	const n = 1 << (2 * s)
	polynomial, _ := BinaryPolynomial(s)
	base, f, delta, omega := orbitFields(s, polynomial)
	spectrum := make([]uint32, n)
	for i := 0; i < n-1; i++ {
		spectrum[i] = 1
	}
	fft := newSparseTransform(f)
	fft.apply(spectrum, false)
	plan := &ScalingOnlinePlan{basis: OrbitBasis{spec: Spec{s, 2, 5}, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: f, extensionFFT: fft}, k: 1, p: []uint32{1, 1}, spectrum: spectrum}
	if cache := os.Getenv("PIR_CUDA_PACKED_TEST_CACHE"); cache != "" {
		if err := saveCUDAOnlineFixture(plan, filepath.Join(cache, "s8-parity-code.bin")); err != nil {
			t.Fatal(err)
		}
	}
	for _, streamed := range []bool{false, true} {
		backend, err := newCUDAOnlineBackendSchedule(plan, CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}, streamed, false, streamed)
		if err != nil {
			t.Fatal(err)
		}
		gpu := &ScalingCUDAEncoder{backend: backend, k: 1, n: n}
		defer gpu.Close()
		for _, value := range []byte{1, 255, 0, 17} {
			var sink bytes.Buffer
			st, err := gpu.EncodePackedTo([]uint32{uint32(value)}, &sink, 257)
			want := bytes.Repeat([]byte{value}, n)
			if err != nil || !bytes.Equal(sink.Bytes(), want) || st.PackedBytes != n {
				gpu.Close()
				t.Fatal("packed global parity/constant word", streamed, value, err)
			}
		}
		gpu.Close()
	}
	if err := os.WriteFile(filepath.Join(out, "parity-check.json"), []byte(`{"s":8,"chunks":16,"modes":2,"messages":4,"passed":true}`), 0644); err != nil {
		t.Fatal(err)
	}
}
