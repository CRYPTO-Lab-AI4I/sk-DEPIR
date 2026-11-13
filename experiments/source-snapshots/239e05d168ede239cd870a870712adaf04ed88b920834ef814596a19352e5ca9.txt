//go:build liftrs_cpp && cgo

package liftrs

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// Preserve the earlier cached worker's CPU packing boundary for the matched
// comparator, but pack every repetition so both methods produce one file/call.
func writeCanonicalPacked(destination io.Writer, word []uint32, s int) (uint64, string, error) {
	hash := sha256.New()
	writer := bufio.NewWriterSize(io.MultiWriter(destination, hash), 1<<20)
	var acc uint64
	var used uint
	var count uint64
	for _, x := range word {
		acc |= uint64(x) << used
		used += uint(s)
		for used >= 8 {
			if err := writer.WriteByte(byte(acc)); err != nil {
				return count, "", err
			}
			acc >>= 8
			used -= 8
			count++
		}
	}
	if used != 0 {
		if err := writer.WriteByte(byte(acc)); err != nil {
			return count, "", err
		}
		count++
	}
	if err := writer.Flush(); err != nil {
		return count, "", err
	}
	return count, hex.EncodeToString(hash.Sum(nil)), nil
}
