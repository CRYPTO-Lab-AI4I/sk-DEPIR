package liftrs

// These opt-in experiment helpers exercise the actual private Go kernels. Normal
// unit tests skip them. Binary fixtures are shared with C++; no public API is added.
import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"
	"time"
)

func benchCheck(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func put32(t *testing.T, w io.Writer, x uint32) {
	benchCheck(t, binary.Write(w, binary.LittleEndian, x))
}
func get32(t *testing.T, r io.Reader) uint32 {
	var x uint32
	benchCheck(t, binary.Read(r, binary.LittleEndian, &x))
	return x
}
func writeProduct(t *testing.T, w io.Writer, p constantProduct) {
	put32(t, w, p.coefficient)
	put32(t, w, uint32(len(p.rows)))
	for _, x := range p.rows {
		put32(t, w, x)
	}
}
func readProduct(t *testing.T, r io.Reader) constantProduct {
	p := constantProduct{coefficient: get32(t, r)}
	p.rows = make([]uint32, get32(t, r))
	for i := range p.rows {
		p.rows[i] = get32(t, r)
	}
	return p
}
func TestExportLanguageFixtures(t *testing.T) {
	dir := os.Getenv("PIR_LANGUAGE_FIXTURES")
	if dir == "" {
		t.Skip("experiment fixture export is opt-in")
	}
	benchCheck(t, os.MkdirAll(dir, 0755))
	if _, err := os.Stat(filepath.Join(dir, "transform-input.bin")); !os.IsNotExist(err) {
		t.Fatal("fixtures already exist; reuse them")
	}
	profile, err := os.Create(filepath.Join(dir, "export-go.pprof"))
	benchCheck(t, err)
	benchCheck(t, pprof.StartCPUProfile(profile))
	defer profile.Close()
	defer pprof.StopCPUProfile()
	base, ext, _, _ := orbitFields(13, 0x201b)
	for _, item := range []struct {
		name string
		f    orbitField
	}{{"base", base}, {"extension", ext}} {
		a := newAdditiveTransform(item.f)
		var products []constantProduct
		for k := range a.terms {
			for _, term := range a.terms[k] {
				if len(term.product.rows) > 0 {
					products = append(products, term.product)
				}
			}
			for _, p := range []constantProduct{a.deltaProducts[k], a.inverseProducts[k]} {
				if len(p.rows) > 0 {
					products = append(products, p)
				}
			}
		}
		file, err := os.Create(filepath.Join(dir, item.name+".bin"))
		benchCheck(t, err)
		w := bufio.NewWriterSize(file, 1<<20)
		put32(t, w, item.f.q)
		put32(t, w, uint32(len(products)))
		put32(t, w, 1<<22)
		for _, p := range products {
			writeProduct(t, w, p)
		}
		random := rand.New(rand.NewSource(1313))
		for i := 0; i < 1<<22; i++ {
			put32(t, w, random.Uint32()&(item.f.q-1))
		}
		benchCheck(t, w.Flush())
		benchCheck(t, file.Close())
	}
	a := newAdditiveTransform(ext)
	file, err := os.Create(filepath.Join(dir, "transform.bin"))
	benchCheck(t, err)
	w := bufio.NewWriterSize(file, 1<<20)
	for _, x := range []uint32{13, 0x201b, ext.extension.Delta(), ext.q, uint32(len(a.terms))} {
		put32(t, w, x)
	}
	for k := range a.terms {
		put32(t, w, uint32(len(a.terms[k])))
		for _, term := range a.terms[k] {
			put32(t, w, uint32(term.degree))
			writeProduct(t, w, term.product)
		}
		for _, x := range a.values[k] {
			put32(t, w, x)
		}
		put32(t, w, a.delta[k])
		put32(t, w, a.inverse[k])
		writeProduct(t, w, a.deltaProducts[k])
		writeProduct(t, w, a.inverseProducts[k])
	}
	benchCheck(t, w.Flush())
	benchCheck(t, file.Close())
	file, err = os.Create(filepath.Join(dir, "transform-input.bin"))
	benchCheck(t, err)
	w = bufio.NewWriterSize(file, 1<<20)
	random := rand.New(rand.NewSource(1313))
	for i := uint32(0); i < ext.q; i++ {
		put32(t, w, random.Uint32()&(ext.q-1))
	}
	benchCheck(t, w.Flush())
	benchCheck(t, file.Close())
}

//go:noinline
func languageKernel(products []constantProduct, f orbitField, in, out []uint32, rotating bool) {
	for start := 0; start < len(in); start += 4096 {
		end := start + 4096
		if end > len(in) {
			end = len(in)
		}
		index := len(products) / 2
		if rotating {
			index = (start / 4096) % len(products)
		}
		p := products[index]
		for i := start; i < end; i++ {
			out[i] = p.mul(f, in[i])
		}
	}
}
func TestRunLanguageBench(t *testing.T) {
	dir := os.Getenv("PIR_LANGUAGE_FIXTURES")
	mode := os.Getenv("PIR_LANGUAGE_MODE")
	if dir == "" || mode == "" {
		t.Skip("language benchmark is opt-in")
	}
	name := os.Getenv("PIR_LANGUAGE_CASE")
	outdir := os.Getenv("PIR_LANGUAGE_OUTPUT")
	benchCheck(t, os.MkdirAll(outdir, 0755))
	base, ext, _, _ := orbitFields(13, 0x201b)
	f := ext
	var in []uint32
	var products []constantProduct
	var a *additiveTransform
	size := 1 << 22
	if value := os.Getenv("PIR_LANGUAGE_SIZE"); value != "" {
		var err error
		size, err = strconv.Atoi(value)
		benchCheck(t, err)
	}
	if mode == "kernel" {
		if name == "base" {
			f = base
		}
		file, err := os.Open(filepath.Join(dir, name+".bin"))
		benchCheck(t, err)
		r := bufio.NewReaderSize(file, 1<<20)
		q, count, n := get32(t, r), get32(t, r), get32(t, r)
		if q != f.q || size > int(n) {
			t.Fatal("fixture dimensions")
		}
		products = make([]constantProduct, count)
		for i := range products {
			products[i] = readProduct(t, r)
		}
		in = make([]uint32, n)
		benchCheck(t, binary.Read(r, binary.LittleEndian, in))
		benchCheck(t, file.Close())
		in = in[:size]
	} else {
		a = newAdditiveTransform(ext)
		file, err := os.Open(filepath.Join(dir, "transform-input.bin"))
		benchCheck(t, err)
		in = make([]uint32, size)
		benchCheck(t, binary.Read(bufio.NewReaderSize(file, 1<<20), binary.LittleEndian, in))
		benchCheck(t, file.Close())
	}
	out := make([]uint32, len(in))
	passes := 1
	if mode == "kernel" {
		passes = (1 << 27) / len(in)
	}
	rotating := os.Getenv("PIR_LANGUAGE_PATTERN") == "rotating"
	var times []int64
	for repeat := 0; repeat < 3; repeat++ {
		copy(out, in)
		runtime.GC()
		profile, err := os.Create(filepath.Join(outdir, fmt.Sprintf("go-%d.pprof", repeat)))
		benchCheck(t, err)
		benchCheck(t, pprof.StartCPUProfile(profile))
		start := time.Now()
		if mode == "kernel" {
			for pass := 0; pass < passes; pass++ {
				languageKernel(products, f, in, out, rotating)
			}
		} else {
			a.applySubspace(out, mode == "inverse")
		}
		times = append(times, time.Since(start).Nanoseconds())
		pprof.StopCPUProfile()
		benchCheck(t, profile.Close())
	}
	file, err := os.Create(filepath.Join(outdir, "output.bin"))
	benchCheck(t, err)
	h := sha256.New()
	w := bufio.NewWriterSize(io.MultiWriter(file, h), 1<<20)
	benchCheck(t, binary.Write(w, binary.LittleEndian, out))
	benchCheck(t, w.Flush())
	benchCheck(t, file.Close())
	// Horner is an independent transform oracle on the small shared fixture.
	if mode == "forward" && len(in) <= 256 {
		for point := range out {
			var expected uint32
			for j := len(in) - 1; j >= 0; j-- {
				expected = f.mul(expected, uint32(point)) ^ in[j]
			}
			if out[point] != expected {
				t.Fatal("Horner mismatch", point)
			}
		}
	}
	// Exercise the inverse against the exact original input, outside timing.
	if mode == "forward" {
		a.applySubspace(out, true)
		for i := range out {
			if out[i] != in[i] {
				t.Fatal("FFT inverse mismatch", i)
			}
		}
	}
	entries := 0
	for _, p := range products {
		entries += len(p.rows)
	}
	result := map[string]any{"language": "go", "mode": mode, "case": name, "size": len(in), "passes": passes, "elapsed_ns": times, "sha256": hex.EncodeToString(h.Sum(nil)), "table_bytes": entries * 4, "field_bits": bits.Len32(f.q) - 1}
	raw, err := json.MarshalIndent(result, "", "  ")
	benchCheck(t, err)
	benchCheck(t, os.WriteFile(filepath.Join(outdir, "result.json"), append(raw, '\n'), 0644))
}
