//go:build liftrs_cpp && cgo

package liftrs

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCPPFileEncoding(t *testing.T) {
	b, err := DiscoverOrbit(Spec{6, 2, 4}, OrbitOptions{Random: rand.New(rand.NewSource(3701))})
	if err != nil {
		t.Fatal(err)
	}
	code, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	p, err := code.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "online.bin")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = WriteScalingOnline(f, p); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(raw)
	digest := hex.EncodeToString(hash[:])
	e, err := OpenScalingCPPFile(path, digest, p.Length(), 4)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if e.cpu.source.ResidentBytes() != 0 || e.WorkspaceBytes() != 0 {
		t.Fatal("file mode retained code/work arrays during open")
	}
	input := make([]uint32, p.k)
	for i := range input {
		input[i] = uint32(i % 64)
	}
	original := append([]uint32(nil), input...)
	var kept, saved []uint32
	for pass := 0; pass < 3; pass++ {
		message := append([]uint32(nil), input...)
		if pass == 1 {
			clear(message)
		}
		want, err := p.Encode(message, p.NewWorkspace())
		if err != nil {
			t.Fatal(err)
		}
		got, st, err := e.EncodeWithStats(message)
		if err != nil || !reflect.DeepEqual(got, want) || st.ReadBytes != uint64(p.k+1+p.Length())*4 || st.ReadChunks != 2 || e.WorkspaceBytes() != uint64(p.Length())*4 {
			t.Fatal("file/reference/counts", err, st)
		}
		if pass == 0 {
			kept = got
			saved = append([]uint32(nil), got...)
		}
		if !reflect.DeepEqual(kept, saved) || !reflect.DeepEqual(input, original) {
			t.Fatal("file input/output ownership")
		}
	}
	if _, _, err = e.EncodeWithStats(input[:len(input)-1]); err == nil {
		t.Fatal("short message")
	}
	bad := append([]uint32(nil), input...)
	bad[0] = 64
	if _, _, err = e.EncodeWithStats(bad); err == nil {
		t.Fatal("bad field input")
	}
	if _, _, err = e.EncodeWithStats(input); err != nil {
		t.Fatal("file error recovery", err)
	}
	pre, full := e.cpu.plans[e.prefix], e.cpu.plans[p.Length()]
	arena, scratch := make([]uint32, p.Length()), make([]uint32, p.Length())
	want, err := p.Encode(input, p.NewWorkspace())
	if err != nil {
		t.Fatal(err)
	}
	for _, tile := range []int{1, 257, p.Length() - 1, p.Length()} {
		st, err := cpp.EncodeOnlineFile(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, 4, p.basis.omega, input, arena, scratch, e.prefix, cpp.OnlineFile{Descriptor: int(e.file.Fd()), POffset: e.pOffset, SpectrumOffset: e.spectrumOffset, TileElements: tile})
		chunks := (p.k+1+tile-1)/tile + (p.Length()+tile-1)/tile
		if err != nil || !reflect.DeepEqual(scratch, want) || st.ReadChunks != uint64(chunks) {
			t.Fatal("file tile/tail", tile, err)
		}
	}
	for _, tile := range []int{0, p.Length() + 1} {
		if _, err = cpp.EncodeOnlineFile(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, 4, p.basis.omega, input, arena, scratch, e.prefix, cpp.OnlineFile{Descriptor: int(e.file.Fd()), POffset: e.pOffset, SpectrumOffset: e.spectrumOffset, TileElements: tile}); err == nil {
			t.Fatal("invalid file tile")
		}
	}
	if _, err = cpp.EncodeOnlineFile(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, 4, p.basis.omega, input, arena, arena, e.prefix, cpp.OnlineFile{Descriptor: int(e.file.Fd()), TileElements: 1}); err == nil {
		t.Fatal("file arena alias")
	}
	for _, test := range []struct {
		digest string
		limit  int
	}{{"", p.Length()}, {hex.EncodeToString(make([]byte, 32)), p.Length()}, {digest, p.Length() / 2}} {
		if x, err := OpenScalingCPPFile(path, test.digest, test.limit, 4); err == nil {
			x.Close()
			t.Fatal("bad digest/limit")
		}
	}
	// A descriptor keeps its verified inode even when the pathname is replaced.
	moved := path + ".saved"
	if err = os.Rename(path, moved); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("unrelated replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err = e.EncodeWithStats(input); err != nil {
		t.Fatal("descriptor identity", err)
	}
	if err = os.Truncate(moved, int64(len(raw)-4)); err != nil {
		t.Fatal(err)
	}
	if _, _, err = e.EncodeWithStats(input); err == nil {
		t.Fatal("backing truncation not detected")
	}
	if _, err = cpp.EncodeOnlineFile(pre.rows, pre.columns, full.rows, full.columns, pre.split, full.split, 4, p.basis.omega, input, arena, scratch, e.prefix, cpp.OnlineFile{Descriptor: int(e.file.Fd()), POffset: e.pOffset, SpectrumOffset: e.spectrumOffset, TileElements: p.Length()}); err == nil {
		t.Fatal("native truncated read accepted")
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	if err = e.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(kept, saved) {
		t.Fatal("file Close overwrote returned word")
	}
	if _, _, err = e.EncodeWithStats(input); !errors.Is(err, ErrCPPClosed) {
		t.Fatal("closed file encoder", err)
	}
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	// Wide-count header must preserve the same payload even on a small fixture.
	wide := append([]byte(nil), raw[:17]...)
	wide[8] = '2'
	counts := make([]byte, 16)
	binary.LittleEndian.PutUint64(counts, uint64(p.k))
	binary.LittleEndian.PutUint64(counts[8:], uint64(p.Length()))
	wide = append(wide, counts...)
	wide = append(wide, raw[25:]...)
	hash = sha256.Sum256(wide)
	if err = os.WriteFile(path, wide, 0600); err != nil {
		t.Fatal(err)
	}
	w, err := OpenScalingCPPFile(path, hex.EncodeToString(hash[:]), p.Length(), 4)
	if err != nil {
		t.Fatal("wide file header", err)
	}
	got, _, err := w.EncodeWithStats(input)
	w.Close()
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("wide header word", err)
	}
	invalidSpectrum := append([]byte(nil), raw...)
	binary.LittleEndian.PutUint32(invalidSpectrum[25+4*(p.k+1):], 1<<12)
	hash = sha256.Sum256(invalidSpectrum)
	if err = os.WriteFile(path, invalidSpectrum, 0600); err != nil {
		t.Fatal(err)
	}
	if x, err := OpenScalingCPPFile(path, hex.EncodeToString(hash[:]), p.Length(), 4); err == nil {
		x.Close()
		t.Fatal("out-of-field spectrum accepted")
	}
	// Recomputed byte identity must not bypass field-range validation.
	raw[25] = 255
	raw[26] = 255
	hash = sha256.Sum256(raw)
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if x, err := OpenScalingCPPFile(path, hex.EncodeToString(hash[:]), p.Length(), 4); err == nil {
		x.Close()
		t.Fatal("malformed P accepted")
	}
}
