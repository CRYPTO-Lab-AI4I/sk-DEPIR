# Bounded distributed GPU additive FFT

Status: implementation ready for validation; no complete MPI/GPU FFT result yet.
The preceding [separable composition78302](../distributed-gpu-vectors/report.md)
passed all outputs/proofs and CUDA checks. This gate adds distributed Taylor
basis conversion around those same column/row stages. Full s18 fields, products,
preprocessing workflow and a production GPU-resident layout remain unvalidated.

## Implementation and invariants

The [rank Taylor planner](../../encoding/liftrs-shared/distributed_taylor.go)
partitions targets by complete owned rows and advances bounded ticket waves.
Every rank derives the same ownership schedule using O(ranks) cursors, not a
full-grid request list. Pairwise source exchange uses the validated rendezvous
round order; local windows are copied directly. Ranks update independent local
targets concurrently where targets exist. Barriers before/after each pass enforce
quotient-before-remainder in forward conversion and the reverse order in inverse
conversion; level order is likewise reversed for inverse. Empty owners still join.
Forward then executes the existing matrix-vector composition; inverse executes
matrix-vector inverse before conversion. Inputs are consumed, outputs own row shards.

Only operative sources are read or sent. Quotient sources lie in the untouched
upper quarter; operative remainder sources lie in the completed upper half.
Unused remainder source prefixes can overlap targets, so they are zero-filled
rather than read concurrently. Beyond-block suffixes are also zero-filled.
Degrees are positive, sorted and at most half/2; these bounds establish the
source/target separation within each pass. Earlier waves cannot change a later
operative source. CPU race tests cover this invariant, partial windows and empty
owners. Sparse source coefficient order comes from the canonical GPU plan.

The GPU [source-window accessor/kernel](../../encoding/liftrs-cuda/taylor_window.cuh)
is extracted verbatim from the existing host-tiled implementation. Both adapters
reuse it and `taylor_update` from `transform.cuh`; no recurrence is copied.
[VectorPreparation](../../encoding/liftrs-cuda/vector_preparation.go) provides
optional `TaylorDegrees` and typed `Taylor32`/`Taylor64` operations on the same
dedicated capped device tile. Host metadata is copied from the shared builder
once; hot window execution allocates no coefficient vectors. Missing optional
plugin symbols return errors; old vector/preprocessing plugins retain their
existing operations. CPU clients still require neither CUDA nor MPI.

Taylor packet headers cover operation/domain/split/level/direction, target and
source extents, term slot, width and source owner. A canonical hash fits those
fields into the existing64-byte control frame. Paired payload proofs check source
stability and peer equality; mismatch snapshots are flushed before peer abort.
These diagnostics cover public experiment data, not adversarial authentication.

## Independent verification and coverage

[CPU tests](../../encoding/liftrs-shared/distributed_taylor_test.go) compare source
windows with serial global-offset updates across32/64-bit slots, 1/2/3/5/8/16/32
ranks, small/tail capacities, empty owners, both conversion directions and wide
virtual coordinates above2^32. Full small-field transforms additionally match
independent direct Horner evaluation and Lagrange inverse interpolation, using
an explicitly multiplied subspace polynomial rather than native/GPU metadata.
Invalid admission and backend errors are checked before/at the operation boundary.
CPU orchestration and race checks pass locally; they rerun before GPU work.
The CUDA plugin and complete CPP/CUDA/MPI test binary also compile locally;
[readiness commands/hashes](readiness/checks.json), [native log](readiness/cuda-build.log),
[MPI build log](readiness/go-mpi-build.log). No GPU/MPI execution occurred locally;
local compiler interface-probe socket warnings did not fail compilation.

The [MPI/GPU test](../../encoding/liftrs-shared/distributed_vectors_gpu_test.go)
and [Taylor transport adapter](../../encoding/liftrs-shared/distributed_taylor_gpu_test.go)
reuse the existing bounded driver via an explicit full-FFT flag. Cases remain
s4 dimensions4/5/8 and s17 dimensions5/9/12, forward/inverse on two ranks:
24 rank/pass rows. A separable baseline is first repeated on the same build;
full FFT normal and all small cases under CUDA memcheck follow. No new long-vector
sanitizer run is needed; the shared vector/origin boundary passed78293.

