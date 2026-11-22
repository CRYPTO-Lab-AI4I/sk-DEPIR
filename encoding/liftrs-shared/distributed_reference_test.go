package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Retained references are test inputs with independently verified provenance,
// not a discovery source or a substitute for the complete polynomial identity.
// Fixed filenames and lengths prevent metadata from selecting arbitrary paths.
// This fixture loader admits N<=2^20, whose canonical header/slots are32-bit.
type pilotReferenceRecord struct {
	Parameters SupportParameters `json:"parameters"`
	Entropy    string            `json:"entropy_sha256"`
	SHA        map[string]string `json:"sha256"`
}

func pilotReferenceSizes(info SupportParameters) map[string]uint64 {
	label := fmt.Sprintf("s%d", info.Spec.FieldDegree)
	return map[string]uint64{
		label + "-rank-0-reference-p.bin":        4 * (info.K + 1),
		label + "-rank-0-reference-g.bin":        4 * (info.N - info.K),
		label + "-rank-0-reference-spectrum.bin": 4 * info.N,
		label + "-reference-code.bin":            25 + 4*(info.K+1+info.N),
		label + "-support.bin":                   uint64(supportWidth(info.Spec.FieldDegree)) * info.K,
	}
}

func readPilotReference(directory string, s int) (pilotReferenceRecord, error) {
	var record pilotReferenceRecord
	data, err := readBoundedPilotFile(filepath.Join(directory, fmt.Sprintf("s%d-reference.json", s)), 64<<10)
	if err != nil {
		return record, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&record); err != nil {
		return record, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return record, fmt.Errorf("reference trailing metadata")
	}
	info := record.Parameters
	if info.Spec != (Spec{s, 2, 2}) || info.N != uint64(1)<<uint(2*s) || info.Samples < 1 || info.Samples > discoverySampleLimit(s) {
		return record, fmt.Errorf("reference shape")
	}
	if err = validSupportParameters(info, PreprocessingLimits{MaxElements: 1 << 20, MaxHostArrayBytes: 32 << 20}); err != nil {
		return record, err
	}
	if _, err = expectedDigest(record.Entropy); err != nil {
		return record, err
	}
	sizes := pilotReferenceSizes(info)
	if len(record.SHA) != len(sizes) {
		return record, fmt.Errorf("reference file set")
	}
	for name, size := range sizes {
		if _, err = expectedDigest(record.SHA[name]); err != nil {
			return record, err
		}
		st, e := os.Stat(filepath.Join(directory, name))
		if e != nil {
			return record, e
		}
		if size > 32<<20 || !st.Mode().IsRegular() || st.Size() != int64(size) {
			return record, fmt.Errorf("reference file size")
		}
	}
	return record, nil
}

// Copy and verify with fixed buffers, never full polynomial/support arrays.
// Support itself is discovered and streamed afresh, so only verify it here.
func retainPilotReference(source, destination string, record pilotReferenceRecord) error {
	for name, size := range pilotReferenceSizes(record.Parameters) {
		target := filepath.Join(destination, name)
		if name == fmt.Sprintf("s%d-support.bin", record.Parameters.Spec.FieldDegree) {
			target = ""
		}
		if err := copyCheckedPilotFile(filepath.Join(source, name), target, size, record.SHA[name]); err != nil {
			return err
		}
	}
	return nil
}
func copyPilotSupport(source, destination, label string, record pilotReferenceRecord) error {
	name := label + "-support.bin"
	return copyCheckedPilotFile(filepath.Join(source, name), filepath.Join(destination, name), pilotReferenceSizes(record.Parameters)[name], record.SHA[name])
}
func copyCheckedPilotFile(source, destination string, size uint64, digest string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	hash := sha256.New()
	var output *os.File
	var sink io.Writer = hash
	if destination != "" {
		output, err = os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer func() {
			if output != nil {
				output.Close()
			}
		}()
		sink = io.MultiWriter(output, hash)
	}
	count, err := io.CopyBuffer(sink, io.LimitReader(input, int64(size)+1), make([]byte, 16<<10))
	if err != nil {
		return err
	}
	if uint64(count) != size || hex.EncodeToString(hash.Sum(nil)) != digest {
		return fmt.Errorf("reference checksum/extent")
	}
	if output != nil {
		if err = output.Sync(); err != nil {
			return err
		}
		err = output.Close()
		output = nil
	}
	return err
}

