package liftrs

// Internal sectioned online-file probe. This is not a public artifact format.
// The caller supplies a digest from a separately trusted prepared-encoder manifest;
// checksums establish byte identity, not mathematical code validity.
import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeScalingOnline(file io.Writer, p *ScalingOnlinePlan) error {
	w := bufio.NewWriterSize(file, 64<<10)
	if _, err := w.WriteString("LRSprobe1"); err != nil {
		return err
	}
	words := []uint32{uint32(p.basis.spec.FieldDegree), uint32(p.basis.spec.CurveDegree), uint32(p.k), uint32(p.basis.Length())}
	for _, x := range words {
		if err := binary.Write(w, binary.LittleEndian, x); err != nil {
			return err
		}
	}
	var block [4096]byte
	for _, values := range [][]uint32{p.p, p.spectrum} {
		for start := 0; start < len(values); start += len(block) / 4 {
			count := minInt(len(block)/4, len(values)-start)
			for i := 0; i < count; i++ {
				binary.LittleEndian.PutUint32(block[4*i:], values[start+i])
			}
			if _, err := w.Write(block[:4*count]); err != nil {
				return err
			}
		}
	}
	return w.Flush()
}
func loadScalingOnline(source io.Reader, digest string, maxElements int) (*ScalingOnlinePlan, error) {
	hash := sha256.New()
	r := bufio.NewReaderSize(io.TeeReader(source, hash), 64<<10)
	var magic [9]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if string(magic[:]) != "LRSprobe1" {
		return nil, fmt.Errorf("bad probe magic")
	}
	var header [4]uint32
	for i := range header {
		if err := binary.Read(r, binary.LittleEndian, &header[i]); err != nil {
			return nil, err
		}
	}
	s, t, k, n := int(header[0]), int(header[1]), int(header[2]), int(header[3])
	if s < 2 || s > 13 || t < 1 || t >= (1<<uint(s))-1 || n != 1<<uint(2*s) || n > maxElements || k < 1 || k >= n {
		return nil, fmt.Errorf("invalid online probe shape/limit")
	}
	polynomial, _ := BinaryPolynomial(s)
	base, extension, delta, omega := orbitFields(s, polynomial)
	b := OrbitBasis{spec: Spec{s, 2, t}, polynomial: polynomial, delta: delta, omega: omega, base: base, extension: extension, extensionFFT: newSparseTransform(extension)}
	p := &ScalingOnlinePlan{basis: b, k: k, p: make([]uint32, k+1), spectrum: make([]uint32, n)}
	for j, values := range [][]uint32{p.p, p.spectrum} {
		limit := base.q
		if j == 1 {
			limit = extension.q
		}
		for i := range values {
			if err := binary.Read(r, binary.LittleEndian, &values[i]); err != nil {
				return nil, err
			}
			if values[i] >= limit {
				return nil, fmt.Errorf("online element outside field")
			}
		}
	}
	if p.p[0] != 1 {
		return nil, fmt.Errorf("invalid P constant")
	}
	if _, err := r.ReadByte(); err != io.EOF {
		return nil, fmt.Errorf("trailing probe data or read error")
	}
	if hex.EncodeToString(hash.Sum(nil)) != digest {
		return nil, fmt.Errorf("online probe checksum differs from manifest")
	}
	return p, nil
}
func TestScalingOnlineFile(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(905))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicScalingPrototype()
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "online.bin")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	if err := writeScalingOnline(io.MultiWriter(file, hash), p); err != nil {
		t.Fatal(err)
	}
	file.Close()
	file, err = os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	loaded, err := loadScalingOnline(file, hex.EncodeToString(hash.Sum(nil)), b.Length())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.p, loaded.p) || !reflect.DeepEqual(p.spectrum, loaded.spectrum) || len(loaded.basis.support) != 0 {
		t.Fatal("minimal plan")
	}
	message := make([]uint32, b.Dimension())
	for i := range message {
		message[i] = uint32(i % 16)
	}
	got, err := loaded.Encode(message, loaded.NewWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	want, err := e.Encode(message)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("online file encoding", err)
	}
	for _, invalid := range []struct {
		digest string
		limit  int
	}{{"bad", b.Length()}, {hex.EncodeToString(hash.Sum(nil)), b.Length() - 1}} {
		file.Seek(0, 0)
		if _, err := loadScalingOnline(file, invalid.digest, invalid.limit); err == nil {
			t.Fatal("invalid manifest/limit accepted")
		}
	}
}
