package liftrs

import (
	"encoding/binary"
	"fmt"
	"io"
)

const rankArchiveTile = 65536

// Counts use bounded pages of eight integers. Only the coordinator contributes;
// every peer checks zero padding, ownership and the final total before payloads.
func archiveCountPage(counts []uint64, page int) [64]byte {
	var frame [64]byte
	for j := 0; j < 8 && page*8+j < len(counts); j++ {
		binary.LittleEndian.PutUint64(frame[8*j:], counts[page*8+j])
	}
	return frame
}
func readArchiveCountPage(frames [][64]byte, ranks, page int, limit uint64) ([]uint64, error) {
	if ranks < 2 || ranks > 64 || len(frames) != ranks || page < 0 || page >= (ranks+7)/8 {
		return nil, fmt.Errorf("archive count ownership")
	}
	for rank, frame := range frames {
		if rank > 0 && frame != [64]byte{} {
			return nil, fmt.Errorf("archive count contributor")
		}
	}
	values := make([]uint64, min(8, ranks-page*8))
	for j := 0; j < 8; j++ {
		v := binary.LittleEndian.Uint64(frames[0][8*j:])
		if j >= len(values) {
			if v != 0 {
				return nil, fmt.Errorf("archive count padding")
			}
		} else {
			if v > limit {
				return nil, fmt.Errorf("archive count extent")
			}
			values[j] = v
		}
	}
	return values, nil
}

// Decode one complete canonical tile,retaining order across tile boundaries.
func readArchiveValues(reader io.Reader, domain uint64, previous *uint64, seen bool, data []byte, values []uint64) error {
	if reader == nil || previous == nil || domain < 2 || len(values) == 0 || len(values) > rankArchiveTile || len(data) != 8*len(values) {
		return fmt.Errorf("archive tile admission")
	}
	if _, err := io.ReadFull(reader, data); err != nil {
		return err
	}
	for j := range values {
		v := binary.LittleEndian.Uint64(data[8*j:])
		if v >= domain-1 || ((seen || j > 0) && v <= *previous) {
			return fmt.Errorf("cached discovery order/range")
		}
		*previous = v
		values[j] = v
	}
	return nil
}
