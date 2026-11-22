package liftrs

import "testing"

func discoveryUnionCases[Word transposeWord](t *testing.T, high Word, limit uint64) {
	t.Helper()
	found := newExponentSet(6)
	found.add(1)
	u := discoveryUnion[Word]{Found: found, Words: 6, Limit: limit}
	if err := u.Consume(0, []Word{0, high, high}); err != nil {
		t.Fatal(err)
	}
	if _, err := u.Complete(); err == nil {
		t.Fatal("partial result accepted")
	}
	if err := u.Consume(3, []Word{0, 0, high}); err != nil {
		t.Fatal(err)
	}
	added, err := u.Complete()
	if err != nil || added != 2 || !found.has(1) || !found.has(2) || !found.has(5) || found.has(0) {
		t.Fatal("union count/membership", added, err)
	}
	// A fresh sample preserves old membership and adds only the new position.
	v := discoveryUnion[Word]{Found: found, Words: 6, Limit: limit}
	if err := v.Consume(0, []Word{high, high, 0, 0, 0, high}); err != nil {
		t.Fatal(err)
	}
	if added, err := v.Complete(); err != nil || added != 1 || !found.has(0) || !found.has(2) {
		t.Fatal("sample union", added, err)
	}
}

func TestDiscoveryUnionCanonicalTiles(t *testing.T) {
	discoveryUnionCases[uint32](t, 1<<15, 1<<16)
	discoveryUnionCases[uint64](t, 1<<35, 1<<36)
}

func TestDiscoveryUnionRejectsBeforeMutation(t *testing.T) {
	for _, input := range []struct {
		offset uint64
		values []uint32
	}{
		{1, []uint32{1}}, {0, nil}, {0, []uint32{1, 0, 1, 0, 1}}, {0, []uint32{1, 16}},
	} {
		u := discoveryUnion[uint32]{Found: newExponentSet(4), Words: 4, Limit: 16}
		if err := u.Consume(input.offset, input.values); err == nil || u.Next != 0 || u.Added != 0 || u.Found[0] != 0 {
			t.Fatal("bad tile changed union")
		}
	}
	u := discoveryUnion[uint32]{Words: 4, Limit: 16}
	if err := u.Consume(0, []uint32{1}); err == nil {
		t.Fatal("short bitmap")
	}
	u.Found = newExponentSet(4)
	if err := u.Consume(0, []uint32{1}); err != nil {
		t.Fatal(err)
	}
	if err := u.Consume(0, []uint32{1}); err == nil {
		t.Fatal("repeated tile")
	}
}

func TestDiscoveryResultMode(t *testing.T) {
	for _, value := range []string{"", "resident", "stream"} {
		mode, err := discoveryResultMode(value)
		if err != nil || (value == "" && mode != "resident") || (value != "" && mode != value) {
			t.Fatal(mode, err)
		}
	}
	if _, err := discoveryResultMode("unknown"); err == nil {
		t.Fatal("unknown mode")
	}
}
