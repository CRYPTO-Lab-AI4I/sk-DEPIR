//go:build liftrs_cuda && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestCUDAHostBacked(t *testing.T) {
	out := os.Getenv("PIR_GPU_BACKED_OUT")
	if out == "" {
		t.Skip("opt-in host-backed GPU transforms")
	}
	cache := os.Getenv("PIR_GPU_BACKED_CACHE")
	if cache == "" {
		t.Fatal("fixture cache required")
	}
	cases := []struct{ s, dimension, tile int }{{4, 7, 48}, {4, 8, 48}, {6, 11, 320}, {6, 12, 320}, {8, 15, 1536}, {8, 16, 1536}, {16, 20, 65536}, {16, 22, 65536}}
	if os.Getenv("PIR_GPU_BACKED_SMALL") == "1" {
		cases = cases[:6]
	}
	var records []map[string]any
	random := rand.New(rand.NewSource(3901))
	digest := func(values []uint32) string {
		hash := sha256.New()
		var block [65536]byte
		for first := 0; first < len(values); {
			count := minInt(len(block)/4, len(values)-first)
			for i := 0; i < count; i++ {
				binary.LittleEndian.PutUint32(block[4*i:], values[first+i])
			}
			hash.Write(block[:4*count])
			first += count
		}
		return hex.EncodeToString(hash.Sum(nil))
	}
	save := func(name string, values []uint32) {
		file, err := os.Create(filepath.Join(cache, name))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		var block [65536]byte
		for first := 0; first < len(values); {
			count := minInt(len(block)/4, len(values)-first)
			for i := 0; i < count; i++ {
				binary.LittleEndian.PutUint32(block[4*i:], values[first+i])
			}
			if _, err = file.Write(block[:4*count]); err != nil {
				t.Fatal(err)
			}
			first += count
		}
	}
	for _, test := range cases {
		polynomial, _ := BinaryPolynomial(test.s)
		_, field, delta, _ := orbitFields(test.s, polynomial)
		table := newPreparationProducts(field.base)
		words, err := cudaMatrixDescription(field, test.dimension)
		if err != nil {
			t.Fatal(err)
		}
		n := 1 << uint(test.dimension)
		input := make([]uint32, n)
		for i := range input {
			input[i] = random.Uint32() & uint32(field.q-1)
		}
		want := append([]uint32(nil), input...)
		start := time.Now()
		newSparseTransform(field).applySubspace(want, false)
		goSeconds := time.Since(start).Seconds()
		save(fmt.Sprintf("s%d-d%d-input.bin", test.s, test.dimension), input)
		save(fmt.Sprintf("s%d-d%d-forward.bin", test.s, test.dimension), want)
		resident, err := cuda.NewTransformBounded(os.Getenv("PIR_GPU_LIBRARY"), test.s, polynomial, delta, table.log, table.powers, n, 512<<20, false)
		if err != nil {
			t.Fatal(err)
		}
		if err = resident.AddPlan(words); err != nil {
			t.Fatal(err)
		}
		residentData := append([]uint32(nil), input...)
		residentStart := time.Now()
		residentStats, err := resident.Apply(residentData, false)
		residentSeconds := time.Since(residentStart).Seconds()
		if err != nil || !reflect.DeepEqual(residentData, want) {
			t.Fatal("resident/reference", test, err)
		}
		residentOwned := resident.OwnedBytes()
		resident.Close()
		residentData = nil
		backed, err := cuda.NewHostBacked(os.Getenv("PIR_GPU_LIBRARY"), test.s, polynomial, delta, table.log, table.powers, n, 512<<20, test.tile)
		if err != nil {
			t.Fatal(err)
		}
		defer backed.Close()
		if err = backed.AddPlan(words); err != nil {
			t.Fatal(err)
		}
		owned := backed.OwnedBytes()
		if residentOwned-owned != 8*uint64(n-minInt(n, test.tile)) {
			t.Fatal("exact backed allocation", test, residentOwned, owned)
		}
		data := append([]uint32(nil), input...)
		start = time.Now()
		forward, err := backed.ApplyBacked(data, false)
		forwardSeconds := time.Since(start).Seconds()
		if err != nil || !reflect.DeepEqual(data, want) {
			t.Fatal("host-backed forward/reference", test, err)
		}
		start = time.Now()
		inverse, err := backed.ApplyBacked(data, true)
		inverseSeconds := time.Since(start).Seconds()
		if err != nil || !reflect.DeepEqual(data, input) {
			t.Fatal("host-backed inverse", test, err)
		}
		for _, st := range [][12]float64{forward, inverse} {
			if st[8] < float64(8*n) || st[9] < float64(8*n) || st[10] > float64(test.tile*4) || st[7] < 0 {
				t.Fatal("backed traffic/staging", test, st)
			}
		}
		if forward[8] != inverse[8] || forward[9] != inverse[9] || backed.OwnedBytes() != owned {
			t.Fatal("stable backed workspace/transfers", test)
		}
		if _, err = backed.Apply(data, false); err == nil {
			t.Fatal("resident Apply accepted backed workspace")
		}
		if _, _, err = backed.Product([]uint32{1}, []uint32{1}, 1, n); err == nil {
			t.Fatal("backed Product accepted")
		}
		if _, err = backed.ApplyBacked(data[:len(data)-1], false); err == nil {
			t.Fatal("unprepared shape")
		}
		if test.s < 16 {
			original := data[0]
			data[0] = uint32(field.q)
			_, err = backed.ApplyBacked(data, false)
			data[0] = original
			if err == nil {
				t.Fatal("backed field range")
			}
		}
		if _, err = backed.ApplyBacked(data, false); err != nil || !reflect.DeepEqual(data, want) {
			t.Fatal("backed recovery", err)
		}
		backed.Close()
		if backed.OwnedBytes() != 0 {
			t.Fatal("backed close leak")
		}
		if _, err = backed.ApplyBacked(data, false); err == nil {
			t.Fatal("backed closed accepted")
		}
		records = append(records, map[string]any{"s": test.s, "dimension": test.dimension, "n": n, "tile_elements": test.tile, "matches": 3, "resident_owned_bytes": residentOwned, "backed_owned_bytes": owned, "go_seconds": goSeconds, "resident_seconds": residentSeconds, "forward_seconds": forwardSeconds, "inverse_seconds": inverseSeconds, "resident_stages": residentStats, "forward_stages": forward, "inverse_stages": inverse, "input_sha256": digest(input), "forward_sha256": digest(want), "closed": true, "recovery": true})
	}
	// Full s16 metadata, bounded allocation only: no full host data array/run.
	polynomial, _ := BinaryPolynomial(16)
	_, field, delta, _ := orbitFields(16, polynomial)
	table := newPreparationProducts(field.base)
	words, err := cudaMatrixDescription(field, 32)
	if err != nil {
		t.Fatal(err)
	}
	full, err := cuda.NewHostBacked(os.Getenv("PIR_GPU_LIBRARY"), 16, polynomial, delta, table.log, table.powers, 1<<32, 64<<20, 65536)
	if err != nil {
		t.Fatal(err)
	}
	if err = full.AddPlan(words); err != nil {
		t.Fatal("full metadata bounded", err)
	}
	if full.OwnedBytes() > 2<<20 {
		t.Fatal("metadata allocated full domain", full.OwnedBytes())
	}
	full.Close()
	small, err := cuda.NewHostBacked(os.Getenv("PIR_GPU_LIBRARY"), 16, polynomial, delta, table.log, table.powers, 1<<32, 64<<20, 32768)
	if err != nil {
		t.Fatal(err)
	}
	if err = small.AddPlan(words); err == nil {
		small.Close()
		t.Fatal("short full vector tile")
	}
	small.Close()
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "backed-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
