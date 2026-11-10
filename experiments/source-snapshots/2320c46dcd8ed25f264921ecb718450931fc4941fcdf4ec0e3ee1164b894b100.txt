// encoder-comparison separates code discovery/preparation, artifact loading,
// database-dependent encoding, and canonical uint32 output file I/O.
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

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
)

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func emit(v any) { fail(json.NewEncoder(os.Stdout).Encode(v)) }
func main() {
	mode := flag.String("mode", "prepare", "prepare, paper or cyclic")
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
	if *mode == "prepare" {
		start := time.Now()
		basis, err := liftrs.DiscoverOrbit(spec, liftrs.OrbitOptions{Random: rand.New(rand.NewSource(int64(1000 + *s)))})
		fail(err)
		discovery := time.Since(start).Nanoseconds()
		raw, err := basis.MarshalBinary()
		fail(err)
		fail(os.WriteFile(filepath.Join(*directory, stem+"-basis.json"), raw, 0644))
		start = time.Now()
		paper, err := basis.PreparePaper()
		fail(err)
		paperNS := time.Since(start).Nanoseconds()
		start = time.Now()
		pa, err := paper.MarshalBinary()
		fail(err)
		paperSave := time.Since(start).Nanoseconds()
		fail(os.WriteFile(filepath.Join(*directory, stem+"-paper.json"), pa, 0644))
		start = time.Now()
		cyclic, err := basis.PrepareCyclic()
		fail(err)
		cyclicNS := time.Since(start).Nanoseconds()
		start = time.Now()
		ca, err := cyclic.MarshalBinary()
		fail(err)
		cyclicSave := time.Since(start).Nanoseconds()
		fail(os.WriteFile(filepath.Join(*directory, stem+"-cyclic.json"), ca, 0644))
		start = time.Now()
		fft, err := basis.PrepareCyclicFFT()
		fail(err)
		fftNS := time.Since(start).Nanoseconds()
		start = time.Now()
		fa, err := fft.MarshalBinary()
		fail(err)
		fftSave := time.Since(start).Nanoseconds()
		fail(os.WriteFile(filepath.Join(*directory, stem+"-cyclic-fft.json"), fa, 0644))
		emit(map[string]any{"s": *s, "m": 2, "t": *t, "n": basis.Length(), "k": basis.Dimension(), "samples": basis.Samples(), "seed": 1000 + *s, "discovery_ns": discovery, "paper_prepare_ns": paperNS, "cyclic_prepare_ns": cyclicNS, "paper_serialize_ns": paperSave, "cyclic_serialize_ns": cyclicSave, "basis_artifact_bytes": len(raw), "paper_artifact_bytes": len(pa), "cyclic_artifact_bytes": len(ca), "cyclic_fft_prepare_ns": fftNS, "cyclic_fft_serialize_ns": fftSave, "cyclic_fft_artifact_bytes": len(fa), "go_version": runtime.Version()})
		return
	}
	if *mode != "paper" && *mode != "cyclic" && *mode != "cyclic-fft" {
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
	// This is an explicit uint32 experimental sink, not the packed production
	// format. File sync is included; cached readback is reported separately.
	buffer := make([]byte, 4*len(word))
	for i, x := range word {
		binary.LittleEndian.PutUint32(buffer[4*i:], x)
	}
	digest := sha256.Sum256(buffer)
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
	if hex.EncodeToString(hash.Sum(nil)) != hex.EncodeToString(digest[:]) {
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
	emit(map[string]any{"method": *mode, "artifact_read_ns": readNS, "artifact_load_ns": loadNS, "encode_ns": timings, "encode_allocated_bytes": allocations, "output_bytes": len(buffer), "output_sha256": hex.EncodeToString(digest[:]), "write_sync_ns": writeNS, "cached_readback_ns": readbackNS, "go_version": runtime.Version()})
}
