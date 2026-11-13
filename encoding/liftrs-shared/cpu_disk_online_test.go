//go:build liftrs_cpp && cgo

package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCPPDiskOnline(t *testing.T) {
	out, cache, fixtures := os.Getenv("PIR_CPU_DISK_ONLINE_OUT"), os.Getenv("PIR_CPU_DISK_ONLINE_CACHE"), os.Getenv("PIR_CPU_DISK_ONLINE_FIXTURES")
	if out == "" {
		t.Skip("opt-in complete disk encoding")
	}
	var records []map[string]any
	for index := 0; index < 9; index++ {
		name := filepath.Join(fixtures, fmt.Sprintf("case-%d.bin", index))
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw)
		digest := hex.EncodeToString(hash[:])
		p, err := LoadScalingOnline(bytes.NewReader(raw), digest, 1<<16)
		if err != nil {
			t.Fatal(err)
		}
		if index == 7 {
			// Both artifact versions have unaligned code ranges (25/33-byte headers).
			var wide bytes.Buffer
			wide.WriteString("LRSprobe2")
			wide.Write(raw[9:17])
			binary.Write(&wide, binary.LittleEndian, []uint64{uint64(p.k), uint64(p.Length())})
			wide.Write(raw[25:])
			name = filepath.Join(cache, "wide-code.bin")
			raw = wide.Bytes()
			if err = os.WriteFile(name, raw, 0600); err != nil {
				t.Fatal(err)
			}
			hash = sha256.Sum256(raw)
			digest = hex.EncodeToString(hash[:])
		}
		work := filepath.Join(cache, fmt.Sprintf("case-%d", index))
		e, err := OpenScalingCPPDisk(name, digest, p.Length(), 16, 1536, work)
		if err != nil {
			t.Fatal(err)
		}
		defer e.Close()
		if e.source.WorkspaceBytes() != 0 || e.source.cpu.source.ResidentBytes() != 0 {
			t.Fatal("disk retained in-memory code/workspace")
		}
		input := make([]uint32, p.k)
		for i := range input {
			input[i] = uint32((17*i + 3) % (1 << uint(p.FieldDegree())))
		}
		saved := append([]uint32(nil), input...)
		var last ScalingCPPDiskStats
		var first []uint32
		var firstCopy []uint32
		for repeat := 0; repeat < 3; repeat++ {
			if repeat == 1 {
				clear(input)
			} else {
				copy(input, saved)
			}
			before := append([]uint32(nil), input...)
			want, err := p.Encode(input, p.NewWorkspace())
			if err != nil {
				t.Fatal(err)
			}
			word, st, err := e.EncodeWithStats(input)
			if err != nil || !reflect.DeepEqual(word, want) || !reflect.DeepEqual(input, before) {
				t.Fatal("disk/Go/input", index, err)
			}
			if st.Calls != 5 || st.BoundaryCalls != 1 || st.BufferBytes > 16*uint64(minInt(1536, p.Length()))+2*65536 || st.BackingBytes != 8*uint64(p.Length()) || st.OutputBytes != 4*uint64(p.Length()) || st.ReadBytes == 0 || st.WriteBytes == 0 || e.source.WorkspaceBytes() != 0 {
				t.Fatal("disk accounting", st)
			}
			if repeat == 0 {
				first = word
				firstCopy = append([]uint32(nil), word...)
			}
			if !reflect.DeepEqual(first, firstCopy) {
				t.Fatal("owning result mutated")
			}
			last = st
		}
		if _, _, err = e.EncodeWithStats(input[:len(input)-1]); err == nil {
			t.Fatal("message length")
		}
		input[0] = 1 << uint(p.FieldDegree())
		if _, _, err = e.EncodeWithStats(input); err == nil {
			t.Fatal("message range")
		}
		copy(input, saved)
		if err = e.scratch.Truncate(0); err != nil {
			t.Fatal(err)
		}
		if _, _, err = e.EncodeWithStats(input); err == nil {
			t.Fatal("short workspace")
		}
		e.scratch.Truncate(int64(p.Length()) * 4)
		if _, _, err = e.EncodeWithStats(input); err != nil {
			t.Fatal("workspace error recovery", err)
		}
		e.Close()
		if _, _, err = e.EncodeWithStats(input); err == nil {
			t.Fatal("closed encode")
		}
		if _, err = OpenScalingCPPDisk(name, digest, p.Length(), 16, 1536, work); err == nil {
			t.Fatal("workspace replacement accepted")
		}
		records = append(records, map[string]any{"case": index, "s": p.FieldDegree(), "n": p.Length(), "matches": 3, "errors_recovery_owning": true, "stats": last})
	}
	raw, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "online-checks.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCPPDiskOnlineWorkflow(t *testing.T) {
	out, cache, reference := os.Getenv("PIR_CPU_DISK_WORKFLOW_OUT"), os.Getenv("PIR_CPU_DISK_WORKFLOW_CACHE"), os.Getenv("PIR_CPU_DISK_WORKFLOW_REFERENCE")
	if out == "" {
		t.Skip("opt-in full s10 disk workflow")
	}
	modes := []string{"file", "disk"}
	if os.Getenv("PIR_TABLE_S") != "" {
		modes = []string{"disk"}
	}
	for _, mode := range modes {
		dest := filepath.Join(cache, mode)
		if err := os.MkdirAll(dest, 0700); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"online.bin", "online-metadata.json"} {
			if err := os.Symlink(filepath.Join(reference, name), filepath.Join(dest, name)); err != nil {
				t.Fatal(err)
			}
		}
		results := filepath.Join(out, mode)
		if err := os.MkdirAll(results, 0700); err != nil {
			t.Fatal(err)
		}
		cfg := scalingFirstUseConfig{spec: Spec{10, 2, 5}, out: results, cache: dest, workers: 16, repeats: 2, memoryLimit: 4 << 30, onlineOnly: true}
		var fast *ScalingCPPFileEncoder
		var disk *ScalingCPPDiskEncoder
		cfg.onlineLoad = func(path, digest string, max int) (*ScalingOnlinePlan, error) {
			var err error
			if mode == "file" {
				fast, err = OpenScalingCPPFile(path, digest, max, 16)
				if err != nil {
					return nil, err
				}
				return fast.cpu.source, nil
			}
			disk, err = OpenScalingCPPDisk(path, digest, max, 16, 65536, filepath.Join(dest, "workspace"))
			if err != nil {
				return nil, err
			}
			return disk.source.cpu.source, nil
		}
		cfg.onlineSetup = func(p *ScalingOnlinePlan) (func([]uint32) ([]uint32, map[string]any, error), func(), map[string]any, error) {
			if mode == "file" {
				return func(input []uint32) ([]uint32, map[string]any, error) {
					word, st, err := fast.EncodeWithStats(input)
					return word, map[string]any{"native_stages": st.ScalingCPPStats, "file_stats": st}, err
				}, func() { fast.Close() }, map[string]any{"mode": mode}, nil
			}
			return func(input []uint32) ([]uint32, map[string]any, error) {
				word, st, err := disk.EncodeWithStats(input)
				return word, map[string]any{"native_stages": st.ScalingCPPStats, "disk_stats": st}, err
			}, func() { disk.Close() }, map[string]any{"mode": mode}, nil
		}
		runScalingFirstUse(t, cfg)
	}
}
