package liftrs

import (
	"encoding/binary"
	"fmt"
	"io"
)

// diagnosticWriteMode preserves bulk writes unless bounded serialization was
// explicitly selected. File layout and retained bytes are identical in both modes.
func diagnosticWriteMode(mode string) (bool, error) {
	switch mode {
	case "", "full":
		return false, nil
	case "bounded":
		return true, nil
	default:
		return false, fmt.Errorf("diagnostic write mode")
	}
}

// writeDiagnosticWords bounds the extra serialization buffer to 64 KiB. The
// standard endian primitives preserve the full uint32/uint64 range and little-endian layout;
// callers retain file creation, durability, resource admission and progress logs.
// report receives completed bytes after each successful write (and at start).
func writeDiagnosticWords[Word transposeWord](w io.Writer, values []Word, report func(uint64, uint64)) error {
	var zero Word
	width := 4
	if _, wide := any(zero).(uint64); wide {
		width = 8
	}
	block := make([]byte, min(65536/width, len(values))*width)
	total := uint64(len(values)) * uint64(width)
	if report != nil {
		report(0, total)
	}
	for first := 0; first < len(values); {
		count := min(len(block)/width, len(values)-first)
		n := count * width
		for j, value := range values[first : first+count] {
			if width == 4 {
				binary.LittleEndian.PutUint32(block[4*j:], uint32(value))
			} else {
				binary.LittleEndian.PutUint64(block[8*j:], uint64(value))
			}
		}
		written, err := w.Write(block[:n])
		if err != nil {
			return err
		}
		if written != n {
			return io.ErrShortWrite
		}
		first += count
		if report != nil {
			report(uint64(first)*uint64(width), total)
		}
	}
	return nil
}
