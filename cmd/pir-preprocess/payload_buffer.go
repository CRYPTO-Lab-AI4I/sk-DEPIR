package main

import (
	"io"
	"time"
)

// payloadWriteStats is opt-in experiment instrumentation around os.File.Write.
// Counts are application calls, not a kernel trace; hashing/serialization, final
// buffer flush and fsync have separate, potentially overlapping clock scopes.
type payloadWriteStats struct {
	Calls               uint64         `json:"calls"`
	Bytes               uint64         `json:"bytes"`
	Sizes               map[int]uint64 `json:"sizes"`
	FileWriteSeconds    float64        `json:"file_write_seconds"`
	ProduceFlushSeconds float64        `json:"produce_flush_seconds"`
	FlushSeconds        float64        `json:"flush_seconds"`
	SyncSeconds         float64        `json:"sync_seconds"`
	CloseSeconds        float64        `json:"close_seconds"`
}

type payloadCounter struct {
	destination io.Writer
	stats       *payloadWriteStats
}

func (w payloadCounter) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := w.destination.Write(p)
	w.stats.FileWriteSeconds += time.Since(start).Seconds()
	w.stats.Calls++
	w.stats.Bytes += uint64(n)
	if w.stats.Sizes == nil {
		w.stats.Sizes = make(map[int]uint64)
	}
	w.stats.Sizes[n]++
	return n, err
}
