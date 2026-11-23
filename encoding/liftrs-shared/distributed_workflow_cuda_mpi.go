//go:build linux && liftrs_cpp && liftrs_cuda && liftrs_mpi && cgo

package liftrs

import (
	"fmt"
	cuda "github.com/CRYPTO-Lab-AI4I/sk-DEPIR/encoding/liftrs-cuda"
	"github.com/CRYPTO-Lab-AI4I/sk-DEPIR/experiments/distributed-root/mpi"
	"io"
	"time"
)

// This internal server workflow is deliberately separate from client encoding.
// Limits cover individual shards/lists,not complete job/RSS/storage admission.
type rankWorkflowOptions struct {
	Spec                         Spec
	Discovery                    rankDiscoveryOptions
	MaxLocalWords                uint64
	PrefixInputs, StreamSpectrum bool
}
type rankWorkflowSinks struct {
	// Only rank0 supplies writers/archive sink. Caller owns exclusive creation,
	// sync/close, coordinated commit and any final publication/retention policy.
	Code, Generator io.Writer
	Archive         func([]uint64) error
	RootObserver    func(string, rankRootMergeStage, []uint32) error
	SupportObserver func([]int) error
	DiscoveryReady  func(ScalingCodeHeader, int) error
	// A checkpoint may return a verified reload. All ranks must execute the
	// same stage sequence;callback success alone is not a durable commit claim.
	Checkpoint       func(string, uint64, []uint32) ([]uint32, error)
	IdentityObserver func([]uint32) error
}
type rankWorkflowResult struct {
	Header           ScalingCodeHeader
	DiscoverySamples int
	RootCounts       []uint64
	Refinement       rankReciprocalStats
	IdentityVerified bool
}

