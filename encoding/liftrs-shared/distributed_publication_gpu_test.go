//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

func samePilotFiles(a, b string) error {
	left, err := os.Open(a)
	if err != nil {
		return err
	}
	defer left.Close()
	right, err := os.Open(b)
	if err != nil {
		return err
	}
	defer right.Close()
	x, y := make([]byte, 16<<10), make([]byte, 16<<10)
	for {
		nx, ex := io.ReadFull(left, x)
		ny, ey := io.ReadFull(right, y)
		if nx != ny || !bytes.Equal(x[:nx], y[:ny]) {
			return fmt.Errorf("artifact byte mismatch")
		}
		if ex == io.EOF || ex == io.ErrUnexpectedEOF {
			if ey != ex {
				return fmt.Errorf("artifact extent")
			}
			return nil
		}
		if ex != nil {
			return ex
		}
		if ey != nil {
			return ey
		}
	}
}

// Only rank0 owns output descriptors. Every node sends just its non-padding
// section in bounded tiles; no full receive/publication array is allocated.
func publishPilotStream[Word transposeWord](t *testing.T, w *mpi.World, directory, label string, h ScalingCodeHeader, root, g []uint32, spectrum []Word, readSpectrum func(uint64, []Word) error, reference *PreparedCode, link operandMPI, spectrumLink typedOperandMPI[Word], operation *uint64) {
	t.Helper()
	var code, generator *os.File
	if w.Rank == 0 {
		var err error
		code, err = os.OpenFile(filepath.Join(directory, label+"-code.bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer code.Close()
		generator, err = os.OpenFile(filepath.Join(directory, label+"-generator.bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer generator.Close()
		if err = writeScalingCodeHeader(code, h); err != nil {
			t.Fatal(err)
		}
	}
	publishPilotValues(t, w, h.N, h.K+1, uint64(1)<<uint(h.Spec.FieldDegree), root, code, link, operation)
	publishPilotValues(t, w, h.N, h.N-h.K, uint64(1)<<uint(h.Spec.FieldDegree), g, generator, link, operation)
	if readSpectrum == nil {
		publishPilotValues(t, w, h.N, h.N, uint64(1)<<uint(2*h.Spec.FieldDegree), spectrum, code, spectrumLink, operation)
	} else {
		publishPilotReadValues(t, w, h.N, uint64(1)<<uint(2*h.Spec.FieldDegree), readSpectrum, code, spectrumLink, operation)
	}
	if w.Rank == 0 {
		if err := code.Sync(); err != nil {
			t.Fatal(err)
		}
		if err := generator.Sync(); err != nil {
			t.Fatal(err)
		}
		if reference != nil {
			expected, err := os.OpenFile(filepath.Join(directory, label+"-reference-code.bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			err = WriteScalingCode(expected, reference.code)
			if err == nil {
				err = expected.Sync()
			}
			closeErr := expected.Close()
			if err != nil || closeErr != nil {
				t.Fatal(err, closeErr)
			}
		}
		if err := samePilotFiles(filepath.Join(directory, label+"-code.bin"), filepath.Join(directory, label+"-reference-code.bin")); err != nil {
			t.Fatal(err)
		}
		if err := samePilotFiles(filepath.Join(directory, label+"-generator.bin"), filepath.Join(directory, label+"-rank-0-reference-g.bin")); err != nil {
			t.Fatal(err)
		}
	}
}

// Base coefficients keep4-byte slots;extension spectra use their canonical Word
// width. This shared section boundary also permits bounded raw-field diagnostics
// without pretending a small transform is a complete full-domain code artifact.
func publishPilotValues[Word transposeWord](t *testing.T, w *mpi.World, domain, count, limit uint64, values []Word, destination io.Writer, link rankValueLink[Word], operation *uint64) {
	t.Helper()
	op := *operation
	*operation = op + 1
	_, words := rankBand(domain, w.Rank, w.Size)
	if err := streamRankValues(rankStreamPlan{op, domain, count, 4096, limit, words, w.Rank, w.Size}, values, destination, link); err != nil {
		t.Fatal(err)
	}
}

// The complete-source boundary shares canonical serialization and operation IDs
// with array publication; only source ownership differs.
func publishPilotReadValues[Word transposeWord](t *testing.T, w *mpi.World, domain, limit uint64, read func(uint64, []Word) error, destination io.Writer, link rankValueLink[Word], operation *uint64) {
	t.Helper()
	op := *operation
	*operation = op + 1
	_, words := rankBand(domain, w.Rank, w.Size)
	if err := streamRankReadValues(rankStreamPlan{op, domain, domain, 4096, limit, words, w.Rank, w.Size}, read, destination, link); err != nil {
		t.Fatal(err)
	}
}

// pilotSpectrumSource downloads every resident spectrum word exactly once.
// Publication reuses each tile for canonical output and this diagnostic file;
// no GPU operation may reset slot0 until Finish has checked and synced it.
type pilotSpectrumSource[Word transposeWord] struct {
	GPU                     *cuda.RankStorage
	File                    *os.File
	Words, Completed, Limit uint64
}

func newPilotSpectrumSource[Word transposeWord](t *testing.T, gpu *cuda.RankStorage, path string, words, limit uint64) *pilotSpectrumSource[Word] {
	t.Helper()
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	return &pilotSpectrumSource[Word]{GPU: gpu, File: file, Words: words, Limit: limit}
}

func (s *pilotSpectrumSource[Word]) Read(offset uint64, tile []Word) error {
	if offset != s.Completed || uint64(len(tile)) > s.Words-s.Completed {
		return fmt.Errorf("spectrum tile sequence/extent")
	}
	if err := storageTransfer(s.GPU, 0, offset, tile, false); err != nil {
		return err
	}
	var err error
	switch values := any(tile).(type) {
	case []uint32:
		err = writeScalingCodeValues(s.File, 4, s.Limit, values, nil)
	case []uint64:
		err = writeScalingCodeValues(s.File, 8, s.Limit, nil, values)
	default:
		return fmt.Errorf("spectrum tile width")
	}
	if err != nil {
		return err
	}
	s.Completed += uint64(len(tile))
	return nil
}

func (s *pilotSpectrumSource[Word]) Finish() error {
	if s.Completed != s.Words {
		return fmt.Errorf("incomplete spectrum download")
	}
	return s.File.Sync()
}
