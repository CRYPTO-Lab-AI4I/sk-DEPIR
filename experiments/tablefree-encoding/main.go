// tablefree-encoding separates code discovery/preparation, artifact loading,
// database-dependent encoding, and packed canonical output file I/O.
package main

import (
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
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func emit(v any) { fail(json.NewEncoder(os.Stdout).Encode(v)) }
func main() {
	mode := flag.String("mode", "discover", "discover, prepare-paper, prepare-cyclic-fft, paper or cyclic-fft")
	s := flag.Int("s", 4, "field degree")
	t := flag.Int("t", 2, "curve degree")
	directory := flag.String("artifacts", "", "artifact directory")
	repeats := flag.Int("repeats", 3, "database encodings")
	flag.Parse()
	if *directory == "" || *repeats < 1 {
		fail(fmt.Errorf("provide -artifacts directory and positive -repeats"))
	}
	spec := liftrs.Spec{FieldDegree: *s, Dimension: 2, CurveDegree: *t}
	stem := fmt.Sprintf("s%02d-m2-t%d", *s, *t)

	if *mode == "discover" {
		start := time.Now()
		basis, err := liftrs.DiscoverOrbit(spec, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(int64(1000 + *s)))})
		fail(err)
		discovery := time.Since(start).Nanoseconds()
		start = time.Now()
		raw, err := basis.MarshalBinary()
		fail(err)
		serialize := time.Since(start).Nanoseconds()
		fail(os.WriteFile(filepath.Join(*directory, stem+"-basis.json"), raw, 0644))
		emit(map[string]any{"s": *s, "m": 2, "t": *t, "n": basis.Length(), "k": basis.Dimension(), "samples": basis.Samples(), "seed": 1000 + *s, "discovery_ns": discovery, "serialize_ns": serialize, "artifact_bytes": len(raw), "go_version": runtime.Version()})
		return
	}
	if *mode == "prepare-paper" || *mode == "prepare-cyclic-fft" {
		start := time.Now()
		data, err := os.ReadFile(filepath.Join(*directory, stem+"-basis.json"))
		fail(err)
		readNS := time.Since(start).Nanoseconds()
		start = time.Now()
		basis, err := liftrs.LoadOrbit(data)
		fail(err)
		loadNS := time.Since(start).Nanoseconds()
		var marshal func() ([]byte, error)
		start = time.Now()
		if *mode == "prepare-paper" {
			encoder, err := basis.PreparePaper()
			fail(err)
			marshal = encoder.MarshalBinary
		} else {
			encoder, err := basis.PrepareCyclicFFT()
			fail(err)
			marshal = encoder.MarshalBinary
		}
		prepare := time.Since(start).Nanoseconds()
		start = time.Now()
		raw, err := marshal()
		fail(err)
		serialize := time.Since(start).Nanoseconds()
		name := (*mode)[len("prepare-"):]
		fail(os.WriteFile(filepath.Join(*directory, stem+"-"+name+".json"), raw, 0644))
		emit(map[string]any{"method": name, "basis_read_ns": readNS, "basis_load_ns": loadNS, "prepare_ns": prepare, "serialize_ns": serialize, "artifact_bytes": len(raw), "go_version": runtime.Version()})
		return
	}
	if *mode != "paper" && *mode != "cyclic-fft" {
		fail(fmt.Errorf("invalid mode"))
	}
	start := time.Now()
	data, err := os.ReadFile(filepath.Join(*directory, stem+"-"+*mode+".json"))
	fail(err)
	readNS := time.Since(start).Nanoseconds()
	var encode func([]uint32) ([]uint32, error)
	var basis *liftrs.OrbitBasis
	start = time.Now()
	if *mode == "paper" {
		e, err := liftrs.LoadPaperEncoder(data)
		fail(err)
		encode = e.Encode
		basis = e.Basis()
	} else {
		e, err := liftrs.LoadCyclicEncoder(data)
		fail(err)
		encode = e.Encode
		basis = e.Basis()
	}
	loadNS := time.Since(start).Nanoseconds()
	message := make([]uint32, basis.Dimension())
	r := rand.New(rand.NewSource(71))
	for i := range message {
		message[i] = uint32(r.Intn(1 << uint(*s)))
	}
	runtime.GC()
	var timings, allocations []uint64
	var word []uint32
	for i := 0; i < *repeats; i++ {
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		start = time.Now()
		word, err = encode(message)
		fail(err)
		timings = append(timings, uint64(time.Since(start).Nanoseconds()))
		runtime.ReadMemStats(&after)
		allocations = append(allocations, after.TotalAlloc-before.TotalAlloc)
	}
	// Retain canonical uint32 hash for direct historical compatibility checks.
	rawWord := make([]byte, 4*len(word))
	for i, x := range word {
		binary.LittleEndian.PutUint32(rawWord[4*i:], x)
	}
	digest := sha256.Sum256(rawWord)
	start = time.Now()
	buffer := packWord(word, *s)
	packNS := time.Since(start).Nanoseconds()
	packedDigest := sha256.Sum256(buffer)
	file, err := os.CreateTemp("", "sk-depir-encoder-output-")
	fail(err)
	defer os.Remove(file.Name())
	start = time.Now()
	_, err = file.Write(buffer)
	fail(err)
	fail(file.Sync())
	fail(file.Close())
	writeNS := time.Since(start).Nanoseconds()
	start = time.Now()
	file, err = os.Open(file.Name())
	fail(err)
	hash := sha256.New()
	_, err = io.Copy(hash, file)
	fail(err)
	fail(file.Close())
	readbackNS := time.Since(start).Nanoseconds()
	if hex.EncodeToString(hash.Sum(nil)) != hex.EncodeToString(packedDigest[:]) {
		fail(fmt.Errorf("output readback mismatch"))
	}
	for j, x := range message {
		position, err := basis.SystematicPosition(j)
		fail(err)
		if word[position] != x {
			fail(fmt.Errorf("not systematic"))
		}
	}
	runtime.KeepAlive(encode)
	emit(map[string]any{"method": *mode, "artifact_read_ns": readNS, "artifact_load_ns": loadNS, "encode_ns": timings, "encode_allocated_bytes": allocations, "output_bytes": len(buffer), "packing_ns": packNS, "packed_sha256": hex.EncodeToString(packedDigest[:]), "output_sha256": hex.EncodeToString(digest[:]), "write_sync_ns": writeNS, "cached_readback_ns": readbackNS, "go_version": runtime.Version()})
}

// packWord follows the contract: element i, bit b occupies file bit i*s+b.
// The accumulator holds fewer than 8+s bits; public encoder support is s<=10.
func packWord(word []uint32, s int) []byte {
	out := make([]byte, (len(word)*s+7)/8)
	var pending uint64
	bits, index := 0, 0
	for _, x := range word {
		pending |= uint64(x) << uint(bits)
		bits += s
		for bits >= 8 {
			out[index] = byte(pending)
			index++
			pending >>= 8
			bits -= 8
		}
	}
	if bits > 0 {
		out[index] = byte(pending)
	}
	return out
}