// prepareRankWorkflow composes validated runtime components without native oracle
// data. The caller must abort peers on any error and never publish partial output.
// Immutable source/secure randomness, durability and full job admission are caller
// responsibilities. This function does not authorize fullsize execution.
func prepareRankWorkflow[Word transposeWord](w *mpi.World, o rankWorkflowOptions, sinks rankWorkflowSinks, emit func(map[string]any)) (result rankWorkflowResult, err error) {
	if err = admitRankDiscovery[Word](w, o.Spec, o.Discovery, emit); err != nil {
		return result, err
	}
	n := uint64(1) << uint(2*o.Spec.FieldDegree)
	_, words := rankBand(n, w.Rank, w.Size)
	if words > o.MaxLocalWords || o.MaxLocalWords == 0 || (w.Rank == 0 && (sinks.Code == nil || sinks.Generator == nil || sinks.Archive == nil)) {
		return result, fmt.Errorf("workflow output/shard admission")
	}
	label := fmt.Sprintf("s%d", o.Spec.FieldDegree)
	discovered, err := discoverRankSupport[Word](w, o.Spec, o.Discovery, emit, sinks.Archive)
	if err != nil {
		return result, err
	}
	if sinks.SupportObserver != nil {
		if err = sinks.SupportObserver(discovered.Roots); err != nil {
			return result, err
		}
	}
	k := uint64(discovered.K)
	if k == 0 || k > n/2 {
		return result, fmt.Errorf("workflow code dimension")
	}
	h, _, err := canonicalScalingCodeHeader(o.Spec.FieldDegree, o.Spec.CurveDegree, k, n)
	if err != nil {
		return result, err
	}
	if sinks.DiscoveryReady != nil {
		if err = sinks.DiscoveryReady(h, discovered.Samples); err != nil {
			return result, err
		}
	}
	base, err := NewBinaryField(o.Spec.FieldDegree, h.Polynomial)
	if err != nil {
		return result, err
	}
	field := orbitField{q: uint64(1) << uint(o.Spec.FieldDegree), base: base, lookup: newBoundedPreparationProducts(base, 18)}
	counts, err := w.Counts(uint64(len(discovered.Roots) + 1))
	if err != nil {
		return result, err
	}
	if _, err = admitRankRootStages(w.Rank, counts, n, o.MaxLocalWords); err != nil {
		return result, err
	}
	// The tree's exact coefficient count must match discovery's K before a GPU leaf allocation.
	stages, err := rankRootMergeStages(w.Rank, counts, n)
	if err != nil || stages[len(stages)-1].Count != k+1 {
		return result, fmt.Errorf("workflow discovery/root count binding")
	}
	options := cuda.PreparationOptions{Path: o.Discovery.Library, S: o.Spec.FieldDegree, Polynomial: h.Polynomial, Delta: h.Delta, Omega: h.Omega, Logs: field.lookup.log, Powers: field.lookup.powers, Devices: []int{o.Discovery.Device}, MaxElements: n, DeviceBytes: o.Discovery.DeviceBytes}
	factor, err := cuda.NewPreparation(options)
	if err != nil {
		return result, err
	}
	lastNotice := map[string]time.Time{}
	local, rootErr := prepareRankRoot(supportSorted, field, n, factor, discovered.Roots, func(phase string) {
		emit(map[string]any{"kind": "workflow-phase", "case": label, "phase": "local-roots/" + phase})
	}, func(phase string, done, total uint64, unit string) {
		if done == 0 || done == total || time.Since(lastNotice[phase]) >= 30*time.Second {
			lastNotice[phase] = time.Now()
			emit(map[string]any{"kind": "workflow-progress", "case": label, "phase": "local-roots/" + phase, "completed": done, "total": total, "unit": unit})
		}
	})
	closeErr := factor.Close()
	if rootErr != nil {
		return result, rootErr
	}
	if closeErr != nil {
		return result, closeErr
	}
	compact, layout, err := mergeRankRootTree[Word](w, label, n, o.MaxLocalWords, options, counts, local, emit, sinks.RootObserver)
	if err != nil {
		return result, err
	}
	session, err := newRankGPUSession[Word](w, options, 2*o.Spec.FieldDegree, label, emit)
	if err != nil {
		return result, err
	}
	defer func() {
		if closeErr := session.Close(); err == nil && closeErr != nil {
			result = rankWorkflowResult{}
			err = closeErr
		}
	}()
	e := &session.Executor
	link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	inputs := [2][]uint32{compact, nil}
	if w.Rank == 0 {
		inputs[1] = []uint32{1}
	}
	sources := [2]rootShardLayout{layout, {0, 1, 1, 1}}
	pair, err := routeShardedOperands(rankShardOperandPlan{1, n, 4096, field.q, words, w.Rank, w.Size, sources}, inputs, link, nil)
	if err != nil {
		return result, err
	}
	e.Plan.Operation = 2
	root := pair[0]
	return finishRankWorkflow(w, o, sinks, rankWorkflowResult{Header: h, DiscoverySamples: discovered.Samples, RootCounts: counts}, field, session, root, nil, 1, emit)
}

