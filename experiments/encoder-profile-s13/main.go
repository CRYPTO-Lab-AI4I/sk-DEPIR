// encoder-profile-s13 profiles reusable setup and encoding, and retains a
// reloadable local JSON cache. Cache save/load are separate from encoding.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

var profileDirectory string

func emit(v any) {
	if err := json.NewEncoder(os.Stdout).Encode(v); err != nil {
		panic(err)
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Phase memory observations are sampled heap values, not isolated process RSS.
// The resource guard is intentionally uniform across stages and repeat calls.
func phase(name string, budget uint64, run func()) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	peak, peakSys := before.HeapAlloc, before.Sys
	var samples uint64
	stop, done := make(chan struct{}), make(chan struct{})
	var lock sync.Mutex
	sample := func() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		lock.Lock()
		defer lock.Unlock()
		samples++
		if m.HeapAlloc > peak {
			peak = m.HeapAlloc
		}
		if m.Sys > peakSys {
			peakSys = m.Sys
		}
		if m.Sys > budget {
			emit(map[string]any{"event": "memory_limit", "phase": name, "go_sys_bytes": m.Sys, "budget_bytes": budget})
			os.Exit(3)
		}
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sample()
			case <-stop:
				sample()
				return
			}
		}
	}()
	emit(map[string]any{"event": "started", "phase": name})
	profilePath := filepath.Join(profileDirectory, name+".pprof")
	file, err := os.Create(profilePath)
	check(err)
	check(pprof.StartCPUProfile(file))
	start := time.Now()
	run()
	elapsed := time.Since(start)
	pprof.StopCPUProfile()
	check(file.Close())
	close(stop)
	<-done
	runtime.ReadMemStats(&after)
	emit(map[string]any{"event": "phase", "phase": name, "elapsed_ns": elapsed.Nanoseconds(), "cpu_profile": profilePath, "baseline_heap_bytes": before.HeapAlloc, "sampled_peak_heap_bytes": peak, "sampled_peak_go_sys_bytes": peakSys, "allocated_bytes": after.TotalAlloc - before.TotalAlloc, "gc_cycles": after.NumGC - before.NumGC, "samples": samples, "sample_period_ns": 100000000})
}

func oracleMul(a, b uint32, s int, polynomial uint32) uint32 {
	var product uint64
	for i := 0; i < s; i++ {
		if b&(1<<uint(i)) != 0 {
			product ^= uint64(a) << uint(i)
		}
	}
	for i := 2*s - 2; i >= s; i-- {
		if product&(uint64(1)<<uint(i)) != 0 {
			product ^= uint64(polynomial) << uint(i-s)
		}
	}
	return uint32(product)
}

// Packed output streams through a bounded buffer. Hashing/I/O are outside encode
// time, and a cached readback checks the final file without another full copy.
func packedOutput(word []uint32, s int) (string, int64, time.Duration) {
	f, err := os.CreateTemp("", "depir-scaled-output-")
	check(err)
	defer os.Remove(f.Name())
	hash := sha256.New()
	writer := bufio.NewWriterSize(io.MultiWriter(f, hash), 1<<20)
	start := time.Now()
	var pending uint64
	bits := 0
	for _, x := range word {
		pending |= uint64(x) << uint(bits)
		bits += s
		for bits >= 8 {
			check(writer.WriteByte(byte(pending)))
			pending >>= 8
			bits -= 8
		}
	}
	if bits > 0 {
		check(writer.WriteByte(byte(pending)))
	}
	check(writer.Flush())
	check(f.Sync())
	check(f.Close())
	elapsed := time.Since(start)
	input, err := os.Open(f.Name())
	check(err)
	readback := sha256.New()
	n, err := io.Copy(readback, input)
	check(err)
	check(input.Close())
	expected := hex.EncodeToString(hash.Sum(nil))
	if hex.EncodeToString(readback.Sum(nil)) != expected {
		panic("packed readback mismatch")
	}
	return expected, n, elapsed
}

