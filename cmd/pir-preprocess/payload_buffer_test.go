package main

import (
	"bytes"
	"errors"
	lrs "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-shared"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBufferedPayloadEquivalenceAndFailure(t *testing.T) {
	raw := bytes.Repeat([]byte{1, 2, 3, 4}, 32769) // Non-aligned tail must be flushed.
	report := &reporter{io.Discard, time.Now()}
	var expected fileRecord
	for _, size := range []int{0, 65536, 4 << 20} {
		dir := t.TempDir()
		var stats payloadWriteStats
		record, err := writePayloadBuffered(dir, "payload.bin", uint64(len(raw)), report, size, &stats, hashed(func(w io.Writer) error {
			for first := 0; first < len(raw); first += 65536 {
				end := first + 65536
				if end > len(raw) {
					end = len(raw)
				}
				if _, err := w.Write(raw[first:end]); err != nil {
					return err
				}
			}
			return nil
		}))
		if err != nil {
			t.Fatal(err)
		}
		if size == 0 {
			expected = record
		} else if record != expected {
			t.Fatal("buffer changes bytes/digest")
		}
		data, err := os.ReadFile(filepath.Join(dir, "payload.bin"))
		if err != nil || !bytes.Equal(data, raw) {
			t.Fatal("tail lost", err)
		}
		if stats.Bytes != uint64(len(raw)) || stats.Calls == 0 || (size == 4<<20 && stats.Calls != 1) {
			t.Fatal("file call counts", stats)
		}
		injected := errors.New("producer interrupted")
		_, err = writePayloadBuffered(dir, "partial.bin", uint64(len(raw)), report, size, nil, func(w io.Writer) (string, error) {
			_, err := w.Write(raw[:7])
			if err != nil {
				return "", err
			}
			return "", injected
		})
		if !errors.Is(err, injected) {
			t.Fatal("producer failure swallowed", err)
		}
	}
	if _, err := writePayloadBuffered(t.TempDir(), "bad", 1, report, (4<<20)+1, nil, func(io.Writer) (string, error) { t.Fatal("invalid buffer called producer"); return "", nil }); err == nil {
		t.Fatal("oversized buffer accepted")
	}
}

func TestBufferedPublicationFailureAndReuse(t *testing.T) {
	for _, size := range []int{0, 4 << 20} {
		cache := t.TempDir()
		spec := lrs.Spec{FieldDegree: 4, Dimension: 2, CurveDegree: 3}
		report := &reporter{io.Discard, time.Now()}
		injected := errors.New("interrupted producer")
		fresh := func() manifest {
			return manifest{Version: 1, Stage: "probe", Spec: spec, Files: map[string]fileRecord{}}
		}
		_, err := publish(cache, "probe", fresh(), func(dir string, m *manifest) error {
			_, err := writePayloadBuffered(dir, "payload.bin", 8, report, size, nil, func(w io.Writer) (string, error) {
				_, err := w.Write([]byte("partial"))
				if err != nil {
					return "", err
				}
				return "", injected
			})
			return err
		})
		if !errors.Is(err, injected) {
			t.Fatal("publication failure lost", err)
		}
		if _, err = os.Stat(filepath.Join(cache, "probe")); !os.IsNotExist(err) {
			t.Fatal("failed stage published")
		}
		complete := []byte("complete")
		_, err = publish(cache, "probe", fresh(), func(dir string, m *manifest) error {
			record, err := writePayloadBuffered(dir, "payload.bin", 8, report, size, nil, hashed(func(w io.Writer) error { _, err := w.Write(complete); return err }))
			m.Files["payload.bin"] = record
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := readManifest(cache, "probe", spec)
		if err != nil {
			t.Fatal(err)
		}
		if err = verifyFile(cache, "probe", "payload.bin", loaded.Files["payload.bin"], report); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(cache, "probe", "payload.bin"), []byte("corrupt!"), 0600); err != nil {
			t.Fatal(err)
		}
		if err = verifyFile(cache, "probe", "payload.bin", loaded.Files["payload.bin"], report); err == nil {
			t.Fatal("corruption accepted")
		}
	}
}
