# Encoding scaling reference

The performance goal is **reasonable encoding performance through s=18 inclusive**
on the HPC system, retaining CPU and optional GPU paths. **s=19 and s=20 remain
exploratory targets**: keep their rows as algorithms, layouts and measurements
improve. This is the maintained capacity and results reference for the
[scaling plan](encoding-scaling.md), separate from TRL 5 acceptance.

Updated 2026-11-17; full s17 encoding source `ed0c026`, reviewed in `284c991`. Standalone mathematical
CPU encoding is validated through **s=17** (resident wide, job76645); GPU encoding
is validated through s=16 plus bounded wider integration. **Full s17 GPU76646 was
attempted and failed its host/cgroup guard before completing encoding.** **Retry76706 is submitted with640 GiB requested host RAM/608 GiB total guard;
its result is pending.** Other s17 modes and s≥18 full encoding remain unvalidated; projections are not support. “Reasonable”
has no agreed numerical latency threshold yet; select it using measured setup,
encoding, memory, I/O and deployment costs before accepting a scaling milestone.

## Current s=17 results

| s | Preprocessing | CPU encoding | GPU encoding |
|---|---|---|---|
| **17** | **HPC: passed, job 75832**<br>Verified resumed worker: **9 h 58 min 38 s**; earlier discovery: **21 min 38 s**<br>Peak sampled process RSS: **348.393 GiB**<br>Hardware: one HPC node, 32 active CPU workers, 4 H200 GPUs | **HPC: passed, job 76645 — resident reuse mode**<br>Encoding: **1 h 45 min 11.087 s** (profiled)<br>Peak sampled process RSS: **427.845 GiB**<br>Hardware: one HPC node, 32 CPU cores, 768 GiB requested RAM, no GPU<br>Verified packed output: **34 GiB**; complete job: **2 h 22 min 49 s** | **HPC: attempted, job 76646 — resident reuse mode; failed host-memory guard**<br>Encoding time: **unknown; did not complete**<br>Setup owned device allocations: **398.463 GiB across 4 H200 GPUs**<br>Peak sampled process RSS before stopping: **221.577 GiB**; file cache additional<br>Hardware: one HPC node, 32 CPU cores, 384 GiB requested RAM<br>**Retry 76706 submitted:** 640 GiB host RAM / 608 GiB total guard; result pending |

