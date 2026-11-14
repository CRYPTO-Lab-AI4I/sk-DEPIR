//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCPUWorkspaceLifetime(t *testing.T) {
	var nilEncoder *ScalingCPPEncoder
	var zeroEncoder ScalingCPPEncoder
	for _, e := range []*ScalingCPPEncoder{nilEncoder, &zeroEncoder} {
		if e.NewWorkspace() != nil {
			t.Fatal("uninitialized encoder allocated a workspace")
		}
		if _, err := e.NewCompactWorkspace(); err != ErrCPPClosed {
			t.Fatal(err)
		}
		if _, err := e.Encode(nil, nil); err != ErrCPPClosed {
			t.Fatal(err)
		}
		if _, _, err := e.EncodeHybridWithStats(nil, nil); err != ErrCPPClosed {
			t.Fatal(err)
		}
		if _, _, err := e.EncodeCompactWithStats(nil, nil); err != ErrCPPClosed {
			t.Fatal(err)
		}
		e.Close()
	}
	var w *ScalingCPPWorkspace
	var c *ScalingCPPCompactWorkspace
	if w.Bytes() != 0 || c.Bytes() != 0 {
		t.Fatal("nil workspace accounting")
	}
	_, plan := wideCodeFixture(t, Spec{4, 2, 2})
	e, err := plan.WithCPPMatrix(2)
	if err != nil {
		t.Fatal(err)
	}
	e.Close()
	if e.NewWorkspace() != nil {
		t.Fatal("closed encoder allocated a workspace")
	}
	if _, err = e.NewCompactWorkspace(); err != ErrCPPClosed {
		t.Fatal(err)
	}
}

