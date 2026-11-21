# Encoding scaling reference

The performance goal is **reasonable encoding performance through s=18 inclusive**
on the HPC system, retaining CPU and optional GPU paths. **s=19 and s=20 remain
exploratory targets**: keep their rows as algorithms, layouts and measurements
improve. This is the maintained capacity and results reference for the
[scaling plan](encoding-scaling.md), separate from TRL 5 acceptance.

Updated 2026-11-21; full s17 encoding source `ed0c026`.
**Resident CPU encoding76645 and four-H200 GPU encoding76706 both passed at s17**,
using the same verified preprocessing and deterministic database. Complete
canonical and tightly packed output hashes match. GPU retry76706 supersedes the
failed first GPU attempt76646. Other s17 modes and s≥18 full encoding remain
unvalidated; projections are not support. “Reasonable” has no agreed numerical
latency threshold yet; select it using measured setup, encoding, memory, I/O and
deployment costs before accepting a scaling milestone.

## Current distributed preprocessing step

We are on **step1: distributed large products and FFTs**. These are bounded
component results, separate from the full s17 runs and fulls18 projections below.

| Step | Status |
|---|---|
| 1. Distributed products and FFTs | **Bounded resident products passed78399**:24 normal +24 sanitized rank/product outputs; host reference also passed.<br>Resident storage78392, vectors78394 and complete FFT78396 passed.<br>**Matched comparison78432 passed** (96 normal +8 small sanitized outputs).<br>**GPU-local Taylor passed78460**, now faster in matched small products.<br>**Larger structured products passed78469** through262144 values; resident3.16× faster at largest size.<br>**Local GPU roots + distributed final merge passed78508**, through65532 roots.<br>Bounded operand ownership routing comparison prepared; fullsize root/Newton scaling remains pending. |
| 2. Distributed discovery | Existing discovery is not yet partitioned across nodes. |
| 3. Generator construction and verification | Existing algorithms still need the distributed executor/data integration. |
| 4. Coordinated checkpoints and publication | Helpers exist; distributed ownership/restart/assembly integration remains. |
| 5. Complete scaling validation | Assembled small pipeline then larger runs remain; fulls18 preprocessing unrun. |

[FFT evidence](../experiments/distributed-gpu-fft/report.md): controller runtime
**2 min15.343s**, normal suite **1.971s**, sanitizer suite **3.726s**. Logged host
peak RSS **372.031 MiB normal /411.016 MiB sanitized per process**; maximum owned
GPU memory **1.571 MiB/card**, excluding context/driver. Largest4096-point single
profiled sample: forward **43.916ms**, inverse **40.807ms**. No fullsize runtime
inferred. [Product evidence and storage costs](../experiments/distributed-gpu-product/report.md).

