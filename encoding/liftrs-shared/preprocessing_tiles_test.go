//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Retain complete independent native outputs; small/sampled scalar Horner and
// schoolbook products ensure agreement is not only between GPU schedules.
func TestPreprocessingTileFixtures(t *testing.T) {
	directory := os.Getenv("PIR_PREPARATION_TILE_FIXTURES")
	if directory == "" {
		t.Skip("explicit fixture generation")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	write := func(name string, build func(*os.File) error) {
		path := filepath.Join(directory, name)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err = build(f); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
	}
	manifest := []map[string]any{}
	digest := func(value any) string {
		h := sha256.New()
		if err := binary.Write(h, binary.LittleEndian, value); err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf("%x", h.Sum(nil))
	}
	notify := func(phase, event string, done, total int, began time.Time) {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"utc": time.Now().UTC().Format(time.RFC3339Nano), "phase": phase, "event": event, "completed": done, "total": total, "unit": "points", "elapsed_seconds": time.Since(began).Seconds()})
	}
	put := func(f *os.File, value any) error { return binary.Write(f, binary.LittleEndian, value) }
	digestWords := func(values []uint64, s int) string {
		if s > 16 {
			return digest(values)
		}
		a := make([]uint32, len(values))
		for j := range a {
			a[j] = uint32(values[j])
		}
		return digest(a)
	}

	record := func(f *os.File, s int, n int, logs, powers []uint32, h ScalingCodeHeader) error {
		for _, v := range []any{uint32(s), h.Polynomial, h.Delta, h.Omega, uint64(n), uint64(len(logs)), logs, uint64(len(powers)), powers} {
			if err := put(f, v); err != nil {
				return err
			}
		}
		return nil
	}
	write("transforms.bin", func(f *os.File) error {
		if err := put(f, uint32(8)); err != nil {
			return err
		}
		for _, s := range []int{4, 17} {
			dims := []int{2, 5, 6, 8}
			if s == 17 {
				dims = []int{2, 7, 12, 22}
			}
			h, field, err := canonicalScalingCodeHeader(s, 3, 1, 1<<uint(2*s))
			if err != nil {
				return err
			}
			base, _ := NewBinaryField(s, h.Polynomial)
			table := newPreparationProducts(base)
			for _, dim := range dims {
				n := 1 << uint(dim)
				began := time.Now()
				phase := fmt.Sprintf("fixture-transform-s%d-d%d", s, dim)
				notify(phase, "started", 0, n, began)
				engine, err := cpp.NewPreparation(s, h.Polynomial, h.Omega, 4, uint64(n), 256<<20)
				if err != nil {
					return err
				}
				input := make([]uint64, n)
				rng := rand.New(rand.NewSource(int64(171103 + dim + s)))
				for j := range input {
					input[j] = rng.Uint64() & ((1 << uint(2*s)) - 1)
				}
				if s == 17 {
					input[0] = 1 << 33
				}
				actual := append([]uint64(nil), input...)
				if s <= 16 {
					a := make([]uint32, n)
					for j := range a {
						a[j] = uint32(input[j])
					}
					err = engine.Apply32(a, false)
					for j := range a {
						actual[j] = uint64(a[j])
					}
				} else {
					err = engine.Apply64(actual, false)
				}
				engine.Close()
				if err != nil {
					return err
				}
				stride := 1
				if n > 256 {
					stride = max(1, n/4)
				}
				checked := 0
				for point := 0; point < n; point += stride {
					var want uint64
					for j := n - 1; j >= 0; j-- {
						want = field.mul(want, uint64(point)) ^ input[j]
					}
					if want != actual[point] {
						return fmt.Errorf("CPU/scalar fixture mismatch s=%d n=%d point=%d", s, n, point)
					}
					if n > 256 {
						checked++
						notify(phase+"-scalar-check", "progress", checked, (n+stride-1)/stride, began)
					}
				}
				if err = record(f, s, n, table.log, table.powers, h); err != nil {
					return err
				}
				if err = put(f, input); err != nil {
					return err
				}
				if err = put(f, actual); err != nil {
					return err
				}
				manifest = append(manifest, map[string]any{"kind": "transform", "index": len(manifest), "s": s, "n": n, "forward_sha256": digestWords(actual, s), "inverse_sha256": digestWords(input, s), "element_bytes": engine.WordBytes()})
				notify(phase, "completed", n, n, began)
			}
		}
		return nil
	})

	write("origin.bin", func(file *os.File) error {
		const s = 17
		const length = 1 << 17
		const first = (1 << 17) - 1
		_, field, err := canonicalScalingCodeHeader(s, 3, 1, 1<<34)
		if err != nil {
			return err
		}
		input, expected := make([]uint64, length), make([]uint64, length)
		for j := 0; j < 16; j++ {
			input[j] = (uint64(170103+j*71317) ^ (1 << 33)) & ((1 << 34) - 1)
		}
		began := time.Now()
		notify("fixture-wide-origin", "started", 0, length, began)
		for point := range expected {
			for j := 15; j >= 0; j-- {
				expected[point] = field.mul(expected[point], uint64(first)*length^uint64(point)) ^ input[j]
			}
		}
		for _, v := range []any{uint64(length), uint64(first), input, expected} {
			if err = put(file, v); err != nil {
				return err
			}
		}
		manifest = append(manifest, map[string]any{"kind": "origin", "index": 0, "s": 17, "n": length, "first_coordinate": uint64(first) * length, "forward_sha256": digest(expected), "inverse_sha256": digest(input)})
		notify("fixture-wide-origin", "completed", length, length, began)
		return nil
	})
	write("products.bin", func(f *os.File) error {
		if err := put(f, uint32(3)); err != nil {
			return err
		}
		for _, item := range []struct{ s, power int }{{4, 6}, {17, 9}, {17, 22}} {
			n := 1 << uint(item.power)
			block := n / 8
			// Same s18 support fraction: ceil(P/B)=2, ceil(G/B)=7, P+G-1=N.
			k := int(uint64(n) * 16300262095 / (1 << 36))
			pc := k + 1
			gc := n - k
			h, _, err := canonicalScalingCodeHeader(item.s, 3, 1, 1<<uint(2*item.s))
			if err != nil {
				return err
			}
			base, _ := NewBinaryField(item.s, h.Polynomial)
			table := newPreparationProducts(base)
			a, b := make([]uint32, pc), make([]uint32, gc)
			rng := rand.New(rand.NewSource(int64(171104 + item.power)))
			for j := range a {
				a[j] = rng.Uint32() & ((1 << uint(item.s)) - 1)
			}
			for j := range b {
				b[j] = rng.Uint32() & ((1 << uint(item.s)) - 1)
			}
			engine, err := cpp.NewPreparation(item.s, h.Polynomial, h.Omega, 4, uint64(n), 256<<20)
			if err != nil {
				return err
			}
			began := time.Now()
			phase := fmt.Sprintf("fixture-product-s%d-n%d", item.s, n)
			notify(phase, "started", 0, n, began)
			expected, err := engine.Product(a, b, n, n)
			engine.Close()
			if err != nil {
				return err
			}
			if n <= 512 {
				scalar := make([]uint32, n)
				for i, x := range a {
					for j, y := range b {
						scalar[i+j] ^= base.Mul(x, y)
					}
				}
				for j := range scalar {
					if scalar[j] != expected[j] {
						return fmt.Errorf("native/schoolbook product mismatch")
					}
				}
			}
			if err = record(f, item.s, n, table.log, table.powers, h); err != nil {
				return err
			}
			for _, v := range []any{uint64(block), uint64(pc), a, uint64(gc), b, expected} {
				if err = put(f, v); err != nil {
					return err
				}
			}
			manifest = append(manifest, map[string]any{"kind": "product", "index": len(manifest) - 9, "s": item.s, "n": n, "p_count": pc, "g_count": gc, "block": block, "output_sha256": digest(expected)})
			notify(phase, "completed", n, n, began)
		}
		return nil
	})
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(directory, "manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
}
