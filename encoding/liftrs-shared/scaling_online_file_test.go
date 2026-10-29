package liftrs

// Internal sectioned online-file probe. This is not a public artifact format.
// The caller supplies a digest from a separately trusted prepared-encoder manifest;
// checksums establish byte identity, not mathematical code validity.
import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
	return LoadScalingOnline(source, digest, maxElements)
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