Sources: [preprocessing result](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088)
and [CPU/GPU encoding results](../experiments/s17-full-encoding/report.md#completed-first-attempts-scaling-093).
The detailed tables put **resident reuse first**, with CPU76645 marked passed
and GPU76646 marked attempted/failed. Other s17 modes are labeled **projection
only; no full measurement for this mode**. They have no measured runtime or peak
memory and must not inherit resident results. Packing
resident CPU output after encoding does not validate a direct packed-output mode.

The [22-case HPC batch](../experiments/encoding-table/report.md) completed the
supported encoding-mode measurements through s16. All results have been
reconciled; laptop entries remain historical evidence, not HPC measurements. CPU jobs74194–74203 and GPU jobs74204–74215 are recorded in its
submission manifest. The [three remaining CPU retries](../experiments/encoding-table/report.md#extended-cpu-retries)
use fresh fill-02 paths and extended deadlines: packed s16 job74251, disk s15
job74252, disk s16 job74253 replace74202/74201/74203 respectively. The
[s16 disk result](../experiments/encoding-table/report.md#completed-disk-s16) now
passes complete source/evidence/large-cache reconciliation in job74447.
All22 cases pass complete source/evidence/cache and cross-mode hash reconciliation
in job74467; their measured HPC cells are now published below (marked§).

The active server-preprocessing target is the
[resident GPU prototype](../experiments/gpu-s17-preprocessing/report.md), with
width-selected kernels and same-node peer shards. Source implementation supports
s17; bounded one-, two- and four-GPU validation passes, including the corrected
peer-lifetime sanitizer retry74466, followed by the passed full s17 run75832.
Larger preprocessing runs require separate admission. Four H200s
model96 GiB of product arrays each (384 GiB aggregate), excluding overhead. This
prototype does not accept three-card shards; the older three-card figure is
only an arithmetic capacity floor. CPU preprocessing is a reference path. GPU
counts in these projections describe capacity, not a performance ceiling. In
particular, preprocessing is not limited to two GPUs: compare more cards and
multiple nodes when they improve complete latency, within reviewed budgets.


## Reading the tables

All rows use m=2, t=5 and one field-element database. Every cell names its run
status (HPC measured, laptop measured, attempted/failed, projection only or not
run), runtime units and memory
measure. `~` marks an estimate rather than a measured peak. Unknown values are
written explicitly; GPU counts in unrun cells are capacity projections.
Preparation excludes discovery, cache loading/saving and database encoding unless
explicitly labeled otherwise (s17 construction includes root-checkpoint I/O).
Ordinary encoding timings exclude output packing/storage unless marked **stored**;
those stored timings include packing, writing and hashing.

Measured CPU memory is sampled process RSS, except explicitly labeled small-case
array payloads. GPU memory is **owned device allocations / host RAM**; context,
driver and staging overhead are extra. Estimates exclude runtime overhead and
filesystem cache. Different revisions, core counts and machines mean these are
reference results, not controlled speedup comparisons. Detailed methodology,
profiles, source identities, scripts and results remain in the linked reports.

## Hardware labels

- **Laptop:** Apple M5 Pro, 18 logical CPUs, 48 GiB unified RAM, macOS arm64.
  Historical native encoding runs generally used 16 workers; individual reports
  give the exact settings.
- **HPC:** Linux cluster; observed GPU nodes have 128 CPU cores, about 907 GiB
  scheduler RAM and four NVIDIA H200 GPUs, each with 140.40 GiB device RAM.
  These are node capacities, not resources allocated to every job. Encoding
  table CPU measurements generally use 32 workers; GPU cells state card counts.
  Exact allocations and software versions are in the linked experiment reports.

For the full s17 attempts, CPU76645 used 32 CPU cores and 768 GiB requested host
RAM without a GPU; GPU76646 used 32 CPU cores, 384 GiB requested host RAM and
four H200s with a 112 GiB owned-memory cap per card. Requested RAM and allocation
caps are separate from measured memory in the result cells.

## HPC capacity and feasibility

Scheduler configuration read2026-11-16 lists17 GPU nodes and **68 configured GPUs**,
superseding the older67-card inventory snapshot. Observed nodes have128 CPU cores,
928610 MiB scheduler RAM (about **907 GiB**) and four H200s. Configured totals do
not certify availability/health; gnode03 remains excluded after its unresolved fault. The measured H200 NVL has 143771 MiB (**140.40 GiB**) device RAM.
Node inventory uses `module load slurm/slurm/25.05` and
`sinfo -p gpuq -N -o '%N %c %m %G'`; current partition totals are recorded in the
[configuration snapshot](../experiments/s17-preprocessing-admission/resource-snapshot.json). Inventory is not a reservation or an experiment
budget; review current availability and agree budgets before larger submissions.

GPU counts below are **capacity floors**, dividing aggregate array bytes by one
card's memory and rounding up. They assume sharded arrays/constants rather than
replication. Communication, staging, runtime and allocation headroom can require
more GPUs. An experimental resident multi-GPU client now exists in source, with bounded
validation passing76426 (server preprocessing is verified separately). Time cannot be estimated
by dividing a single-GPU timing by the card count.

- s≤16: measured paths below fit the HPC system; unrun variants are capacity
  projections, with runtime unknown. Existing job budgets still apply.
- s=17: CPU array models fit one large-memory node. Resident GPU modes project
  two to four GPUs; the four-H200 resident reuse attempt failed its host-memory
  guard. CPU resident reuse passed; other modes remain unvalidated. Host-backed
  encoding projects one GPU with about 349 GiB host arrays.
- s=18: disk-backed CPU encoding projects about 317 GiB private arrays and 1 TiB
  work files. Streamed resident GPU encoding projects eight GPUs across at least
  two four-GPU nodes. Preprocessing and communication remain separate gates.
- s=19–20: the modeled owning CPU outputs alone exceed one node's RAM; GPU
  capacities appear below. New layouts, distributed execution or further memory
  savings may change these projections. They remain exploratory, not ruled out.

## Aim for s17

The initial full-size CPU target is complete: **resident reuse encoding passes
at s17 on HPC**, using verified GPU preprocessing. See the current-results table
above for timings, memory and hardware. CPU output retains 8N backing bytes
(4N logical uint32 bytes); packing produces a separate tight 34 GiB file.

The [CPU refactor comparison](../experiments/cpu-encoding-regression/report.md)
passed at s14: encoding was 1.28% slower, with setup/loading overhead reported
separately. The [shared CPU implementation](cpu-encoding.md),
[width-aware code data](scaling-code-data.md),
[wide scalar reference](../experiments/s17-arithmetic/report.md),
[native kernels](../experiments/s17-native/report.md) and
[matrix validation](../experiments/s17-matrix/report.md) supply the independent
checks behind the full run. CPU-only client builds remain CUDA-free.

The [resident GPU integration](../experiments/gpu-code-encoding/report.md) passes
bounded correctness checks. Its full s17 attempt completed setup on four H200s
but stopped at the 360 GiB total host-memory guard before the first transform;
process RSS and source-file cache together exceeded the admitted budget. The authorized retry is being prepared with **640 GiB requested host RAM and a
608 GiB total-memory guard**, superseding the512/480 GiB proposal and keeping
the320 GiB process-RSS limit and112 GiB/card owned-memory cap. **Job76706 is
submitted; result pending**; [budget and recipe](../experiments/s17-full-encoding/report.md#gpu-retry-with-conservative-host-headroom-scaling-098). Full GPU encoding time and CPU/GPU output agreement remain
unknown. This is a host/cache admission issue observed during this attempt, not
a demonstrated full GPU encoding success.

| s17 route | Current evidence or projected payload | Next action |
|---|---|---|
| CPU resident reuse | **Measured: 427.845 GiB peak process RSS; encoding 1 h 45 min 11.087 s** | Full run passed; no repeat needed for correctness |
| CPU file-streamed | Unrun mode; ~270 GiB host arrays, reusable code file ~142 GiB (projections) | Optional lower-RAM alternative; measure repeated reads/cache costs if needed |
| GPU resident reuse | **Attempted: 398.463 GiB setup device allocations; full encoding incomplete** | Retry with reviewed host/cache headroom |
| GPU fast-layout streamed | Unrun mode; ~256 GiB aggregate device, ~221 GiB host arrays (projections) | Separate wider port/validation; two H200s are only a capacity floor |
| GPU host-backed | Unrun mode; few MiB device tiles, ~349 GiB host arrays (projections) | Separate wider port/validation; transfer cost unknown |

The HPC node capacity is about 907 GiB schedulable RAM and 140.4 GiB per H200.
Array models exclude runtime, cache, context and allocation headroom. Keep the
faster validated path when it fits the admitted budget; slower memory modes stay
explicit options. Same-node resident GPU execution supports 1/2/4/8 devices in
source, with bounded 1/2/4-device validation; arithmetic capacity floors do not
imply support for arbitrary card counts or distributed execution.

## Preprocessing and CPU encoding

Reuse aliases buffers after their last use. Compact bounds transform scratch.
File streams code constants; packed file also writes packed output directly.
Disk keeps transform work arrays in files but still returns an owning output.
These are distinct modes; their savings cannot simply be added together.

### Preprocessing

These are **reusable preparation** timings, not database encoding time. The s17
row separates discovery, construction (including root checkpoint I/O) and the
resumed verified worker; other rows use their historical boundaries. Host memory
excludes device allocations unless stated.

| s | Run status | Host memory | Preparation runtime | Feasibility / evidence |
|---|---|---|---|---|
| 10 | HPC measured | Whole-command peak RSS: ≤0.24 GiB | CPU: 6.65 s; GPU: 1.53 s | Measured on HPC |
| 14 | Laptop measured | Peak host process RSS: 13.73 GiB | 33 min 38 s | Measured on laptop |
| 15 | Laptop measured | Peak host process RSS: 20.55 GiB | 3 h 16 min 7 s† | Measured on laptop |
| 16 | HPC measured | Peak host process RSS: 72.575 GiB | GPU: 2 h 20 min 15.469 s | Measured on HPC |
| 17 | **HPC measured; passed75832**<br>K:3876566203; support from75819 | **Peak sampled host RSS:348.393GiB**<br>Device peak:not sampled; product arrays384GiB aggregate (model)<br>Final device owned:256.009GiB observed | **Construction/identity/spectrum:9h16min46s**, includes root checkpoints<br>Verified resumed worker:9h58min38s<br>Earlier discovery:21min38s | **Demonstrated on HPC:4H200,32 active CPU workers, one node**<br>Canonical preprocessing:235.324GiB<br>[Full evidence/boundaries](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088) |
| 18 | Full preprocessing:not run<br>**Exact dimension measured76647:K16300262095** | Projected final host arrays:889.446GiB<br>Projected discovery arrays:770.893GiB<br>Peak RSS:unknown | Full preprocessing:unknown<br>Count-only GPU sum:37.4ms, excludes setup |1536GiB device product arrays (model)<br>16GPU power-of-two projection across4nodes at112GiB/card<br>Distributed host/GPU implementation and larger storage admission required |
| 19 | Not run | Peak RAM: unknown | unknown | Implementation/admission pending |
| 20 | Not run | Peak RAM: unknown | unknown | Implementation/admission pending |

### CPU encoding: resident modes

| s | **Resident CPU reuse (current baseline)** | Original CPU allocation (historical comparison) | Compact CPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.313 s–0.319 s<br>Array payload: ~13.7 MiB | HPC measured<br>Runtime: 0.308 s–0.31 s<br>Array payload: ~17.7 MiB | HPC measured§<br>Runtime: 0.317 s–0.319 s<br>Peak workflow process RSS: 30.387 MiB |
| 14 | HPC measured<br>Runtime: 58.667 s–58.853 s<br>Peak process RSS: 3.595 GiB | HPC measured<br>Runtime: 1 min 31.164 s<br>Peak process RSS: 5.594 GiB | HPC measured<br>Runtime: 1 min 4.517 s–1 min 8.413 s<br>Peak process RSS: 2.526 GiB |
| 15 | HPC measured§<br>Runtime: 4 min 34.704 s–4 min 41.715 s<br>Peak workflow process RSS: 14.460 GiB | HPC measured<br>Runtime: 10 min 55.937 s<br>Peak process RSS: 18.465 GiB | HPC measured§<br>Runtime: 4 min 39.900 s–4 min 40.038 s<br>Peak workflow process RSS: 9.901 GiB |
| 16 | HPC measured<br>Runtime: 20 min 47.173 s–21 min 21.328 s<br>Peak process RSS: 58.351 GiB | HPC measured<br>Runtime: 20 min 52.503 s<br>Peak process RSS: 74.370 GiB | HPC measured<br>Runtime: 21 min 29.759 s<br>Peak process RSS: 39.409 GiB |
| 17 | **HPC: passed, job 76645; resident reuse**<br>Encoding: **1 h 45 min 11.087 s**, profiled<br>Peak sampled process RSS: **427.845 GiB**<br>Complete job: **2 h 22 min 49 s**<br>Separate packed output: **34 GiB**, verified<br>**Feasible: one HPC node, 32 CPU cores, 768 GiB requested RAM** | **Projection only: original CPU**<br>No full s17 measurement for this mode<br>Runtime:unknown<br>Estimated arrays:~491 GiB<br>Not validated at s17 | **Projection only: compact CPU**<br>No full s17 measurement for this mode; wider port unvalidated<br>Runtime:unknown<br>Estimated arrays:~287 GiB<br>Capacity projection only |
| 18 | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM |
| 19 | Not run<br>Runtime: unknown<br>Output alone: 1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Output alone: 1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Output alone: 1024 GiB<br>Capacity: exceeds one node's RAM |
| 20 | Not run<br>Runtime: unknown<br>Output alone: 4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Output alone: 4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Output alone: 4096 GiB<br>Capacity: exceeds one node's RAM |

### CPU encoding: file and disk modes

Packed-file runtime includes writing and hashing packed output. Other runtimes
measure encoding; load, verification and separate packing are additional.

| s | File CPU | Packed-file CPU (stored output) | Disk-backed CPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.311 s–0.332 s<br>Array payload: ~8.6 MiB | HPC measured<br>Runtime: 0.325 s–0.327 s<br>Array payload: ~9.6 MiB | HPC measured<br>Runtime: 2.86 s first; 1.136 s reused<br>Estimated private arrays: ~5.7 MiB<br>Work files: 8 MiB |
| 14 | HPC measured<br>Runtime: 59.357 s–59.377 s<br>Peak process RSS: 2.214 GiB | HPC measured§<br>Runtime: 1 min 0.382 s–1 min 1.263 s<br>Peak workflow process RSS: 2.651 GiB<br>Peak tracked file residency: 1.633 GiB (additional) | HPC measured§<br>Runtime: 11 min 24.200 s–12 min 38.350 s<br>Peak workflow process RSS: 1.213 GiB<br>Work files: 2.000 GiB<br>Peak tracked file residency: 3.511 GiB (additional) |
| 15 | HPC measured§<br>Runtime: 4 min 41.053 s–4 min 41.738 s<br>Peak workflow process RSS: 8.838 GiB<br>Peak tracked file residency: 5.256 GiB (additional) | HPC measured§<br>Runtime: 4 min 45.251 s–4 min 47.455 s<br>Peak workflow process RSS: 10.715 GiB<br>Peak tracked file residency: 6.688 GiB (additional) | HPC measured§<br>Runtime: 1 h 16 min 13.550 s–1 h 25 min 3.987 s<br>Peak workflow process RSS: 4.841 GiB<br>Work files: 8.000 GiB<br>Peak tracked file residency: 13.364 GiB (additional) |
| 16 | HPC measured<br>Runtime: 20 min 49.616 s<br>Peak process RSS: 35.484 GiB | HPC measured§<br>Runtime: 21 min 32.972 s<br>Peak workflow process RSS: 43.487 GiB<br>Peak tracked file residency: 27.427 GiB (additional) | HPC measured§<br>Runtime: 10 h 12 min 39.871 s<br>Peak workflow process RSS: 19.469 GiB<br>Work files: 32.000 GiB<br>Peak tracked file residency: 59.271 GiB (additional) |
| 17 | **Projection only: file-streamed CPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated array payload: ~270 GiB<br>Capacity: one-node RAM projected; full run unvalidated | **Projection only: direct packed-file CPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated array payload: ~270 GiB<br>Capacity: one-node RAM projected; full run unvalidated | **Projection only: disk-backed CPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated private arrays: ~78.4 GiB<br>Work files: 256 GiB<br>Capacity: one-node RAM projected; full run unvalidated |
| 18 | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~317 GiB<br>Work files: 1024 GiB<br>Capacity: one-node RAM projected; disk capacity unverified |
| 19 | Not run<br>Runtime: unknown<br>Output alone: 1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~1279 GiB<br>Work files: 4096 GiB<br>Capacity: exceeds one node's RAM |
| 20 | Not run<br>Runtime: unknown<br>Output alone: 4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥16384 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~5162 GiB<br>Work files: 16384 GiB<br>Capacity: exceeds one node's RAM |

† Historical s15 preparation has pause/clock complications; it is not a clean
latency baseline or an HPC forecast. Discovery was additional: about 1.1 s at
s10, 255 s at s14, 1131 s at s15, and 10597.545 s at s16. The s16 CPU-only
preparation attempt was stopped after about nine hours; it did not complete.
GPU-assisted preparation completed with a 48 GiB device-array model under a
64 GiB device cap; 48 GiB is not a measured device peak. s10 GPU preparation used
one GPU under a 512 MiB cap; its device peak was not isolated.

Older CPU measurements use16 cores at s10/s15 and32 at s14/s16; batch§
allocations use32 cores with per-harness worker settings documented below. Small-case payloads
omit runtime/metadata. File-backed RSS does not include all filesystem cache:
s16 file CPU reached **61.820 GiB RSS plus tracked cache** across the workflow,
versus 35.484 GiB peak RSS. Disk-backed latency can vary substantially with
filesystem/cache state; its full encoder is now measured at s10/s14/s15/s16 (the s16 retained-cache reread passes job74447). File and disk
capacity models do not establish storage throughput or runtime at larger sizes.

Sources: [preparation s10](../experiments/gpu-products/report.md),
[laptop s14](../experiments/parallel-s14/report.md),
[laptop s15](../experiments/parallel-s15/report.md),
[s16 discovery/CPU attempt](../experiments/hpc-s16/report.md),
[s16 GPU preparation](../experiments/gpu-preparation/report.md);
[CPU s10 baseline](../experiments/cuda-encoding/report.md),
[CPU s14 baseline](../experiments/cpu-encoding-profile/report.md),
[CPU s15](../experiments/hpc-cached-s15/report.md),
[CPU s16](../experiments/cached-s16-encoding/report.md),
[CPU reuse](../experiments/cpu-buffer-reuse/report.md),
[CPU compact](../experiments/cpu-output-memory/report.md),
[CPU file](../experiments/cpu-streamed-constants/report.md),
[CPU packed file](../experiments/cpu-packed-output/report.md),
[CPU disk](../experiments/cpu-disk-encoding/report.md).
The s10 reuse comparison is from the matched CPU run in the GPU host-backed report below.

## GPU encoding

Every cell labels encoding runtime, device memory, host memory and GPU count.
Stream uses the fast computation layout while streaming code constants. Packed
stream also writes packed output directly. Host-backed keeps full arithmetic
arrays in host RAM and transfers tiles; it is a slower memory alternative.

### GPU encoding: resident modes

Measured GPU memory is explicitly owned device allocation; measured host memory
is sampled process RSS. `~` denotes estimated array payload. Multi-GPU payloads
are aggregate, not per GPU. GPU counts in unrun cells are capacity projections,
not implemented support; the feasibility discussion above still applies.

| s | **Resident GPU reuse (current full-size attempt)** | Original GPU allocation (historical comparison) | Compact GPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.103 s<br>GPU array payload: 12.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.104 s<br>GPU array payload: 16.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.108 s–0.109 s<br>GPU array payload: 10.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 |
| 14 | HPC measured<br>Runtime: 5.176 s–5.179 s<br>Owned GPU allocations: 3.190 GiB<br>Estimated host arrays: ~2.38 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 5.155 s–5.257 s<br>Owned GPU allocations: 4.190 GiB<br>Peak workflow host RSS: 2.521 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 5.800 s–5.860 s<br>Owned GPU allocations: 2.192 GiB<br>Peak workflow host RSS: 2.525 GiB<br>GPUs: 1 |
| 15 | HPC measured§<br>Runtime: 16.917 s–17.756 s<br>Owned GPU allocations: 12.809 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 16.853 s–17.497 s<br>Owned GPU allocations: 16.808 GiB<br>Peak workflow host RSS: 9.768 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 20.487 s–20.873 s<br>Owned GPU allocations: 8.811 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 |
| 16 | HPC measured<br>Runtime: 56.975 s–57.418 s<br>Owned GPU allocations: 51.427 GiB<br>Peak host process RSS: 39.031 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 0.216 s–1 min 25.954 s<br>Owned GPU allocations: 67.423 GiB<br>Peak host process RSS: 39.024 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 33.332 s–1 min 41.526 s<br>Owned GPU allocations: 35.429 GiB<br>Peak host process RSS: 39.028 GiB<br>GPUs: 1 |
| 17 | **HPC: attempted, job 76646; failed host-memory guard**<br>Encoding time: **unknown; did not complete**<br>Setup owned device allocation: **398.463 GiB aggregate, 4 H200 GPUs**<br>Peak sampled process RSS: **221.577 GiB**; file cache additional<br>Host request: **384 GiB**; total-memory guard: **360 GiB**, insufficient<br>**Retry 76706 submitted:** 640 GiB host RAM / 608 GiB total guard<br>Result pending; full GPU encoding unvalidated | **Projection only: original GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~526.44 GiB<br>Estimated host arrays: 220.88 GiB<br>GPUs: 4<br>Capacity: one node projected; implementation pending | **Projection only: compact GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~270.44 GiB<br>Estimated host arrays: 220.88 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending |
| 18 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1597 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 12<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~2109 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 16<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1085 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 8<br>Capacity: multiple nodes required |
| 19 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~6399 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 46<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~8447 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 61<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4351 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 31<br>Capacity: multiple nodes required |
| 20 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~33834 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 241*<br>Capacity: exceeds HPC GPU inventory or per-node host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~33834 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 241<br>Capacity: exceeds HPC GPU inventory or per-node host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~25642 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 183*<br>Capacity: exceeds HPC GPU inventory or per-node host RAM |

### GPU encoding: streamed and host-backed modes

| s | Stream GPU | Packed-stream GPU (stored output) | Host-backed GPU |
|---|---|---|---|
| 10 | HPC measured§<br>Runtime: 0.098 s–0.106 s<br>Owned GPU allocations: 8.035 MiB<br>Peak workflow host RSS: 20.645 MiB<br>GPUs: 1 | HPC measured§<br>Runtime: 0.107 s–0.114 s<br>Owned GPU allocations: 8.035 MiB<br>Peak workflow host RSS: 28.113 MiB<br>GPUs: 1<br>Peak tracked file residency: 10.578 MiB (additional) | HPC measured<br>Runtime: 1.606 s–1.617 s<br>GPU array payload: 0.534 MiB<br>Host array payload: 13.4 MiB<br>GPUs: 1 |
| 14 | HPC measured<br>Runtime: 5.467 s–5.527 s<br>Owned GPU allocations: 2.000 GiB<br>Estimated host arrays: ~2.38 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 5.804 s–5.827 s<br>Owned GPU allocations: 2.000 GiB<br>Estimated host arrays: ~1.39 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 1 min 46.837 s–1 min 48.589 s<br>Owned GPU allocations: 0.710 MiB<br>Peak workflow host RSS: 3.521 GiB<br>GPUs: 1 |
| 15 | HPC measured§<br>Runtime: 18.537 s–18.564 s<br>Owned GPU allocations: 8.001 GiB<br>Peak workflow host RSS: 9.768 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 20.451 s–21.448 s<br>Owned GPU allocations: 8.001 GiB<br>Peak workflow host RSS: 7.646 GiB<br>GPUs: 1<br>Peak tracked file residency: 6.689 GiB (additional) | HPC measured§<br>Runtime: 8 min 45.294 s–9 min 7.208 s<br>Owned GPU allocations: 0.898 MiB<br>Peak workflow host RSS: 13.777 GiB<br>GPUs: 1 |
| 16 | HPC measured<br>Runtime: 58.851 s–1 min 3.487 s<br>Owned GPU allocations: 32.005 GiB<br>Peak host process RSS: 39.029 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 5.485 s–1 min 6.227 s<br>Owned GPU allocations: 32.005 GiB<br>Peak host process RSS: 23.014 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 1 h 1 min 17.511 s<br>Owned GPU allocations: 1.273 MiB<br>Peak workflow host RSS: 55.049 GiB<br>GPUs: 1 |
| 17 | **Projection only: streamed GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~256 GiB<br>Estimated host arrays: 220.88 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending | **Projection only: direct packed-stream GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~256 GiB<br>Estimated host arrays: 156.88 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending | **Projection only: host-backed GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~348.88 GiB<br>GPUs: 1<br>Capacity: one node projected; implementation pending |
| 18 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1024 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 8<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1024 GiB<br>Estimated host arrays: 633 GiB<br>GPUs: 8<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~1401 GiB<br>GPUs: 1 (host exceeds node RAM) |
| 19 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4096 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 30<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4096 GiB<br>Estimated host arrays: 2557 GiB<br>GPUs: 30<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~5629 GiB<br>GPUs: 1 (host exceeds node RAM) |
| 20 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~24576 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 176*<br>Capacity: exceeds HPC GPU inventory or per-node host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~24576 GiB<br>Estimated host arrays: 10324 GiB<br>GPUs: 176*<br>Capacity: exceeds HPC GPU inventory or per-node host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~30804 GiB<br>GPUs: 1 device, host exceeds node<br>Capacity: exceeds HPC GPU inventory or per-node host RAM |

Host estimates in multi-GPU rows are **aggregate**, not per node; distributing
arrays or avoiding replication requires an explicit host as well as device design.
At s18 the ~889 GiB host model leaves little room on one node; distribute it or
reduce it rather than assuming the payload fits with runtime headroom.

* At s20, the estimated K makes the prefix workspace exceed the current N-sized
arena. These entries allow an expanded arena; the current modes reject that
shape. A redesign could change the counts. Host-backed s20 likewise needs a
larger arena, included in its host estimate. Unrun s≥17 modes require separate wider-mode implementation/validation; resident
CPU reuse has passed, and resident GPU reuse has been attempted as recorded above.

The packed s16 host peak is encoding-only; verification reaches 31.014 GiB RSS,
and RSS plus tracked cache reaches 60.395 GiB. Host-backed full encoding is now
validated through s16 (61 min17.511 s,55.049 GiB workflow host RSS); tiny tiles substantially increase transfers and runtime.
Neither backed encoder currently combines its storage strategy with direct packed
output. Compact plus streaming plus packed output is also not an integrated mode.

Sources: [original s10](../experiments/cuda-encoding/report.md),
[original s16](../experiments/cuda-s16-encoding/report.md),
[reuse](../experiments/cuda-buffer-reuse/report.md),
[compact](../experiments/gpu-output-memory/report.md),
[stream](../experiments/gpu-streamed-constants/report.md),
[packed stream](../experiments/gpu-packed-output/report.md),
[host-backed](../experiments/gpu-host-encoding/report.md).

## Preprocessing GPU capacity projections

Array counts below are **models excluding device/runtime headroom**. The s17
four-card run is now demonstrated; s≥18 is not run. Three resident extension-field
arrays are modeled; cross-node execution is not implemented. Streaming/external
memory may change requirements. s≥18 runtime/peak host memory remains unknown.

| s | Aggregate device arrays | GPU capacity floor | Host/machine feasibility |
|---|---|---|---|
| 17 |384GiB modeled arrays |3 arithmetic floor;4 implemented |**HPC passed on4H200, one node**; sampled host peak348.393GiB; device metadata/headroom extra |
| 18 |1536GiB modeled arrays |11physical floor;14at112GiB cap;16power-of-two |At least4nodes for16cards; final host arrays889.446GiB, distributed host/transport unimplemented; peak unknown |
| 19 | ~6144 GiB | 44 | At least 11 four-GPU nodes; host distribution and peak unknown |
| 20 | ~24576 GiB | 176 | Exceeds configured68-GPU inventory; needs a different layout or more hardware |

## Projection assumptions and upkeep

For q=2^s, N=q², use exact K=918672421 at s16, **3876566203 at s17** and
**16300262095 at s18**. Only s19–20 still use catalogue estimates
K=68331898887 and286146801617. Rounded projections below use these counts;
measured rows retain their original evidence and methodology. Base values and
canonical outputs use four-byte slots. Internal extension values use four bytes
through s16 and eight-byte extension slots thereafter (implemented at s17). Let
D=internal bytes × N, P=4(K+1), M=4K and W=4N. Principal-array models are:

- CPU original: P+D+2M+2D+W; reuse: P+D+2M+2D.
- CPU compact: P+2D+M+bounded column scratch; file: 2D+M;
  packed file adds a writer buffer; disk private: W+M+bounded buffers, files: 2D.
- GPU original: 4D+P; reuse: 3D+P; compact: 2D+P plus bounded scratch;
  stream/packed stream: 2D. GPU preparation: 3D.
- GPU host harness: P+D+M+W; packed stream removes W; host-backed adds its
  host arena D. At s20 the expanded arena adds D to reuse, compact, stream and
  host-backed models; the original already has that space.

These are extensions of the [source-linked memory audit](../experiments/encoding-memory/report.md)
and subsequent mode reports, not new benchmark results. The prefix requirement
is 2L≤N with L=nextpow2(2K−1); s20 fails it. Packed output itself occupies Ns/8
bytes, separately from four-byte canonical output and internal workspace.

**Update this reference with every relevant scaling result or design change.**
Update the affected cells and date/checkpoint; link the exact experiment report
and results, with hardware, threads, parameters and measurement boundaries.
Preserve measured versus projected status, revise GPU/node counts and host/disk
needs together, and explain changed assumptions. Keep s19–20 visible even when
projections look unfavorable. The scaling plan records implementation steps;
this reference records current evidence and capacity.

Next work is the host-budget correction for the attempted s17 GPU resident run
and the [s18 preprocessing memory prototypes](#s18-lower-memory-preprocessing-work).
Full s17 preprocessing and the resident CPU baseline have passed; wider
arithmetic is implemented and validated for those completed paths. Present memory–time/I/O trade-offs before adopting them;
retain the fastest validated mode by default when it fits agreed budgets. Slower
memory-saving modes remain explicit command-line choices when memory is available.
This target does not authorize larger allocations or change the rule to hand off
jobs expected to exceed 30 seconds without monitoring.

‡ Job74253, gnode17:32 cores allocated,16 native workers. One profiled s16 disk
encoding took36759.871s; load92.555s, verification382.655s, packing60.636s and
whole controller37504.672s are separate. Checks cover all918672421 systematic
values and16 curves; canonical hash matches previous s16 encodings. Sampled
tracked file residency peaked59.271GiB, and sampled RSS-plus-tracked-cache peaked
78.736GiB; this sum is not exact physical or cgroup usage.32GiB work files and
8GiB packed output are retained. Source and sealed logs/profiles/phase evidence
reconcile; independent large retained-cache rereads pass jobs74447 and74467.
Logical reads3.878TiB/writes0.808TiB explain a substantial I/O deployment cost;
these are interface counts, not measured physical disk traffic. This slower
memory mode remains explicit; it changes no default or full s17 admission.

## s18 lower-memory preprocessing work

The [concrete plan](encoding-scaling.md#s18-preprocessing-memory-design-and-comparison-scaling-096)
now starts with host lifetimes and a bounded streamed-spectrum prototype, then
measures a host-backed tiled transform and a split product before comparing
four-H200 hybrid versus sixteen-H200 distributed preprocessing.
[Source-derived model and validation recipe](../experiments/preprocessing-memory/report.md):
streaming the final spectrum projects final s18 host arrays **889.446 → 377.447 GiB**
with a 1 MiB tile. This is a final-stage projection, **not peak RSS or demonstrated
s18 feasibility**: discovery (~770.893 GiB) and device product arrays (1536 GiB)
are unchanged, and other intermediates/cache/headroom remain. **Bounded validation76704 passed** on HPCgnode04 (4H200,4 CPU cores,8 GiB host
request;3 min59 s whole job). Independent scalar s8/s17 cases and CUDA memcheck
pass;561 source/43 evidence/74cache files reconcile. For a32 MiB spectrum with
1 MiB tiles, resident transform/download/hash-write takes222–242 ms; stream takes
215–218 ms (two samples each). This is a component measurement with a memory hash
sink, not disk output or full preprocessing. No full-size runtime penalty or RSS
saving is established. [Results and boundaries](../experiments/preprocessing-memory/report.md#completed-bounded-validation-scaling-099).

Next: integrate explicit verified canonical publication and scoped host lifetimes,
then measure host-backed tiled transforms and split products. Full preprocessing
limits, faster defaults and all full-size measurement cells are unchanged.


## Measurement boundaries for batch entries

§ [Completed22-case HPC batch](../experiments/encoding-table/report.md#completed-all-case-review-scaling-073)
uses32 allocated CPU cores and retained preprocessing, with original native
worker settings captured in the run logs. The jobs are profiled, at different
revisions/nodes from older baselines; these rows are not controlled speedups.
One encoding sample at s16, two at smaller sizes. New memory entries are sampled
**whole-workflow** process RSS (including verification/readback), not isolated
encoding peaks. File residency is additional and can overlap RSS; it is not an
exact physical-RAM sum. All22 cases complete on the HPC system under their
recorded budgets. These batch measurements stop at s16; the separate s17
resident CPU pass and GPU guard failure are recorded in the current tables.


Historical intermediate plans and calibration tables are preserved in the
[scaling history](encoding-history.md). They do not override current result cells.