// Both binaries use this identical benchmark file and the same verified cached
// s10 code. Setup/checking are outside encoding timers; allocation, workspace
// replacement and one complete call are inside. No GPU/preprocessing is needed.
func BenchmarkCPUEncodingModules(b *testing.B) {
	path := os.Getenv("PIR_CPU_MODULES_CODE")
	digest := os.Getenv("PIR_CPU_MODULES_SHA256")
	if path == "" {
		b.Skip("explicit verified small code cache required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != digest {
		b.Fatal("code digest")
	}
	plan, err := LoadScalingOnline(bytes.NewReader(raw), digest, 1<<20)
	if err != nil {
		b.Fatal(err)
	}
	data, err := LoadScalingCode(bytes.NewReader(raw), digest, ScalingCodeLimits{1 << 20, 8 << 20})
	if err != nil {
		b.Fatal(err)
	}
	message := make([]uint32, plan.Dimension())
	for i := range message {
		message[i] = uint32(i*7+3) & 1023
	}
	want, err := plan.Encode(message, plan.NewWorkspace())
	if err != nil {
		b.Fatal(err)
	}
	var packed bytes.Buffer
	_, packedHash, err := writeCanonicalPacked(&packed, want, 10)
	if err != nil {
		b.Fatal(err)
	}
	canonicalHash := sha256.Sum256(rawUint32LE(want))
	b.Logf("code=%s canonical=%x packed=%s N=%d K=%d", digest, canonicalHash, packedHash, plan.Length(), plan.Dimension())
	cpu, err := plan.WithCPPMatrix(4)
	if err != nil {
		b.Fatal(err)
	}
	defer cpu.Close()
	wide, err := data.WithWideCPP(4)
	if err != nil {
		b.Fatal(err)
	}
	defer wide.Close()
	w := cpu.NewWorkspace()
	hybrid := cpu.NewWorkspace()
	compact, err := cpu.NewCompactWorkspace()
	if err != nil {
		b.Fatal(err)
	}
	ww, err := wide.NewWorkspace(wide.Resources().WorkspaceBytes)
	if err != nil {
		b.Fatal(err)
	}
	file, err := OpenScalingCPPFile(path, digest, 1<<20, 4)
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()
	disk, err := OpenScalingCPPDisk(path, digest, 1<<20, 4, 65536, filepath.Join(b.TempDir(), "workspace"))
	if err != nil {
		b.Fatal(err)
	}
	defer disk.Close()
	cases := []struct {
		name   string
		encode func() ([]uint32, error)
	}{
		{"resident", func() ([]uint32, error) { return cpu.Encode(message, w) }},
		{"hybrid", func() ([]uint32, error) { word, _, err := cpu.EncodeHybridWithStats(message, hybrid); return word, err }},
		{"wide", func() ([]uint32, error) { return wide.Encode(message, ww) }},
		{"compact", func() ([]uint32, error) {
			word, _, err := cpu.EncodeCompactWithStats(message, compact)
			return word, err
		}},
		{"file", func() ([]uint32, error) { word, _, err := file.EncodeWithStats(message); return word, err }},
		{"disk", func() ([]uint32, error) { word, _, err := disk.EncodeWithStats(message); return word, err }},
	}
	for _, test := range cases {
		b.Run(test.name, func(b *testing.B) {
			// Establish reusable workspace state consistently before samples.
			if word, err := test.encode(); err != nil || !reflect.DeepEqual(word, want) {
				b.Fatal(test.name, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				word, err := test.encode()
				b.StopTimer()
				if err != nil || !reflect.DeepEqual(word, want) {
					b.Fatal(test.name, err)
				}
				b.StartTimer()
			}
		})
	}
	b.Run("load-resident", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loaded, err := LoadScalingOnline(bytes.NewReader(raw), digest, 1<<20)
			if err != nil || loaded.Dimension() != plan.Dimension() {
				b.Fatal(err)
			}
		}
	})
	b.Run("open-file", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loaded, err := OpenScalingCPPFile(path, digest, 1<<20, 4)
			if err != nil {
				b.Fatal(err)
			}
			if err = loaded.Close(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("packed", func(b *testing.B) {
		// Include writer/hash work while keeping the destination bounded and synchronous.
		if _, err := file.EncodePackedTo(message, io.Discard, 65536); err != nil {
			b.Fatal(err)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			stats, err := file.EncodePackedTo(message, io.Discard, 65536)
			b.StopTimer()
			if err != nil || stats.SHA256 != packedHash {
				b.Fatal(fmt.Sprint("packed identity ", err))
			}
			b.StartTimer()
		}
	})
}

func rawUint32LE(values []uint32) []byte {
	out := make([]byte, 4*len(values))
	for i, v := range values {
		for j := 0; j < 4; j++ {
			out[4*i+j] = byte(v >> uint(8*j))
		}
	}
	return out
}

// Exercise the same code/message through every maintained native CPU memory
// strategy, plus the independent Go schedule. Keep real small verified code
// fixtures so a shared scheduling bug cannot hide behind mode agreement alone.
func TestCPUEncodingModules(t *testing.T) {
	for _, spec := range []Spec{{4, 2, 2}, {6, 2, 4}, {8, 2, 5}} {
		data, plan := wideCodeFixture(t, spec)
		for _, workers := range []int{1, 4} {
			t.Run(fmt.Sprintf("s%d/workers%d", spec.FieldDegree, workers), func(t *testing.T) {
				var raw bytes.Buffer
				if err := WriteScalingCode(&raw, data); err != nil {
					t.Fatal(err)
				}
				// Also exercise legacy wide-count headers at a small field degree.
				if workers == 4 {
					old := append([]byte(nil), raw.Bytes()...)
					raw.Reset()
					raw.WriteString("LRSprobe2")
					raw.Write(old[9:17])
					for _, v := range []uint64{uint64(plan.Dimension()), uint64(plan.Length())} {
						for j := 0; j < 8; j++ {
							raw.WriteByte(byte(v >> uint(8*j)))
						}
					}
					raw.Write(old[25:])
				}
				path := filepath.Join(t.TempDir(), "online.bin")
				if err := os.WriteFile(path, raw.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256(raw.Bytes())
				digest := hex.EncodeToString(sum[:])
				cpu, err := plan.WithCPPMatrix(workers)
				if err != nil {
					t.Fatal(err)
				}
				defer cpu.Close()
				wide, err := data.WithWideCPP(workers)
				if err != nil {
					t.Fatal(err)
				}
				defer wide.Close()
				wideWork, err := wide.NewWorkspace(wide.Resources().WorkspaceBytes)
				if err != nil {
					t.Fatal(err)
				}
				file, err := OpenScalingCPPFile(path, digest, plan.Length(), workers)
				if err != nil {
					t.Fatal(err)
				}
				defer file.Close()
				disk, err := OpenScalingCPPDisk(path, digest, plan.Length(), workers, 1536, filepath.Join(t.TempDir(), "work"))
				if err != nil {
					t.Fatal(err)
				}
				defer disk.Close()
				compact, err := cpu.NewCompactWorkspace()
				if err != nil {
					t.Fatal(err)
				}
				residentWork, hybridWork := cpu.NewWorkspace(), cpu.NewWorkspace()
				modes := []struct {
					name   string
					encode func([]uint32) ([]uint32, error)
				}{
					{"resident", func(m []uint32) ([]uint32, error) { return cpu.Encode(m, residentWork) }},
					{"hybrid", func(m []uint32) ([]uint32, error) { w, _, e := cpu.EncodeHybridWithStats(m, hybridWork); return w, e }},
					{"wide", func(m []uint32) ([]uint32, error) { return wide.Encode(m, wideWork) }},
					{"compact", func(m []uint32) ([]uint32, error) { w, _, e := cpu.EncodeCompactWithStats(m, compact); return w, e }},
					{"file", func(m []uint32) ([]uint32, error) { w, _, e := file.EncodeWithStats(m); return w, e }},
					{"disk", func(m []uint32) ([]uint32, error) { w, _, e := disk.EncodeWithStats(m); return w, e }},
				}
				message := make([]uint32, plan.Dimension())
				for j := range message {
					message[j] = uint32(j*7+3) & ((1 << uint(spec.FieldDegree)) - 1)
				}
				want, err := plan.Encode(message, plan.NewWorkspace())
				if err != nil {
					t.Fatal(err)
				}
				for _, mode := range modes {
					first, err := mode.encode(message)
					if err != nil || !reflect.DeepEqual(first, want) {
						t.Fatal(mode.name, err)
					}
					saved := append([]uint32(nil), first...)
					zero, err := mode.encode(make([]uint32, len(message)))
					if err != nil || !reflect.DeepEqual(zero, make([]uint32, plan.Length())) {
						t.Fatal(mode.name, "zero", err)
					}
					bad := append([]uint32(nil), message...)
					bad[0] = 1 << uint(spec.FieldDegree)
					if _, err = mode.encode(bad); err == nil {
						t.Fatal(mode.name, "invalid field value")
					}
					got, err := mode.encode(message)
					if err != nil || !reflect.DeepEqual(got, want) || !reflect.DeepEqual(first, saved) {
						t.Fatal(mode.name, "recovery/ownership", err)
					}
				}
				var packed, expected bytes.Buffer
				if _, _, err = writeCanonicalPacked(&expected, want, spec.FieldDegree); err != nil {
					t.Fatal(err)
				}
				st, err := file.EncodePackedTo(message, &packed, 31)
				if err != nil || !bytes.Equal(packed.Bytes(), expected.Bytes()) || st.WrittenBytes != uint64(expected.Len()) {
					t.Fatal("packed representation", err)
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(after, raw.Bytes()) {
					t.Fatal("immutable code", err)
				}
			})
		}
	}
}
