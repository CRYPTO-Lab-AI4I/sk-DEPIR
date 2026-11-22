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
	rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
	label := fmt.Sprintf("s18-d%d", dim)
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
	gpu, err := cuda.NewRankStorage(cuda.PreparationOptions{Path: os.Getenv("PIR_GPU_PREPARATION_LIBRARY"), S: s, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: tables.log, Powers: tables.powers, Devices: []int{0}, MaxElements: n, DeviceBytes: 64 << 20}, rows, cols, w.Rank, w.Size, 4096)
	if err != nil {
		t.Fatal(err)
	}
	defer gpu.Close()
	plan, err := newRankTransposePlan(1, rows, cols, w.Rank, w.Size, n, 4096)
	if err != nil {
		t.Fatal(err)
	}
	capacity := min(n, 3*max(rows, cols))
	call := label
	pass := 0
	start := uint64(1)
	span := uint64(2 + 2*(dim-dim/2))
	e := rankGPUExecutor[uint64]{GPU: gpu, Plan: plan, Capacity: capacity, Limit: 1 << s}
	e.Notice = func(phase string, done, total uint64) {
		if done == 0 {
			switch phase {
			case "rank-product-left-forward":
				pass = 0
			case "rank-product-right-forward":
				pass = 1
			case "rank-product-inverse":
				pass = 2
			}
		}
		if done == 0 || done == total {
			emit(map[string]any{"kind": "progress", "case": call, "pass": pass, "phase": phase, "completed": done, "total": total})
		}
	}
	e.Transpose = transposeMPI[uint64]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, data []uint64, own, remote [64]byte) {
		emit(map[string]any{"kind": "tile-proof", "case": call, "pass": pass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	e.Windows = taylorMPI[uint64]{world: w, record: func(direction string, peer int, p taylorPacket, data []uint64, own, remote [64]byte) {
		emit(map[string]any{"kind": "taylor-proof", "case": call, "pass": pass, "packet": p, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	e.Degrees = func(level int) ([]uint64, error) {
		degrees, err := gpu.TaylorDegrees(level)
		if err == nil {
			emit(map[string]any{"kind": "taylor-plan", "case": call, "pass": pass, "operation_start": start + uint64(pass)*span, "level": level, "degrees": degrees, "capacity": capacity})
		}
		return degrees, err
	}
	proof := func(width int, direction string, p operandPacket, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "word_bytes": width, "packet": p, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}
	link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
		proof(4, direction, p, own, remote)
	}}
	wide := typedOperandMPI[uint64]{world: w, record: func(direction string, p operandPacket, data []uint64, own, remote [64]byte) {
		proof(8, direction, p, own, remote)
	}}
	product := func(name string, a, b []uint32, ac, bc, keep uint64) []uint32 {
		call = label + "/" + name
		start = e.Plan.Operation
		pass = 0
		result, err := e.Product(a, b, ac, bc, keep)
		if err != nil {
			t.Fatal(err)
		}
		emit(map[string]any{"kind": "tail-call", "case": call, "s": s, "dimension": dim, "operation_start": start, "mode": "product", "a_count": ac, "b_count": bc, "keep": keep})
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
		left := append([]uint32(nil), root...)
		for j := range left {
			if first+uint64(j) >= uint64(ac) {
				left[j] = 0
			}
		}
		g = product(fmt.Sprintf("newton-%d", size), left, squared, uint64(ac), uint64(min(2*known-1, size)), uint64(size))
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
	identity := product("identity", root, g, uint64(count), uint64(target), uint64(count+target-1))
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
	call = label + "/spectrum"
	start = e.Plan.Operation
	pass = 0
	spectrum, err := e.Spectrum(g, uint64(target))
	if err != nil {
		t.Fatal(err)
	}
	emit(map[string]any{"kind": "tail-call", "case": call, "s": s, "dimension": dim, "operation_start": start, "mode": "spectrum", "keep": n})
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
	if !reflect.DeepEqual(spectrum, native[first:first+words]) {
		t.Fatal("wide spectrum/native FFT")
	}
	high := make([]uint64, words)
	for j := range high {
		high[j] = ((uint64(1) << 35) ^ ((first + uint64(j) + 1) * 0x9e3779b9)) & ((uint64(1) << 36) - 1)
	}
	for suffix, values := range map[string][]uint64{"spectrum": spectrum, "expected-spectrum": native[first : first+words], "high": high} {
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
	publish("stream-spectrum", func(file *os.File) { publishPilotValues(t, w, n, n, 1<<36, spectrum, file, wide, &e.Plan.Operation) })
	publish("stream-high", func(file *os.File) { publishPilotValues(t, w, n, n, 1<<36, high, file, wide, &e.Plan.Operation) })
	emit(map[string]any{"kind": "wide-tail", "case": label, "s": s, "dimension": dim, "n": n, "target": target, "root_count": count, "first": first, "words": words, "slot_bytes": 8, "coefficient_bytes": 4, "owned_device_bytes": gpu.OwnedBytes(), "matches": true})
}
