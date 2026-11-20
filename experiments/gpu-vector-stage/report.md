# Persistent GPU vector stages

Status: bounded HPC validation **78285 submitted**, source `0711805`; no GPU
result inspected. [Submission receipt](submission.json), results
`runs/vectors-01/`, job log `slurm-78285.log`. No monitoring; await completion ping.
This is the next distributed-preprocessing component after the
[validated sharded transpose](../distributed-transpose/report.md). It is not a
complete distributed FFT, product, or s18 preprocessing implementation.

## Implementation and boundaries

The internal [VectorStage adapter](../../encoding/liftrs-cuda/vector_stage.cuh)
borrows one fresh resident executor exclusively. The executor owns one bounded
GPU tile, field tables and global matrix metadata. No host grid, transpose array,
second GPU operand, or per-call replacement allocation is introduced. Upload,
row/column transform and download are separate synchronous operations. Forward
and inverse calls can reuse the resident tile without transfers. The adapter
reuses `local_vectors` and `vector_transform`; arithmetic is not copied. Global
plan size and first-vector coordinates are independent of local storage capacity.
Storage is interleaved as `[coefficient * count + vector]`, matching existing
local-vector kernels. Neither this adapter nor its tests change the public ABI,
client build dependencies, supported full-run parameters or production defaults.
The borrowed executor retains its existing s≤17 field admission; s18 arithmetic
admission and full-size pipeline integration remain separate validation gates.

Requests reject empty/overflowing/out-of-range extents and partial vectors;
upload validates field values before mutation. Invalid requests preserve a live
tile. Failed device execution invalidates its loaded state. The caller must keep
the executor alive and must not use other executor operations while borrowing it.
Each launch processes at most 4096 complete vectors and joins before the next
launch. Timestamped phase events report completed vectors; the driver emits and
flushes case/CPU-oracle progress and RSS observations. A 30-second observation
thread emits explicitly labelled heartbeats (not new progress) and an RSS guard
during long native calls. These bounded checks are
not throughput tuning or evidence of copy/compute overlap.

## Reproduction and evidence

[Controller](run.py), [Slurm entry](run.sbatch), [rank entry](rank.sh),
[driver](check.cu), [reviewer](analyze.py), [review tests](test_review.py),
[canonical field fixture builder](../../encoding/liftrs-shared/gpu_vector_fixture_test.go).
Revision and every build-input hash are frozen from the committed Git tree by
the shared controller and recorded in `runs/<name>/results.json`. Command logs,
per-rank stdout/stderr and Go/native CPU profiles are retained there. Fields,
binaries, actual/reference output arrays and native profiling input remain in
`.cache/liftrs/gpu-vector-stage/<name>/` outside Git, sealed by hashes. No expensive
code preprocessing is discarded; these cases need only small canonical fields.

```sh
python3 -B experiments/gpu-vector-stage/test_review.py
python3 -B experiments/gpu-vector-stage/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/gpu-vector-stage/slurm-%j.log \
  experiments/gpu-vector-stage/run.sbatch vectors-01
# After the user's completion ping:
python3 -B experiments/gpu-vector-stage/analyze.py vectors-01
```

The admitted validation envelope is unchanged: two HPC nodes, one H200 and
2 CPU cores / 4 GiB requested host RAM per node (2 GPUs / 4 cores / 8 GiB total),
128 MiB owned device cap/card, tightened 1 GiB RSS guard/process, no pinned staging,
128 MiB retained cache, 2 GiB scratch, 32 MiB/file, 720-second controller deadline,
15-minute scheduler limit. gnode03 remains excluded; other eligible nodes allowed.
The two ranks run the same stage suite independently; this test adds no MPI
payloads or GPU communication. Normal GPU execution has a 150-second step limit, with 480 seconds for the
complete sanitizer suite; the controller retains its 720-second total deadline.
Sanitizer failures fail the run. Capacity estimates do not authorize larger runs.

## Correctness and timing

The driver checks 24 cases/process: s4 dimensions 4/5/8 and s17 dimensions 7/12/34,
row/column stages, zero and nonzero/tail origins, 32/64-bit slots. Small cases use
three interleaved vectors; dimension34 uses one complete vector to keep retained
outputs bounded. The largest virtual grid is 2^34, but its local vector has only
2^17 words; the last row starts above 2^32. No full virtual grid is allocated.
Native recursive CPU transforms use polynomial multiplication independent of GPU
lookup arithmetic. For dimension34, every output is compared with independent
Horner evaluation of a degree-three polynomial at the correct affine/basis point.
This wide witness does not cover every possible dense polynomial.

Every case checks forward output, inverse reuse with no upload, inverse of CPU
oracle values, repeated upload/execution, unchanged owned allocation size, and
invalid origin/count/extent/value rejection preserving the resident output.
All input, CPU expected, forward and inverse arrays are saved; the reviewer checks
every byte and required case/shape/width/completion. The entire suite repeats
under CUDA memcheck/leak checks, not just a narrow case subset. The controller
also retains profiled adapter-constructor setup (executor setup remains in complete command time) and forward/inverse kernel wall times;
these exclude transfer, verification and persistence. Complete command/controller
elapsed time includes those costs and is recorded separately. They are single
profiled diagnostic samples, not end-to-end baselines or s18 runtime estimates.

## Memory and time trade-offs

The component owns one `capacity_words * sizeof(Word)` device tile plus existing
field/plan metadata, with no adapter-owned host array. The largest test reserves
3 MiB for its GPU tile and transfers 1 MiB per upload/download. Caller test arrays
and saved evidence are diagnostic costs, not selected production storage. Two
full host shards at a 2^36 domain are not adopted by this adapter.

A retained GPU tile avoids transfers between consecutive vector calls; later
MPI exchange still needs bounded staging and explicit packing/communication.
Extra launch joins and validation/transfer costs remain unmeasured until this
run. No lower-memory route becomes a default or is assumed free. Existing faster
validated resident paths remain unchanged. Next, after this gate passes, compose
the stages with the sharded transpose and implement distributed Taylor source
windows and ordering; complete GPU FFT/product/workflow costs remain unknown.
