# Resident GPU preprocessing through s17

Status: source prototype and bounded validation prepared; CUDA/peer execution
has not yet been validated. No full s17 run, measured speedup or admission.
GPU execution is the primary server-preprocessing target; CPU supplies arithmetic
oracles and a reference implementation rather than the next full preparation run.

## Reproduction and budgets

[Controller](run.py), [Slurm entry](run.sbatch),
[shared-recurrence driver](check.cpp), [wide GPU monomial windows](check.cu),
[completion analyzer](analyze.py).
Results go to `runs/<name>/results.json`, command logs, `progress.jsonl`, profiles
and `tiles.json`. Matching binaries, shared library and reusable small code data
remain in `.cache/liftrs/gpu-s17-preprocessing/<name>/`, outside Git. Source revision,
exact hashes (including .cu/.cuh), environment, GPU identity/topology and all
commands are sealed in results. No previous job's output is consulted.

```sh
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
# Existing one-GPU budget:
sbatch --output=experiments/gpu-s17-preprocessing/slurm-%j.log \
  experiments/gpu-s17-preprocessing/run.sbatch gpu-01 --devices 1
# Requires approval for a second allocated GPU; not automatically authorized:
sbatch --gres=gpu:h200:2 --output=experiments/gpu-s17-preprocessing/slurm-%j.log \
  experiments/gpu-s17-preprocessing/run.sbatch peer-01 --devices 2
# Only after the completion ping:
python3 -B experiments/gpu-s17-preprocessing/analyze.py gpu-01
```

The job requests4 CPU cores and8 GiB host RAM, with6 GiB sampled process RSS,
4 GiB Go soft limit,2 GiB new retained cache and4 GiB scratch guards. Commands
have600-second limits, controller900 seconds and Slurm1200 seconds. Owned device
payload is capped at512 MiB aggregate:512 MiB on one GPU or256 MiB on each of
two. CUDA contexts/peer mappings are additional; each card requires64 MiB free
headroom outside its cap. A second GPU increases allocated hardware, so its
submission requires agreement even though aggregate payload memory is unchanged.
Full s17 allocations are not admitted by these tests or maximum-size metadata.
No slower memory mode is enabled. The topology must allow all-pairs CUDA peer
access; an unsupported request fails without a host-backed or CPU fallback.

## Implementation and tests

The server library links `encoding/liftrs-cuda/matrix.cu` and `preparation.cu`.
The shared preprocessing session selects it with `Backend: "cuda"`, an explicit
library and nonempty `GPUDevices`. `GPUBytes` is a per-device cap in this mode.
The old empty-device CUDA32 adapter and ordinary client APIs retain their choices.
GPU counts1/2/4/8 are supported in source for power-of-two slab division; actual
same-node card availability and validation remain additional constraints.

`transform.cuh` shares arithmetic, per-vector FFT and Taylor recurrences between
legacy client kernels and both preparation widths. `preparation_plan.hpp` reuses
the native sparse builder; `preparation_kernels.cuh` supplies launches, sharded
access, bit-weight classification/expansion and root warp scans. The CPU/CUDA
membership recurrence is shared in `liftrs-cpp/monomial.hpp`, with Go's criterion
retained as an independent oracle. The root warp scan reuses short exponent gaps
and computes one initial power per warp. Host code retains the single existing
sampling/support, factor-product/Newton and full P·G verification flow.

The GPU suite checks complete s3/4/6/8 discovery, P/G, canonical spectrum/file
bytes, checkpoints, lifetimes and caps against Go, using seed170069. Short s17
transforms/products/root batches include values/indices above2^32 and are checked
against independent quadratic-field Horner/schoolbook/power arithmetic. A bounded
full-size metadata/cap failure must reject huge discovery without allocating its
2 GiB bitmap. No genuine s17 support is produced. A separate CUDA driver reuses the production
weight/expansion kernels with the full123 MiB class cache but only17 bitmap
words at a time. Three1088-position windows include indices above2^32 and near
2^34, comparing bits/counts with CPU membership (independently checked against
Go in the regression suite). It stays below256 MiB on one card and never
allocates the full bitmap. Tiny physical buffers also probe the production
shard accessor with logical indices crossing2^32; all buffers are on GPU0, so
this is address-mapping validation, not peer proof. These large-class windows
are not memchecked, while
the same kernels are memchecked in complete small domains.

