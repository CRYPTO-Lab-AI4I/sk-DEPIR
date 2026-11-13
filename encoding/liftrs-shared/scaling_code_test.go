package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math/rand"
	"reflect"
	"testing"
	"testing/iotest"
)

func codeDigest(raw []byte) string { h := sha256.Sum256(raw); return hex.EncodeToString(h[:]) }
func smallScalingCode(t *testing.T) (*ScalingOnlinePlan, []byte) {
	t.Helper()
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(170057))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	p, err := e.ScalingPlan()
	if err != nil {
		t.Fatal(err)
	}
	var raw bytes.Buffer
	if err = WriteScalingOnline(&raw, p); err != nil {
		t.Fatal(err)
	}
	return p, raw.Bytes()
}
func codeLimits(p *ScalingOnlinePlan) ScalingCodeLimits {
	return ScalingCodeLimits{uint64(p.Length()), p.ResidentBytes()}
}

func TestScalingCodeLegacy(t *testing.T) {
	p, raw := smallScalingCode(t)
	limits := codeLimits(p)
	d, err := LoadScalingCode(iotest.OneByteReader(bytes.NewReader(raw)), codeDigest(raw), limits)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(d.p, p.p) || !reflect.DeepEqual(d.spectrum32, p.spectrum) || d.spectrum64 != nil || d.ResidentBytes() != p.ResidentBytes() {
		t.Fatal("legacy arrays widened or changed")
	}
	h, err := VerifyScalingCode(bytes.NewReader(raw), codeDigest(raw), limits)
	if err != nil || h != d.Header() {
		t.Fatal("stream verification", err)
	}
	for j := 0; j < p.Dimension(); j++ {
		old, _ := p.SystematicPosition(j)
		got, err := d.SystematicPosition(uint64(j))
		if err != nil || old != got {
			t.Fatal("legacy ordering", j, err)
		}
	}
	if _, err = d.SystematicPosition(h.K); err == nil {
		t.Fatal("index cap")
	}
	var stored bytes.Buffer
	if err = WriteScalingCode(&stored, d); err != nil || !bytes.Equal(raw, stored.Bytes()) {
		t.Fatal("legacy bytes changed", err)
	}
	// Read a historical wide-count header even at a smaller field size; writing
	// normalizes its counts back to the canonical width for s.
	var alternate bytes.Buffer
	alternate.WriteString("LRSprobe2")
	binary.Write(&alternate, binary.LittleEndian, [2]uint32{4, 2})
	binary.Write(&alternate, binary.LittleEndian, [2]uint64{uint64(p.Dimension()), uint64(p.Length())})
	alternate.Write(raw[25:])
	converted, err := LoadScalingCode(bytes.NewReader(alternate.Bytes()), codeDigest(alternate.Bytes()), limits)
	if err != nil {
		t.Fatal(err)
	}
	if converted.Header().CountBytes != 8 || converted.Header().POffset != 33 {
		t.Fatal("actual old layout lost")
	}
	stored.Reset()
	if err = WriteScalingCode(&stored, converted); err != nil || !bytes.Equal(stored.Bytes(), raw) {
		t.Fatal("old count normalization", err)
	}
}

func TestScalingCodeHeaderWidths(t *testing.T) {
	for _, s := range []int{15, 16, 17, 18, 19, 20} {
		n := uint64(1) << uint(2*s)
		h, f, err := canonicalScalingCodeHeader(s, 5, n/4, n)
		if err != nil {
			t.Fatal(err)
		}
		var raw bytes.Buffer
		if err = writeScalingCodeHeader(&raw, h); err != nil {
			t.Fatal(err)
		}
		headerBytes := 25
		if s == 16 {
			headerBytes = 33
		}
		if s >= 17 {
			headerBytes = 53
		}
		if raw.Len() != headerBytes || h.POffset != uint64(headerBytes) || h.Bytes != h.SpectrumOffset+uint64(h.ExtensionBytes)*n {
			t.Fatal("offset/width", s)
		}
		got, err := ReadScalingCodeHeader(bytes.NewReader(raw.Bytes()))
		if err != nil || got != h {
			t.Fatal("header roundtrip", s, err)
		}
		if s >= 17 {
			if got.CountBytes != 8 || got.ExtensionBytes != 8 {
				t.Fatal("wide slot/count")
			}
			d := &ScalingCodeData{header: h, field: f}
			index := h.K - 1
			coordinate, err := d.SystematicPosition(index)
			if err != nil {
				t.Fatal(err)
			}
			if coordinate != f.pack(f.tower.Pow(f.unpack(h.Omega), index+1)) {
				t.Fatal("wide exponent truncation", s)
			}
			high := false
			for j := uint64(0); j < 128; j++ {
				x, _ := d.SystematicPosition(j)
				high = high || x > uint64(^uint32(0))
			}
			if !high {
				t.Fatal("no high-coordinate witness", s)
			}
			// An authentic header is rejected before resident payload allocation under
			// a small cap. Streaming verification with full extent caps just hits EOF.
			if _, err = LoadScalingCode(bytes.NewReader(raw.Bytes()), codeDigest(raw.Bytes()), ScalingCodeLimits{256, 4096}); err == nil {
				t.Fatal("wide allocation cap ignored")
			}
			if _, err = VerifyScalingCode(bytes.NewReader(raw.Bytes()), codeDigest(raw.Bytes()), ScalingCodeLimits{n, h.Bytes - h.POffset}); err == nil {
				t.Fatal("header-only wide file accepted")
			}
			if _, err = LoadScalingOnline(bytes.NewReader(raw.Bytes()), codeDigest(raw.Bytes()), 256); err == nil {
				t.Fatal("old encoder admitted new code format")
			}
		}
	}
}

