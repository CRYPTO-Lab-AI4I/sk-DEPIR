# Resident GPU preprocessing through s17

Status: one-GPU bounded validation passes; two-GPU arithmetic agrees but a
peer-lifetime memcheck failure requires a corrected retry. No full s17 run or admission.
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
# Two-GPU bounded validation explicitly approved by the user:
sbatch --gres=gpu:h200:2 --output=experiments/gpu-s17-preprocessing/slurm-%j.log \
  experiments/gpu-s17-preprocessing/run.sbatch peer-01 --devices 2
# Only after the completion ping:
python3 -B experiments/gpu-s17-preprocessing/analyze.py gpu-01
```

The job requests4 CPU cores and8 GiB host RAM, with6 GiB sampled process RSS,
4 GiB Go soft limit,2 GiB new retained cache and4 GiB scratch guards. Commands
have600-second limits, controller900 seconds and Slurm1200 seconds. Owned device
payload is capped at512 MiB aggregate:512 MiB on one GPU,256 MiB on each of
two, or128 MiB on each of four. CUDA contexts/peer mappings are additional; each card requires64 MiB free
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

Resident tiles contain2^22 s17 extension values (32 MiB each), comparing all
available1/2/4-card prefixes within one allocation.
Untimed sparse degree15 Horner witnesses, timed dense field input, independent
evaluation-at0/1 checks, full inverse identity and full forward hashes
are checked; each approved allocation compares the same data across its power-of-two
device prefixes. Timings in `tiles.json` are profiled diagnostic whole-call
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
cases, C bridge/native syntax and table-label checks pass. Completed execution
evidence and the peer-lifetime retry are recorded below. No full preprocessing
table cell changes to measured status.

The cluster requires a typed GPU GRES request. An initial untyped submission
was rejected before allocating a job; the script now requests `gpu:h200:1`
without changing count, memory or deadline.

## Submission

One-H200 validation job74453 is submitted, source8828856 ([record](submission.json)).
Results/logs/profiles: `runs/gpu-01/` and `slurm-74453.log`; reusable small caches
and binaries: `.cache/liftrs/gpu-s17-preprocessing/gpu-01/`. No monitoring or
result analysis until the completion ping. At this initial submission, two-GPU approval was pending. The user subsequently
approved bounded peer validation, recorded below. No full s17 job is admitted.

## Approved two-GPU peer validation

The user explicitly approved the two-card budget. Job74454 (`peer-01`) is now
submitted with typed `gpu:h200:2`, source d5d2a3b ([record](peer-submission.json)).
The same source kernels,256MiB/card (512MiB aggregate),4CPU/8GiB host and existing
600/900/1200s guards apply. No implementation, timing boundary or full-size
budget changes. The runner compares the same dense tile on one and two cards
inside this allocation, verifies peer-layout hashes and records topology.
Results/logs/profiles: `runs/peer-01/`, `slurm-74454.log`; retained small artifacts
and matching binaries: `.cache/liftrs/gpu-s17-preprocessing/peer-01/`.
No state polling, result analysis or monitoring until the completion ping.

## GPU-count performance policy

Preprocessing has no two-GPU design limit. Choose the count that minimizes
measured complete latency, including preparation setup, communication, host
work and persistence; fitting in memory is only an admission condition.
The runner accepts `--devices 1`, `2` or `4` for the current same-node hardware.
A four-card run compares1/2/4 in that one allocation with identical dense input
and hashes. The512MiB aggregate cap becomes128MiB/card; the wide-window driver's
actual payload is below that bound, checked by `run.py --check` and the analyzer.
No current job or budget is changed and no four-card job is submitted here.

Existing one/two-card result files remain compatible with the expanded analyzer.
Real four-card runtime/topology validation remains pending. Larger groups and
multiple nodes are future measured choices requiring distributed implementation
and reviewed resources, rather than an arbitrary two- or four-card ceiling.

[GPU-count preparation checks](gpu-count-checks.json) verify CLI/schema and
small-test cap consistency only; no four-card runtime is claimed.

## Completed bounded GPU review and peer-lifetime correction (SCALING-072)

User completion ping authorizes this review. [One-GPU analysis](runs/gpu-01/analysis.json)
passes job74453 on gnode02:496 source,35 evidence,36 retained-cache identities;
small complete preparation through s8 versus Go, s17 high-bit and address/window
checks, legacy CUDA/client/CPU regressions, race/vet and sanitizer checks pass.
CUDA memcheck reports zero errors/leaks. See [commands/provenance](runs/gpu-01/results.json),
[runner](run.py), [analyzer](analyze.py), [Slurm log](slurm-74453.log).

[Two-GPU forensic analysis](runs/peer-01/failure-analysis.json), job74454 on the
same node, reconciles496 source,23 evidence,29 cache identities. Arithmetic,
checkpoint and tile hashes agree, but **the job failed**: CUDA memcheck reports
one `cudaErrorPeerAccessAlreadyEnabled` (704) at `cudaGetLastError` while a second
engine enables an existing peer mapping. There are zero leaked bytes; later
regression stages were skipped after memcheck exit99. Preserve [original results](runs/peer-01/results.json)
and [Slurm log](slurm-74454.log); arithmetic success does not make this a clean pass.

Same-allocation warm/profiled/logged dense-tile comparison, s17 width,2^22 values
(32 MiB), H200 NVL pair with NV6 topology, one sample per direction:

| GPUs | Forward whole call | Inverse whole call | Aggregate owned GPU allocations |
|---|---|---|---|
| 1 | 183.296 ms | 158.537 ms | 65.57 MiB |
| 2 | 215.943 ms | 189.873 ms | 67.14 MiB |

[Tile results](runs/peer-01/tiles.json) include the common forward SHA256
`60dd1c3cb32ce4eca3d4f55c2cf4b4ad111c80a08c6f90b4786987ec09fface7`.
Two cards are17.8%/19.8% slower here. These are transform component measurements,
including transfers; not full preparation timings or full-size speedup estimates.
Driver/context headroom is extra. Whole-command sampled host RSS remains below
1 GiB, including build commands; isolated preprocessing host peak is unmeasured.

The correction shares process-scoped, reference-counted directional peer leases
across32/64-bit engines, releases owned links after the last engine's buffers,
and leaves external mappings intact. Constructor failure unwinds leases. This
changes lifecycle bookkeeping, not arithmetic, layouts or ABI; setup overhead is
unmeasured. New cross-width lifetime/reopen tests and existing failed-constructor
checks run under memcheck, without suppressions. [Shared CUDA linking and local checks](peer-lease-checks.json) precede a fresh `peer-02 --devices 2` retry under the same approved
512 MiB aggregate/256 MiB per-card,4CPU/8 GiB,600/900/1200-second guards.
No full s17 run or four-card allocation is admitted by this retry.

Next performance candidate: assign complete vectors to each GPU for local FFTs,
using bulk peer redistribution rather than fine-grained remote recurrence reads.
Source inspection suggests peer access costs; no profile yet attributes the
measured slowdown. A proposed three-array schedule keeps the3N device payload,
with four extra local transpose passes per FFT (about8N word transfers, roughly
1 TiB local memory traffic at s17). Host arrays and retained storage stay the same;
setup/latency benefit is unknown. Compare explicitly before adopting it; current
defaults stay unchanged. More GPUs remain eligible when complete latency improves.

## Corrected peer retry submission

Job74466, source10a362b, is submitted as `peer-02 --devices 2`
([submission and exact resource policy](peer-retry-submission.json)). Results:
`runs/peer-02/`; log: `slurm-74466.log`; cache:
`.cache/liftrs/gpu-s17-preprocessing/peer-02/`. The existing approved two-H200
budget is unchanged. No monitoring or new-result inspection until completion ping.

## Completed corrected peer retry (SCALING-073)

Job74466 on gnode02 **passes**. [Analysis](runs/peer-02/analysis.json) reconciles
498 source,35 evidence and36 retained-cache identities against source10a362b.
[Results/commands](runs/peer-02/results.json), [Slurm log](slurm-74466.log) and
[runner](run.py) preserve the unchanged approved two-H200 budget and profiles.
Cross-width engine overlap/close/reopen tests pass normally and under memcheck;
CUDA reports **zero errors and zero bytes leaked**. Small complete preparation,
s17 high-bit/windows/addresses, full tile forward hashes/inverses, legacy CUDA
online/client and Go/native regression/vet/shared-recurrence sanitizers all pass.
The prior peer-01 failure stays archived. No new lifetime failure is observed.

Same32 MiB s17-width dense tile:1GPU forward184.709 ms/inverse158.943 ms;
2GPU forward216.765 ms/inverse190.268 ms,17.4%/19.7% slower. Owned device totals
remain65.57/67.14 MiB. Same complete forward SHA256 as peer-01. These warm/profiled
component samples include transfers and do not forecast full-size preprocessing.
Next is local-vector FFT/bulk peer communication optimization and measured GPU-
count scaling, with the previously recorded traffic trade-off made explicit.
No full s17 cache, complete run or reviewed full-size allocation exists yet.

## Local-vector comparison (SCALING-074)

After the clean peer retry, implement local vector ownership without duplicating
FFT arithmetic. `run.py --vector-layout local` snapshots the explicit plugin
selector; default remains `peer`. The new kernel calls the shared vector recurrence
with the correct global vector offset. Gather/restore use existing peer transpose
kernels; local transposes reuse existing value/scratch slabs. Taylor stages keep
their current synchronization. Tiny shapes with fewer vectors than cards use the
existing peer schedule. No ABI, canonical output, CPU build or default changes.

**Trade-off before adoption:** same2N transform/3N product payload, same host
arrays, cache/output storage and no additional transforms or host transfers.
The straightforward schedule adds8 transpose passes per FFT (4 local,4 peer
gather/restore),16N word reads/writes, ~2 TiB at s17. Earlier discussion's~1 TiB
count covered only the four local transposes; gather/restore adds another~1 TiB
of aggregate array traffic, with actual peer-link reads a subset. Metadata/setup
are otherwise shared; overhead and latency benefit are unknown. This remains
explicit, never an automatic switch or claimed speedup. Later fusion can remove
redundant passes if measurements justify it.

[Host checks](local-vector-host-checks.jsonl), [progress check](local-vector-progress-checks.jsonl),
[static check commands/source hashes](local-vector-checks.json), [driver](check.cpp),
[runner](run.py) and [analyzer](analyze.py) make the bounded comparison reproducible.
18 optimized host checks pass: full dense independent Horner and inverse oracles
at32/64-bit widths, s4/s17,1/2/4 simulated peer slabs and both layouts. CUDA shared
linking passes under the25-second local check. These are not GPU runtime evidence.

Run `sbatch --gres=gpu:h200:2 --output=experiments/gpu-s17-preprocessing/slurm-%j.log experiments/gpu-s17-preprocessing/run.sbatch local-01 --devices 2 --vector-layout local`.
Use the same approved two-H200,512 MiB aggregate/256 MiB per-device,4CPU/8 GiB,
6 GiB RSS,2 GiB new cache/4 GiB scratch,600/900/1200-second guards. Full small
preparation/high-width checks and memcheck use local mode; the32 MiB dense tile
compares both layouts on each1/2 GPU prefix in the same allocation, with equal
owned payload asserted and identical full forward hash/inverse checks. Shared
host recurrence sanitizer coverage expands from9 to18 cases. Profiling/logging
stay enabled; samples are component timings, not full-size forecasts.

After completion ping: `python3 -B experiments/gpu-s17-preprocessing/analyze.py local-01`.
Keep source revision, logs, profile and reusable small caches; no fulls17 run is
admitted. Larger/more-card calibration and checkpointed full preparation follow
only after correctness/performance review and full resource agreement.

Job74468 is submitted, source62e626b ([submission](local-submission.json)).
Results/log/profile: `runs/local-01/`, `slurm-74468.log`; caches/binaries:
`.cache/liftrs/gpu-s17-preprocessing/local-01/`. No monitoring/result inspection
until user completion ping. All approved numeric resource caps are unchanged.

## Completed local-vector validation (SCALING-075)

Job74468 passes on source62e626b, gnode02. [Analysis](runs/local-01/analysis.json)
reconciles498 source,35 evidence and36 retained-cache identities. [Results](runs/local-01/results.json),
[Slurm log](slurm-74468.log), [controller](run.py), [analyzer](analyze.py) preserve
exact environment, commands, approved limits, binaries, profiles and small caches.
Small complete GPU preparation matches Go, high-bit/window/address checks and
full tile hashes/inverses pass, as do legacy CUDA/client and Go/native regressions,
vet and18 optimized/18 sanitized shared-recurrence cases. CUDA memcheck reports
**zero errors and zero leaked bytes**, including peer lifetime tests.

Same-allocation warm/profiled/logged s17-width dense tile,2^22 values (32 MiB),
one forward/inverse sample per layout/card count; wall times include transfers:

| Vector layout | GPUs | Forward runtime | Inverse runtime | Owned GPU memory (aggregate) |
|---|---|---|---|---|
| Peer | 1 | 197.515 ms | 170.759 ms | 65.57 MiB |
| Peer | 2 | 233.538 ms | 197.709 ms | 67.14 MiB |
| Local | 1 | 194.450 ms | 183.013 ms | 65.57 MiB |
| Local | 2 | 189.407 ms | 178.511 ms | 67.14 MiB |

Local is18.9% faster forward and9.7% faster inverse than peer on two cards;
owned memory is equal. Local two-card versus local one-card gains only2.6%/2.5%.
Single-card local inverse is7.2% slower than peer. All four full forward hashes
match previous tile references. These fixed-order single samples establish
correctness and a candidate benefit, not a stable speedup or fulls17 forecast.
Keep peer default; do not silently adopt extra traffic or change capacity models.

### Prepared larger repeated comparison

[Exact resource/command plan](calibration-plan.json), [static checks](calibration-checks.json).
Runner controls now explicitly accept tile powers22/25, repetitions1/3 and
aggregate owned GPU caps512/4096 MiB, retaining original defaults. Before creating
outputs it checks the largest one-GPU prefix fits its per-device cap with16 MiB
conservative metadata allowance; allocation accounting still enforces actual caps.
The test preserves sparse independent evaluations, dense full hashes and exact
inverses per repetition; analyzer checks every layout/count/repetition, equal
owned payload across layouts and cap compliance. The retained32 MiB job remains
compatible and reconciles with the updated analyzer. This changes future timing
methodology explicitly: three repeated samples per layout/count, fixed order,
profiled/logged; no unprofiled or statistically controlled speedup claim.

Proposed `calibration-01`: four H200s,4 GiB aggregate owned cap (1 GiB/card),
256 MiB input (2^25 values),1/2/4GPU prefixes, peer/local,3 repetitions each.
Host4CPU/8 GiB requested/6 GiB RSS/4 GiB Go soft limit,2 GiB new cache/4 GiB scratch,
and600/900/1200-second deadlines stay unchanged. Two host input arrays consume
512 MiB before runtime/GC/profiles; reusable device transform arrays total512 MiB
plus metadata, within per-card admission even on the one-GPU comparison. No large
new retained input/output files, fulls17 support or preparation is created.

**Await resource agreement before submission.** AGENTS.md requires reviewed
experiment budgets before increasing devices/memory; CLI capacity is not approval.
After approval, use the exact command in calibration-plan.json, then hand off
without polling until completion ping. Performance choice, larger calibration,
checkpointed full preparation and fulls17 resource admission remain separate.

## Four-GPU resource approval (SCALING-076)

User explicitly approves the prepared4H200/4GiB aggregate owned-memory budget,
1GiB/card, for256MiB tiles with3repetitions on1/2/4GPU prefixes and both layouts.
All host/storage/deadline limits in calibration-plan.json stay unchanged. This
approval is only bounded calibration; fulls17 preprocessing remains unadmitted.

Job74635 submitted, sourcefc15997 ([submission/resource record](calibration-submission.json)).
Results: `runs/calibration-01/`; log: `slurm-74635.log`; reusable caches/binaries:
`.cache/liftrs/gpu-s17-preprocessing/calibration-01/`. sbatch returned successfully.
No monitoring, polling or new-result inspection until user completion ping.