One or two-card resident tiles contain2^22 s17 extension values (32 MiB each).
Untimed sparse degree15 Horner witnesses, timed dense field input, independent
evaluation-at0/1 checks, full inverse identity and full forward hashes
are checked; if two cards are approved, compare the same data with one card in
the same allocation. Timings in `tiles.json` are profiled diagnostic whole-call
seconds, including host range checks/uploads/downloads and peer synchronization.
They are not complete preprocessing latency or an s17 speedup estimate. CPU
profiles sample host activity and cannot substitute for a GPU kernel profile.

Compute-sanitizer memcheck/leak checks cover small complete prep and wide kernels.
The legacy single-GPU online encoder, ordinary Go/native regressions and vet guard
shared-code compatibility. A C++ host driver executes the same recurrences over
simulated1/2/4-device slabs, both widths at small s and uint64 at s17, including
full-domain metadata with34-bit dimensions and offset constants above2^32. Optimized and ASan/UBSan
runs are included. Host simulation does not validate peer visibility or hardware.

## Memory, communication and logging

The fast resident product owns two operands and transpose scratch:3·N·wordBytes,
partitioned rather than replicated. At s17 the aggregate is384 GiB,96 GiB/card
on four GPUs, plus tables/plans/root buffers. Three H200s are only an arithmetic
capacity floor; the current power-of-two implementation would use four. This is
source support/capacity, not successful execution or a larger resource budget.
Single- and multi-card implementations use the same kernels and layout. Full
cross-node execution remains unimplemented.

Peer enablement and phase ordering use the documented CUDA peer-access and
synchronization model ([NVIDIA CUDA12.9 guide](https://docs.nvidia.com/cuda/archive/12.9.0/cuda-c-programming-guide/index.html#peer-to-peer-memory-access),
[memory synchronization domains](https://docs.nvidia.com/cuda/archive/12.9.0/cuda-c-programming-guide/index.html#memory-synchronization-domains)).
Taylor passes join all cards before dependent work. Independent vector waves bound progress granularity when native logging is
enabled; uninstrumented calls launch whole vector partitions. Peer reads/writes and two coalesced shared-memory distributed transposes per
row phase are real communication costs. Tables/plan metadata are small replicas;
large arithmetic slabs are resident once per logical array. Input widening and
base-result staging reuse dead transpose scratch, so they add no full allocation.
GPU classification uses a3^s-byte class cache per card and sharded N/8 bitmap;
the class caches are released before subsequent transforms. Sorted bitmap download
uses at most8 MiB additional native host staging, then fills the Go-owned support
array. That staging/temporary host plan construction is additional to the native
CPU executor's payload cap; aggregate host RSS guards still apply. Go support,
root/generator, final spectrum, GC and filesystem cache remain additional.

No memory savings are exchanged for host arithmetic backing, disk I/O or repeated
transforms. Whether distribution is faster is unknown until measured; link topology,
peer traffic, barriers and setup can outweigh added GPUs. Retain the validated
s16 path until wider/peer results pass. Full s17 still needs measured headroom,
reviewed host/device/storage/time budgets and a guarded persistent full-run harness.

`LIFTRS_CPU_PROGRESS=1` records named GPU phases and advancing completed vectors,
classes or bytes, plus UTC and monotonic elapsed time. Vector waves join on device
work boundaries; no per-element host polling or full data passes are added for
logging. Small kernel events are suppressed to avoid flooding factor-product
logs. The controller writes30-second resource heartbeats separately. Instrumented
measurements include that cost and must not replace unprofiled latency baselines.
On errors, queued device work is drained before releasing borrowed Go pointers;
CUDA device selection is restored so ordinary client calls can coexist.

## Current evidence

[Static/shared checks](static-checks.json) record host simulation and syntax checks.
Both CUDA translation units compile to compute90 PTX; the shared host recurrence
cases, C bridge/native syntax and table-label checks pass. CUDA execution, peer
communication, sanitizer and regression results remain pending. No full preprocessing table cell changes to measured status.

The cluster requires a typed GPU GRES request. An initial untyped submission
was rejected before allocating a job; the script now requests `gpu:h200:1`
without changing count, memory or deadline.
