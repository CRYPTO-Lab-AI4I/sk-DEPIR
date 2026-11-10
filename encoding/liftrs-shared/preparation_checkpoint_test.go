//go:build liftrs_cpp && cgo

package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Root files are private trusted public-code checkpoints, bound to the exact
// support digest/spec. They are never substituted for final identity checks.
func loadPreparationRoot(cache string, spec Spec, k, n int, support string) ([]uint32, error) {
	raw, err := os.ReadFile(filepath.Join(cache, "root-metadata.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		S, T, K, N    int
		SHA256        string
		SupportSHA256 string `json:"support_sha256"`
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.S != spec.FieldDegree || m.T != spec.CurveDegree || m.K != k || m.N != n || m.SupportSHA256 != support {
		return nil, fmt.Errorf("root checkpoint provenance")
	}
	f, err := os.Open(filepath.Join(cache, "root.bin"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() != int64(k+1)*4 {
		return nil, fmt.Errorf("root checkpoint size")
	}
	h := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(f, h), 1<<20)
	p := make([]uint32, k+1)
	var block [65536]byte
	for offset := 0; offset < len(p); {
		count := minInt(len(block)/4, len(p)-offset)
		if _, err = io.ReadFull(r, block[:4*count]); err != nil {
			return nil, err
		}
		for j := 0; j < count; j++ {
			p[offset+j] = binary.LittleEndian.Uint32(block[4*j:])
		}
		offset += count
	}
	if hex.EncodeToString(h.Sum(nil)) != m.SHA256 {
		return nil, fmt.Errorf("root checkpoint digest")
	}
	return p, nil
}

func TestPreparationRootCheckpoint(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(2101))})
	if err != nil {
		t.Fatal(err)
	}
	// Match private scaling preparation's cutoff setup, preserving reference fields.
	b.base.products = &polynomialProducts{transform: b.extensionFFT, cutoff: 32768}
	b.extension.products = &polynomialProducts{transform: b.extensionFFT, cutoff: 1024}
	var root []uint32
	original, err := b.prepareCyclicScalingCheckpointed(b.extensionFFT, nil, nil, func(p []uint32) error { root = append([]uint32(nil), p...); return nil })
	if err != nil {
		t.Fatal(err)
	}
	replay, err := b.prepareCyclicScalingCheckpointed(b.extensionFFT, nil, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original.p, replay.p) || !reflect.DeepEqual(original.g, replay.g) || !reflect.DeepEqual(original.spectrum, replay.spectrum) {
		t.Fatal("checkpoint replay differs")
	}
	for _, bad := range [][]uint32{{1}, make([]uint32, len(root)), append([]uint32{1, 1 << 16}, root[2:]...)} {
		if _, err = b.prepareCyclicScalingCheckpointed(b.extensionFFT, nil, bad, nil); err == nil {
			t.Fatal("bad root accepted")
		}
	}
	directory := t.TempDir()
	data := make([]byte, 4*len(root))
	for i, v := range root {
		binary.LittleEndian.PutUint32(data[4*i:], v)
	}
	digest := sha256.Sum256(data)
	metadata := map[string]any{"s": 4, "t": 2, "k": b.Dimension(), "n": b.Length(), "sha256": hex.EncodeToString(digest[:]), "support_sha256": "reference"}
	raw, _ := json.Marshal(metadata)
	os.WriteFile(filepath.Join(directory, "root.bin"), data, 0600)
	os.WriteFile(filepath.Join(directory, "root-metadata.json"), raw, 0600)
	got, err := loadPreparationRoot(directory, b.spec, b.Dimension(), b.Length(), "reference")
	if err != nil || !reflect.DeepEqual(got, root) {
		t.Fatal("load", err)
	}
	if _, err = loadPreparationRoot(directory, b.spec, b.Dimension(), b.Length(), "other"); err == nil {
		t.Fatal("support mismatch accepted")
	}
	data[0] ^= 1
	os.WriteFile(filepath.Join(directory, "root.bin"), data, 0600)
	if _, err = loadPreparationRoot(directory, b.spec, b.Dimension(), b.Length(), "reference"); err == nil {
		t.Fatal("digest mismatch accepted")
	}
}
