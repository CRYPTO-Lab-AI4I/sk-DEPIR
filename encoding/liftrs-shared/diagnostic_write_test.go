package liftrs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

type diagnosticSink struct {
	bytes.Buffer
	largest int
	fail    error
	short   bool
}

func (w *diagnosticSink) Write(b []byte) (int, error) {
	w.largest = max(w.largest, len(b))
	if w.fail != nil {
		return 0, w.fail
	}
	if w.short {
		return len(b) - 1, nil
	}
	return w.Buffer.Write(b)
}
func checkDiagnosticWords[Word transposeWord](t *testing.T, values []Word) {
	t.Helper()
	var expected bytes.Buffer
	if err := binary.Write(&expected, binary.LittleEndian, values); err != nil {
		t.Fatal(err)
	}
	sink := &diagnosticSink{}
	last, calls := uint64(0), 0
	err := writeDiagnosticWords(sink, values, func(done, total uint64) {
		if done < last || done > total || total != uint64(expected.Len()) {
			t.Fatal("progress")
		}
		last = done
		calls++
	})
	if err != nil || !bytes.Equal(sink.Bytes(), expected.Bytes()) || sink.largest > 65536 || last != uint64(expected.Len()) || calls == 0 {
		t.Fatal("bounded canonical bytes", err)
	}
	if len(values) > 16384 && calls < 3 {
		t.Fatal("not chunked")
	}
}
func TestDiagnosticWrite(t *testing.T) {
	narrow := make([]uint32, 40001)
	wide := make([]uint64, 20003)
	for i := range narrow {
		narrow[i] = ^uint32(i)
	}
	for i := range wide {
		wide[i] = ^uint64(i)
	}
	checkDiagnosticWords(t, narrow)
	checkDiagnosticWords(t, wide)
	checkDiagnosticWords(t, []uint32{})
	checkDiagnosticWords(t, []uint64{1, 1 << 63, ^uint64(0)})
	for _, mode := range []string{"", "full", "bounded"} {
		bounded, err := diagnosticWriteMode(mode)
		if err != nil || bounded != (mode == "bounded") {
			t.Fatal(mode, err)
		}
	}
	if _, err := diagnosticWriteMode("typo"); err == nil {
		t.Fatal("mode accepted")
	}
	failure := errors.New("sink failure")
	if err := writeDiagnosticWords(&diagnosticSink{fail: failure}, wide, nil); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if err := writeDiagnosticWords(&diagnosticSink{short: true}, narrow, nil); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}

// This isolates serialization CPU/allocation cost; disk/sync and progress overhead
// remain in the GPU composition experiment's matched arms.
func BenchmarkDiagnosticWrite(b *testing.B) {
	narrow := make([]uint32, 1<<20)
	wide := make([]uint64, 1<<20)
	for _, bounded := range []bool{false, true} {
		mode := "full"
		if bounded {
			mode = "bounded"
		}
		b.Run(mode+"/uint32", func(b *testing.B) {
			b.SetBytes(int64(len(narrow)) * 4)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if bounded {
					err = writeDiagnosticWords(io.Discard, narrow, nil)
				} else {
					err = binary.Write(io.Discard, binary.LittleEndian, narrow)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(mode+"/uint64", func(b *testing.B) {
			b.SetBytes(int64(len(wide)) * 8)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if bounded {
					err = writeDiagnosticWords(io.Discard, wide, nil)
				} else {
					err = binary.Write(io.Discard, binary.LittleEndian, wide)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
