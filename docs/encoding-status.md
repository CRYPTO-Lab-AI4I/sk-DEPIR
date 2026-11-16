# Encoding scaling reference

The performance goal is **reasonable encoding performance through s=18 inclusive**
on this HPC system, retaining CPU and optional GPU paths. **s=19 and s=20 remain
exploratory targets**: keep their rows as algorithms, layouts and measurements
improve. This is the maintained capacity and results reference for the
[scaling plan](encoding-scaling.md), separate from TRL 5 acceptance.

Updated 2026-11-16; validated code-data checkpoint `83c8469`. Standalone mathematical
encoding is validated through s=16. **Full encoding at s≥17 is unvalidated**; verified larger
preprocessing, layout admission and any multi-GPU execution must be implemented and
validated before those projections become demonstrated capabilities. “Reasonable”
has no agreed numerical latency threshold yet; select it using measured setup,
encoding, memory, I/O and deployment costs before accepting a scaling milestone.

A [22-case HPC batch](../experiments/encoding-table/report.md) was submitted to fill
the supported unrun encoding cells through s16. Entries remain “Not run” until completed
results are reconciled; laptop entries are historical evidence, not HPC
measurements. CPU jobs74194–74203 and GPU jobs74204–74215 are recorded in its
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
peer-lifetime sanitizer retry74466. Full-size admission remains separate. Four H200s
model96 GiB of product arrays each (384 GiB aggregate), excluding overhead. This
prototype does not accept three-card shards; the older three-card figure is
only an arithmetic capacity floor. CPU preprocessing is a reference path. GPU
counts in these projections describe capacity, not a performance ceiling. In
particular, preprocessing is not limited to two GPUs: compare more cards and
multiple nodes when they improve complete latency, within reviewed budgets.


## Reading the tables

All rows use m=2, t=5 and one field-element database. Every cell names its run
status (HPC measured, laptop measured or not run), runtime units and memory
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

## Machine capacity and feasibility

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
validation pending (server preprocessing has validated peer kernels). Time cannot be estimated
by dividing a single-GPU timing by the card count.

- s≤16: measured paths below fit this machine; unrun variants are capacity
  projections, with runtime unknown. Existing job budgets still apply.
- s=17: CPU array models fit one large-memory node. Resident GPU modes project
  two to four GPUs; host-backed encoding projects one GPU with about 349 GiB
  host arrays. These are possible routes after the port, not validated runs.
- s=18: disk-backed CPU encoding projects about 317 GiB private arrays and 1 TiB
  work files. Streamed resident GPU encoding projects eight GPUs across at least
  two four-GPU nodes. Preprocessing and communication remain separate gates.
- s=19–20: the modeled owning CPU outputs alone exceed one node's RAM; GPU
  capacities appear below. New layouts, distributed execution or further memory
  savings may change these projections. They remain exploratory, not ruled out.

## Aim for s17

