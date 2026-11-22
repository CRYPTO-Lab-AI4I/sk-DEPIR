//go:build liftrs_mpi && cgo

package mpi

import "testing"

// Invalid/closed handles reject before touching MPI; these tests do not initialize
// MPI and can run during a local tagged build. Group isolation is tested in-job.
func TestCommunicatorAdmission(t *testing.T) {
	for _, w := range []*World{nil, {}, {Rank: 0, Size: 2, ChunkWords: 1024}} {
		if _, err := w.Split(0, 0); err == nil {
			t.Fatal("closed split")
		}
		if _, err := w.Counts(1); err == nil {
			t.Fatal("closed counts")
		}
		if err := w.Close(); err == nil {
			t.Fatal("closed release")
		}
		if w.validChunks() {
			t.Fatal("closed transport")
		}
		if err := w.EnableStaging(); err == nil {
			t.Fatal("closed staging")
		}
	}
}