func main() {
	s := flag.Int("s", 13, "field degree")
	t := flag.Int("t", 5, "curve degree")
	repeats := flag.Int("repeats", 1, "database encodings")
	memory := flag.Uint64("memory-bytes", 12<<30, "maximum observed Go runtime memory")
	profiles := flag.String("profiles", "experiments/encoder-profile-s13", "CPU profile directory (profiling is enabled)")
	cache := flag.String("cache", ".cache/liftrs/s13-m2-t5", "retained local cache directory")
	reuse := flag.Bool("reuse-cache", false, "load the retained cache without repeating discovery/preparation")
	provenance := flag.String("source-manifest", "", "source provenance identifier recorded with the cache")
	flag.Parse()
	profileDirectory = *profiles
	check(os.MkdirAll(profileDirectory, 0755))
	check(os.MkdirAll(*cache, 0755))
	if *s < 2 || *s > 13 || *repeats < 1 || *memory < 1<<28 {
		panic("invalid benchmark bounds")
	}
	// Leave headroom below the hard sampled Go-runtime guard. This is a soft GC
	// target, not an OS address-space limit (Go reserves a large virtual arena).
	debug.SetMemoryLimit(int64(*memory * 3 / 4))
	spec := liftrs.Spec{FieldDegree: *s, Dimension: 2, CurveDegree: *t}
	var basis *liftrs.OrbitBasis
	var encoder *liftrs.CyclicEncoder
	const cacheLimit int64 = 2 << 30
	cachePath := filepath.Join(*cache, "cyclic.json")
	metadataPath := filepath.Join(*cache, "metadata.json")
	if !*reuse {
		if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
			panic("cache already exists; use -reuse-cache or a different directory")
		}
		phase("discovery", *memory, func() {
			var err error
			basis, err = liftrs.DiscoverOrbit(spec, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(int64(1000 + *s)))})
			check(err)
		})
		phase("preparation", *memory, func() { var err error; encoder, err = basis.PrepareCyclicFFT(); check(err) })
		phase("cache_save", *memory, func() {
			raw, err := encoder.MarshalBinary()
			check(err)
			if int64(len(raw)) > cacheLimit {
				panic("local cache exceeds 2 GiB retention budget")
			}
			// Save atomically and sync; retain on disk after the benchmark process exits.
			file, err := os.CreateTemp(*cache, "cyclic-*.tmp")
			check(err)
			temporary := file.Name()
			defer os.Remove(temporary)
			_, err = file.Write(raw)
			check(err)
			check(file.Sync())
			check(file.Close())
			check(os.Rename(temporary, cachePath))
			digest := sha256.Sum256(raw)
			metadata := map[string]any{"s": *s, "m": 2, "t": *t, "n": basis.Length(), "k": basis.Dimension(), "sha256": hex.EncodeToString(digest[:]), "bytes": len(raw), "source_manifest": *provenance, "format": "cyclic-json-v1", "cache_limit_bytes": cacheLimit}
			encoded, err := json.MarshalIndent(metadata, "", "  ")
			check(err)
			check(os.WriteFile(metadataPath, append(encoded, '\n'), 0644))
			emit(map[string]any{"event": "cache_saved", "path": cachePath, "metadata": metadata})
		})
	}
	// Drop the original preparation before loading. The following encoding must
	// genuinely use the retained file, not an accidentally live in-memory encoder.
	encoder, basis = nil, nil
	phase("cache_load", *memory, func() {
		rawMetadata, err := os.ReadFile(metadataPath)
		check(err)
		var metadata struct {
			S, M, T int
			N, K    int
			SHA256  string
			Bytes   int64
		}
		check(json.Unmarshal(rawMetadata, &metadata))
		if metadata.S != *s || metadata.M != 2 || metadata.T != *t || metadata.Bytes <= 0 || metadata.Bytes > cacheLimit {
			panic("cache metadata/spec mismatch")
		}
		file, err := os.Open(cachePath)
		check(err)
		raw, err := io.ReadAll(io.LimitReader(file, cacheLimit+1))
		check(err)
		check(file.Close())
		if int64(len(raw)) != metadata.Bytes {
			panic("cache byte size mismatch")
		}
		digest := sha256.Sum256(raw)
		if hex.EncodeToString(digest[:]) != metadata.SHA256 {
			panic("cache checksum mismatch")
		}
		encoder, err = liftrs.LoadCyclicEncoderWithLimit(raw, cacheLimit)
		check(err)
		basis = encoder.Basis()
		if basis.Length() != metadata.N || basis.Dimension() != metadata.K {
			panic("loaded cache dimensions differ")
		}
		emit(map[string]any{"event": "cache_loaded", "path": cachePath, "bytes": len(raw), "sha256": metadata.SHA256})
	})
	emit(map[string]any{"event": "code", "s": *s, "m": 2, "t": *t, "n": basis.Length(), "k": basis.Dimension(), "go_version": runtime.Version(), "reused_cache": *reuse})

	random := rand.New(rand.NewSource(71))
	message := make([]uint32, basis.Dimension())
	q := uint32(1) << uint(*s)
	for i := range message {
		message[i] = uint32(random.Intn(int(q)))
	}
	var word []uint32
	for repeat := 0; repeat < *repeats; repeat++ {
		word = nil
		phase(fmt.Sprintf("encode_%d", repeat), *memory, func() { var err error; word, err = encoder.Encode(message); check(err) })
	}
	// The first information-set coordinate is omega. Iterating it checks every
	// position in linear time; sample the public exponentiation mapper as a cross-check.
	start := time.Now()
	polynomial, err := liftrs.BinaryPolynomial(*s)
	check(err)
	base, err := liftrs.NewBinaryField(*s, polynomial)
	check(err)
	tower, err := liftrs.NewQuadraticField(base)
	check(err)
	origin, err := basis.SystematicPosition(0)
	check(err)
	omega := liftrs.QuadraticElement{A: uint32(origin) & (q - 1), B: uint32(origin) >> uint(*s)}
	point := omega
	for j, x := range message {
		position := int(point.A | point.B<<uint(*s))
		if position >= len(word) || word[position] != x {
			panic("systematic mismatch")
		}
		if j == 0 || j == len(message)/2 || j == len(message)-1 {
			direct, err := basis.SystematicPosition(j)
			check(err)
			if direct != position {
				panic("coordinate mapper mismatch")
			}
		}
		point = tower.Mul(point, omega)
	}
	for trial := 0; trial < 16; trial++ {
		curve := [2][]uint32{make([]uint32, *t+1), make([]uint32, *t+1)}
		for axis := range curve {
			for j := range curve[axis] {
				curve[axis][j] = uint32(random.Intn(int(q)))
			}
		}
		var sum uint32
		for z := uint32(0); z < q; z++ {
			var p [2]uint32
			for axis := range curve {
				for j := *t; j >= 0; j-- {
					p[axis] = oracleMul(p[axis], z, *s, polynomial) ^ curve[axis][j]
				}
			}
			sum ^= word[p[0]+q*p[1]]
		}
		if sum != 0 {
			panic("independent curve sum mismatch")
		}
	}
	verification := time.Since(start)
	// Hash canonical uint32 output through a fixed buffer rather than a 4*N copy.
	canonical := sha256.New()
	buffer := make([]byte, 64<<10)
	for i := 0; i < len(word); {
		count := len(buffer) / 4
		if len(word)-i < count {
			count = len(word) - i
		}
		for j := 0; j < count; j++ {
			binary.LittleEndian.PutUint32(buffer[4*j:], word[i+j])
		}
		_, err := canonical.Write(buffer[:4*count])
		check(err)
		i += count
	}
	packed, n, ioTime := packedOutput(word, *s)
	emit(map[string]any{"event": "verified", "s": *s, "m": 2, "t": *t, "k": basis.Dimension(), "n": basis.Length(), "seed": 71, "systematic_positions_checked": len(message), "curve_checks": 16, "verification_ns": verification.Nanoseconds(), "canonical_sha256": hex.EncodeToString(canonical.Sum(nil)), "packed_sha256": packed, "packed_bytes": n, "packing_write_sync_ns": ioTime.Nanoseconds()})
	runtime.KeepAlive(encoder)
	runtime.KeepAlive(message)
}