// Cold and restarted execution share every refinement, identity and publication
// operation. The caller owns session cleanup before this function is entered.
func finishRankWorkflow[Word transposeWord](w *mpi.World, o rankWorkflowOptions, sinks rankWorkflowSinks, result rankWorkflowResult, field orbitField, session *rankGPUSession[Word], root, g []uint32, known uint64, emit func(map[string]any)) (_ rankWorkflowResult, err error) {
	h := result.Header
	n, k := h.N, h.K
	_, words := rankBand(n, w.Rank, w.Size)
	label := fmt.Sprintf("s%d", o.Spec.FieldDegree)
	e := &session.Executor
	link := operandMPI{world: w, record: func(direction string, p operandPacket, data []uint32, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	wideLink := typedOperandMPI[Word]{world: w, record: func(direction string, p operandPacket, data []Word, own, remote [64]byte) {
		emit(map[string]any{"kind": "operand-proof", "case": label, "direction": direction, "packet": p, "own_proof": fmt.Sprintf("%x", own), "remote_proof": fmt.Sprintf("%x", remote)})
	}}
	checkpoint := func(stage string, keep uint64, values []uint32) ([]uint32, error) {
		if sinks.Checkpoint == nil {
			return values, nil
		}
		restored, err := sinks.Checkpoint(stage, keep, values)
		if err != nil {
			return nil, err
		}
		if uint64(len(restored)) != words {
			return nil, fmt.Errorf("workflow checkpoint extent")
		}
		return restored, nil
	}
	root, err = checkpoint("root", k+1, root)
	if err != nil {
		return result, err
	}
	if g == nil {
		g = make([]uint32, words)
		if w.Rank == 0 {
			g[0] = 1
		}
	}
	if known > 1 {
		g, err = checkpoint(fmt.Sprintf("newton-%d", known), known, g)
		if err != nil {
			return result, err
		}
	}
	g, stats, err := refineRankReciprocal(rankReciprocalPlan{Domain: n, RootCount: k + 1, Target: n - k, Known: known, MaxLocalWords: o.MaxLocalWords, Rank: w.Rank, Ranks: w.Size, PrefixInputs: o.PrefixInputs}, root, g, rankReciprocalOps{
		Square: func(values []uint32, keep uint64) ([]uint32, error) {
			op := e.Plan.Operation
			e.Plan.Operation++
			return squarePolynomialShard(rankSquarePlan{op, n, 4096, field.q, keep, words, w.Rank, w.Size}, values, func(v uint32) uint32 { return field.mul(v, v) }, link, func(done, total uint64) { e.Notice("newton-square", done, total) })
		},
		Product: func(name string, a, b []uint32, ac, bc, keep uint64, prefix bool) ([]uint32, uint64, error) {
			before := e.UploadedInputWords
			out, err := session.product(name, a, b, ac, bc, keep, prefix)
			return out, e.UploadedInputWords - before, err
		},
		Checkpoint: func(keep uint64, values []uint32) ([]uint32, error) {
			return checkpoint(fmt.Sprintf("newton-%d", keep), keep, values)
		},
	})
	if err != nil {
		return result, err
	}
	identity, err := session.Product("identity", root, g, k+1, n-k, n)
	if err != nil {
		return result, err
	}
	if err = verifyRankCodeIdentity(n, w.Rank, w.Size, identity); err != nil {
		return result, err
	}
	if sinks.IdentityObserver != nil {
		if err = sinks.IdentityObserver(identity); err != nil {
			return result, err
		}
	}
	var spectrum []Word
	if o.StreamSpectrum {
		err = session.PrepareSpectrum("spectrum", g, n-k)
	} else {
		spectrum, err = session.Spectrum("spectrum", g, n-k)
	}
	if err != nil {
		return result, err
	}
	if w.Rank == 0 {
		if err = writeScalingCodeHeader(sinks.Code, h); err != nil {
			return result, err
		}
	}
	publish := func(count, limit uint64, values []uint32, writer io.Writer) error {
		op := e.Plan.Operation
		e.Plan.Operation++
		return streamRankValues(rankStreamPlan{op, n, count, 4096, limit, o.MaxLocalWords, w.Rank, w.Size}, values, writer, link)
	}
	if err = publish(k+1, field.q, root, sinks.Code); err != nil {
		return result, err
	}
	if err = publish(n-k, field.q, g, sinks.Generator); err != nil {
		return result, err
	}
	if o.StreamSpectrum {
		op := e.Plan.Operation
		e.Plan.Operation++
		err = streamRankReadValues(rankStreamPlan{op, n, n, 4096, n, o.MaxLocalWords, w.Rank, w.Size}, func(offset uint64, tile []Word) error { return storageTransfer(e.GPU, 0, offset, tile, false) }, sinks.Code, wideLink)
	} else {
		op := e.Plan.Operation
		e.Plan.Operation++
		err = streamRankValues(rankStreamPlan{op, n, n, 4096, n, o.MaxLocalWords, w.Rank, w.Size}, spectrum, sinks.Code, wideLink)
	}
	if err != nil {
		return result, err
	}
	if err = session.Close(); err != nil {
		return result, err
	}
	w.Barrier()
	result.Refinement, result.IdentityVerified = stats, true
	return result, nil
}
