//go:build liftrs_cpp && liftrs_cuda && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
)

type spectrumShortWriter struct{}

func (spectrumShortWriter) Write(b []byte) (int, error) { return len(b) - 1, nil }

// Horner is independent of both GPU output strategies, including the true s17
// extension field. Tiny tiles cross GPU shard boundaries and exercise tail slots.
func TestGPUPreparationStreamedSpectrum(t *testing.T) {
	devices := gpuPreparationDevices(t)
	for _, s := range []int{8, 17} {
		h, f, err := canonicalScalingCodeHeader(s, 5, 1, 1<<uint(2*s))
		if err != nil {
			t.Fatal(err)
		}
		base, _ := NewBinaryField(s, h.Polynomial)
		table := newPreparationProducts(base)
		for _, layout := range []string{"peer", "local"} {
			t.Setenv("LIFTRS_GPU_PREPARATION_VECTOR_LAYOUT", layout)
			for count := 1; count <= len(devices); count *= 2 {
				t.Run(fmt.Sprintf("s%d/%s/%d", s, layout, count), func(t *testing.T) {
					engine, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: devices[:count], MaxElements: 4096, DeviceBytes: gpuPreparationCap(t)})
					if err != nil {
						t.Fatal(err)
					}
					defer engine.Close()
					g := make([]uint32, 33)
					for j := range g {
						g[j] = uint32(j*71317+170069) & ((1 << uint(s)) - 1)
					}
					values := make([]uint64, 4096)
					for point := range values {
						for j := len(g) - 1; j >= 0; j-- {
							values[point] = f.mul(values[point], uint64(point)) ^ uint64(g[j])
						}
					}
					width := uint8(4)
					if s > 16 {
						width = 8
					}
					var expected bytes.Buffer
					if err = writeScalingCodeValues(&expected, width, 1<<uint(2*s), nil, values); err != nil {
						t.Fatal(err)
					}
					for _, tile := range []int{8, 24, 1000, 64 << 10} {
						var actual bytes.Buffer
						if err = engine.SpectrumTo(g, 4096, &actual, tile); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
							t.Fatal("stream vs scalar", tile, err)
						}
					}
					if err = engine.SpectrumTo(g, 4096, spectrumShortWriter{}, 24); !errors.Is(err, io.ErrShortWrite) {
						t.Fatal("short writer", err)
					}
					sentinel := errors.New("sink failed")
					if err = engine.SpectrumTo(g, 4096, codeFailWriter{sentinel}, 24); !errors.Is(err, sentinel) {
						t.Fatal("writer failure", err)
					}
					if err = engine.SpectrumTo(g, 4096, io.Discard, 7); err == nil {
						t.Fatal("underfunded tile")
					}
					if err = engine.SpectrumTo(g, 4095, io.Discard, 24); err == nil {
						t.Fatal("nonpower shape")
					}
					bad := []uint32{1 << uint(s)}
					if err = engine.SpectrumTo(bad, 4096, io.Discard, 24); err == nil {
						t.Fatal("field range")
					}
					var actual bytes.Buffer
					if err = engine.SpectrumTo(g, 4096, &actual, 24); err != nil || !bytes.Equal(actual.Bytes(), expected.Bytes()) {
						t.Fatal("reuse after failure", err)
					}
					engine.Close()
					if err = engine.SpectrumTo(g, 4096, io.Discard, 24); err == nil {
						t.Fatal("closed engine")
					}
				})
			}
		}
	}
}

// This component compares complete transform+download+hash/write boundaries, not
// full preprocessing or disk bandwidth. Dense G avoids sparse-input shortcuts.
func TestGPUPreparationSpectrumTransfer(t *testing.T) {
	if os.Getenv("PIR_GPU_SPECTRUM_MEASURE") != "1" {
		t.Skip("explicit bounded comparison")
	}
	h, _, err := canonicalScalingCodeHeader(17, 5, 1, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	base, _ := NewBinaryField(17, h.Polynomial)
	table := newPreparationProducts(base)
	const n = 1 << 22
	engine, err := cuda.NewPreparation(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: 17, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: table.log, Powers: table.powers, Devices: gpuPreparationDevices(t), MaxElements: n, DeviceBytes: gpuPreparationCap(t)})
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	g := make([]uint32, n/2)
	for j := range g {
		g[j] = uint32(j*71317+170069) & ((1 << 17) - 1)
	}
	data := make([]uint64, n)
	var records []map[string]any
	for sample := 0; sample < 2; sample++ {
		for _, stream := range []bool{sample%2 == 1, sample%2 != 1} {
			hash := sha256.New()
			began := time.Now()
			if stream {
				err = engine.SpectrumTo(g, n, hash, 1<<20)
			} else {
				clear(data)
				for j, v := range g {
					data[j] = uint64(v)
				}
				if err = engine.Apply64(data, false); err == nil {
					err = writeScalingCodeValues(hash, 8, 1<<34, nil, data)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			records = append(records, map[string]any{"sample": sample, "streamed": stream, "seconds": time.Since(began).Seconds(), "sha256": fmt.Sprintf("%x", hash.Sum(nil)), "n": n, "stream_tile_bytes": 1 << 20, "resident_output_bytes": 8 * n})
		}
	}
	for _, r := range records {
		if r["sha256"] != records[0]["sha256"] {
			t.Fatal("stream/resident hash mismatch")
		}
	}
	raw, _ := json.Marshal(records)
	fmt.Println("SPECTRUM_TRANSFER " + string(raw))
}