[Product evidence](../experiments/distributed-gpu-product/report.md): controller
runtime **2 min21.151s**, normal suite **2.172s**, sanitizer suite **7.084s**.
Logged host peak RSS **372.293 MiB normal /415.930 MiB sanitized per process**;
maximum owned GPU memory **1.572 MiB/card**. Largest4096-point normal product
samples **111.395–111.792ms**, profiled. This validates bounded arithmetic and
communication; fullsize feasibility/runtime remain unknown.
[Next resident-storage design and trade-offs](distributed-preprocessing-storage.md).
[Resident-storage gate](../experiments/distributed-gpu-storage/report.md): **78392
passed208 normal +208 sanitized rank/pass outputs**,544 proof endpoints/mode,
zero CUDA errors/leaks; **2min3.480s** controller, **4.924s normal /4.024s sanitized**.
Host peak RSS **376.137/413.750 MiB per process**; owned GPU **1.583 MiB/card**.
HPC gnode14/gnode15,2 nodes /2 H200 /4 CPU cores /8 GiB requested RAM. These are
small storage tests, not fullsize throughput. [Resident vector composition](../experiments/distributed-resident-vectors/report.md)
**passed78394**:24 normal +24 sanitized rows, zero CUDA errors/leaks; controller
**2min2.265s**, normal **2.026s**, sanitizer **2.772s**. Host RSS **372.121/416.020
MiB per process**, owned GPU **1.617 MiB/card**. [Complete resident FFT](../experiments/distributed-resident-fft/report.md)
**passed78396**:24 normal +24 sanitized rows,240 transpose +1788 Taylor proof
endpoints/arm, zero CUDA errors/leaks. Controller **2min21.091s**, normal **2.371s**,
sanitizer **3.691s**; host RSS **373.934/411.531 MiB per process**, owned GPU
**1.618 MiB/card**. [Resident products](../experiments/distributed-resident-product/report.md)
**passed78399**:24 normal +24 sanitized products,720 transpose +5364 Taylor
endpoints/arm, clean CUDA summaries. Controller **2min33.300s**, normal arm
**3.121s**, sanitizer **6.476s**; host RSS **374.566/414.418 MiB per process**,
owned GPU **1.618 MiB/card**. Largest4096-value resident samples **0.2347–0.2367s**
versus host **0.1080–0.1111s**, with diagnostic boundaries described in the report.
[Matched size/tile comparison](../experiments/resident-product-scaling/report.md)
**78432 passed**:96 normal outputs +8 small sanitized outputs, zero CUDA errors/
leaks; controller **3min50.175s**. Both backends use GPU arithmetic. At16384 values/
4096-word tiles,3-sample median net time **355.987–365.759 ms host-stored** versus
**616.580–619.547 ms resident** (two product shapes). Normal peak RSS
**377.359 MiB host /375.188 MiB resident per process**; maximum owned GPU
**1.575/1.791 MiB per card**, respectively. Larger tiles improve resident only
**3.6–4.1%** at this size. [GPU-local Taylor access](../experiments/resident-taylor-local/report.md)
**passed78460**:legacy products, local full-extension FFT/products normal and
sanitized, raw-write guards and all matched samples; zero CUDA errors/leaks.
Controller **3min20.380s**. N16384 median net time **222.795–224.370 ms local GPU**
versus **352.875–361.130 ms host-stored GPU**, **1.58–1.61x faster** (two shapes).
Matched local RSS **375.418 MiB/process**, owned GPU **1.791 MiB/card**, unchanged
resident payload; sanitizer peak RSS **418.320 MiB/process**.
[Larger structured products](../experiments/resident-product-larger/report.md)
**passed78469**: all22 arms/52 rank outputs, exact linear/prefix oracles and clean
small sanitizers. HPC gnode13/gnode14,2 H200s,4 CPU cores,8 GiB requested RAM;
controller **4 min13.462 s**, normal host RSS **383.594 MiB/process**, small sanitized
RSS **414.480 MiB/process**, maximum owned GPU **4.612 MiB/card**.

| Structured evaluation values | Host-stored GPU median net product time | Local-access resident GPU median net product time | Resident speedup |
|---|---|---|---|
| 16,384 | 0.349606 seconds | 0.225291 seconds | 1.55× |
| 65,536 | 1.245009 seconds | 0.591319 seconds | 2.11× |
| 262,144 | 4.911230 seconds | 1.553557 seconds | 3.16× |

Three profiled samples/backend/size; both use GPU arithmetic. Setup/oracle and
persistence are outside net product time, inside controller time. Detailed ranges
and boundaries are in the report. These are **s17-field synthetic components**,
not s18 PIR preprocessing. [Root-merge integration](../experiments/distributed-root-merge/report.md)
**passed78508** within the same resource caps.

| Evaluation values | Synthetic roots | Final distributed product net time |
|---|---|---|
| 4,096 | 4,092 | 0.087347 seconds |
| 16,384 | 16,380 | 0.227747 seconds |
| 65,536 | 65,532 | 0.602521 seconds |

Five arms/12 rank outputs passed including small root sanitizer; complete native
coefficient references and scalar witnesses. HPC gnode13/gnode14,2H200/4CPU/8GiB
requested RAM. Controller **2 min28.668 s**; normal peak RSS **390.387 MiB/process**,
small sanitized **416.391 MiB/process**. Post-call owned GPU observations: product
**2.356 MiB/card**, factor engine **3.103 MiB/card**; product stays allocated while
factors run, so these are separate workspace observations, not aggregate peak
GPU memory. One profiled sample/size; setup/oracles/persistence outside product
clock. [Bounded input routing](../experiments/distributed-root-routing/report.md)
is prepared with explicit peer/routed selection. Partitioned support,
production root/Newton execution and fulls18 preprocessing remain.
No fulls18 or production preprocessing speedup claim.

## Current s=17 results

