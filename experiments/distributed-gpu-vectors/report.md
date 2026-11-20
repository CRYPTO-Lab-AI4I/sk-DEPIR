# Distributed GPU matrix-vector composition

Status: bounded MPI/GPU validation **78302 submitted**; implementation `9cb59b6`.
[Receipt](submission.json), results `runs/composed-01/`, log `slurm-78302.log`.
Exact frozen job-start revision/hashes will be in raw results. No monitoring or
composed result inspected. This is a bounded
integration gate after [transpose78279](../distributed-transpose/report.md) and
[GPU vector78293](../gpu-vector-stage/report.md) component validation. It is a
separable matrix transform, **not a complete additive FFT**: distributed Taylor
basis conversion remains the next gate, followed by full forward/inverse products,
s18 field admission and complete preprocessing integration.

## Shared implementation

The [rank orchestrator](../../encoding/liftrs-shared/distributed_vectors.go) uses
existing `transposeRank` twice. Forward execution transposes row-owned input to
column ownership, packs complete columns into bounded interleaved tiles, runs the
GPU column stage, transposes back, then runs affine rows at global row origins.
Inverse executes row inverse before the two transposes/column inverse. Input is
consumed; no global matrix or full coordinator reference is gathered.

[VectorPreparation](../../encoding/liftrs-cuda/vector_preparation.go) owns a
separate single-GPU executor. Its optional ABI reuses `VectorStage`, `local_vectors`
and the existing field/metadata builders; arithmetic is not copied. Dedicated
ownership prevents product/encoding operations from reusing its storage. Global
shape and local capacity are immutable; typed calls validate width, ownership,
complete-vector extents and closed state. Caller pointers are borrowed only for
the synchronous call. Each batch uploads/transforms/downloads; persistent storage
is reused, with no implicit CPU fallback. Existing plugins remain compatible for
other operations; the new owner requires the optional stage symbol. CPU clients
and ordinary builds remain CUDA/MPI-free; no default or full-run support changes.

The [CPU orchestration tests](../../encoding/liftrs-shared/distributed_vectors_test.go)
use noncommuting reversible operations independent of GPU math, blocking channels,
32/64-bit slots, odd/tail/empty owners, 1/2/3/8/16 ranks, inverse recovery, invalid
admission and propagated backend errors. Their purpose is ownership/order, not a
second field implementation.

## Reproduction and admission

[Controller](run.py), [Slurm entry](run.sbatch),
[MPI GPU tests](../../encoding/liftrs-shared/distributed_vectors_gpu_test.go),
[reviewer](analyze.py), [review tests](test_review.py). Exact committed revision
and build-input hashes are frozen and recorded in `runs/<name>/results.json`.
Logs, CPU profiles, CPU-first gate results, phase records and outputs are retained;
cache/binaries/shards outside Git at `.cache/liftrs/distributed-gpu-vectors/<name>/`
are sealed by checksum. This test needs small deterministic fixtures, not a full
preprocessing cache; existing expensive caches are untouched.

```sh
python3 -B experiments/distributed-gpu-vectors/test_review.py
python3 -B experiments/distributed-gpu-vectors/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-gpu-vectors/slurm-%j.log \
  experiments/distributed-gpu-vectors/run.sbatch composed-01
# After the user's completion ping:
python3 -B experiments/distributed-gpu-vectors/analyze.py composed-01
```

Existing admitted envelope: two HPC nodes / two H200s (one per node), two CPU
cores and 4 GiB requested host RAM per node, 128 MiB owned device cap/card,
1 GiB RSS guard/rank, no pinned staging, 128 MiB retained cache, 2 GiB scratch,
32 MiB/file, 720-second controller / 15-minute scheduler. gnode03 excluded; other
eligible HPC nodes allowed. Native allocations/context and measured RSS are
separate. No increase in resources is authorized by this component design.

CPU orchestration/race checks precede GPU work. Rank entry, binary, library and
writable shared working directory are checked on both nodes before execution;
node-local frozen source paths are never used as remote rank entries. Staged MPI
uses the existing descriptor/proof lanes and explicit validated
`OMPI_MCA_pml=ucx, UCX_RCACHE_ENABLE=n` configuration, not a production default.
Both worker streams remain separate; sanitizer summaries are checked jointly and
must be exactly one clean error/leak pair per rank. The focused regression covers
shared launch paths, required preflight, missing packet and corrupt output.

## Correctness and measurements

Six cases: s4 dimensions4/5/8 and s17 dimensions5/9/12, each forward and inverse
on two ranks. Thus24 measurement rows per execution. Both widths, nonzero affine
row origins, odd dimensions, short GPU batches and multi-chunk MPI tiles are used.
The largest grid is4096 words (32 KiB aggregate at s17); longest vector64 words.
No new 2^17-word sanitizer run is needed: that kernel/origin boundary passed78293.
All cases repeat under CUDA memcheck/leak checks. Prior measured small-vector
sanitizer costs are milliseconds per call; each new MPI/sanitizer step has240s
plus controller allowance. New composed runtime remains unmeasured. No estimate
from a single long-vector case is substituted for the full suite.

The native CPU oracle builds row/column bases independently. Column basis entries
are explicit vanishing products over the low-coordinate subspace, rather than
metadata copied from the GPU builder. One CPU column at a time supplies each
rank's expected rows, then affine row transforms produce its reference. Every
result is checked and saved in explicit little-endian form. Inverse additionally
matches the original deterministic pattern at every owned coordinate. The reviewer
checks output sizes, hashes, every saved reference byte, all rank/pass/operation/
offset/peer proof coverage and completion. Transfer proofs are diagnostics for
public test data, not authentication. Mismatch tiles are synchronously saved
before peer abort. No full reference is gathered on rank0.

Timing rows include host packing, MPI framing/proofs, GPU upload/compute/download
and transpose allocation. Excluded from each transform clock: CPU oracle/setup,
output verification/fsync and profile reports; those remain in complete command/
controller time. Runs are CPU-profiled, phase-logged diagnostics, not unprofiled
baselines. Each phase logs completed/total work; native small-vector calls are
bounded and Go records progress around batches/transfers. CPU profiles are reported
immediately after normal validation, before later sanitizer work. Completed arm
results are persisted independently so a later failure cannot hide them.

## Explicit storage and time limits

This control implementation consumes caller input plus up to two owned host
shards and bounded vector/transport tiles; caller-retained oracle/input copies
are additional diagnostic storage. It allocates no full global matrix, but it
is **not** the chosen full-size GPU-resident layout. At a hypothetical2^36-element
64-bit grid /16 ranks, three principal host shards alone project96 GiB/rank,
1536 GiB aggregate, before runtime/headroom. That is a projection, not measured
capacity, implementation support or resource admission.

Each vector stage uploads/downloads all owned values. Two stages per forward
transform imply2N slots in each host/device direction, plus two MPI transposes,
CPU packing and framing/proof overhead. For a hypothetical full s18 grid, those
GPU transfers alone total1 TiB H2D plus1 TiB D2H aggregate per forward transform;
communication/local GPU packing and full workflow costs remain unknown. No extra
canonical cache/output format is introduced. Existing faster resident defaults
are unchanged; no slower memory-saving mode is adopted. Production composition
still targets reused GPU arenas with bounded host exchange, and must measure
that storage/communication layout separately before a full s18 trial.
