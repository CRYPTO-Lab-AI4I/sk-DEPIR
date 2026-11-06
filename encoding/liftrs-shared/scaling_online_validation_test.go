package liftrs

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"testing"
)

// Malformed/truncated files must fail even when their byte checksum is correct.
// This exercises shape/field checks independently of manifest authentication.
func TestScalingOnlineValidation(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(1502))})
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
	var buf bytes.Buffer
	if err := writeScalingOnline(&buf, p); err != nil {
		t.Fatal(err)
	}
	valid := buf.Bytes()
	cases := map[string]func([]byte) []byte{
		"magic":           func(x []byte) []byte { x[0] = '?'; return x },
		"shape":           func(x []byte) []byte { binary.LittleEndian.PutUint32(x[21:], 999); return x },
		"prefix capacity": func(x []byte) []byte { binary.LittleEndian.PutUint32(x[17:], uint32(b.Length()-1)); return x },
		"P constant":      func(x []byte) []byte { binary.LittleEndian.PutUint32(x[25:], 0); return x },
		"base field":      func(x []byte) []byte { binary.LittleEndian.PutUint32(x[29:], 16); return x },
		"extension field": func(x []byte) []byte { binary.LittleEndian.PutUint32(x[25+4*(p.k+1):], 256); return x },
		"truncated":       func(x []byte) []byte { return x[:len(x)-1] },
		"trailing":        func(x []byte) []byte { return append(x, 0) },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			raw := edit(append([]byte(nil), valid...))
			digest := sha256.Sum256(raw)
			if _, err := LoadScalingOnline(bytes.NewReader(raw), hex.EncodeToString(digest[:]), b.Length()); err == nil {
				t.Fatal("invalid minimal file accepted")
			}
		})
	}
}

// Admission checks do not allocate the full s=15 domain. Caller caps remain
// mandatory even when the experimental implementation understands its field.
func TestScalingOnlineLargeShape(t *testing.T) {
	for _, s := range []int{14, 15} {
		n := 1 << uint(2*s)
		if !validScalingOnlineShape(s, 5, n/4, n, n) ||
			validScalingOnlineShape(s, 5, n/4, n, n-1) ||
			validScalingOnlineShape(s, 5, n/2+1, n, n) ||
			validScalingOnlineShape(s, 5, n/4, n-1, n) {
			t.Fatal("large online shape guard", s)
		}
	}
	if validScalingOnlineShape(16, 5, 1, 1<<32, 1<<32) {
		t.Fatal("s=16 admitted despite packed arithmetic limit")
	}
}
