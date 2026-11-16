//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func cudaEncodingFixtures(t *testing.T, spec Spec) (*ScalingCodeData, *ScalingOnlinePlan, []uint32, []uint32) {
	t.Helper()
	d, p := wideCodeFixture(t, spec)
	message := make([]uint32, p.Dimension())
	r := rand.New(rand.NewSource(170087))
	for i := range message {
		message[i] = uint32(r.Intn(1 << uint(spec.FieldDegree)))
	}
	word, err := p.Encode(message, p.NewWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	cpu, err := d.WithWideCPP(2)
	if err != nil {
		t.Fatal(err)
	}
	defer cpu.Close()
	workspace, err := cpu.NewWorkspace(cpu.Resources().WorkspaceBytes)
	if err != nil {
		t.Fatal(err)
	}
	native, err := cpu.Encode(message, workspace)
	if err != nil || !reflect.DeepEqual(native, word) {
		t.Fatal("CPU/Go fixture", err)
	}
	return d, p, message, word
}
func TestCUDACodeSmall(t *testing.T) {
	library := os.Getenv("PIR_GPU_PREPARATION_LIBRARY")
	if library == "" {
		t.Skip("explicit resident plugin required")
	}
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {4, 2, 3}, {6, 2, 4}, {8, 2, 5}} {
		d, _, message, want := cudaEncodingFixtures(t, spec)
		for _, count := range []int{1, 2, 4} {
			for _, layout := range []string{"peer", "local"} {
				t.Run(fmt.Sprintf("s%d-t%d/%d/%s", spec.FieldDegree, spec.CurveDegree, count, layout), func(t *testing.T) {
					devices := make([]int, count)
					for i := range devices {
						devices[i] = i
					}
					original := append([]uint32(nil), message...)
					r, _ := d.CUDAResources(count)
					e, err := d.WithCUDA(CUDACodeOptions{LibraryPath: library, Devices: devices, DeviceMemoryLimit: 1 << 30, VectorLayout: layout})
					if err != nil {
						t.Fatal(err)
					}
					defer e.Close()
					owned := e.OwnedBytes()
					if owned < r.CodeBytes+r.DeviceWorkspaceBytes || owned > uint64(count)<<30 {
						t.Fatal("device accounting", owned, r)
					}
					word, stats, err := e.EncodeWithStats(message)
					if err != nil || !reflect.DeepEqual(word, want) {
						t.Fatal("GPU/CPU/Go codeword mismatch", err)
					}
					if stats.TimingSource != "synchronized-host-wall" || stats.NativeTotalNS <= 0 {
						t.Fatal("timing boundary", stats)
					}
					saved := append([]uint32(nil), word...)
					zero, err := e.Encode(make([]uint32, len(message)))
					if err != nil {
						t.Fatal(err)
					}
					for _, value := range zero {
						if value != 0 {
							t.Fatal("nonzero zero-word")
						}
					}
					bad := append([]uint32(nil), message...)
					bad[0] = 1 << uint(spec.FieldDegree)
					if _, err = e.Encode(bad); err == nil {
						t.Fatal("invalid field accepted")
					}
					if _, err = e.Encode(nil); err == nil {
						t.Fatal("invalid shape accepted")
					}
					again, err := e.Encode(message)
					if err != nil || !reflect.DeepEqual(again, want) || !reflect.DeepEqual(saved, word) || !reflect.DeepEqual(message, original) {
						t.Fatal("retry/ownership/immutability", err)
					}
					if e.OwnedBytes() != owned {
						t.Fatal("unexpected repeat allocation")
					}
					if count == 4 && layout == "local" {
						// Bound aggregate live caps before creating a second peer lease. Closing
						// one client must not disable links still used by the other client.
						if owned+4*(64<<20) > 4<<30 {
							t.Fatal("coexisting engine admission")
						}
						second, err := d.WithCUDA(CUDACodeOptions{LibraryPath: library, Devices: devices, DeviceMemoryLimit: 64 << 20, VectorLayout: layout})
						if err != nil {
							t.Fatal(err)
						}
						defer second.Close()
						if err = e.Close(); err != nil {
							t.Fatal(err)
						}
						shared, err := second.Encode(message)
						if err != nil || !reflect.DeepEqual(shared, want) {
							t.Fatal("peer lease released while another client lives", err)
						}
						if err = second.Close(); err != nil {
							t.Fatal(err)
						}
					}
					if err = e.Close(); err != nil || e.OwnedBytes() != 0 {
						t.Fatal("close/accounting", err)
					}
					if !reflect.DeepEqual(word, want) {
						t.Fatal("word changed after close")
					}
					if _, err = e.Encode(message); err != ErrCUDAClosed {
						t.Fatal("closed encoder accepted", err)
					}
				})
			}
		}
	}
}

// The raw checker forces both storage widths on genuine small code data, then
// separately checks true34-bit positions. It never pretends a small s17 domain
// is a genuine full codeword. binary.Write is bounded to these tiny fixtures.
func TestCUDACodeNativeFixtures(t *testing.T) {
	dir := os.Getenv("PIR_GPU_CODE_FIXTURES")
	if dir == "" {
		t.Skip("explicit fixture destination")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	var data bytes.Buffer
	write := func(v any) {
		if err := binary.Write(&data, binary.LittleEndian, v); err != nil {
			t.Fatal(err)
		}
	}
	write(uint32(5))
	for _, spec := range []Spec{{3, 2, 2}, {4, 2, 2}, {4, 2, 3}, {6, 2, 4}, {8, 2, 5}} {
		d, _, message, word := cudaEncodingFixtures(t, spec)
		h := d.header
		base, _ := NewBinaryField(spec.FieldDegree, h.Polynomial)
		table := newPreparationProducts(base)
		write(uint32(spec.FieldDegree))
		write(h.Polynomial)
		write(h.Delta)
		write(h.Omega)
		write(h.K)
		write(h.N)
		write(table.log)
		write(table.powers)
		write(d.p)
		for _, v := range d.spectrum32 {
			write(uint64(v))
		}
		write(message)
		write(word)
	}
	h, f, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := NewBinaryField(17, h.Polynomial)
	table := newPreparationProducts(base)
	write(h.Polynomial)
	write(h.Delta)
	write(h.Omega)
	write(table.log)
	write(table.powers)
	indices := []uint64{0, 1, 65536, (1 << 32) - 1, 1 << 32, (1 << 33) + 179, (1 << 34) - 2}
	write(uint32(len(indices)))
	high := false
	for _, i := range indices {
		write(i)
		position := f.pow(h.Omega, i+1)
		write(position)
		if position > 1<<32-1 {
			high = true
		}
	}
	if !high {
		t.Fatal("fixture did not exercise high canonical positions")
	}
	if err = os.WriteFile(filepath.Join(dir, "native.bin"), data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
}
