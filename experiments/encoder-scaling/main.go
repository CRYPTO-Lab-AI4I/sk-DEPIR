// encoder-scaling verifies the moderate-instance checkpoint and samples encoding
// heap separately from uninstrumented timings, preparation and output I/O.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func emit(v any) { check(json.NewEncoder(os.Stdout).Encode(v)) }

// oracleMul multiplies carryless integers then performs polynomial long division,
// independently of liftrs.BinaryField's shift/reduce and extension arithmetic.
func oracleMul(a, b uint32) uint32 {
	var p uint64
	for bit := uint(0); bit < 10; bit++ {
		if b&(1<<bit) != 0 {
			p ^= uint64(a) << bit
		}
	}
	for bit := 19; bit >= 10; bit-- {
		if p&(uint64(1)<<uint(bit)) != 0 {
			p ^= uint64(0x409) << uint(bit-10)
		}
	}
	return uint32(p)
}

func wordHash(word []uint32) string {
	raw := make([]byte, 4*len(word))
	for i, x := range word {
		binary.LittleEndian.PutUint32(raw[4*i:], x)
	}
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func load(directory, method string) (*liftrs.OrbitBasis, func([]uint32) ([]uint32, error)) {
	data, err := os.ReadFile(filepath.Join(directory, "s10-m2-t6-"+method+".json"))
	check(err)
	if method == "paper" {
		e, err := liftrs.LoadPaperEncoder(data)
		check(err)
		return e.Basis(), e.Encode
	}
	if method != "cyclic-fft" {
		panic("invalid method")
	}
	e, err := liftrs.LoadCyclicEncoder(data)
	check(err)
	return e.Basis(), e.Encode
}

func main() {
	mode := flag.String("mode", "verify", "verify or memory")
	method := flag.String("method", "cyclic-fft", "paper or cyclic-fft for memory sampling")
	directory := flag.String("artifacts", "experiments/encoder-scaling/artifacts", "checkpoint directory")
	flag.Parse()
	basis, encode := load(*directory, *method)
	metadata, err := basis.MarshalBinary()
	check(err)
	var description struct {
		Spec liftrs.Spec `json:"spec"`
	}
	check(json.Unmarshal(metadata, &description))
	if description.Spec != (liftrs.Spec{FieldDegree: 10, Dimension: 2, CurveDegree: 6}) {
		panic("unexpected specification")
	}
	if basis.Length() != 1048576 {
		panic("unexpected encoded length")
	}
	random := rand.New(rand.NewSource(7310))
	message := make([]uint32, basis.Dimension())
	for i := range message {
		message[i] = uint32(random.Intn(1024))
	}
	if *mode == "memory" {
		// Only cached encoder/message remain live at baseline. Sampling perturbs this
		// call; it is never used as the uninstrumented latency measurement.
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		stop, done := make(chan struct{}), make(chan struct{})
		peak, peakInuse, count := before.HeapAlloc, before.HeapInuse, uint64(0)
		go func() {
			defer close(done)
			ticker := time.NewTicker(2 * time.Millisecond)
			defer ticker.Stop()
			sample := func() {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				if m.HeapAlloc > peak {
					peak = m.HeapAlloc
				}
				if m.HeapInuse > peakInuse {
					peakInuse = m.HeapInuse
				}
				count++
			}
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
		start := time.Now()
		word, err := encode(message)
		check(err)
		elapsed := time.Since(start)
		close(stop)
		<-done
		runtime.ReadMemStats(&after)
		runtime.KeepAlive(word)
		runtime.KeepAlive(encode)
		emit(map[string]any{"method": *method, "k": basis.Dimension(), "n": basis.Length(), "seed": 7310, "baseline_heap_alloc_bytes": before.HeapAlloc, "sampled_peak_heap_alloc_bytes": peak, "sampled_incremental_heap_alloc_bytes": peak - before.HeapAlloc, "sampled_peak_heap_inuse_bytes": peakInuse, "sample_period_ns": 2000000, "samples": count, "output_unpacked_bytes": 4 * len(word), "allocated_bytes": after.TotalAlloc - before.TotalAlloc, "gc_cycles": after.NumGC - before.NumGC, "instrumented_encode_ns": elapsed.Nanoseconds(), "sha256": wordHash(word), "go_version": runtime.Version()})
		return
	}
	if *mode != "verify" {
		panic("invalid mode")
	}
	word, err := encode(message)
	check(err)
	otherMethod := "paper"
	if *method == "paper" {
		otherMethod = "cyclic-fft"
	}
	otherBasis, otherEncode := load(*directory, otherMethod)
	if otherBasis.Dimension() != basis.Dimension() {
		panic("dimension disagreement")
	}
	otherWord, err := otherEncode(message)
	check(err)
	if !reflect.DeepEqual(word, otherWord) {
		panic("paper/cyclic mismatch")
	}
	seen := make([]bool, basis.Length())
	for j, x := range message {
		pos, err := basis.SystematicPosition(j)
		check(err)
		if pos < 0 || pos >= len(word) || seen[pos] || word[pos] != x {
			panic("invalid information set")
		}
		seen[pos] = true
	}
	// A base-field curve has 1024 points. Reduced degree <=1022 implies that its
	// value sum is zero. This samples actual degree-at-most-six curves, without
	// relying on orbit discovery or the FFT implementation to evaluate them.
	for trial := 0; trial < 64; trial++ {
		var curve [2][7]uint32
		for axis := range curve {
			for j := range curve[axis] {
				curve[axis][j] = uint32(random.Intn(1024))
			}
		}
		var sum uint32
		for z := uint32(0); z < 1024; z++ {
			var point [2]uint32
			for axis := range curve {
				for j := 6; j >= 0; j-- {
					point[axis] = oracleMul(point[axis], z) ^ curve[axis][j]
				}
			}
			sum ^= word[point[0]+1024*point[1]]
		}
		if sum != 0 {
			panic(fmt.Sprintf("failed sampled curve %d", trial))
		}
	}
	emit(map[string]any{"s": 10, "m": 2, "t": 6, "k": basis.Dimension(), "n": basis.Length(), "seed": 7310, "systematic_positions_checked": len(message), "curve_checks": 64, "curve_points": 1024, "sha256": wordHash(word), "paper_cyclic_equal": true, "go_version": runtime.Version()})
}
