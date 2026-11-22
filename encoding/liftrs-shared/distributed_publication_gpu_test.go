//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

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
func publishPilotStream(t *testing.T, w *mpi.World, directory, label string, h ScalingCodeHeader, root, g, spectrum []uint32, reference *PreparedCode, link operandMPI, operation *uint64) {
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
	section := func(values []uint32, count, limit uint64, destination io.Writer) {
		op := *operation
		*operation = op + 1
		_, words := rankBand(h.N, w.Rank, w.Size)
		if err := streamRankValues(rankStreamPlan{op, h.N, count, 4096, limit, words, w.Rank, w.Size}, values, destination, link); err != nil {
			t.Fatal(err)
		}
	}
	section(root, h.K+1, uint64(1)<<uint(h.Spec.FieldDegree), code)
	section(g, h.N-h.K, uint64(1)<<uint(h.Spec.FieldDegree), generator)
	section(spectrum, h.N, uint64(1)<<uint(2*h.Spec.FieldDegree), code)
	if w.Rank == 0 {
		if err := code.Sync(); err != nil {
			t.Fatal(err)
		}
		if err := generator.Sync(); err != nil {
			t.Fatal(err)
		}
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
		if err = samePilotFiles(filepath.Join(directory, label+"-code.bin"), filepath.Join(directory, label+"-reference-code.bin")); err != nil {
			t.Fatal(err)
		}
		if err = samePilotFiles(filepath.Join(directory, label+"-generator.bin"), filepath.Join(directory, label+"-rank-0-reference-g.bin")); err != nil {
			t.Fatal(err)
		}
	}
}
