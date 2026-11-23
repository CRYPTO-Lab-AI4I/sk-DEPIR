package liftrs

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestArchiveCountPages(t *testing.T) {
	counts := make([]uint64, 16)
	for i := range counts {
		counts[i] = uint64(i)
	}
	var got []uint64
	for page := 0; page < 2; page++ {
		frames := make([][64]byte, 16)
		frames[0] = archiveCountPage(counts, page)
		values, err := readArchiveCountPage(frames, 16, page, 16)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, values...)
		frames[3][0] = 1
		if _, err = readArchiveCountPage(frames, 16, page, 16); err == nil {
			t.Fatal("noncoordinator counts")
		}
	}
	for i, v := range got {
		if v != counts[i] {
			t.Fatal("page order")
		}
	}
	frames := make([][64]byte, 2)
	frames[0] = archiveCountPage([]uint64{1, 2, 3}, 0)
	if _, err := readArchiveCountPage(frames, 2, 0, 8); err == nil {
		t.Fatal("nonzero padding")
	}
	frames[0] = archiveCountPage([]uint64{9, 0}, 0)
	if _, err := readArchiveCountPage(frames, 2, 0, 8); err == nil {
		t.Fatal("count extent")
	}
}

func TestArchiveValues(t *testing.T) {
	var previous uint64
	encode := func(values ...uint64) []byte {
		var b bytes.Buffer
		for _, v := range values {
			_ = binary.Write(&b, binary.LittleEndian, v)
		}
		return b.Bytes()
	}
	read := func(data []byte, seen bool) error {
		return readArchiveValues(bytes.NewReader(data), 16, &previous, seen, make([]byte, 16), make([]uint64, 2))
	}
	if err := read(encode(0, 7), false); err != nil {
		t.Fatal(err)
	}
	if err := read(encode(8, 14), true); err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{encode(14, 15), encode(7, 8), encode(15, 16), encode(1)} {
		if err := read(data, true); err == nil {
			t.Fatal("invalid order/range/extent")
		}
	}
}