Each rank independently generates a small native full FFT reference (at most4096
words plus equally sized scratch, 32 KiB each in uint64 storage), then retains
its owned rows. Independent Go Horner checks verify the first/last owned output
on each rank in both fields, complementing all-word native comparison. These are
bounded diagnostic replicas, never a coordinator/network
gather or production full-grid allocation. Every forward output matches native
CPU, every inverse output matches the original deterministic pattern, and every
saved output/reference byte is reviewed. Source snapshots must remain unchanged
by the GPU update; owned device allocations must not grow on inverse reuse.

The [reviewer](analyze.py) validates exact transpose operations and independently
enumerates target intervals/operative remote sources to verify every Taylor
packet, paired proof, plan and completion. The [review tests](test_review.py)
reject omitted/wrong source windows and check inactive-prefix/tail examples.
Complete CUDA error/leak summaries are required from both streams of each rank.

## Reproduction, resources and retention

[Runner](run.py) reuses the admitted [composition controller](../distributed-gpu-vectors/run.py);
[Slurm entry](run.sbatch). Exact Git revision/build hashes, commands, environment,
phase logs and results are retained in `runs/<name>/results.json`. Source snapshots
are frozen before build. Cache, binaries, output shards and profiles remain under
`.cache/liftrs/distributed-gpu-fft/<name>/`, outside Git and sealed by hashes.
Existing expensive preprocessing caches are untouched.

```sh
python3 -B experiments/distributed-gpu-fft/test_review.py
python3 -B experiments/distributed-gpu-fft/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-gpu-fft/slurm-%j.log \
  experiments/distributed-gpu-fft/run.sbatch fft-01
# After the user's completion ping:
python3 -B experiments/distributed-gpu-fft/analyze.py fft-01
```

Same admitted HPC envelope: two nodes, one H200 and2 CPU cores /4 GiB requested
host RAM per node,128 MiB owned device cap/card,1 GiB RSS guard/rank, no pinned
staging,128 MiB retained cache,2 GiB scratch,32 MiB/file,720-second controller /
15-minute scheduler. gnode03 excluded; other eligible nodes allowed. CPU/race
checks and binary/wrapper/library/shared-cwd preflight precede GPU work. Remote
TMPDIR is valid node-local `/tmp`. Staged UCX uses the prior explicit cache control.
Completed arms/profile reports persist before later sanitizer execution. The
shared host-tiled adapter is additionally compiled as a compatibility check; its
GPU arithmetic source was moved verbatim, not rewritten.

The previous4096-word separable suite took2.722s normal /3.826s sanitized; Taylor
adds bounded short kernels, source packets, synchronizations and snapshots. New
complete FFT latency is unknown. Each MPI step has240s (220s test timeout), within
the unchanged720s controller/15min allocation, with substantial margin over those
measured small-vector costs. Job-side logs flush pass/window/proof and increasing
target/vector counts; no monitoring after submission. Native kernel phase events
and Go events bracket bounded calls. No ETA is inferred from heartbeat-only output.

## Measurement and memory/time boundaries

Per-transform clocks include conversion, host packing, MPI headers/proofs/barriers,
GPU snapshot/vector transfers and computation. They exclude CPU reference/setup,
output checks/fsync and profile reports; complete command/controller time includes
those costs. Runs remain profiled/logged diagnostics, not unprofiled baselines.

The Taylor planner reuses one bounded host snapshot across conversion passes,
plus O(ranks) cursors/tickets per pass. It adds no full host shard. The surrounding
control layout still retains caller input plus up to two host shards and bounded
vector/transport tiles; diagnostic oracle/caller copies and runtime are additional.
GPU Taylor and vector stages reuse the same admitted device capacity. Constant
host plan metadata is additional; RSS/context and owned device payload differ.

A window with t sparse terms uploads/downloads up to(t+1)*target_count slots;
source windows overlap across tickets/levels, repeating traffic. Full snapshot
readback validates source immutability; production might avoid it after separate
validation. More windows add launches, packing, barriers and proof work. None of
this overhead is assumed free or extrapolated into a fulls18 runtime estimate.

The three-host-shard full2^36/16-rank/uint64 model still projects96 GiB/rank,
1536 GiB aggregate before headroom. This is a control-layout cost, not the selected
full-size server design, demonstrated feasibility or budget admission. Production
must reuse GPU arenas with bounded host exchange and measure communication/setup/
verification/persistence. Faster validated resident paths/defaults remain unchanged.
After this gate: validate distributed products/identities, admit s18 arithmetic,
and integrate discovery/Newton/checkpoint/publication before any full preprocessing.
