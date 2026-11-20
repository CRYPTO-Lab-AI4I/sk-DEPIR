//go:build liftrs_mpi && cgo

package mpi

import (
	"crypto/sha256"
	"encoding/binary"
)

// PayloadFingerprint hashes raw canonical-width little-endian storage, including
// high bits of invalid data before validation. Scratch is bounded independently
// of payload size. This is diagnostic integrity, not authentication.
func PayloadFingerprint[Word uint32 | uint64](values []Word) (result [32]byte) {
	h := sha256.New()
	var block [1024]byte
	var zero Word
	width := 8
	if _, ok := any(zero).(uint32); ok {
		width = 4
	}
	for first := 0; first < len(values); {
		take := min(len(values)-first, len(block)/width)
		for j := 0; j < take; j++ {
			if width == 4 {
				binary.LittleEndian.PutUint32(block[j*width:], uint32(values[first+j]))
			} else {
				binary.LittleEndian.PutUint64(block[j*width:], uint64(values[first+j]))
			}
		}
		_, _ = h.Write(block[:take*width])
		first += take
	}
	copy(result[:], h.Sum(nil))
	return
}
