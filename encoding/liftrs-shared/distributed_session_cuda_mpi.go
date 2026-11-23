//go:build linux && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"time"
)

// rankGPUSession is internal server runtime: no testing framework or oracle data.
// Callers serialize session operations and Close on their owning rank.
// It shares the same workspace,operation identities and proof logging
// for compact root subgroups and the later full-domain pipeline. It owns no
// support/native-reference arrays. Word describes extension slots,not base P/G.
type rankGPUSession[Word transposeWord] struct {
	Executor       rankGPUExecutor[Word]
	closed, failed bool
	label, call    string
	start, span    uint64
	pass, s, dim   int
	emit           func(map[string]any)
}

func newRankGPUSession[Word transposeWord](w *mpi.World, options cuda.PreparationOptions, dim int, label string, emit func(map[string]any)) (*rankGPUSession[Word], error) {
	if w == nil || w.Size < 2 || w.Size > 64 || w.Size&(w.Size-1) != 0 || w.Rank < 0 || w.Rank >= w.Size || options.S < 2 || options.S > 18 || dim < 2 || dim > 2*options.S || label == "" || emit == nil {
		return nil, fmt.Errorf("GPU session admission")
	}
	// RankStorage uses four-byte extension slots through s16,then eight-byte
	// slots. Reject a generic instantiation that disagrees before GPU allocation.
	var zero Word
	_, wide := any(zero).(uint64)
	if wide != (options.S > 16) {
		return nil, fmt.Errorf("GPU session extension slot width")
	}
	rows, cols := uint64(1)<<uint(dim-dim/2), uint64(1)<<uint(dim/2)
	p, err := newRankTransposePlan(1, rows, cols, w.Rank, w.Size, rows*cols, 4096)
	if err != nil {
		return nil, err
	}
	options.MaxElements = rows * cols
	gpu, err := cuda.NewRankStorage(options, rows, cols, w.Rank, w.Size, 4096)
	if err != nil {
		return nil, err
	}
	d := &rankGPUSession[Word]{label: label, s: options.S, dim: dim, span: uint64(2 + 2*(dim-dim/2)), emit: emit}
	e := &d.Executor
	*e = rankGPUExecutor[Word]{GPU: gpu, Plan: p, Capacity: min(rows*cols, 3*max(rows, cols)), Limit: uint64(1) << uint(options.S)}
	lastNotice := map[string]time.Time{}
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
		key := d.call + "/" + phase
		if done == 0 || done == total || time.Since(lastNotice[key]) >= 30*time.Second {
			lastNotice[key] = time.Now()
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
	return d, nil
}

func (d *rankGPUSession[Word]) begin(name string) {
	d.call, d.start, d.pass = d.label+"/"+name, d.Executor.Plan.Operation, 0
}
func (d *rankGPUSession[Word]) productDone(ac, bc, keep uint64) {
	d.emit(map[string]any{"kind": "executor-call", "case": d.call, "s": d.s, "dimension": d.dim, "operation_start": d.start, "mode": "product", "a_count": ac, "b_count": bc, "keep": keep, "matches": true})
}
func (d *rankGPUSession[Word]) Product(name string, a, b []uint32, ac, bc, keep uint64) ([]uint32, error) {
	return d.product(name, a, b, ac, bc, keep, false)
}

func (d *rankGPUSession[Word]) ProductPrefixInputs(name string, a, b []uint32, ac, bc, keep uint64) ([]uint32, error) {
	return d.product(name, a, b, ac, bc, keep, true)
}

func (d *rankGPUSession[Word]) product(name string, a, b []uint32, ac, bc, keep uint64, prefix bool) ([]uint32, error) {
	if err := d.usable(); err != nil {
		return nil, err
	}
	d.begin(name)
	run := d.Executor.Product
	if prefix {
		run = d.Executor.ProductPrefixInputs
	}
	result, err := run(a, b, ac, bc, keep)
	if err != nil {
		d.failed = true
		return nil, err
	}
	d.productDone(ac, bc, keep)
	return result, nil
}
func (d *rankGPUSession[Word]) Spectrum(name string, g []uint32, count uint64) ([]Word, error) {
	if err := d.usable(); err != nil {
		return nil, err
	}
	d.begin(name)
	result, err := d.Executor.Spectrum(g, count)
	if err != nil {
		d.failed = true
		return nil, err
	}
	d.spectrumDone()
	return result, nil
}

// PrepareSpectrum keeps slot0 live for tile publication instead of materializing
// the full host spectrum. Operation IDs and completion proofs match Spectrum.
func (d *rankGPUSession[Word]) PrepareSpectrum(name string, g []uint32, count uint64) error {
	if err := d.usable(); err != nil {
		return err
	}
	d.begin(name)
	if err := d.Executor.PrepareSpectrum(g, count); err != nil {
		d.failed = true
		return err
	}
	d.spectrumDone()
	return nil
}

func (d *rankGPUSession[Word]) spectrumDone() {
	d.emit(map[string]any{"kind": "executor-call", "case": d.call, "s": d.s, "dimension": d.dim, "operation_start": d.start, "mode": "spectrum", "keep": d.Executor.Plan.Rows * d.Executor.Plan.Columns, "matches": true})
}

// Sessions own GPU storage, not MPI communicators. Close is safe after errors and
// repeat calls; collective callers must abort peers when an operation fails.
func (d *rankGPUSession[Word]) Close() error {
	if d == nil || d.closed {
		return nil
	}
	d.closed = true
	return d.Executor.GPU.Close()
}
func (d *rankGPUSession[Word]) usable() error {
	if d == nil || d.closed || d.failed || d.Executor.GPU == nil {
		return fmt.Errorf("GPU session is closed,failed or uninitialized")
	}
	return nil
}