The [approved execution sequence](encoding-scaling.md#current-execution-plan-for-s17)
now starts with the [wide scalar reference](../experiments/s17-arithmetic/report.md);
job74290 passed independent scalar and legacy-ordering checks, with all recorded
source/binary/evidence hashes reconciled. The [native64-bit kernel prototype](../experiments/s17-native/report.md)
passes120 optimized and120 sanitizer cases with shared CPU regressions and
source/evidence reconciliation (job74295). Separate [wide Go/C plans and parallel
matrix execution](../experiments/s17-matrix/report.md) pass corrected validation
job74330: bounded independent oracles, parallel Taylor, lifetime race,112 optimized
and112 sanitized direct C matrix cases, and CPU/legacy regressions. All421 source,
31 evidence and10 cache identities reconcile. A [width-aware code-data prototype](scaling-code-data.md)
now passes bounded legacy byte/slot compatibility, wide-header/value-codec,
cap/digest/range/error and shared Go/CPU regressions (job74396). Its426 source,13
evidence and2 binary identities reconcile. An explicit experimental
[complete wide CPU schedule](../experiments/s17-encoding/report.md) now connects
resident code data to five native transforms and canonical output; bounded
validation passes job74412 with438 source,24 evidence and18 cache identities
reconciled. The [CPU implementation refactor](cpu-encoding.md) now shares that native
sequence and supporting operations; bounded cross-mode/race/sanitizer validation
passes job74448. Small s10 encoding medians are1–3% slower; setup costs increase
27–44% (about2–3 ms), an unresolved regression requiring follow-up. Fulls17 preprocessing now passes75832 and its genuine cache is retained;
full-size CPU/GPU encoding latency remains unknown. Its reuse model still retains 8N
output backing bytes (4N logical uint32 bytes), and older returned words require
additional live RAM. No full-size capacity or timing estimate changes.
No full-encoding table cell changes from these kernel steps.

The [shared preprocessing implementation](preprocessing.md) now extends the
sampling/root/reciprocal/identity procedure through s17 with explicit native
execution and support-bound restart artifacts. Targeted bounded checks passed;
full bounded native/sanitizer validation passes corrected job74447;
the full GPU run75832 now provides genuine support/cache and measured host
RSS/runtime (see the preprocessing row/report).
Legacy CUDA execution remains s≤16; the resident CUDA17 peer prototype is
implemented in source; one-GPU validation passes job74453, followed by
the corrected peer-lifetime sanitizer retry74466, which passes. Full GPU preprocessing75832 has completed under its separate admitted budget.
The native fast product payload is384 GiB at s17, with Go arrays and overhead
additional; a host peak around600 GiB or higher is an unmeasured scenario, not
a capacity guarantee. No full-encoding measurement changes.

Both CPU and GPU encoding are **potentially feasible on this system after the
s17 port**, with preprocessing admitted separately. First target one large-RAM
CPU node for a correctness and performance baseline, then a resident multi-GPU
encoder on one node. These are recommendations, not demonstrated support or
approved larger allocations.

| Route | Projected encoding payload | Intended role |
|---|---|---|
| CPU resident reuse | ~427 GiB host | First full CPU target; avoid external-memory work when sufficient RAM is admitted |
| CPU file-streamed | ~270 GiB host, reusable code file ~142 GiB | Lower-RAM alternative; repeated reads/cache costs must be compared with resident reuse |
| GPU fast-layout streamed | ~256 GiB aggregate device, ~221 GiB host | Two H200s are a capacity floor; first assess per-device headroom, partitioning and communication |
| GPU resident reuse | ~398 GiB aggregate device, ~221 GiB host | Three H200s are a capacity floor; compare fewer repeated code transfers against the extra GPU |
| GPU host-backed | Few MiB device tiles, ~349 GiB host | One-GPU fallback; repeated transfers may make it substantially slower |

The CPU node advertises about907GiB schedulable RAM; device capacity is about
140.4GiB per H200. Models exclude runtime/cache/context and allocation headroom.
Do not force a slower memory mode when the faster validated path fits the admitted
budget. Two/three GPU counts assume sharding; the resident multi-GPU client is implemented in source with runtime
validation pending. Test communication and full overlapping lifetimes before selecting a
card count or claiming latency. s17 output is34GiB packed, versus64GiB for a
canonical uint32 array; packed output does not eliminate transform workspace.

The prerequisite work is wider34-bit internal arithmetic with boundary tests,
compatible coordinates/counts/cache formats and independently verified s17 code
preprocessing. Base-field values still fit32-bit slots. GPU preprocessing now has a measured348.393GiB sampled host peak and full cache
from75832. Its product arrays model384GiB across four implemented shards;
CPU-only preprocessing and client encoding still need separate full measurements.
Use small correctness tests and existing s16 regressions before full s17 runs.
Runtime estimates remain unknown until wider arithmetic and CPU/GPU kernels are
measured; capacity alone cannot predict useful performance.

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
| 10 | HPC measured | Whole-command peak RSS: ≤0.24 GiB | CPU: 6.65 s; GPU: 1.53 s | Measured on this machine |
| 14 | Laptop measured | Peak host process RSS: 13.73 GiB | 33 min 38 s | Measured on laptop |
| 15 | Laptop measured | Peak host process RSS: 20.55 GiB | 3 h 16 min 7 s† | Measured on laptop |
| 16 | HPC measured | Peak host process RSS: 72.575 GiB | GPU: 2 h 20 min 15.469 s | Measured on this machine |
| 17 | **HPC measured; passed75832**<br>K:3876566203; support from75819 | **Peak sampled host RSS:348.393GiB**<br>Device peak:not sampled; product arrays384GiB aggregate (model)<br>Final device owned:256.009GiB observed | **Construction/identity/spectrum:9h16min46s**, includes root checkpoints<br>Verified resumed worker:9h58min38s<br>Earlier discovery:21min38s | **Demonstrated on this machine:4H200, one node**<br>Canonical preprocessing:235.324GiB<br>[Full evidence/boundaries](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088) |
| 18 | Not run | Peak RAM: unknown | unknown | Implementation/admission pending |
| 19 | Not run | Peak RAM: unknown | unknown | Implementation/admission pending |
| 20 | Not run | Peak RAM: unknown | unknown | Implementation/admission pending |

### CPU encoding: resident modes

| s | Original CPU | Reuse CPU | Compact CPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.308 s–0.31 s<br>Array payload: ~17.7 MiB | HPC measured<br>Runtime: 0.313 s–0.319 s<br>Array payload: ~13.7 MiB | HPC measured§<br>Runtime: 0.317 s–0.319 s<br>Peak workflow process RSS: 30.387 MiB |
| 14 | HPC measured<br>Runtime: 1 min 31.164 s<br>Peak process RSS: 5.594 GiB | HPC measured<br>Runtime: 58.667 s–58.853 s<br>Peak process RSS: 3.595 GiB | HPC measured<br>Runtime: 1 min 4.517 s–1 min 8.413 s<br>Peak process RSS: 2.526 GiB |
| 15 | HPC measured<br>Runtime: 10 min 55.937 s<br>Peak process RSS: 18.465 GiB | HPC measured§<br>Runtime: 4 min 34.704 s–4 min 41.715 s<br>Peak workflow process RSS: 14.460 GiB | HPC measured§<br>Runtime: 4 min 39.900 s–4 min 40.038 s<br>Peak workflow process RSS: 9.901 GiB |
| 16 | HPC measured<br>Runtime: 20 min 52.503 s<br>Peak process RSS: 74.370 GiB | HPC measured<br>Runtime: 20 min 47.173 s–21 min 21.328 s<br>Peak process RSS: 58.351 GiB | HPC measured<br>Runtime: 21 min 29.759 s<br>Peak process RSS: 39.409 GiB |
| 17 | Not run<br>Runtime: unknown<br>Estimated array payload: ~491 GiB<br>Capacity: one-node RAM projected; full run unvalidated | Not run<br>Runtime: unknown<br>Estimated array payload: ~427 GiB<br>Capacity: one-node RAM projected; full run unvalidated | Not run<br>Runtime: unknown<br>Estimated array payload: ~287 GiB<br>Capacity: one-node RAM projected; full run unvalidated |
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
| 17 | Not run<br>Runtime: unknown<br>Estimated array payload: ~270 GiB<br>Capacity: one-node RAM projected; full run unvalidated | Not run<br>Runtime: unknown<br>Estimated array payload: ~270 GiB<br>Capacity: one-node RAM projected; full run unvalidated | Not run<br>Runtime: unknown<br>Estimated private arrays: ~78.4 GiB<br>Work files: 256 GiB<br>Capacity: one-node RAM projected; full run unvalidated |
| 18 | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~317 GiB<br>Work files: 1024 GiB<br>Capacity: one-node RAM projected; disk capacity unverified |
| 19 | Not run<br>Runtime: unknown<br>Output alone: 1024 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~1279 GiB<br>Work files: 4096 GiB<br>Capacity: exceeds one node's RAM |
| 20 | Not run<br>Runtime: unknown<br>Output alone: 4096 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Transform arrays: ≥16384 GiB<br>Capacity: exceeds one node's RAM | Not run<br>Runtime: unknown<br>Estimated private arrays: ~5162 GiB<br>Work files: 16384 GiB<br>Capacity: exceeds one node's RAM |

* **Historical low-confidence forecast, superseded by the measured GPU row:** the
pre-trial scenario allowed
roughly12–24h after discovery, or roughly24–48h for fresh discovery plus
preparation/verification/persistence. Baselines are measured s16 GPU-assisted
preparation8415.469s (2h20m15s, excludes discovery) and CPU discovery10597.545s
(2h56m38s). N grows4x; an illustrative N·log2(N)^2 factor is
4(34/32)^2≈4.52, giving10.56h preparation and13.29h discovery before extra
width/communication effects. The12–24h construction range allows approximately
1.1–2.3x over that model for wider values and integration costs; this allowance
is a judgment, not a measured bound. The end-to-end range rounds their sum and
allows persistence/checks, not queue time. Actual runtime may exceed it.

This historical GPU scenario is not a runtime estimate for the new native
executor. Native weight-class enumeration replaces per-monomial carry evaluation,
so the old discovery extrapolation is no longer a calibrated model for that
implementation. Its GPU discovery/preparation is now measured by75819/75832;
CPU-only full preparation remains unmeasured.

This assumes a working wider server backend with adequate resident memory and
useful parallelism. The three-array GPU payload is~384GiB: three H200s are a
capacity floor, with four on one node a layout/headroom option, not implemented
support at the time of that forecast. Exact K and the four-card GPU workflow
are now measured; cross-node communication remains unmeasured. No linear
speedup from GPU count is assumed. Faster GPU-assisted discovery/native host
work could lower latency; streaming or poor inter-GPU scaling could raise it.
These scenarios are not a promise that the present code can run s17. Use bounded
wider preprocessing calibration before assigning full-run limits.

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
filesystem/cache state; its full encoder is now measured at s10 and s16 (the s16 retained-cache reread passes job74447). File and disk
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

| s | Original GPU | Reuse GPU | Compact GPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.104 s<br>GPU array payload: 16.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.103 s<br>GPU array payload: 12.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.108 s–0.109 s<br>GPU array payload: 10.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 |
| 14 | HPC measured§<br>Runtime: 5.155 s–5.257 s<br>Owned GPU allocations: 4.190 GiB<br>Peak workflow host RSS: 2.521 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 5.176 s–5.179 s<br>Owned GPU allocations: 3.190 GiB<br>Estimated host arrays: ~2.38 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 5.800 s–5.860 s<br>Owned GPU allocations: 2.192 GiB<br>Peak workflow host RSS: 2.525 GiB<br>GPUs: 1 |
| 15 | HPC measured§<br>Runtime: 16.853 s–17.497 s<br>Owned GPU allocations: 16.808 GiB<br>Peak workflow host RSS: 9.768 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 16.917 s–17.756 s<br>Owned GPU allocations: 12.809 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 20.487 s–20.873 s<br>Owned GPU allocations: 8.811 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 |
| 16 | HPC measured<br>Runtime: 1 min 0.216 s–1 min 25.954 s<br>Owned GPU allocations: 67.423 GiB<br>Peak host process RSS: 39.024 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 56.975 s–57.418 s<br>Owned GPU allocations: 51.427 GiB<br>Peak host process RSS: 39.031 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 33.332 s–1 min 41.526 s<br>Owned GPU allocations: 35.429 GiB<br>Peak host process RSS: 39.028 GiB<br>GPUs: 1 |
| 17 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~526.42 GiB<br>Estimated host arrays: 220.84 GiB<br>GPUs: 4<br>Capacity: one node projected; implementation pending | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~398.42 GiB<br>Estimated host arrays: 220.84 GiB<br>GPUs: 3<br>Capacity: one node projected; implementation pending | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~270.42 GiB<br>Estimated host arrays: 220.84 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending |
| 18 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~2109 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 16<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1597 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 12<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1085 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 8<br>Capacity: multiple nodes required |
| 19 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~8447 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 61<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~6399 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 46<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4351 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 31<br>Capacity: multiple nodes required |
| 20 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~33834 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 241<br>Capacity: exceeds this machine's GPUs or host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~33834 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 241*<br>Capacity: exceeds this machine's GPUs or host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~25642 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 183*<br>Capacity: exceeds this machine's GPUs or host RAM |

### GPU encoding: streamed and host-backed modes

| s | Stream GPU | Packed-stream GPU (stored output) | Host-backed GPU |
|---|---|---|---|
| 10 | HPC measured§<br>Runtime: 0.098 s–0.106 s<br>Owned GPU allocations: 8.035 MiB<br>Peak workflow host RSS: 20.645 MiB<br>GPUs: 1 | HPC measured§<br>Runtime: 0.107 s–0.114 s<br>Owned GPU allocations: 8.035 MiB<br>Peak workflow host RSS: 28.113 MiB<br>GPUs: 1<br>Peak tracked file residency: 10.578 MiB (additional) | HPC measured<br>Runtime: 1.606 s–1.617 s<br>GPU array payload: 0.534 MiB<br>Host array payload: 13.4 MiB<br>GPUs: 1 |
| 14 | HPC measured<br>Runtime: 5.467 s–5.527 s<br>Owned GPU allocations: 2.000 GiB<br>Estimated host arrays: ~2.38 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 5.804 s–5.827 s<br>Owned GPU allocations: 2.000 GiB<br>Estimated host arrays: ~1.39 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 1 min 46.837 s–1 min 48.589 s<br>Owned GPU allocations: 0.710 MiB<br>Peak workflow host RSS: 3.521 GiB<br>GPUs: 1 |
| 15 | HPC measured§<br>Runtime: 18.537 s–18.564 s<br>Owned GPU allocations: 8.001 GiB<br>Peak workflow host RSS: 9.768 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 20.451 s–21.448 s<br>Owned GPU allocations: 8.001 GiB<br>Peak workflow host RSS: 7.646 GiB<br>GPUs: 1<br>Peak tracked file residency: 6.689 GiB (additional) | HPC measured§<br>Runtime: 8 min 45.294 s–9 min 7.208 s<br>Owned GPU allocations: 0.898 MiB<br>Peak workflow host RSS: 13.777 GiB<br>GPUs: 1 |
| 16 | HPC measured<br>Runtime: 58.851 s–1 min 3.487 s<br>Owned GPU allocations: 32.005 GiB<br>Peak host process RSS: 39.029 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 5.485 s–1 min 6.227 s<br>Owned GPU allocations: 32.005 GiB<br>Peak host process RSS: 23.014 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 1 h 1 min 17.511 s<br>Owned GPU allocations: 1.273 MiB<br>Peak workflow host RSS: 55.049 GiB<br>GPUs: 1 |
| 17 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~256 GiB<br>Estimated host arrays: 220.84 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~256 GiB<br>Estimated host arrays: 156.84 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~348.84 GiB<br>GPUs: 1<br>Capacity: one node projected; implementation pending |
| 18 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1024 GiB<br>Estimated host arrays: 889 GiB<br>GPUs: 8<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~1024 GiB<br>Estimated host arrays: 633 GiB<br>GPUs: 8<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~1401 GiB<br>GPUs: 1 (host exceeds node RAM) |
| 19 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4096 GiB<br>Estimated host arrays: 3581 GiB<br>GPUs: 30<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~4096 GiB<br>Estimated host arrays: 2557 GiB<br>GPUs: 30<br>Capacity: multiple nodes required | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~5629 GiB<br>GPUs: 1 (host exceeds node RAM) |
| 20 | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~24576 GiB<br>Estimated host arrays: 14420 GiB<br>GPUs: 176*<br>Capacity: exceeds this machine's GPUs or host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: ~24576 GiB<br>Estimated host arrays: 10324 GiB<br>GPUs: 176*<br>Capacity: exceeds this machine's GPUs or host RAM | Not run<br>Runtime: unknown<br>Estimated GPU arrays: few MiB<br>Estimated host arrays: ~30804 GiB<br>GPUs: 1 device, host exceeds node<br>Capacity: exceeds this machine's GPUs or host RAM |

Host estimates in multi-GPU rows are **aggregate**, not per node; distributing
arrays or avoiding replication requires an explicit host as well as device design.
At s18 the ~889 GiB host model leaves little room on one node; distribute it or
reduce it rather than assuming the payload fits with runtime headroom.

* At s20, the estimated K makes the prefix workspace exceed the current N-sized
arena. These entries allow an expanded arena; the current modes reject that
shape. A redesign could change the counts. Host-backed s20 likewise needs a
larger arena, included in its host estimate. All s≥17 rows assume a future port.

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
| 18 | ~1536 GiB | 11 | At least 3 four-GPU nodes for this layout; host distribution and peak unknown |
| 19 | ~6144 GiB | 44 | At least 11 four-GPU nodes; host distribution and peak unknown |
| 20 | ~24576 GiB | 176 | Exceeds configured68-GPU inventory; needs a different layout or more hardware |

## Projection assumptions and upkeep

For q=2^s, N=q², use exact measured K=918672421 at s16; catalogue estimates
K=3870933765, 16298885492, 68331898887, 286146801617 at s17–20. Base values and
canonical outputs use four-byte slots. Internal extension values use four bytes
through s16 and a hypothetical eight-byte representation thereafter. Let
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

Next work remains mode combinations and overlapping buffer lifetimes, followed
by wider arithmetic and preprocessing admission for s17, then progressive
validation toward s18. Present memory–time/I/O trade-offs before adopting them;
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

## Bounded preprocessing components (HPC, not full table rows)

[GPU review](../experiments/gpu-s17-preprocessing/report.md#completed-corrected-peer-retry-scaling-073)
measures s17-width dense transforms at only2^22 values (32 MiB), on gnode02 H200
NVL cards with NV6 peer connectivity. Warm/profiled/logged whole-call timings,
one sample, include transfers. Full s17 preparation remains **Not run**.

| GPUs | Forward runtime | Inverse runtime | Owned GPU memory (aggregate) | Validation |
|---|---|---|---|---|
| 1 | 184.709 ms | 158.943 ms | 65.57 MiB | Complete bounded job74453 passes; comparison measured within74466 |
| 2 | 216.765 ms | 190.268 ms | 67.14 MiB | Corrected bounded job74466 passes, zero memcheck errors/leaks |

Two GPUs are17.4%/19.7% slower on this small component. Extra cards are not ruled
out for larger workloads. Driver/context memory is additional; sampled command
host RSS below1 GiB includes builds and is not isolated preparation memory.

§ [Completed22-case HPC batch](../experiments/encoding-table/report.md#completed-all-case-review-scaling-073)
uses32 allocated CPU cores and retained preprocessing, with original native
worker settings captured in the run logs. The jobs are profiled, at different
revisions/nodes from older baselines; these rows are not controlled speedups.
One encoding sample at s16, two at smaller sizes. New memory entries are sampled
**whole-workflow** process RSS (including verification/readback), not isolated
encoding peaks. File residency is additional and can overlap RSS; it is not an
exact physical-RAM sum. All22 cases complete on this HPC system under their
recorded budgets. Fulls17 remains unrun.

The [local-vector comparison](../experiments/gpu-s17-preprocessing/report.md#completed-local-vector-validation-scaling-075)
passes bounded GPU validation74468, including zero memcheck errors/leaks. Same
s17-width32 MiB dense tile, one warm/profiled/logged whole-call sample per case:

| Layout | GPUs | Forward runtime | Inverse runtime | Owned GPU memory (aggregate) |
|---|---|---|---|---|
| Peer | 1 | 197.515 ms | 170.759 ms | 65.57 MiB |
| Peer | 2 | 233.538 ms | 197.709 ms | 67.14 MiB |
| Local | 1 | 194.450 ms | 183.013 ms | 65.57 MiB |
| Local | 2 | 189.407 ms | 178.511 ms | 67.14 MiB |

Local reduces two-card time18.9%/9.7% versus peer, but two local cards gain only
2.6%/2.5% over one local card. Single-card local inverse is7.2% slower. These
fixed-order component samples are not fulls17 forecasts or a settled layout
choice. Default peer and all capacity cells stay unchanged. Same2N transform/3N
product payload; extra8 transpose passes add16N word reads/writes (~2 TiB at
s17), with full-size runtime/link traffic unknown. A256 MiB repeated1/2/4GPU
comparison now has explicit4GPU/4 GiB resource approval; submission and results
are recorded in the experiment report. No capacity or full-size timing changes.

## Repeated GPU preprocessing calibration (HPC components)

[Job74635](../experiments/gpu-s17-preprocessing/report.md#completed-four-gpu-calibration-scaling-077)
passes complete bounded validation on four H200s (gnode02, NV6 peers), with zero
CUDA errors/leaks, full tile hash/inverse agreement and reconciled evidence.
256 MiB tiles (2^25 s17-width values),3 fixed-order warm/profiled/logged samples
per case, whole-call wall time including transfers. Not fulls17 preparation.

| Layout | GPUs | Forward median (range) | Inverse median (range) | Owned GPU memory, aggregate |
|---|---|---|---|---|
| peer | 1 | 0.865674 s (0.864669–0.865862 s) | 0.754588 s (0.754539–0.755842 s) | 513.57 MiB |
| peer | 2 | 0.995857 s (0.995585–0.996306 s) | 0.884177 s (0.883658–0.884894 s) | 515.14 MiB |
| peer | 4 | 1.088009 s (1.087734–1.089616 s) | 0.978584 s (0.976949–0.978659 s) | 518.28 MiB |
| local | 1 | 0.855847 s (0.855641–0.855908 s) | 0.830580 s (0.830201–0.831150 s) | 513.57 MiB |
| local | 2 | 0.915740 s (0.915194–0.915805 s) | 0.890322 s (0.889293–0.890732 s) | 515.14 MiB |
| local | 4 | 0.778385 s (0.778306–0.778699 s) | 0.753873 s (0.751693–0.754376 s) | 518.28 MiB |

Four-card local is28.5%/23.0% faster forward/inverse than four-card peer, but versus
one-card peer gains10.1% forward and essentially nothing inverse. Sum of reported
medians improves5.4%, not end-to-end preprocessing. Sampled whole GPU-test process-
group RSS peaks1.615 GiB; device context/driver overhead is additional. Current
small vector grids may limit scaling (source geometry, not measured occupancy).
Full-domain latency and peak host memory remain unknown. Default/capacity rows
are unchanged; finish persistent full-workflow admission before any fulls17 job.

## Persistent preprocessing readiness

Historical admission checkpoint, superseded by [completed full preprocessing](#full-s17-preprocessing-verified-scaling-088).

The [server CLI and complete-path validation](../experiments/gpu-preprocessing-workflow/report.md)
now implement atomic support/root/code cache stages and verified restarts, retaining
profiles/provenance/logs. Local native/Go-oracle and failure-path checks pass;
complete GPU executable validation is pending. No s17 full-run table cell changes.
Projected retained s17 public preprocessing files total~235 GiB, with additional
profiles, failed stages and storage headroom; fresh writes/readback add unmeasured
I/O latency. Full host peak and budget admission remain outstanding.

| s17 preprocessing route | Modeled product arrays per GPU | Implementation status | Full latency |
|---|---|---|---|
| 4 GPUs, one node | 96 GiB | Same-node kernels boundedly validated; persistent workflow validation pending | Unknown |
| 8 GPUs, at least two nodes here | 48 GiB | User-authorized option; cross-node execution not implemented | Unknown |
| 16 GPUs, at least four nodes here | 24 GiB | User-authorized option; cross-node execution not implemented | Unknown |

These divide the384 GiB array model evenly, excluding runtime/staging/headroom,
communication and possible host replication. They are not speedup estimates or
full-resource admission; CPU/GPU client encoding support stays unchanged.

Persistent workflow75194 passes the complete one-GPU s8 path and CPU/race/ASan/vet,
but four-GPU discovery stalls on gnode03 with a later GPU-reset report. It is
failed/incomplete, distinct from prior successful four-card kernels on gnode02.
No fulls17 cell changes. [Forensic report and unchanged-budget retry](../experiments/gpu-preprocessing-workflow/report.md#job75194-failure-and-bounded-retry-scaling-079)
retain unknown fault cause, partial caches and bounded-shutdown/preflight changes.

Persistent complete-path validation75342 passes ongnode04 in3min42s including
builds/tests. Peer/local1/4GPU fresh/root/completed-cache cases, CUDA-free
readback, independent Go bytes, peer preflight, race/ASan/vet and zero CUDA errors/
leaks pass;509/37/73 source/evidence/cache identities reconcile. This completes
the small persistent-runner gate, not any fulls17 measurement or capacity claim.
Full support/cardinality, concurrent RAM/device/storage/time budget and guarded
launch admission remain next. [Report](../experiments/gpu-preprocessing-workflow/report.md#completed-persistent-workflow-validation-scaling-080).

[Exact count-only admission](../experiments/s17-preprocessing-admission/report.md)
is implemented with local references passing; full s17 GPU count remains pending.
No full-preparation cell changes. Format-derived successful retained files are
12K+12N+61 bytes at s17 (≤288 GiB+61 forK≤N/2), before logs/profiles/partial stages
and headroom. New count-only mode avoids the large host list/grid; its~2.5 GiB
modeled GPU enumeration payload is not a full preprocessing memory reduction.
Final RAM/storage/runtime admission and cross-node support remain separate.

[Current configuration/free-space snapshot](../experiments/s17-preprocessing-admission/resource-snapshot.json) supports the admission draft; configuration, filesystem free space and blank QoS fields are not quota or resource approval.

## Exact s17 count and full admission readiness

Historical admission checkpoint, superseded by [completed full preprocessing](#full-s17-preprocessing-verified-scaling-088).

[Count75810](../experiments/s17-preprocessing-admission/report.md#completed-count-admission-scaling-082)
passes all checks, with exact K3876566203,N17179869184 and no full host grid/list.
Count-only executable6.276s includes setup/profiling/metadata; it is not discovery
or preparation latency.513source/41evidence/85cache identities reconcile, zero
CUDA errors/leaks. No s17 full-preparation/encoding row becomes measured.

Successful public preprocessing files calculate to235.324GiB: support28.883,
root14.441, canonical P/spectrum142.441 and G49.559GiB. Discovery principal host
model189.765GiB (two index lists/grid/two bitmaps); final retained host220.883GiB,
with product/Newton/identity/GC/runtime/driver/cache overhead additional and peak
unknown. Existing~220.84GiB/~235GiB estimates are superseded by these exact-K
calculations, not measured peaks.

The [full controller](../experiments/gpu-s17-full-preprocessing/report.md) uses
one exclusive128-core node/32workers,864GiB host request,4H200112GiB owned/card,
512GiB retained and16GiB scratch. Eight/sixteen GPU options require cross-node
support; no linear speedup is assumed.

The first full s17 local-layout trial is now approved under the concrete
[experiment budget](../experiments/gpu-s17-full-preprocessing/budget.json):
4H200112GiBowned/card, exclusive128cores/32workers,864GiBhost request,
512GiBretained/16GiBscratch,48h scheduler cutoff. Submission is recorded in its
report; table rows stay unmeasured until completion and evidence review.

Full s17 public preprocessing is now submitted as job75819 (local layout), with
[recorded approved limits](../experiments/gpu-s17-full-preprocessing/submission.json).
Status is **submitted, not yet verified**; all full s17 runtime/peak-memory entries
remain unmeasured until completion and reconciliation. No database encoding job.

The approved replacement extends the scheduler cutoff to100h (97h worker/98h
controller), with optional2h/adaptive recovery writes targeting about1% overhead.
Two snapshots per discovery/root/reciprocal phase project about132GiB additional
retained payload; compute/memory/storage limits remain unchanged. No FFT workspace
is saved, and an interrupted transform is repeated. These are recovery/capacity
changes, **not measured s17 preprocessing performance**. See the
[recovery contract](preprocessing.md#incremental-recovery) and full-run report.

Replacement **75832 submitted** with100h scheduler cutoff and incremental recovery;
original75819 received a cancellation request only after the replacement was
accepted. Existing completed support is reusable. Results/logs and exact budget:
[submission receipt](../experiments/gpu-s17-full-preprocessing/submission-resume-02.json).
No new-job monitoring or full result review; s17 remains unmeasured.

## Matched CPU refactor comparison at s14

[Job76055 passes](../experiments/cpu-encoding-regression/report.md#completed-comparison-scaling-086)
on the HPC hostgnode02,32 CPU cores, no GPU: four balanced pairs, identical
verified code/database, full canonical output hashes and separate profiles.
Both versions log progress, with the old671bf5e baseline receiving an explicit
logging-only patch; candidate revisione07f308. These matched measurements are
separate from older table entries with different revisions/methodology.

| Operation | Pre-refactor median | Refactored median | Difference |
|---|---|---|---|
| Resident encoding | Runtime60.346 seconds | Runtime61.119 seconds | +1.28% |
| Resident cache loading | Runtime1.894 seconds | Runtime2.540 seconds | +34.1%;0.646 seconds |
| File-backed encoder opening | Runtime1.501 seconds | Runtime2.042 seconds | +36.1%;0.541 seconds |

Encoding-process peak sampled RSS is3.593–3.595GiB in both implementations,
with filesystem cache additional. Workspace allocation varies2.6–282ms and the
first old code load is an I/O outlier; all samples/ranges remain in the report.
The workspace median ratio is unstable, not evidence of a fixed38x penalty.
The shared decoder is the setup profiling hotspot. This small larger-size
encoding difference does not block GPU client integration; setup optimization
remains a follow-up. No s17 performance or memory-capacity update follows.

## Resident GPU client integration (SCALING-087)

[Source/API](gpu-encoding.md) now implements resident code-data encoding through
s17 with explicit1/2/4/8 same-node GPUs, common CPU/GPU cyclic order and shared
width-selected GPU transforms. CPU defaults remain CUDA-free; legacy s16 GPU
modes are unchanged. [Bounded validation](../experiments/gpu-code-encoding/report.md)
is pending. Genuine s17 preprocessing/encoding measurements remain unknown;
new code availability does not change any measured table cell.

For exacts17K, principal resident code/work/parity payload projects398.457GiB
aggregate and104.004GiB maximum on one of four GPUs, before metadata/context/
headroom. Keeping source/message/canonical output projects220.883GiB host arrays;
output is64GiB uint32. This refines the earlier~398GiB resident model, not measured
peaks or full admission. The new path provides canonical words; s17 direct packed,
streamed-code, compact and host-backed modes remain unimplemented. Compare full
setup/transfers/peer routing/output latency after correctness and budget gates.

The bounded resident-client integration suite is submitted as **job76426**,
sourcebf2bffc, within the prior4GPU/4GiB test budget. [Receipt](../experiments/gpu-code-encoding/submission.json)
links results/logs. No monitoring or fulls17 encoding submission; runtime status
remains pending until the completion ping and evidence reconciliation.

## Full s17 preprocessing verified (SCALING-088)

**75832 passes**:4H200,348.393GiB peak sampled host RSS; worker9h58min38s,
including cache I/O/readback and reused support. Earlier discovery21min38s,
2samples; complete P·G identity and canonical-file readback pass. Final reusable
payload235.324GiB; clients need the142.441GiB `code/online.bin`. Four periodic
root saves add78.525s (0.219% worker wall), with23.426GiB latest snapshots retained.
Full device peak was not sampled;384GiB is the product-array model, while the
final post-release observed owned total is256.009GiB. Do not conflate these.

[Report/summary](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088)
reconciles523source/17evidence/14cache identities. This is a resumed run, not an
uninterrupted fresh baseline; profiled/logged root construction accounts for
about86% of worker wall. Raw3.162GiB progress remains locally retained outside Git,
with sealed digest/selected events; streaming review passes without a redundant
large-payload rehash. Encoding rows stay unmeasured at s17. The preprocessing
prerequisite is complete; full CPU baseline/GPU validation, output and resource
admission remain separate. GPU validation76426 was not inspected here.
