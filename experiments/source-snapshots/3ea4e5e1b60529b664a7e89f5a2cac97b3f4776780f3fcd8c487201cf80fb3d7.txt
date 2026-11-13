//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type cpuPackedFailure struct{ short bool }

func (w cpuPackedFailure) Write(p []byte) (int, error) {
	if w.short {
		return 0, nil
	}
	return len(p) / 2, io.ErrClosedPipe
}

func TestCPPFilePacked(t *testing.T) {
	out, cache := os.Getenv("PIR_CPU_PACKED_OUT"), os.Getenv("PIR_CPU_PACKED_CACHE")
	if out == "" {
		t.Skip("opt-in CPU packed sink")
	}
	var records []map[string]any
	for index := 0; index < 9; index++ {
		s := index + 2
		var p *ScalingOnlinePlan
		if index < 7 {
			polynomial, _ := BinaryPolynomial(s)
			base, f, delta, omega := orbitFields(s, polynomial)
			spectrum := make([]uint32, 1<<uint(2*s))
			for i := 0; i < len(spectrum)-1; i++ {
				spectrum[i] = 1
			}
			fft := newSparseTransform(f)
			fft.apply(spectrum, false)
			p = &ScalingOnlinePlan{basis: OrbitBasis{spec: Spec{s, 2, 1}, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: f, extensionFFT: fft}, k: 1, p: []uint32{1, 1}, spectrum: spectrum}
		} else {
			s = 6
			if index == 8 {
				s = 8
			}
			raw, err := os.ReadFile(filepath.Join(os.Getenv("PIR_CPU_PACKED_FIXTURES"), fmt.Sprintf("s%d-code.bin", s)))
			if err != nil {
				t.Fatal(err)
			}
			hash := sha256.Sum256(raw)
			p, err = LoadScalingOnline(bytes.NewReader(raw), hex.EncodeToString(hash[:]), 1<<uint(2*s))
			if err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(cache, fmt.Sprintf("case-%d.bin", index))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = WriteScalingOnline(f, p); err != nil {
			t.Fatal(err)
		}
		f.Close()
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw)
		digest := hex.EncodeToString(hash[:])
		for _, workers := range []int{1, 16} {
			e, err := OpenScalingCPPFile(path, digest, p.Length(), workers)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			input := make([]uint32, p.k)
			for i := range input {
				input[i] = uint32((17*i + 3) % (1 << uint(s)))
			}
			saved := append([]uint32(nil), input...)
			var last CPUFilePackedStats
			for repeat := 0; repeat < 3; repeat++ {
				if repeat == 1 {
					clear(input)
				} else {
					copy(input, saved)
				}
				want, err := p.Encode(input, p.NewWorkspace())
				if err != nil {
					t.Fatal(err)
				}
				var expected bytes.Buffer
				_, digest, err := writeCanonicalPacked(&expected, want, s)
				if err != nil {
					t.Fatal(err)
				}
				var sink bytes.Buffer
				buffer := []int{1, 7, 257}[repeat]
				st, err := e.EncodePackedTo(input, &sink, buffer)
				if err != nil || !bytes.Equal(sink.Bytes(), expected.Bytes()) || st.SHA256 != digest || st.WrittenBytes != uint64(expected.Len()) || st.BufferBytes > buffer || st.WorkspaceBytes != 8*uint64(p.Length()) || st.Calls != 5 || st.BoundaryCalls != 1 || st.ReadBytes != 4*uint64(p.k+1+p.Length()) || e.WorkspaceBytes() != st.WorkspaceBytes {
					t.Fatal("packed/Go/accounting", index, workers, repeat, st, err)
				}
				last = st
			}
			copy(input, saved)
			own, _, err := e.EncodeWithStats(input)
			if err != nil {
				t.Fatal(err)
			}
			before := append([]uint32(nil), own...)
			for _, short := range []bool{false, true} {
				st, err := e.EncodePackedTo(input, cpuPackedFailure{short}, 17)
				if err == nil || st.SHA256 != "" {
					t.Fatal("writer error", err)
				}
			}
			if _, err = e.EncodePackedTo(input, io.Discard, 0); err == nil {
				t.Fatal("buffer bound")
			}
			if _, err = e.EncodePackedTo(input[:len(input)-1], io.Discard, 8); err == nil {
				t.Fatal("short input")
			}
			input[0] = 1 << uint(s)
			if _, err = e.EncodePackedTo(input, io.Discard, 8); err == nil {
				t.Fatal("field range")
			}
			copy(input, saved)
			if _, err = e.EncodePackedTo(input, io.Discard, 8); err != nil || !reflect.DeepEqual(input, saved) || !reflect.DeepEqual(own, before) {
				t.Fatal("recovery/input/detached word", err)
			}
			e.Close()
			if _, err = e.EncodePackedTo(input, io.Discard, 8); err == nil {
				t.Fatal("closed sink")
			}
			records = append(records, map[string]any{"case": index, "s": s, "workers": workers, "matches": 3, "writer_errors_recovery_owning": true, "stats": last})
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

func TestCPPFilePackedCached(t *testing.T) {
	out, cache, reference := os.Getenv("PIR_CPU_PACKED_MATCHED_OUT"), os.Getenv("PIR_CPU_PACKED_MATCHED_CACHE"), os.Getenv("PIR_CPU_PACKED_REFERENCE")
	if out == "" {
		t.Skip("opt-in matched CPU packed pipelines")
	}
	raw, err := os.ReadFile(filepath.Join(reference, "online-metadata.json"))
	if err != nil {
		t.Fatal(err)
	}
	var meta struct {
		SHA256 string
		K, N   int
	}
	if err = json.Unmarshal(raw, &meta); err != nil {
		t.Fatal(err)
	}
	input := make([]uint32, meta.K)
	rng := rand.New(rand.NewSource(71))
	for i := range input {
		input[i] = uint32(rng.Intn(1024))
	}
	var records []map[string]any
	var expected string
	for _, mode := range []string{"owning", "packed"} {
		start := time.Now()
		e, err := OpenScalingCPPFile(filepath.Join(reference, "online.bin"), meta.SHA256, meta.N, 16)
		if err != nil {
			t.Fatal(err)
		}
		setup := time.Since(start).Seconds()
		for repeat := 0; repeat < 2; repeat++ {
			path := filepath.Join(cache, fmt.Sprintf("%s-%d.bin", mode, repeat))
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			start = time.Now()
			var digest string
			var written uint64
			var packedStats CPUFilePackedStats
			var native ScalingCPPFileStats
			if mode == "owning" {
				word, st, err := e.EncodeWithStats(input)
				if err != nil {
					t.Fatal(err)
				}
				native = st
				written, digest, err = writeCanonicalPacked(f, word, 10)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				packedStats, err = e.EncodePackedTo(input, f, 1<<20)
				if err != nil {
					t.Fatal(err)
				}
				written, digest = packedStats.WrittenBytes, packedStats.SHA256
			}
			pipeline := time.Since(start).Seconds()
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			if err = f.Close(); err != nil {
				t.Fatal(err)
			}
			inclusive := time.Since(start).Seconds()
			if expected == "" {
				expected = digest
			}
			if digest != expected || written != uint64(meta.N*10/8) {
				t.Fatal("matched digest/size")
			}
			records = append(records, map[string]any{"mode": mode, "repeat": repeat, "setup_seconds": setup, "pipeline_seconds": pipeline, "sync_close_seconds": inclusive, "packed_sha256": digest, "packed_bytes": written, "native": native, "packed_stats": packedStats})
		}
		e.Close()
	}
	raw, err = json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "matched.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
