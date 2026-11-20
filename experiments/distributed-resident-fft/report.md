# Complete bounded resident GPU FFT

Status: **job78396 submitted**, implementation `2ec5458`; [receipt](submission.json).
Local CPU/race and native/CPP/CUDA/MPI builds pass; GPU results remain pending.
[Results destination](runs/fft-01/), [job log](slurm-78396.log). Actual job-start
revision/source identities are saved by the controller.
[Readiness commands and artifacts](readiness/checks.json).

[Resident vectors78394](../distributed-resident-vectors/report.md) passed24 normal
and24 sanitized rank/pass results, the24-row host-tile regression and all paired
proofs, with zero CUDA errors/leaks. This next gate composes the complete Taylor
conversion with those stages. We remain on step1 (distributed products/FFTs),
not distributed discovery or fulls18 preprocessing.

## One algorithm, storage adapters

The [Taylor scheduler](../../encoding/liftrs-shared/distributed_taylor.go) now
separates Read/Send/Apply storage access from its existing ticket cursors, trimmed
source windows, peer rounds, quotient/remainder barriers and level order. The
host adapter retains its borrowed send slices and snapshot callback behavior;
the [resident adapter](../../encoding/liftrs-shared/distributed_storage_cuda.go)
reads bounded logical ranges from GPU and uses a reusable send buffer. No domain-
sized request list or complete host shard is introduced. Storage errors propagate
without a successful final barrier; focused CPU regressions cover read/send/apply.

Host and resident FFTs share canonical parameter/operation plans and stage order:
Taylor then columns/affine rows forward, reversed inverse. Existing CPU independent
Horner/Lagrange/schoolbook, ownership and race regressions pass. A second FFT or
field recurrence is not copied. The resident data/scratch slots retain the same
three-array allocation and row/interleaved addressing; the third slot is preserved.
Inverse Taylor reads logical coefficients even when physical output is interleaved.

[`RankStorage.TaylorDegrees/Taylor32/Taylor64`](../../encoding/liftrs-cuda/rank_storage.go)
share plan metadata and the extracted native Taylor launcher with VectorPreparation.
Sparse coefficients, `TaylorWindow` source access and `window_taylor/taylor_update`
arithmetic are unchanged. A read-only host snapshot contains Count target words
then Count words per sparse term. Canonical values, snapshot extent and target
ownership are checked. One lazily retained, cap-accounted device window receives
the snapshot. Updated targets copy directly from that window into the resident
slot; no target readback or complete shard staging occurs. Incomplete results must
be discarded/reset after error. Missing optional ABI returns an explicit error,
without a CPU fallback. No client/default/field-admission changes.

## Independent bounded validation

The [shared resident driver](../../encoding/liftrs-shared/distributed_resident_gpu_test.go)
selects `TestDistributedResidentFFT` with Taylor enabled. Both slot widths run
s4 dimensions4/5/8 and s17 dimensions5/9/12, N16–4096, on2 ranks. Every forward
word is checked against the native full FFT, whose first/last owned values are
also checked by scalar Horner evaluation. Inverse must recover every original
word. Third-slot sentinel, unchanged retained allocation on inverse, rejected
row-major vector views and out-of-field nonmutation checks remain active.

The [runner](run.py) first reruns the **complete host-tile FFT reference**, not
just separable vectors, because its Taylor launcher was shared in this change.
Then resident normal and all small cases under CUDA memcheck/leak checks run:
24 rank/pass rows per arm. [Independent review](analyze.py) checks full output
bytes/hashes, exact Taylor metadata and trimmed source routes, transpose coverage
and paired proofs, preservation/reuse flags, source/evidence/cache seals and both
stdout/stderr CUDA summaries. [Reviewer/selection tests](test_review.py) reuse
four FFT fixtures and reject missing resident claims; selecting the complete
reference and new driver is explicit.

## Resources, reproduction and measurement boundaries

[Slurm entry](run.sbatch) reuses the validated composition controller, staged MPI,
UCX cache control, shared executable/wrapper/library/cwd and remote TMPDIR
preflight. Same admission: **2 HPC nodes /2 H200s /4 CPU cores /8 GiB requested
host RAM** (one GPU,2 cores,4 GiB/node), **128 MiB owned GPU/card**,
**1 GiB RSS guard/rank**, no pinned buffers, **128 MiB retained cache /2 GiB
scratch /32 MiB file**, **720-second controller /15-minute scheduler**.
gnode03 excluded; other eligible nodes allowed. No resource increase or monitoring.

```sh
python3 -B experiments/distributed-resident-fft/test_review.py
python3 -B experiments/distributed-resident-fft/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-resident-fft/slurm-%j.log \
  experiments/distributed-resident-fft/run.sbatch fft-01
# After the user's completion ping:
python3 -B experiments/distributed-resident-fft/analyze.py fft-01
```

Exact revision, source identities, commands/exits, resource observations, flushed
stage/proof logs, CPU profiles and intermediate completed arms persist in
`runs/<name>/`. Sealed binary/library/output/reference cache remains outside Git
at `.cache/liftrs/distributed-resident-fft/<name>/`. No previous cache is deleted.
Per-pass diagnostic clocks include plan setup, source/snapshot staging, Taylor,
vector range scans, kernels, packing, MPI/proofs and barriers. They exclude initial
input/oracle construction/upload, final readback/check/fsync and profile reports;
command/controller times include those. Profiling and phase/proof logging remain
enabled; these are not unprofiled latency baselines. Logs identify operation,
level, quotient/remainder and completed work; native long launches retain bounded
progress reporting, and guards/resources remain job-side.

78394 controller took122.265 seconds, normal2.026 seconds /sanitizer2.772 seconds.
78307's same-size complete host FFT suite took1.971 seconds normal /3.726 seconds
sanitized; no larger domain or sanitizer shape is introduced here. Three MPI
arms keep240 seconds each (220-second test timeout) inside720-second controller;
previous complete-controller costs leave substantial expected headroom. Setup
builds the old host-tiled adapter as a compatibility check. Resident complete-FFT
latency remains unknown until this run, not an inferred speedup.

## Memory–time and scope limits

Let C=min(N,3*max(rows,columns)) <=1 Miword in this gate. Taylor adds **two C-word
host buffers** (snapshot/send) and **one C-word GPU window**, counted alongside
existing arrays/transport/plan under the device cap. Largest C192 at eight-byte
slots adds **3072 bytes/rank host payload** and **1536 bytes/card GPU payload**.
At the supported bounded maximum of1 Miword, these become16 MiB/rank host and
8 MiB/card GPU; context/runtime, metadata and caller/diagnostic arrays are separate.
The existing127-word transport tile is unchanged. No full host shard is retained
by the execution adapter; small test inputs/oracles/readback arrays are additional.

This first shared-storage adapter still copies locally supplied Taylor sources
through bounded host snapshots, as well as staging remote sources. Repeated
D2H/H2D traffic, field checks, barriers and plan/setup costs are included in measured
boundaries; they are not presumed free. Further local-window GPU gathering may
reduce transfers after correctness and matched costs are established. No slower
memory-saving mode becomes a default, and no claimed full-workflow memory saving
follows from the isolated buffer model.

Resident pointwise products and three-transform lifetimes, realistic larger
throughput, s18 field admission, root/Newton/generator/discovery integration and
coordinated persistence remain separate gates. Fulls18 preprocessing is unrun.
