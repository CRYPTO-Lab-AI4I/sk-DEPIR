//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestCPPDiskTransform(t *testing.T) {
	out, cache, reference := os.Getenv("PIR_CPU_DISK_OUT"), os.Getenv("PIR_CPU_DISK_CACHE"), os.Getenv("PIR_CPU_DISK_REFERENCE")
	if out == "" {
		t.Skip("opt-in bounded RAM disk transforms")
	}
	cases := []struct{ s, dimension, tile int }{{4, 7, 48}, {4, 8, 48}, {6, 11, 320}, {6, 12, 320}, {8, 15, 1536}, {8, 16, 1536}, {16, 20, 65536}, {16, 22, 65536}}
	if os.Getenv("PIR_CPU_DISK_SMALL") == "1" {
		cases = cases[:6]
	}
	readWords := func(path string) []uint32 {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		values := make([]uint32, len(raw)/4)
		for i := range values {
			values[i] = binary.LittleEndian.Uint32(raw[4*i:])
		}
		return values
	}
	var records []map[string]any
	for _, test := range cases {
		stem := fmt.Sprintf("s%d-d%d", test.s, test.dimension)
		input := readWords(filepath.Join(reference, stem+"-input.bin"))
		want := readWords(filepath.Join(reference, stem+"-forward.bin"))
		n := 1 << uint(test.dimension)
		if len(input) != n || len(want) != n {
			t.Fatal("fixture shape")
		}
		polynomial, _ := BinaryPolynomial(test.s)
		_, field, _, _ := orbitFields(test.s, polynomial)
		plan, err := newMatrixCPP(field, test.dimension)
		if err != nil {
			t.Fatal(err)
		}
		resident := append([]uint32(nil), input...)
		scratch := make([]uint32, n)
		start := time.Now()
		var native cpp.MatrixStats
		if err = cpp.ApplyMatrixWithStats(plan.rows, plan.columns, resident, scratch, plan.split, 16, false, &native); err != nil || !reflect.DeepEqual(resident, want) {
			t.Fatal("resident/oracle", err)
		}
		residentSeconds := time.Since(start).Seconds()
		resident = nil
		scratch = nil
		files := []*os.File{}
		guard := bytes.Repeat([]byte{0x5a}, 16)
		for _, suffix := range []string{"data.bin", "scratch.bin"} {
			f, err := os.Create(filepath.Join(cache, stem+"-"+suffix))
			if err != nil {
				t.Fatal(err)
			}
			if err = f.Truncate(int64(n)*4 + 32); err != nil {
				t.Fatal(err)
			}
			f.WriteAt(guard, 0)
			f.WriteAt(guard, int64(n)*4+16)
			files = append(files, f)
		}
		data, temp := files[0], files[1]
		saveInput := func() {
			var block [65536]byte
			for first := 0; first < n; {
				count := minInt(n-first, len(block)/4)
				for i := 0; i < count; i++ {
					binary.LittleEndian.PutUint32(block[4*i:], input[first+i])
				}
				if _, err = data.WriteAt(block[:4*count], 16+int64(first)*4); err != nil {
					t.Fatal(err)
				}
				first += count
			}
		}
		readData := func() []uint32 {
			raw := make([]byte, 4*n)
			if _, err = data.ReadAt(raw, 16); err != nil {
				t.Fatal(err)
			}
			result := make([]uint32, n)
			for i := range result {
				result[i] = binary.LittleEndian.Uint32(raw[4*i:])
			}
			return result
		}
		call := func(inverse bool) (cpp.DiskStats, error) {
			st, err := cpp.ApplyMatrixFile(plan.rows, plan.columns, int(data.Fd()), int(temp.Fd()), 16, 16, n, test.tile, plan.split, 16, inverse)
			runtime.KeepAlive(data)
			runtime.KeepAlive(temp)
			return st, err
		}
		saveInput()
		start = time.Now()
		forward, err := call(false)
		forwardSeconds := time.Since(start).Seconds()
		if err != nil || !reflect.DeepEqual(readData(), want) {
			t.Fatal("disk forward/oracle", test, err)
		}
		start = time.Now()
		inverse, err := call(true)
		inverseSeconds := time.Since(start).Seconds()
		if err != nil || !reflect.DeepEqual(readData(), input) {
			t.Fatal("disk inverse", test, err)
		}
		for _, st := range []cpp.DiskStats{forward, inverse} {
			if st.BufferBytes > 8*uint64(test.tile)+65536 || st.ReadBytes < uint64(16*n) || st.WriteBytes < uint64(16*n) || st.Transposes != 2 {
				t.Fatal("disk counters", st)
			}
		}
		if forward.ReadBytes != inverse.ReadBytes || forward.WriteBytes != inverse.WriteBytes {
			t.Fatal("direction-independent transfers")
		}
		if _, err = cpp.ApplyMatrixFile(plan.rows, plan.columns, int(data.Fd()), int(data.Fd()), 16, 16, n, test.tile, plan.split, 16, false); err == nil {
			t.Fatal("alias file")
		}
		if _, err = cpp.ApplyMatrixFile(plan.rows, plan.columns, int(data.Fd()), int(temp.Fd()), 1, 16, n, test.tile, plan.split, 16, false); err == nil {
			t.Fatal("unaligned offset")
		}
		short := max(1<<uint(plan.split), n>>uint(plan.split)) - 1
		if _, err = cpp.ApplyMatrixFile(plan.rows, plan.columns, int(data.Fd()), int(temp.Fd()), 16, 16, n, short, plan.split, 16, false); err == nil {
			t.Fatal("short tile")
		}
		readonly, err := os.Open(data.Name())
		if err != nil {
			t.Fatal(err)
		}
		if _, err = cpp.ApplyMatrixFile(plan.rows, plan.columns, int(readonly.Fd()), int(temp.Fd()), 16, 16, n, test.tile, plan.split, 16, false); err == nil {
			t.Fatal("readonly")
		}
		readonly.Close()
		if err = temp.Truncate(16); err != nil {
			t.Fatal(err)
		}
		if _, err = call(false); err == nil {
			t.Fatal("short file")
		}
		temp.Truncate(int64(n)*4 + 32)
		temp.WriteAt(guard, int64(n)*4+16)
		if test.s < 16 {
			var bad [4]byte
			binary.LittleEndian.PutUint32(bad[:], uint32(field.q))
			data.WriteAt(bad[:], 16+int64(n-1)*4)
			if _, err = call(false); err == nil {
				t.Fatal("invalid field value")
			}
		}
		saveInput()
		if _, err = call(false); err != nil || !reflect.DeepEqual(readData(), want) {
			t.Fatal("restored recovery", err)
		}
		for _, f := range files {
			for _, offset := range []int64{0, int64(n)*4 + 16} {
				got := make([]byte, 16)
				if _, err = f.ReadAt(got, offset); err != nil || !bytes.Equal(got, guard) {
					t.Fatal("canary", err)
				}
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
		}
		// Forward data is retained as experiment evidence; scratch is a reusable file.
		raw, err := os.ReadFile(data.Name())
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw[16 : 16+4*n])
		digest := hex.EncodeToString(hash[:])
		plan.rows.Close()
		plan.columns.Close()
		if _, err = call(false); err == nil {
			t.Fatal("closed plan")
		}
		for _, f := range files {
			f.Close()
		}
		records = append(records, map[string]any{"s": test.s, "dimension": test.dimension, "n": n, "tile_elements": test.tile, "matches": 3, "errors_recovery_canaries": true, "resident_seconds": residentSeconds, "forward_seconds": forwardSeconds, "inverse_seconds": inverseSeconds, "forward": forward, "inverse": inverse, "forward_sha256": digest, "backing_bytes": 8*uint64(n) + 64})
		runtime.GC()
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "disk-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