func TestRetainedPilotReference(t *testing.T) {
	source, dest := t.TempDir(), t.TempDir()
	h, _, err := canonicalScalingCodeHeader(4, 2, 97, 256)
	if err != nil {
		t.Fatal(err)
	}
	record := pilotReferenceRecord{Parameters: SupportParameters{Spec: h.Spec, K: h.K, N: h.N, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Samples: 1}, Entropy: fmt.Sprintf("%064x", 0), SHA: map[string]string{}}
	for name, size := range pilotReferenceSizes(record.Parameters) {
		data := make([]byte, size)
		sum := sha256.Sum256(data)
		record.SHA[name] = hex.EncodeToString(sum[:])
		if err = os.WriteFile(filepath.Join(source, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	metadata := filepath.Join(source, "s4-reference.json")
	save := func() {
		data, e := json.Marshal(record)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(metadata, data, 0600); e != nil {
			t.Fatal(e)
		}
	}
	save()
	loaded, err := readPilotReference(source, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err = retainPilotReference(source, dest, loaded); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dest, "s4-support.bin")); !os.IsNotExist(err) {
		t.Fatal("support copied before discovery")
	}
	if err = copyPilotSupport(source, dest, "s4", loaded); err != nil {
		t.Fatal(err)
	}
	name := "s4-rank-0-reference-p.bin"
	data, _ := os.ReadFile(filepath.Join(source, name))
	data[0] ^= 1
	if err = os.WriteFile(filepath.Join(source, name), data, 0600); err != nil {
		t.Fatal(err)
	}
	if retainPilotReference(source, t.TempDir(), loaded) == nil {
		t.Fatal("corrupt reference accepted")
	}
	if err = os.WriteFile(filepath.Join(source, name), data[:1], 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = readPilotReference(source, 4); err == nil {
		t.Fatal("short reference accepted")
	}
	record.SHA["../escape"] = "bad"
	save()
	if _, err = readPilotReference(source, 4); err == nil {
		t.Fatal("extra path accepted")
	}
}

// Global support intervals arrive in increasing order. The writer retains only
// the preceding exponent and emits canonical width through the existing codec.
type orderedSupportWriter struct {
	Writer         io.Writer
	S              int
	N, Count, Last uint64
}

func (w *orderedSupportWriter) append(values []uint64) error {
	last := w.Last
	for j, v := range values {
		if v >= w.N-1 || ((w.Count > 0 || j > 0) && v <= last) {
			return fmt.Errorf("support order/range")
		}
		last = v
	}
	if err := writeScalingCodeValues(w.Writer, supportWidth(w.S), w.N, nil, values); err != nil {
		return err
	}
	w.Last = last
	w.Count += uint64(len(values))
	return nil
}

type shortPilotWriter struct{}

func (shortPilotWriter) Write(data []byte) (int, error) { return len(data) - 1, nil }

func TestOrderedSupportWriter(t *testing.T) {
	broken := orderedSupportWriter{Writer: shortPilotWriter{}, S: 4, N: 256}
	if broken.append([]uint64{1}) != io.ErrShortWrite || broken.Count != 0 {
		t.Fatal("short support write")
	}
	for _, s := range []int{4, 18} {
		n := uint64(1) << uint(2*s)
		var out bytes.Buffer
		w := orderedSupportWriter{Writer: &out, S: s, N: n}
		if err := w.append(nil); err != nil {
			t.Fatal(err)
		}
		for _, v := range [][]uint64{{0, 1}, {2}, {n - 2}} {
			if err := w.append(v); err != nil {
				t.Fatal(err)
			}
		}
		if w.Count != 4 || out.Len() != 4*int(supportWidth(s)) {
			t.Fatal("support count/width")
		}
		before := append([]byte(nil), out.Bytes()...)
		for _, v := range [][]uint64{{n - 2}, {n - 1}, {3, 2}} {
			if w.append(v) == nil || !bytes.Equal(before, out.Bytes()) {
				t.Fatal("invalid stream mutated output")
			}
		}
	}
}