func TestScalingCodeValidation(t *testing.T) {
	p, raw := smallScalingCode(t)
	limits := codeLimits(p)
	malformed := map[string]func([]byte) []byte{
		"magic":          func(x []byte) []byte { x[0] = '?'; return x },
		"P constant":     func(x []byte) []byte { binary.LittleEndian.PutUint32(x[25:], 0); return x },
		"base range":     func(x []byte) []byte { binary.LittleEndian.PutUint32(x[29:], 16); return x },
		"spectrum range": func(x []byte) []byte { binary.LittleEndian.PutUint32(x[25+4*(p.k+1):], 256); return x },
		"truncated":      func(x []byte) []byte { return x[:len(x)-1] },
		"trailing":       func(x []byte) []byte { return append(x, 0) },
	}
	for name, edit := range malformed {
		t.Run(name, func(t *testing.T) {
			bad := edit(append([]byte(nil), raw...))
			digest := codeDigest(bad)
			if _, err := LoadScalingCode(bytes.NewReader(bad), digest, limits); err == nil {
				t.Fatal("invalid resident accepted")
			}
			if _, err := VerifyScalingCode(bytes.NewReader(bad), digest, limits); err == nil {
				t.Fatal("invalid streamed accepted")
			}
		})
	}
	for _, digest := range []string{"", "nothex", codeDigest([]byte("wrong"))} {
		if _, err := LoadScalingCode(bytes.NewReader(raw), digest, limits); err == nil {
			t.Fatal("unverified data accepted")
		}
	}
	for _, cap := range []ScalingCodeLimits{{0, limits.MaxPayloadBytes}, {limits.MaxElements, 0}, {limits.MaxElements - 1, limits.MaxPayloadBytes}, {limits.MaxElements, limits.MaxPayloadBytes - 1}} {
		if _, err := LoadScalingCode(bytes.NewReader(raw), codeDigest(raw), cap); err == nil {
			t.Fatal("cap ignored")
		}
	}
	h, _, err := canonicalScalingCodeHeader(17, 5, 1<<32, 1<<34)
	if err != nil {
		t.Fatal(err)
	}
	var header bytes.Buffer
	writeScalingCodeHeader(&header, h)
	for _, offset := range []int{0, 25, 33, 37, 41, 49, 50, 51, 52} {
		bad := append([]byte(nil), header.Bytes()...)
		bad[offset] ^= 0xff
		if _, err := ReadScalingCodeHeader(bytes.NewReader(bad)); err == nil {
			t.Fatal("bad wide header accepted", offset)
		}
	}
	bad := append([]byte(nil), header.Bytes()...)
	binary.LittleEndian.PutUint64(bad[17:], ^uint64(0))
	if _, err := ReadScalingCodeHeader(bytes.NewReader(bad)); err == nil {
		t.Fatal("overflow count")
	}
	for n := 0; n < len(header.Bytes()); n++ {
		if _, err := ReadScalingCodeHeader(bytes.NewReader(header.Bytes()[:n])); err == nil {
			t.Fatal("short wide header", n)
		}
	}
	d, err := LoadScalingCode(bytes.NewReader(raw), codeDigest(raw), limits)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("writer failed")
	if err = WriteScalingCode(codeFailWriter{failure}, d); !errors.Is(err, failure) {
		t.Fatal("writer error lost", err)
	}
	d.p[0] = 0
	if err = WriteScalingCode(io.Discard, d); err == nil {
		t.Fatal("invalid data writer")
	}
	d.p[0] = 1
	d.header.Omega ^= 1
	if err = WriteScalingCode(io.Discard, d); err == nil {
		t.Fatal("writer reinterpreted coordinates")
	}
}

type codeFailWriter struct{ err error }

func (w codeFailWriter) Write([]byte) (int, error) { return 0, w.err }

func TestScalingCodeWideSlots(t *testing.T) {
	count := 16391
	values := make([]uint64, count)
	for i := range values {
		values[i] = (uint64(i)<<21 | uint64(i)) & ((1 << 34) - 1)
	}
	values[0] = 1 << 32
	values[1] = 1 << 33
	values[2] = (1 << 34) - 1
	var raw bytes.Buffer
	if err := writeScalingCodeValues(&raw, 8, 1<<34, nil, values); err != nil {
		t.Fatal(err)
	}
	got := make([]uint64, count)
	if _, err := readScalingCodeValues(bytes.NewReader(raw.Bytes()), uint64(count), 8, 1<<34, nil, got); err != nil || !reflect.DeepEqual(got, values) {
		t.Fatal("wide chunk/slot truncation", err)
	}
	if _, err := readScalingCodeValues(bytes.NewReader(raw.Bytes()), uint64(count), 8, 1<<34, nil, nil); err != nil {
		t.Fatal("streamed slots", err)
	}
	if _, err := readScalingCodeValues(bytes.NewReader(raw.Bytes()), uint64(count), 8, 1<<34, make([]uint32, count), nil); err == nil {
		t.Fatal("narrow truncation")
	}
	if err := writeScalingCodeValues(io.Discard, 4, 1<<34, nil, values); err == nil {
		t.Fatal("narrow write truncation")
	}
	if err := writeScalingCodeValues(io.Discard, 8, 1<<34, nil, []uint64{1 << 34}); err == nil {
		t.Fatal("wide range")
	}
	if _, err := readScalingCodeValues(bytes.NewReader([]byte{}), 1, 3, 1<<34, nil, nil); err == nil {
		t.Fatal("unknown width")
	}
}
