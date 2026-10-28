package liftrs

import (
	"bytes"
	"math/rand"
	"testing"
)

// Explicit local-cache limits must retain default behavior and malformed-data
// checks rather than bypassing validation to make large checkpoints reloadable.
func TestCyclicExplicitCacheLimit(t *testing.T) {
	b, err := DiscoverOrbit(Spec{4, 2, 2}, OrbitOptions{Random: rand.New(rand.NewSource(710))})
	if err != nil {
		t.Fatal(err)
	}
	e, err := b.PrepareCyclicFFT()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := e.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	for _, limit := range []int64{-1, 0, int64(len(raw) - 1)} {
		if _, err := LoadCyclicEncoderWithLimit(raw, limit); err == nil {
			t.Fatal("invalid byte limit accepted")
		}
	}
	loaded, err := LoadCyclicEncoderWithLimit(raw, int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	message := make([]uint32, b.Dimension())
	for j := range message {
		message[j] = uint32(j % 16)
	}
	want, err := e.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	for j := range want {
		if want[j] != got[j] {
			t.Fatal("cache output changed")
		}
	}
	roundtrip, err := loaded.MarshalBinary()
	if err != nil || !bytes.Equal(raw, roundtrip) {
		t.Fatal("cache bytes changed", err)
	}
	if _, err := LoadCyclicEncoderWithLimit([]byte(`{"version":99}`), 1<<20); err == nil {
		t.Fatal("malformed cache accepted")
	}
	// An embedded basis can independently exceed the standalone 2 MiB cap.
	marker := []byte(`"basis":{`)
	location := bytes.Index(raw, marker) + len(marker)
	if location < len(marker) {
		t.Fatal("missing embedded basis")
	}
	padded := append(append(append([]byte(nil), raw[:location]...), bytes.Repeat([]byte(" "), 2<<20)...), raw[location:]...)
	if _, err := LoadCyclicEncoder(padded); err == nil {
		t.Fatal("default embedded basis cap changed")
	}
	if _, err := LoadCyclicEncoderWithLimit(padded, int64(len(padded))); err != nil {
		t.Fatal(err)
	}
	// Whitespace makes a valid artifact larger without allocating a large code.
	// The opt-in limit propagates to the embedded basis, while defaults still fail.
	large := append(append([]byte(nil), raw...), bytes.Repeat([]byte(" "), (64<<20)-len(raw)+1)...)
	if _, err := LoadCyclicEncoder(large); err == nil {
		t.Fatal("default cap changed")
	}
	if _, err := LoadCyclicEncoderWithLimit(large, int64(len(large))); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCyclicEncoderLimit(raw, int64(len(raw)), 1); err == nil {
		t.Fatal("embedded basis limit ignored")
	}
}
