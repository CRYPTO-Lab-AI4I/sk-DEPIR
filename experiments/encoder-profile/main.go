// encoder-profile samples one reusable-preparation or database-encoding stage.
// Loading, deterministic message generation, serialization and verification are
// outside CPU sampling. Profiles include allocation and GC during that stage.
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
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	stage := flag.String("stage", "encode", "prepare or encode")
	method := flag.String("method", "cyclic-fft", "paper or cyclic-fft")
	duration := flag.Duration("duration", 4*time.Second, "minimum sampled duration; complete the final call")
	artifacts := flag.String("artifacts", "experiments/tablefree-encoding/artifacts", "prepared checkpoint directory")
	profile := flag.String("profile", "cpu.pprof", "CPU profile output")
	flag.Parse()
	if (*stage != "prepare" && *stage != "encode") || (*method != "paper" && *method != "cyclic-fft") || *duration <= 0 {
		panic("invalid stage, method or duration")
	}
	stem := filepath.Join(*artifacts, "s09-m2-t7-")
	var run func()
	var marshal func() ([]byte, error)
	var basis *liftrs.OrbitBasis
	var word []uint32
	var message []uint32
	if *stage == "prepare" {
		data, err := os.ReadFile(stem + "basis.json")
		check(err)
		basis, err = liftrs.LoadOrbit(data)
		check(err)
		if *method == "paper" {
			run = func() { e, err := basis.PreparePaper(); check(err); marshal = e.MarshalBinary }
		} else {
			run = func() { e, err := basis.PrepareCyclicFFT(); check(err); marshal = e.MarshalBinary }
		}
	} else {
		data, err := os.ReadFile(stem + *method + ".json")
		check(err)
		var encode func([]uint32) ([]uint32, error)
		if *method == "paper" {
			e, err := liftrs.LoadPaperEncoder(data)
			check(err)
			encode = e.Encode
			basis = e.Basis()
		} else {
			e, err := liftrs.LoadCyclicEncoder(data)
			check(err)
			encode = e.Encode
			basis = e.Basis()
		}
		message = make([]uint32, basis.Dimension())
		random := rand.New(rand.NewSource(71))
		for i := range message {
			message[i] = uint32(random.Intn(512))
		}
		run = func() { var err error; word, err = encode(message); check(err) }
	}
	runtime.GC()
	file, err := os.Create(*profile)
	check(err)
	check(pprof.StartCPUProfile(file))
	start := time.Now()
	calls := 0
	for calls == 0 || time.Since(start) < *duration {
		run()
		calls++
	}
	elapsed := time.Since(start)
	pprof.StopCPUProfile()
	check(file.Close())
	// Check final output against the checkpoint outside the sampled stage.
	var raw []byte
	if *stage == "prepare" {
		raw, err = marshal()
		check(err)
		expected, err := os.ReadFile(stem + *method + ".json")
		check(err)
		if sha256.Sum256(raw) != sha256.Sum256(expected) {
			panic("prepared artifact differs from checkpoint")
		}
	} else {
		raw = make([]byte, len(word)*4)
		for i, x := range word {
			binary.LittleEndian.PutUint32(raw[4*i:], x)
		}
		for j, x := range message {
			position, err := basis.SystematicPosition(j)
			check(err)
			if word[position] != x {
				panic("not systematic")
			}
		}
	}
	digest := sha256.Sum256(raw)
	check(json.NewEncoder(os.Stdout).Encode(map[string]any{"stage": *stage, "method": *method, "s": 9, "m": 2, "t": 7, "n": basis.Length(), "k": basis.Dimension(), "seed": 71, "calls": calls, "elapsed_ns": elapsed.Nanoseconds(), "sha256": hex.EncodeToString(digest[:]), "go_version": runtime.Version(), "gomaxprocs": runtime.GOMAXPROCS(0)}))
	fmt.Fprintln(os.Stderr, "CPU profile:", *profile)
}
