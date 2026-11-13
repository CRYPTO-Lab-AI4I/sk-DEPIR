//go:build liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCUDAHostOnline(t *testing.T) {
	out := os.Getenv("PIR_CUDA_HOST_OUT")
	if out == "" {
		t.Skip("opt-in complete host-backed GPU encoder")
	}
	fixtures := os.Getenv("PIR_CUDA_HOST_FIXTURES")
	var records []map[string]any
	for _, s := range []int{2, 3, 4, 6, 8} {
		raw, err := os.ReadFile(filepath.Join(fixtures, fmt.Sprintf("s%d-code.bin", s)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		p, err := LoadScalingOnline(bytes.NewReader(raw), hex.EncodeToString(sum[:]), 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		_, k, n, l, err := cudaShape(p)
		if err != nil {
			t.Fatal(err)
		}
		options := CUDAOptions{LibraryPath: os.Getenv("PIR_GPU_LIBRARY"), DeviceMemoryLimit: 512 << 20}
		if l > n/2 {
			if _, err = p.WithCUDAHostBacked(options, 65536); err == nil {
				t.Fatal("unsupported prefix accepted")
			}
			records = append(records, map[string]any{"s": s, "rejected_prefix": true})
			// The one-symbol subcode has generator 1+X and constant outputs.
			// It exercises successful small fields without another discovery.
			spectrum := make([]uint32, n)
			for i := 0; i < n-1; i++ {
				spectrum[i] = 1
			}
			newSparseTransform(p.basis.extension).apply(spectrum, false)
			p = &ScalingOnlinePlan{basis: p.basis, k: 1, p: []uint32{1, 1}, spectrum: spectrum}
			k = 1
		}
		for _, tile := range []int{256, 1536} {
			gpu, err := p.WithCUDAHostBacked(options, tile)
			if err != nil {
				t.Fatal(err)
			}
			defer gpu.Close()
			owned := gpu.OwnedBytes()
			input := make([]uint32, k)
			for i := range input {
				input[i] = uint32((17*i + 3) % (1 << uint(s)))
			}
			saved := append([]uint32(nil), input...)
			var last CUDAEncodeStats
			for repeat := 0; repeat < 3; repeat++ {
				if repeat == 1 {
					clear(input)
				} else {
					copy(input, saved)
				}
				before := append([]uint32(nil), input...)
				want, err := p.Encode(input, p.NewWorkspace())
				if err != nil {
					t.Fatal(err)
				}
				word, st, err := gpu.EncodeWithStats(input)
				if err != nil || !reflect.DeepEqual(word, want) || !reflect.DeepEqual(input, before) {
					t.Fatal("host-backed/Go/input", s, tile, err)
				}
				if st.Transforms != 5 || st.HostArenaBytes != uint64(4*n) || st.HostStageBytes > uint64(4*tile) || st.HostToDeviceBytes == 0 || st.DeviceToHostBytes == 0 || gpu.OwnedBytes() != owned {
					t.Fatal("host-backed accounting", st)
				}
				if s <= 4 {
					for _, x := range word {
						if x != input[0] {
							t.Fatal("one-symbol constant word", s)
						}
					}
				}
				last = st
			}
			if _, err = gpu.Encode(input[:len(input)-1]); err == nil {
				t.Fatal("message length")
			}
			input[0] = 1 << uint(s)
			if _, err = gpu.Encode(input); err == nil {
				t.Fatal("message range")
			}
			copy(input, saved)
			if _, err = gpu.Encode(input); err != nil {
				t.Fatal("error recovery", err)
			}
			if _, err = gpu.EncodePackedTo(input, &bytes.Buffer{}, 32); err == nil {
				t.Fatal("unsupported packed sink")
			}
			b := gpu.backend.(*cudaHostBackend)
			if _, err = b.Engine.EncodeOnlineBacked(b.p, b.spectrum, input, b.arena, b.arena, p.basis.omega); err == nil {
				t.Fatal("aliased buffers accepted")
			}
			old, err := p.WithCUDAHostBacked(CUDAOptions{LibraryPath: os.Getenv("PIR_CUDA_HOST_OLD_LIBRARY"), DeviceMemoryLimit: 512 << 20}, tile)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = old.Encode(input); err == nil {
				t.Fatal("old ABI accepted online")
			}
			old.Close()
			gpu.Close()
			if gpu.OwnedBytes() != 0 {
				t.Fatal("close device leak")
			}
			if _, err = gpu.Encode(input); err == nil {
				t.Fatal("closed encode")
			}
			records = append(records, map[string]any{"s": s, "tile": tile, "matches": 3, "errors_recovery_close": true, "stats": last})
		}
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "host-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
