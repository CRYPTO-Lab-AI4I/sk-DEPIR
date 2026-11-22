//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"testing"
)

// pipelineGPU shares the same workspace,operation identities and proof logging
// for compact root subgroups and the later full-domain pipeline. It owns no
// support/native-reference arrays. Word describes extension slots,not base P/G.
type pipelineGPU[Word transposeWord] struct {
	Executor     rankGPUExecutor[Word]
	t            *testing.T
	label, call  string
	start, span  uint64
	pass, s, dim int
	emit         func(map[string]any)
}

func openPipelineGPU[Word transposeWord](t *testing.T, w *mpi.World, options cuda.PreparationOptions, dim int, label string, emit func(map[string]any)) *pipelineGPU[Word] {
	t.Helper()
	rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
	p, err := newRankTransposePlan(1, rows, cols, w.Rank, w.Size, rows*cols, 4096)
	if err != nil {
		t.Fatal(err)
	}
	options.MaxElements = rows * cols
	gpu, err := cuda.NewRankStorage(options, rows, cols, w.Rank, w.Size, 4096)
	if err != nil {
		t.Fatal(err)
	}
	d := &pipelineGPU[Word]{t: t, label: label, s: options.S, dim: dim, span: uint64(2 + 2*(dim-dim/2)), emit: emit}
	e := &d.Executor
	*e = rankGPUExecutor[Word]{GPU: gpu, Plan: p, Capacity: min(rows*cols, 3*max(rows, cols)), Limit: uint64(1) << uint(options.S)}
	e.Notice = func(phase string, done, total uint64) {
		if done == 0 {
			switch phase {
			case "rank-product-left-forward":
				d.pass = 0
			case "rank-product-right-forward":
				d.pass = 1
			case "rank-product-inverse":
				d.pass = 2
			}
		}
		if done == 0 || done == total {
			emit(map[string]any{"kind": "progress", "case": label, "call": d.call, "phase": phase, "pass": d.pass, "completed": done, "total": total})
		}
	}
	e.Transpose = transposeMPI[Word]{world: w, rank: w.Rank, record: func(direction string, peer int, p transposePacket, data []Word, own, remote [64]byte) {
		emit(map[string]any{"kind": "tile-proof", "case": d.call, "pass": d.pass, "operation": p.Operation, "round": p.Round, "offset": p.Offset, "words": p.Words, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	e.Windows = taylorMPI[Word]{world: w, record: func(direction string, peer int, p taylorPacket, data []Word, own, remote [64]byte) {
		emit(map[string]any{"kind": "taylor-proof", "case": d.call, "pass": d.pass, "packet": p, "peer": peer, "direction": direction, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	e.Degrees = func(level int) ([]uint64, error) {
		degrees, err := gpu.TaylorDegrees(level)
		if err == nil {
			emit(map[string]any{"kind": "taylor-plan", "case": d.call, "pass": d.pass, "operation_start": d.start + uint64(d.pass)*d.span, "level": level, "degrees": degrees, "capacity": e.Capacity})
		}
		return degrees, err
	}
	return d
}

func (d *pipelineGPU[Word]) begin(name string) {
	d.call, d.start, d.pass = d.label+"/"+name, d.Executor.Plan.Operation, 0
}
func (d *pipelineGPU[Word]) productDone(ac, bc, keep uint64) {
	d.emit(map[string]any{"kind": "executor-call", "case": d.call, "s": d.s, "dimension": d.dim, "operation_start": d.start, "mode": "product", "a_count": ac, "b_count": bc, "keep": keep, "matches": true})
}
func (d *pipelineGPU[Word]) Product(name string, a, b []uint32, ac, bc, keep uint64) []uint32 {
	d.t.Helper()
	d.begin(name)
	result, err := d.Executor.Product(a, b, ac, bc, keep)
	if err != nil {
		d.t.Fatal(name, err)
	}
	d.productDone(ac, bc, keep)
	return result
}
func (d *pipelineGPU[Word]) Spectrum(name string, g []uint32, count uint64) []Word {
	d.t.Helper()
	d.begin(name)
	result, err := d.Executor.Spectrum(g, count)
	if err != nil {
		d.t.Fatal(name, err)
	}
	d.spectrumDone()
	return result
}

// PrepareSpectrum keeps slot0 live for tile publication instead of materializing
// the full host spectrum. Operation IDs and completion proofs match Spectrum.
func (d *pipelineGPU[Word]) PrepareSpectrum(name string, g []uint32, count uint64) {
	d.t.Helper()
	d.begin(name)
	if err := d.Executor.PrepareSpectrum(g, count); err != nil {
		d.t.Fatal(err)
	}
	d.spectrumDone()
}

func (d *pipelineGPU[Word]) spectrumDone() {
	d.emit(map[string]any{"kind": "executor-call", "case": d.call, "s": d.s, "dimension": d.dim, "operation_start": d.start, "mode": "spectrum", "keep": d.Executor.Plan.Rows * d.Executor.Plan.Columns, "matches": true})
}
