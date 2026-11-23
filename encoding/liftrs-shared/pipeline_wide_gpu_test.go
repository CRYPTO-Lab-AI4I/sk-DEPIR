//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	cpp "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cpp"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
)

// This bounds the field18 adapter workflow independently of the2^36-point code
// grid. Base coefficients/prefixes stay32-bit; all extension I/O is64-bit.
func TestWidePipelineTail(t *testing.T) {
	output := os.Getenv("PIR_GPU_MATRIX_STAGE_OUTPUT")
	if output == "" {
		t.Skip("explicit wide pipeline adapter validation")
	}
	withRankGPUStages(t, output, 4, func(w *mpi.World, emit func(map[string]any)) {
		for _, dim := range []int{5, 12} {
			checkWidePipelineTail(t, w, output, dim, emit)
		}
		emit(map[string]any{"kind": "completed", "cases": 2})
	})
}

func checkWidePipelineTail(t *testing.T, w *mpi.World, output string, dim int, emit func(map[string]any)) {
	t.Helper()
	const s = 18
	n := uint64(1) << uint(dim)
	label := fmt.Sprintf("s18-d%d", dim)
	inputMode, inputErr := newtonInputMode(os.Getenv("PIR_PIPELINE_NEWTON_INPUT"))
	if inputErr != nil {
		t.Fatal(inputErr)
	}
	newtonUploads, newtonCloneWords := uint64(0), uint64(0)
	newtonSteps := 0
	count, target := 17, 64
	if dim == 5 {
		count, target = 9, 16
	}
	h, _, err := canonicalScalingCodeHeader(s, 2, 1, 1<<36)
	if err != nil {
		t.Fatal(err)
	}
	base, err := NewBinaryField(s, h.Polynomial)
	if err != nil {
		t.Fatal(err)
	}
	tables := newBoundedPreparationProducts(base, 18)
	p := make([]uint32, count)
	p[0] = 1
	for j := 1; j < count; j++ {
		p[j] = uint32((j+1)*0x1337) & ((1 << s) - 1)
	}
	// Independent coefficient recurrence,not the Newton schedule under test.
	expected := make([]uint32, target)
	expected[0] = 1
	for k := 1; k < target; k++ {
		for j := 1; j < count && j <= k; j++ {
			expected[k] ^= base.Mul(p[j], expected[k-j])
		}
	}
	first, words := rankBand(n, w.Rank, w.Size)
	own := func(full []uint32) []uint32 {
		out := make([]uint32, words)
		for j := range out {
			if k := first + uint64(j); k < uint64(len(full)) {
				out[j] = full[k]
			}
		}
		return out
	}
	record := func(row map[string]any) {
		// Keep the wide oracle's established event vocabulary while sharing all
		// execution, transport, operation identities and progress with runtime.
		if row["kind"] == "executor-call" {
			row["kind"] = "tail-call"
			delete(row, "matches")
		}
		emit(row)
	}
	session, err := newRankGPUSession[uint64](w, cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 64 << 20}, dim, label, record)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	}()
	e := &session.Executor
	gpu := e.GPU
	proof := func(width int, direction string, p operandPacket, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "word_bytes": width, "packet": p, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}
	link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
		proof(4, direction, p, own, remote)
	}}
	wide := typedOperandMPI[uint64]{world: w, record: func(direction string, p operandPacket, data []uint64, own, remote [64]byte) {
		proof(8, direction, p, own, remote)
	}}
	product := func(name string, a, b []uint32, ac, bc, keep uint64, prefix bool) []uint32 {
		run := session.Product
		if prefix {
			run = session.ProductPrefixInputs
		}
		result, err := run(name, a, b, ac, bc, keep)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	root := own(p)
	g := own([]uint32{1})
	known := 1
	err = reciprocalSchedule(1, target, func(size int) error {
		op := e.Plan.Operation
		e.Plan.Operation++
		squared, err := squarePolynomialShard(rankSquarePlan{op, n, 4096, 1 << s, uint64(size), words, w.Rank, w.Size}, g, func(v uint32) uint32 { return base.Mul(v, v) }, link, nil)
		if err != nil {
			return err
		}
		ac := min(count, size)
		left := root
		if inputMode == "dense" {
			left = append([]uint32(nil), root...)
			newtonCloneWords = uint64(len(left))
			for j := range left {
				if first+uint64(j) >= uint64(ac) {
					left[j] = 0
				}
			}
		}
		before := e.UploadedInputWords
		g = product(fmt.Sprintf("newton-%d", size), left, squared, uint64(ac), uint64(min(2*known-1, size)), uint64(size), inputMode == "prefix")
		newtonUploads += e.UploadedInputWords - before
		newtonSteps++
		if !reflect.DeepEqual(g, own(expected[:size])) {
			return fmt.Errorf("wide independent reciprocal prefix")
		}
		saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-prefix-%d.bin", label, w.Rank, size)), g)
		known = size
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	identity := product("identity", root, g, uint64(count), uint64(target), uint64(count+target-1), false)
	convolution := make([]uint32, count+target-1)
	for i, a := range p {
		for j, b := range expected {
			convolution[i+j] ^= base.Mul(a, b)
		}
	}
	if !reflect.DeepEqual(identity, own(convolution)) {
		t.Fatal("wide full nonaliasing convolution")
	}
	for j, v := range identity {
		k := first + uint64(j)
		if k < uint64(target) {
			var want uint32
			if k == 0 {
				want = 1
			}
			if v != want {
				t.Fatal("wide reciprocal identity prefix")
			}
		}
	}
	saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-identity.bin", label, w.Rank)), identity)
	spectrumMode := os.Getenv("PIR_PIPELINE_SPECTRUM")
	if spectrumMode == "" {
		spectrumMode = "resident"
	}
	var spectrum []uint64
	var source *pilotSpectrumSource[uint64]
	if spectrumMode == "stream" {
		if err := session.PrepareSpectrum("spectrum", g, uint64(target)); err != nil {
			t.Fatal(err)
		}
		source = newPilotSpectrumSource[uint64](t, gpu, filepath.Join(output, fmt.Sprintf("%s-rank-%d-spectrum.bin", label, w.Rank)), words, 1<<36)
		defer source.File.Close()
	} else if spectrumMode == "resident" {
		spectrum, err = session.Spectrum("spectrum", g, uint64(target))
		if err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal("wide spectrum output mode")
	}
	matrix, err := cpp.NewWideMatrix(s, h.Polynomial, dim, dim/2)
	if err != nil {
		t.Fatal(err)
	}
	native := make([]uint64, n)
	scratch := make([]uint64, n)
	for j, v := range expected {
		native[j] = uint64(v)
	}
	if _, err = matrix.Apply(native, scratch, 1, false); err != nil {
		t.Fatal(err)
	}
	if err = matrix.Close(); err != nil {
		t.Fatal(err)
	}
	if source == nil && !reflect.DeepEqual(spectrum, native[first:first+words]) {
		t.Fatal("wide spectrum/native FFT")
	}
	high := make([]uint64, words)
	for j := range high {
		high[j] = ((uint64(1) << 35) ^ ((first + uint64(j) + 1) * 0x9e3779b9)) & ((uint64(1) << 36) - 1)
	}
	outputs := map[string][]uint64{"expected-spectrum": native[first : first+words], "high": high}
	if source == nil {
		outputs["spectrum"] = spectrum
	}
	for suffix, values := range outputs {
		saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-%s.bin", label, w.Rank, suffix)), values)
	}
	saveMatrixStageWords(t, filepath.Join(output, fmt.Sprintf("%s-rank-%d-generator.bin", label, w.Rank)), g)
	publish := func(name string, fn func(*os.File)) {
		var file *os.File
		if w.Rank == 0 {
			var err error
			file, err = os.OpenFile(filepath.Join(output, label+"-"+name+".bin"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}
		fn(file)
		if file != nil {
			if err := file.Sync(); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	publish("stream-generator", func(file *os.File) {
		publishPilotValues(t, w, n, uint64(target), 1<<s, g, file, link, &e.Plan.Operation)
	})
	publish("stream-spectrum", func(file *os.File) {
		if source == nil {
			publishPilotValues(t, w, n, n, 1<<36, spectrum, file, wide, &e.Plan.Operation)
			return
		}
		read := func(offset uint64, tile []uint64) error {
			if err := source.Read(offset, tile); err != nil {
				return err
			}
			if !reflect.DeepEqual(tile, native[first+offset:first+offset+uint64(len(tile))]) {
				return fmt.Errorf("wide spectrum source/native FFT")
			}
			return nil
		}
		publishPilotReadValues(t, w, n, 1<<36, read, file, wide, &e.Plan.Operation)
		if err := source.Finish(); err != nil {
			t.Fatal(err)
		}
	})
	highDownloaded := uint64(0)
	publish("stream-high", func(file *os.File) {
		if source == nil {
			publishPilotValues(t, w, n, n, 1<<36, high, file, wide, &e.Plan.Operation)
			return
		}
		// These small transforms lie within Fq,so their actual spectra need not have
		// high extension bits. Exercise those bits with a separate GPU upload/read
		// fixture after spectrum publication;this is validation-only transfer work.
		for offset := uint64(0); offset < words; {
			take := min(uint64(4096), words-offset)
			if err := storageTransfer(gpu, 0, offset, high[offset:offset+take], true); err != nil {
				t.Fatal(err)
			}
			offset += take
		}
		highSource := newPilotSpectrumSource[uint64](t, gpu, filepath.Join(output, fmt.Sprintf("%s-rank-%d-high-source.bin", label, w.Rank)), words, 1<<36)
		defer highSource.File.Close()
		publishPilotReadValues(t, w, n, 1<<36, highSource.Read, file, wide, &e.Plan.Operation)
		if err := highSource.Finish(); err != nil {
			t.Fatal(err)
		}
		highDownloaded = highSource.Completed
	})
	downloaded := uint64(len(spectrum))
	if source != nil {
		downloaded = source.Completed
	}
	emit(map[string]any{"kind": "wide-tail", "newton_input_mode": inputMode, "newton_root_clone_words_per_step": newtonCloneWords, "newton_input_upload_words": newtonUploads, "newton_steps": newtonSteps, "high_source_downloaded_words": highDownloaded, "spectrum_mode": spectrumMode, "spectrum_resident_host_words": len(spectrum), "spectrum_downloaded_words": downloaded, "case": label, "s": s, "dimension": dim, "n": n, "target": target, "root_count": count, "first": first, "words": words, "slot_bytes": 8, "coefficient_bytes": 4, "owned_device_bytes": gpu.OwnedBytes(), "matches": true})
}