| s | Preprocessing | CPU encoding | GPU encoding |
|---|---|---|---|
| **17** | **HPC: passed, job 75832**<br>Verified resumed worker: **9 h 58 min 38 s**; earlier discovery: **21 min 38 s**<br>Peak sampled process RSS: **348.393 GiB**<br>Hardware: one HPC node, 32 active CPU workers, 4 H200 GPUs | **HPC: passed, job 76645 — resident reuse mode**<br>Encoding: **1 h 45 min 11.087 s** (profiled)<br>Peak sampled process RSS: **427.845 GiB**<br>Hardware: one HPC node, 32 CPU cores, 768 GiB requested RAM, no GPU<br>Verified packed output: **34 GiB**; complete job: **2 h 22 min 49 s** | **HPC: passed, job 76706 — resident reuse mode**<br>Encoding: **7 min 14.930 s**, profiled<br>Peak sampled process RSS: **221.591 GiB**<br>Owned device allocations: **398.463 GiB aggregate**, measured at setup and return<br>Hardware: one HPC node, **4 H200 GPUs, 32 CPU cores, 640 GiB requested RAM**<br>Verified packed output: **34 GiB**, identical to CPU<br>Complete job: **22 min 11 s**<br>**Feasible: one HPC node** |

Sources: [preprocessing result](../experiments/gpu-s17-full-preprocessing/report.md#completed-full-preprocessing-scaling-088)
and [CPU/GPU encoding results](../experiments/s17-full-encoding/report.md#completed-first-attempts-scaling-093).
The detailed tables put **resident reuse first**, with CPU76645 marked passed
and GPU76706 marked passed. Other s17 modes are labeled **projection
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
four H200s with a 112 GiB owned-memory cap per card. That was the failed first
GPU attempt; successful retry76706 requested640 GiB host RAM with the same cores
and device caps. Requested RAM and allocation
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
  two to four GPUs; **four-H200 resident reuse passed76706** after the host/cache
  budget correction. CPU resident reuse also passed; other modes remain unvalidated. Host-backed
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
bounded correctness checks, and **full GPU encoding76706 now passes** with CPU
canonical/packed hash agreement. Encoding takes **434.930 s (7 min14.930 s)**
versus CPU6311.087 s: an observed **14.51× encoding-stage ratio**, based on single
profiled/logged runs on different HPC nodes. The142.441 GiB code loads in318.948 s
on GPU retry versus1752.498 s on CPU; filesystem/cache state differs, so these
are not controlled end-to-end speedups. GPU setup is106.844 s and complete job
22 min11 s, including correctness/output checks.

GPU sampled process RSS is221.591 GiB. Owned device allocations are398.463 GiB
aggregate at setup and return; CUDA context/driver and unobserved transient peak
are separate. Requested640 GiB host RAM/608 GiB total guard provides admission
headroom; observed cgroup peak261.071 GiB includes charged cache and differs from
process RSS. The earlier76646 guard failure is retained in the report as history,
not the current result. [Full GPU evidence and boundaries](../experiments/s17-full-encoding/report.md#completed-full-gpu-retry-scaling-102).

| s17 route | Current evidence or projected payload | Next action |
|---|---|---|
| CPU resident reuse | **Measured: 427.845 GiB peak process RSS; encoding 1 h 45 min 11.087 s** | Full run passed; no repeat needed for correctness |
| CPU file-streamed | Unrun mode; ~270 GiB host arrays, reusable code file ~142 GiB (projections) | Optional lower-RAM alternative; measure repeated reads/cache costs if needed |
| GPU resident reuse | **Measured: 7 min14.930 s encoding; 221.591 GiB sampled host RSS; 398.463 GiB owned device allocations** | Passed on4H200; CPU canonical/packed hashes match |
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
| 18 | Full preprocessing:not run<br>**Exact dimension measured76647:K16300262095**<br>Bounded export76710 and transform/split components76756 passed; see [component results](#s18-lower-memory-preprocessing-work) | Projected final host arrays:889.446GiB<br>Projected discovery arrays:770.893GiB<br>Peak RSS:unknown | Full preprocessing:unknown<br>Count-only GPU sum:37.4ms, excludes setup |1536GiB device product arrays (model)<br>16GPU power-of-two projection across4nodes at112GiB/card<br>Distributed host/GPU implementation and larger storage admission required<br>Alternative: split final-product arrays project384 GiB on4 GPUs; [complete route comparison](#s18-complete-preprocessing-route-decision) records Newton/discovery host limits |
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

| s | **Resident GPU reuse (current baseline)** | Original GPU allocation (historical comparison) | Compact GPU |
|---|---|---|---|
| 10 | HPC measured<br>Runtime: 0.103 s<br>GPU array payload: 12.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.104 s<br>GPU array payload: 16.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 | HPC measured<br>Runtime: 0.108 s–0.109 s<br>GPU array payload: 10.6 MiB<br>Host array payload: 9.1 MiB<br>GPUs: 1 |
| 14 | HPC measured<br>Runtime: 5.176 s–5.179 s<br>Owned GPU allocations: 3.190 GiB<br>Estimated host arrays: ~2.38 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 5.155 s–5.257 s<br>Owned GPU allocations: 4.190 GiB<br>Peak workflow host RSS: 2.521 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 5.800 s–5.860 s<br>Owned GPU allocations: 2.192 GiB<br>Peak workflow host RSS: 2.525 GiB<br>GPUs: 1 |
| 15 | HPC measured§<br>Runtime: 16.917 s–17.756 s<br>Owned GPU allocations: 12.809 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 16.853 s–17.497 s<br>Owned GPU allocations: 16.808 GiB<br>Peak workflow host RSS: 9.768 GiB<br>GPUs: 1 | HPC measured§<br>Runtime: 20.487 s–20.873 s<br>Owned GPU allocations: 8.811 GiB<br>Peak workflow host RSS: 9.770 GiB<br>GPUs: 1 |
| 16 | HPC measured<br>Runtime: 56.975 s–57.418 s<br>Owned GPU allocations: 51.427 GiB<br>Peak host process RSS: 39.031 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 0.216 s–1 min 25.954 s<br>Owned GPU allocations: 67.423 GiB<br>Peak host process RSS: 39.024 GiB<br>GPUs: 1 | HPC measured<br>Runtime: 1 min 33.332 s–1 min 41.526 s<br>Owned GPU allocations: 35.429 GiB<br>Peak host process RSS: 39.028 GiB<br>GPUs: 1 |
| 17 | **HPC: passed, job 76706 — resident reuse mode**<br>Encoding: **7 min 14.930 s**, profiled<br>Peak sampled process RSS: **221.591 GiB**<br>Owned device allocations: **398.463 GiB aggregate**, measured at setup and return<br>Hardware: one HPC node, **4 H200 GPUs, 32 CPU cores, 640 GiB requested RAM**<br>Verified packed output: **34 GiB**, identical to CPU<br>Complete job: **22 min 11 s**<br>**Feasible: one HPC node** | **Projection only: original GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~526.44 GiB<br>Estimated host arrays: 220.88 GiB<br>GPUs: 4<br>Capacity: one node projected; implementation pending | **Projection only: compact GPU**<br>No full s17 measurement for this mode<br>Runtime: unknown<br>Estimated GPU arrays: ~270.44 GiB<br>Estimated host arrays: 220.88 GiB<br>GPUs: 2<br>Capacity: one node projected; implementation pending |
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

Next work is the [s18 preprocessing memory prototypes](#s18-lower-memory-preprocessing-work).
Full s17 preprocessing, resident CPU and four-H200 GPU encoding have passed;
wider arithmetic is validated for those completed paths. Present memory–time/I/O trade-offs before adopting them;
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

The [explicit export integration](../experiments/preprocessing-export/report.md)
is implemented with shared P/G verification, atomic publication and command-owned
support/recovery release; **bounded complete export validation76710 passed** on HPCgnode04
(4H200,4CPU,8 GiB host request;5 min28 s whole job). Independent/resident canonical
and G bytes, atomic publication, restart/corruption/failure checks and four-GPU
API/CLI memcheck pass;568source/74evidence/169cache identities reconcile. With that release, final s18
export-stage arrays project **256.001 GiB** with1 MiB tile, versus377.447 GiB
when support stays referenced. Earlier-stage/device requirements are unchanged;
this is not measured RSS or full s18 feasibility. Single small complete fresh
commands take3.755–6.174 s streamed, including setup/I/O/profile; fixed sample
ordering prevents a speedup claim. The subsequent bounded transform/product
results are below. Full preprocessing limits, faster defaults and all full-size
measurement cells remain unchanged.


The [tiled-transform and split-product comparison](../experiments/preprocessing-tiles/report.md#completed-component-validation-scaling-104)
**passed bounded validation76756** on HPC gnode04 (4 H200 GPUs, 4 CPU cores,
8 GiB requested host RAM; complete job11 min44 s). All independent CPU/output
hash checks and CUDA memcheck pass;578source/45evidence/377cache identities
reconcile. For **2^22 points in the s17 field**, warm profiled measurements are:

| Component, four H200 GPUs | Resident | Lower-memory prototype | Memory–time trade-off |
|---|---|---|---|
| Forward transform, 32 MiB host array | Runtime:235.820–236.090 ms<br>Owned GPU allocations:70.276 MiB aggregate | Host-backed,8 MiB/card tile<br>Runtime:1.303–1.305 s<br>Owned GPU allocations:38.276 MiB aggregate<br>Extra host staging:32 MiB | About5.53 times the forward latency; each direction uploads3230.969 MiB and downloads327.914 MiB versus32 MiB each resident |
| Final-product geometry,16 MiB host output | Runtime:539.435–539.450 ms<br>Owned GPU allocations:102.276 MiB aggregate | Fourteen split products<br>Runtime:3.019–3.020 s<br>Owned GPU allocations:30.414 MiB aggregate<br>Extra host temporary:4 MiB minus4 bytes | About5.60 times the latency; about3.19 times coefficient transfers; host output unchanged |

Each range contains two timed samples, with profiling/logging included and
setup/first-call/persistence reported separately. **These are bounded component
results, not complete s17/s18 preprocessing measurements.** Lower memory is
measured, but full-size runtime penalties and workflow peak remain unknown.
The split final-product geometry projects device arrays **1536 → 384 GiB** at
s18, retaining the256 GiB host output and adding up to64 GiB minus4 bytes of host
temporary. Earlier discovery/root/reciprocal lifetimes still need admission.
No full s18 feasibility or default change follows. The complete design comparison is below; addressing repeated Taylor traffic and
launch overhead remains necessary for the one-node hybrid.

## s18 complete preprocessing route decision

The [complete route comparison](../experiments/s18-preprocessing-routes/report.md)
recommends distributing **host/root work as well as GPUs** for speed. The model
is reproducible and distinguishes measured inputs, principal payload projections
and unknown communication/runtime. Neither route is fully implemented or admitted.

| s18 design (projection) | Device arrays | Host arrays and headroom | Runtime / implementation status |
|---|---|---|---|
| Four H200 GPUs, one HPC node, hybrid | Largest split/root product:384 GiB aggregate,96 GiB/card | Discovery:770.893 GiB; last Newton product:697.446 GiB including split temporary, before overhead; uncollected old prefix can raise retention to825.446 GiB. Earlier support release projects576 GiB stage payload, unimplemented. Tiled spectrum also retains a512 GiB host arena; post-transform streamed-export payload256.001 GiB is not its peak. | Complete runtime:unknown. Bounded prototypes are slower. Missing tiled base-grid discovery, prefix-aware Newton splitting and final-spectrum integration; device capacity alone does not establish feasibility. |
| Sixteen H200 GPUs, four HPC nodes, distributed | Resident full product:1536 GiB aggregate,96 GiB/card | Balanced sharding projects discovery192.723 GiB/node and Newton158.362 GiB/node, before staging/runtime/cache. Requires genuinely sharded support/polynomial state, not a full coordinator copy. | Complete runtime:unknown; four-node/GPU communication unmeasured. Two-node host-memory exchange77127 passed; distributed preprocessing, parallel root work and distributed checkpoints/publication unimplemented. **Recommended design direction for speed**, not proven faster or supported. |

The last **three Newton products**, not just the final identity, exceed the
four-card resident capacity. Prefix-aware splitting needs3/7/13 contributing
block pairs, versus14 for the full identity. Host lifetime improvements benefit
both designs. Canonical files remain950.170 GiB aggregate; conservative recovery
retention brings the payload envelope to1478.170 GiB before storage headroom.
No GPU count reduces canonical data size. No full-size runtime forecast follows
from the small-component ratios; root construction dominated fulls17.
Distributed prototyping is selected. The [shared root boundary](../experiments/preprocessing-root-boundary/report.md)
**passed bounded backend/recovery validation77126** (HPCgnode13,4H200,4CPU,
8GiB host request,4min9s). Native/GPU canonical/root/recovery/race and CUDA memcheck
pass. This establishes correctness, not a measured memory saving.
A [two-node host-memory exchange](../experiments/preprocessing-transport/report.md)
**passed job77127** on HPC gnode13/gnode14 (zero GPUs,2CPU and4GiB requested RAM
per node,5s complete job). At32MiB input/rank,16MiB is sent remotely/node per pass;
median forward time2.447ms gives6.385GiB/s remote payload/node. Full pattern and
roundtrip checks pass;3source/4evidence/binary seals reconcile. Peak process RSS
is297.734/287.867MiB by rank, including MPI. Three measured samples/size; the
shared slowest-rank time is not two independent samples. No GPU/four-node
bandwidth or fulls18 runtime follows. Intermediates
stay in RAM/direct messages; buffered persistent output and bounded writer
concurrency are a separate validation gate. No new full-size measurement or RAM
saving is established; existing CPU/GPU resident results remain unchanged.

### Current root-worker result: matched subset validation77213

The [matched bitmap/search comparison](../experiments/root-subset-validation/report.md#completed-matched-comparison-scaling-112)
**passed job77213** on HPCgnode02 (4H200,4CPU,8GiB requested host RAM;8min11s
whole job). Complete CPU coefficients, scalar-root witnesses, output hashes,
exhaustive/admission/wide-index/race tests and both small CUDA memchecks pass.
601source/51evidence/164cache identities reconcile, with the single explicit
ignored-source snapshot documented in the report. Earlier [worker comparison77148](../experiments/parallel-root-workers/report.md)
remains historical evidence with a different CPU-oracle validation setup.

| Root component | Four workers, bitmap validation | Four workers, sorted validation | Status |
|---|---:|---:|---|
| s8 genuine support,7381 exponents |Mean runtime:0.016320s |Mean runtime:0.016540s |HPC measured; no general search benefit |
| s12 synthetic,65536 exponents |Mean runtime:0.205891s |Mean runtime:0.209219s |HPC measured; about1.6% search overhead, sample ranges overlap |
| s17 synthetic,65536 exponents |Mean runtime:1.480394s |Mean runtime:0.304781s |HPC measured; sparse root subset, not fulls17 preparation |
| s17 synthetic,262144 exponents |Mean runtime:3.427332s |Mean runtime:1.039159s |HPC measured; observed bitmap/search ratio3.30, not fullsize speedup |

Four timed samples per cell use bitmap/search/search/bitmap process order and
matched sorted CPU-oracle setup; profile/logging enabled. Partition, full worker
validation/root work and complete merge are included; setup and final output
checks/persistence are separate. For the largest sorted case, whole-root mean
is2.505141s versus1.039159s with four workers (observed ratio2.41). Four-worker
subtimes are0.062647s partition,0.667263s joined roots,0.309244s merge.

Sampled process RSS: **bitmap4.455–4.685GiB /sorted0.5075–0.5097GiB**, including
runtime/fixtures/CUDA contexts. Owned GPU allocations remain43.102MiB aggregate
for largest four-worker calls in both modes; no device-memory reduction claimed.
Per-validator N-sized bitmap backing is2GiB ats17/8GiB ats18projection; sorted
membership avoids it while retaining every range/order/Frobenius check, adding
O(KlogK) searches in the already sorted indices. The K-index partition copy remains.
Allocation payload, sampled RSS and permanent Newton state are distinct measures.

**Next:** use the explicit sorted option for these measured sparse root-worker
workloads. GPU node-to-node staging and bounded root-polynomial transfer/merge
passed below; next separately validate buffered persistent output. Ordinary full-support/cache/discovery
validation stays bitmap; dense/fullsize search cost and complete s18 performance
remain unknown. No automatic density/field switch, fullsize job or final GPU-count
selection follows. Fulls17 CPU/GPU and fulls18-unrun cells are unchanged.

The [GPU staged-transport gate](../experiments/gpu-staged-transport/report.md)
**retry 77281 passed**, following the corrected pre-compute snapshot failure
77241. [Verified results](../experiments/gpu-staged-transport/runs/staged-02/analysis.json).
HPC:2 nodes (gnode04/05),1 H200 NVL and2 CPU/4 GiB requested host RAM per node;
128 MiB owned-device cap/card,64 MiB pinned staging/rank,2 GiB RSS guard/rank,
15-minute scheduler cutoff. Actual job elapsed **34 seconds**.

| Bounded input per rank | Matched host-only exchange | Complete GPU/staging/exchange pipeline | Remote payload rate per node (MPI / complete pipeline) |
|---|---|---|---|
| 1 MiB | 0.0817 ms | 0.4518 ms | 2.717 / 1.081 GiB/s |
| 8 MiB | 0.5849 ms | 3.0172 ms | 2.497 / 1.295 GiB/s |
| 32 MiB | 2.2691 ms | 11.8274 ms | 2.568 / 1.321 GiB/s |

Three-pass medians after warm-up; slowest-rank timing maxima. Remote payload is
half the input per node; complete-pipeline rate is not network bandwidth.
Normal peak rank RSS:421.016 MiB /411.207 MiB; largest owned device allocation:
33.500 MiB/card; pinned host buffers:64 MiB/rank. This is not total job memory.
Full CPU arithmetic/result, GPU ownership, reverse exchange and both CUDA
memchecks pass. Profiling/logging are included; independent correctness checks
and builds are outside pipeline timing but inside job duration.
Actual normalized-root polynomial computation/transfer/merge now passes below.
No distributed FFT/root pipeline, GPU-direct, four-node/full-size throughput or
fulls18 capacity/runtime is established. Fulls17 CPU/GPU results and fulls18-unrun
cells remain unchanged; ordinary CPU/GPU defaults remain unchanged.

The [cross-node root/merge gate](../experiments/distributed-root/report.md)
**passed77285**, HPC gnode16/17:1 H200 and2 CPU/4 GiB requested host RAM per node.
Job elapsed2 minutes36 seconds including builds, CPU oracles, race checks,
memchecks and profile reporting. All complete CPU/scalar-root, cross-rank
sent/received array, receiving-GPU merged/reference and both CUDA memchecks pass;
621 source/20 evidence/35 cache seals verify without overrides.

| Bounded root input | Complete partition/root/exchange/GPU-merge/verify pipeline | Local root phase | Exchange/wait/peer checks | GPU merge phase |
|---|---|---|---|---|
| s=8 genuine,7,381 exponents | 12.816 ms | 6.090 ms | 0.326 ms | 5.392 ms |
| s=17 synthetic,65,536 exponents | 274.363 ms | 216.357 ms | 8.748 ms | 42.891 ms |
| s=17 synthetic,262,144 exponents | 1.179829 s | 1.010393 s | 23.986 ms | 101.660 ms |

Two-pass medians after warm-up, slowest-rank maxima; phase maxima are not
additive. CPU oracle/setup and persistence are outside pipeline timing. Exchange
includes rank-arrival skew, so these clocks do not measure isolated link bandwidth.
Peak normal rank process RSS:401.129 MiB /410.523 MiB; owned GPU payload:
15.241 MiB/card; no new pinned buffer. Sealed retained cache20.498 MiB, outside Git.
These process/card measurements are not total concurrent job memory. Both ranks
replicate bounded support/CPU oracle/final merge; no full-size memory saving follows.

The [bounded persistence gate](../experiments/preprocessing-persistence/report.md)
**passed77352**, HPCgnode02 (Xeon6710E),2 CPU/4 GiB requested host RAM,0GPU, NFS.
Job elapsed1 minute45 seconds including builds, tests and profile reporting.
All atomic publication/full readback/real small CLI restart/reference/corruption
and concurrent race checks pass;628 source/14 evidence/106 cache seals verify.
Each trial writes96MiB across four24MiB files; mirrored two-sample medians:

| Write mode | Writers | Publication | Full readback | Complete publication/readback | Application payload-write calls |
|---|---|---|---|---|---|
| Direct64KiB codec | 1 | 0.336030 s | 1.296711 s | 1.632754 s | 1,536 |
| Direct64KiB codec | 2 | 0.268352 s | 1.042680 s | 1.311052 s | 1,536 |
| Direct64KiB codec | 4 | 0.198815 s | 1.284555 s | 1.483389 s | 1,536 |
|4MiB coalescing | 1 | 0.346113 s | 0.895899 s | 1.242028 s | 24 |
|4MiB coalescing | 2 | 0.212306 s | 1.391488 s | 1.603806 s | 24 |
|4MiB coalescing | 4 | 0.184447 s | 1.459786 s | 1.644246 s | 24 |

Calls fall64-fold, but elapsed benefits are mixed; **no default change**. Readback
exceeds publication in every trial.4MiB buffers add up to16MiB for4 writers plus
copies, with the same canonical bytes/checks. Profiling/logging/generation/hash
costs are included; two small cached samples do not establish sustained bandwidth.
Persistence sampled group RSS41.137MiB; final retained payload residency96MiB;
concurrent RSS+retained-cache proxy109.609MiB (partial accounting). Persistence
max sampled cgroup charge362.695MiB; whole job cgroup high-water718.566MiB, including
compilation. These are distinct measures. Retained cache111.946MiB, outside Git.

Earlier two-rank and persistence gates pass. The four-rank gate below failed;
transfer diagnosis now blocks scalable ownership/full generator/Newton/transform
and rank-owned publication integration. Investigate coalesced verification reads without removing
checks; larger buffer benefit remains a candidate. Fulls18 runtime/capacity and
multi-node I/O remain unknown, with separate admission required for larger trials.
Fulls17 CPU/GPU results and fulls18-unrun table cells remain unchanged; CPU builds
and faster resident defaults remain unchanged.

The bounded [rank-owned GPU preprocessing tree](../experiments/ucx-cache-gpu-tree/report.md#review-of-78261)
**passed its full four-round crossover in job 78261**. All CPU/state/ownership/
transfer checks passed, with 16 clean CUDA memory-check processes. The validated
configuration explicitly selects UCX with `UCX_RCACHE_ENABLE=n`; earlier direct
failures under the default cache remain recorded, and no runtime default changes.

HPC: **2 H200s across 2 nodes, 4 logical ranks/CPU cores, 8 GiB requested host RAM**.
The largest partial synthetic s=17 fixture (262,144 exponents) took **1.153–1.247 s**
per-round median; normal process peak RSS was **410.836–412.703 MiB/rank**, owned
GPU arrays **15.241 MiB/rank** (contexts additional). Mode advantage reversed
between pairs, so no consistent staging/direct speed winner is established.
These are profiled, setup-excluded diagnostic timings, not full preprocessing.

Next: collective large-product/distributed FFT integration, followed by distributed
discovery and generator/checkpoint/publication composition. The coordinator still
holds a full source/reference and the final product still uses one executor.
Full s=18 preprocessing remains unrun; current full s=17 results and CPU-client
support are unchanged. [Original failed gate](../experiments/root-merge-tree/report.md)
and [partial logging-failed retry](../experiments/ucx-cache-gpu-tree/report.md#review-of-78249)
remain preserved in the evidence.

Latest bounded HPC GPU-tree measurements (job78261; same allocation/build, explicit
UCX cache control; resources and exclusions above):

| Round | Mode | Median pipeline, largest partial fixture | Normal process peak RSS/rank |
|---|---|---:|---:|
| staged-a | Staged | 1.247468 s | 412.055 MiB |
| direct-a | Direct | 1.153021 s | 411.613 MiB |
| direct-b | Direct | 1.246852 s | 410.836 MiB |
| staged-b | Staged | 1.155454 s | 412.703 MiB |

These rows do not fill any full s=18 preprocessing/encoding result cell.

The first collective-FFT data-movement component,
[rank-sharded transpose](../experiments/distributed-transpose/report.md#review-of-78279),
**passed MPI validation 78279**: 16 operations across both widths/modes, 7,312
endpoint checks, exact output shards and legacy root/support lanes. HPC: 2 nodes /
4 CPU cores / 8 GiB requested host / zero GPUs; **1 min 18 s** scheduler elapsed.
Largest 64-bit transpose: **5.476 ms staged / 4.697 ms direct**, one profiled sample;
logged process peak RSS **248.145–249.312 MiB/rank**. This validates the packing/
transport reference, not a distributed GPU FFT or full-size bandwidth estimate.
Stage-scoped GPU row/column operations with global offsets and locally bounded
device buffers passed78293, and vector/transpose composition passed78302.
Taylor-complete composition passed bounded MPI/GPU validation78307. Two full host
shards remain a reference-layout cost, not an adopted production default.
At a full2^36-element grid /16 ranks /64-bit slots, this reference would retain
64 GiB/rank (1 TiB aggregate) for two host shards alone; it is not the selected
production route or a replacement for the GPU capacity models above.


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
resident CPU pass and successful GPU retry are recorded in the current tables.


Historical intermediate plans and calibration tables are preserved in the
[scaling history](encoding-history.md). They do not override current result cells.

The [persistent GPU vector-stage adapter](../experiments/gpu-vector-stage/report.md)
now separates global transform coordinates from bounded local GPU storage and
reuses the existing arithmetic. **78293 independently passes48 normal and42 split-sanitized GPU cases** with
zero CUDA errors/leaks. Raw controller failed on a summary-stream check, now fixed;
component results were verified without rerunning GPU work. Normal suites took
31.210 seconds; sanitizer287.750 seconds; host peak RSS131.109 MiB normal /
167.059 MiB sanitized per process, owned GPU memory4.569 MiB/card. Full-size distributed FFT and full s18 preprocessing remain unvalidated.

The [GPU vector/transpose composition](../experiments/distributed-gpu-vectors/report.md) is implemented with
bounded local tiles and independent ownership/order tests. **78302 passed**:
24 normal /24 sanitized rank/pass results and480 transfer-proof endpoints,
2.722s normal /3.826s sanitizer,372.113 MiB normal /412.508 MiB sanitized host RSS
per process and1.571 MiB owned GPU/card on two HPC H200 nodes. The control layout
adds transfers and is not selected for full-size preprocessing or changed defaults.

The [Taylor-complete distributed FFT gate](../experiments/distributed-gpu-fft/report.md) is now implemented, with
independent CPU/scalar and CUDA checks; **78307 passed** all small forward/inverse
results and transfer checks. Host-shard products **passed78321**; bounded resident
FFT **passed78396** and resident products **passed78399**. GPU-resident large
products, s18 arithmetic and preprocessing integration remain pending.
