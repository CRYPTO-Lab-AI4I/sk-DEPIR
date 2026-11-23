# Encoding scaling reference

The performance goal is **reasonable encoding performance through s=18 inclusive**
on the HPC system, retaining CPU and optional GPU paths. **s=19 and s=20 remain
exploratory targets**: keep their rows as algorithms, layouts and measurements
improve. This is the maintained capacity and results reference for the
[scaling plan](encoding-scaling.md), separate from TRL 5 acceptance.

Updated 2026-11-23; full s17 encoding source `ed0c026`.
**Resident CPU encoding76645 and four-H200 GPU encoding76706 both passed at s17**,
using the same verified preprocessing and deterministic database. Complete
canonical and tightly packed output hashes match. GPU retry76706 supersedes the
failed first GPU attempt76646. Other s17 modes and s≥18 full encoding remain
unvalidated; projections are not support. “Reasonable” has no agreed numerical
latency threshold yet; select it using measured setup, encoding, memory, I/O and
deployment costs before accepting a scaling milestone.

## Current distributed preprocessing step

The distributed algorithms and assembled cold workflow pass genuine small-case
checks on **four workers across two HPC nodes/two H200s**. Final publication
**82100 passed**; durable source-bound checkpoint commits and fresh-process file
reload **82101 passed**. Canonical field18 FFT/products also pass at bounded sizes;
these are not full s18 runs.

**Assembled arithmetic resume82102, actual command82181 and eight-worker82194 passed. Current focus: completed-discovery/root recovery and larger representative s10 validation (SCALING-195), submitted82199; results pending.**
Then parameterized launch/full working-set admission and representative larger
GPU/node scaling remain. Full s18 preprocessing is **not run / not admitted**.
The measurements below distinguish small-case integration from fullsize projections
and the separate full s17 preprocessing/encoding results.

| Step | Status |
|---|---|
| 1. Distributed products and FFTs | **Bounded resident products passed78399**:24 normal +24 sanitized rank/product outputs; host reference also passed.<br>Resident storage78392, vectors78394 and complete FFT78396 passed.<br>**Matched comparison78432 passed** (96 normal +8 small sanitized outputs).<br>**GPU-local Taylor passed78460**, now faster in matched small products.<br>**Larger structured products passed78469** through262144 values; resident3.16× faster at largest size.<br>**Local GPU roots + distributed final merge passed78508**, through65532 roots.<br>**Bounded operand ownership routing passed78662** (all5 arms,10 rank outputs).<br>**Integrated genuine s4/s6 pilot passed78898**: repeated sharded root/Newton/identity/spectrum and canonical artifacts.<br>**Canonical s18 bounded FFT/products passed78917**, with clean small CUDA memchecks. **Four-worker canonical s18 FFT/products passed79239 on two GPUs**,with clean small sanitizers. [Hierarchical sharded root execution passed79376](../experiments/distributed-root-hierarchy/report.md), including unequal child domains and clean small CUDA checks. Fullsize scaling remains pending. |
| 2. Distributed discovery | Local monomial/grid/union ownership and direct orbit routing passed78909 for genuine s4/s6. Four-worker genuine discovery-to-hierarchy passed79541. Coordinator entropy source remains; streamed support/retained-reference cold and fresh-process resume passed79213; full coordinator support/oracle arrays removed in the explicit pilot mode. No fullsize result. |
| 3. Generator construction and verification | Sharded Newton/identity/spectrum integration passed78898 for genuine s4/s6. Four-worker complete genuine cold/restart passed79983;fullsize integration remains. |
| 4. Coordinated checkpoints and publication | Completed-prefix commit/reload and byte-identical canonical artifacts passed78898; partitioned discovery passed78909. **Fresh-process prefix8 restart and bounded streamed publication passed78915** for genuine s4/s6, with identical canonical code/generator output. Four-worker complete cold/restart passed79983. **Optional compact prefix checkpoint cold/restart passed81190**, with exact stored lengths/hashes and identical final artifacts. **Optional GPU-backed spectrum publication passed81365**,including matched resident output and bounded wide/high-bit checks. **Assembled final commit82100 and source-bound durable prefixes/file reload82101 passed**; assembled arithmetic resume82102 passed. Fullsize/node-failure recovery remains. |
| 5. Complete scaling validation | Assembled genuine s4/s6 normal pipeline and s4 sanitizer passed78898. [Canonical s18 bounded FFT/products passed78917](../experiments/s18-field-admission/report.md). Complete four-worker cold/restart passed79983. **Larger genuine s8/curve5 cold/restart passed81286**,two nodes/two H200s;11.147s cold/8.003s restart. Fulls18 remains unrun. |

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
**passed78662**, all5arms/10rank outputs with clean small sanitizer and exact
source-derived input proofs. Largest routed product **0.601303 seconds** versus
peer **0.597966 seconds**; input assembly/routing **32.968/30.703 milliseconds**,
respectively. One sample includes arrival skew; no stable speedup/penalty inferred.
Routed input removes **131,068 bytes** of peer-polynomial host storage per rank
and halves aggregate input traffic (**262,136→131,068 bytes**).
HPC gnode08/gnode10,2H200/4CPU/8GiB requested RAM; controller **2 min58.915 s**.
Normal RSS **390.715 MiB/process**, small sanitized **415.520 MiB/process**;
post-call owned product/factor **2.356/3.103 MiB/card**, separate workspace
observations rather than aggregate physical peak. Cache **21.002 MiB**.

The next implementation step is a reusable distributed executor for complete
rank-owned root reduction and repeated/truncated Newton products. Partitioned
discovery/support, explicit s18 arithmetic admission, coordinated checkpoint/
publication and assembled scaling validation also remain. Current GPU integration
uses two ranks with s17-field synthetic supports, not fulls18 preprocessing.
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
| 18 | Full preprocessing:not run<br>**Exact dimension measured76647:K16300262095**<br>Bounded export76710 and transform/split components76756 passed<br>Genuine s4/s6 pipeline/restart passed78915; streamed coordinator copies removed/verified79213<br>Canonical s18 bounded GPU arithmetic passed78917;four-worker components passed79239 on2GPUs;full distributed workflow still unrun<br>See [component results](#s18-lower-memory-preprocessing-work) | Projected final host arrays:889.446GiB<br>Projected discovery arrays:770.893GiB<br>Peak RSS:unknown | Full preprocessing:unknown<br>Count-only GPU sum:37.4ms, excludes setup |1536GiB device product arrays (model)<br>16GPU power-of-two projection across4nodes at112GiB/card<br>Distributed full-size host/GPU integration and larger storage admission required<br>Alternative: split final-product arrays project384 GiB on4 GPUs; [complete route comparison](#s18-complete-preprocessing-route-decision) records Newton/discovery host limits |
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

The [integrated pilot](../experiments/distributed-preprocessing-pilot/report.md)
reuses the validated kernels instead of adding another isolated component series.
CPU/native checks and both integrated GPU arms passed78898. Discovery
and final artifact assembly still use the coordinator, and two ranks exercise
one root merge. Its prefix reload does not establish fresh-process/node recovery.

Integrated78898: HPC gnode14/gnode15,2H200/4CPU/8GiB RAM, controller **2 min9.545 s**.
Normal peak RSS **372.371 MiB/process**, smallest sanitizer **414.961 MiB/process**;
post-call executor owned GPU **0.075245 MiB/card**, excluding context and earlier
engine peaks; retained cache **14.700 MiB**. Complete measured s4/s6 cases
**2.112602/0.426373 seconds**, including reference/checkpoint/publication work and
cold/warm setup effects, not unprofiled scaling estimates. Both canonical artifact
hashes match independent native preprocessing; all prefix/route checks passed.

[Partitioned discovery integration](../experiments/distributed-discovery-pilot/report.md)
**passed78909** in the same-budget assembled gate. It retains local monomial
lists, grids and union bits, and routes closed orbits directly; only bounded
sampling traffic and a diagnostic support archive use the coordinator. This does not admit s18 parameters or fullsize workloads.

78909 verified partitioned discovery through the complete genuine pipeline: exact
entropy/sample counts, root routing, prefix checks and retained canonical hashes;
small sanitizer clean. HPC gnode15/gnode16,2H200/4CPU/8GiB RAM, controller
**2 min21.458 s**. Normal peak RSS **374.676 MiB/process**, sanitized **410.871
MiB/process**; observed post-call discovery/final-executor owned GPU **0.075245
MiB/card**, excluding context/transients/factor peaks. Retained cache **14.764 MiB**.
Case times s4/s6 **1.156618/0.502099 seconds** include diagnostics and setup; no
repeatable speedup inferred. [Fresh-process restart/publication gate](../experiments/distributed-restart-pilot/report.md)
**passed78915** for the same cases: fresh MPI processes skipped discovery/roots,
resumed prefix8 and wrote identical canonical code/generator files. Controller
**2 min2.635 seconds**; peak process RSS **373.039 MiB cold /372.965 MiB resume**;
observed post-call owned GPU **0.075245 MiB/card**, excluding context/transients
and factor peaks. HPC gnode14/gnode15,2H200/4CPU/8GiB requested RAM. Cold/resume
case times s4 **1.928594/1.403786 seconds**, s6 **0.539676/0.486390 seconds** include
diagnostics/setup/persistence; no stable speedup inferred. Cache **15.216 MiB**.
Fullsize/node-failure recovery and removal of coordinator diagnostic arrays remain.


### Canonical s18 field admission (SCALING-165)

[Bounded gate passed78917](../experiments/s18-field-admission/report.md) reuses the
maintained resident FFT/product and transport schedules at s18, with36-bit
extension values in8-byte slots and18-bit base coefficients in4-byte slots.
Host/native metadata and support codecs now admit s18, retaining all full-domain
allocation checks. Existing CPU lookup defaults remain limited through s17;
explicit GPU server tables add less than3 MiB/set on host/device, versus less
than1.5 MiB at s17. No lower-memory/slower mode is selected. GPU validation passed; fulls18 preprocessing and encoding cells remain unrun. The normal case
uses32/4096 values; sanitizer uses32 only. These are component sizes, not genuine
s18 code lengths (2^36). Existing two-node resource admission is unchanged.


78917: HPC gnode14/gnode15,2H200/4CPU/8GiB requested RAM; controller **2 min25.876
seconds**. All16 normal and8 small sanitized outputs passed; zero CUDA errors/
leaks in four sanitized processes. Normal FFT/product host peak RSS **378.676 /
376.270 MiB per process**, small sanitized **414.059 /411.086 MiB**. Maximum
reported owned GPU **3.118 MiB/card**, excluding context/driver/transients; cache
**16.478 MiB**. At4096 values, one-sample profiled forward/inverse **38.957/36.906
ms**, product shapes **95.070/93.402 ms**, with diagnostic boundaries in report.
No fulls18 memory/runtime or stable speedup follows from these bounded samples.

[Next coordinator-memory integration](../experiments/distributed-coordinator-stream/report.md)
streams canonical support and reads retained, verified native reference files
instead of constructing complete reference/support copies on rank0. Original **79202 failed before GPU work**: snapshot omitted the runtime manifest.
Packaging was fixed; **79213 passed cold and fresh-process resume** in the same
two-node envelope, including exact streamed support and canonical outputs.
No CUDA kernel/default change or repeated sanitizer. Native input generation is
separate measured setup, not a production/fullsize reference requirement.


79213 result: HPC gnode09/gnode11,2H200/4CPU/8GiB requested RAM; controller
**2 min15.180 seconds**,cache **15.355 MiB**. Cold/resume peak process RSS
**373.082/371.145 MiB**; observed post-call ownedGPU **0.075245 MiB/card**
(excluding contexts/transients/factor peaks). Complete profiled s4 cases
**1.656894/1.531600 seconds**,s6 **0.511963/0.511681 seconds**. No stable speedup
or fullsize memory/runtime inferred. [Evidence](../experiments/distributed-coordinator-stream/report.md).

[Four-worker gate](../experiments/distributed-four-worker/report.md) **passed79239**:
4 MPI workers share2H200s on2nodes,4CPU cores/8GiB host unchanged. Each worker
owns at most64MiB GPU payload,128MiB/card aggregate. This validates multi-peer correctness; four-worker full root/discovery/persistence integration and
fulls18 remain unvalidated. Existing two-worker/client defaults remain.


79239: HPC gnode08/gnode09,4MPI workers /2H200s /4CPU cores /8GiB requested RAM.
Controller **2 min27.292 seconds**,cache **16.531 MiB**,32 normal and16 small
sanitized outputs passed;8 clean CUDA error/leak summaries. Normal peak process
RSS **378.047 MiB**,sanitized **413.871 MiB**;reported ownedGPU **3.095 MiB per
worker**,two workers/card (context/driver/transients additional). No fullsize
runtime/peak memory or four-GPU speedup follows. [Evidence](../experiments/distributed-four-worker/report.md).

### Remaining work before full s18 preprocessing (SCALING-169)

| Work | Remaining boundary |
|---|---|
| Multi-level rank-owned root reduction | Hierarchical stage plan,sharded routing and subgroup GPU composition passed79376 for synthetic s18 supports16/4092; genuine discovery-to-tree integration passed79541. The product/FFT backend already passes four-worker tests. |
| Full pipeline at the intended topology | Four-worker discovery/hierarchy passed79541. Complete genuine s4/s6 four-worker cold/restart passed79983,including roots/Newton/identity/spectrum/streamed output and coordinated commits. Fullsize diagnostic-free ownership and wide full-domain execution remain; s18 evidence is component-only. |
| Restart and final artifacts at that topology | Four-worker coordinated commits,fresh-process prefix8 resume and canonical publication passed79983. Representative larger interruption/relaunch and final topology validation remain;no node-loss claim. |
| Representative scaling and full-run admission | Validate one assembled smaller case,then meaningful larger workloads with host/device peaks,communication,I/O and progress logging. Use results to choose GPU/node count,storage,deadline and headroom;obtain full-run resource agreement,then launch fulls18. |

The exact s18 dimension is already measured;the arithmetic and reciprocal/identity
algorithms exist. Current work is connecting their distributed ownership and
validating complete-scale resource behavior. Historical16H200/four-node capacity
projections are not a finalized implementation,measured runtime or an admitted
full-run budget. Fulls18 preprocessing and encoding table cells remain unrun.


### Hierarchical root integration passed (SCALING-170)

[Prototype and trade-offs](../experiments/distributed-root-hierarchy/report.md):
concurrent2-worker sibling products feed a4-worker final merge via bounded shard
redistribution. Explicit MPI group handles share existing codecs/proofs and close
before parent reuse. CPU/race/native shape checks and integrated GPU execution
**passed79376**. Normal synthetic s18 cases16/4092 roots include unequal4096/2048 child
transform domains;small16-root memcheck covers the new layout. Same4workers/
2H200/2nodes/4CPU/8GiB request,64MiBowned/worker,128MiB/card aggregate.
Controller runtime **2 min20.307 seconds**;normal arm **6.189 seconds**;peak logged
RSS **386.105 MiB/worker normal /415.605 MiB/worker sanitized**;maximum reported
owned GPU memory **3.118 MiB/worker**,excluding context/driver/transients. These are
bounded diagnostic measurements;no production default or
fullsize capacity/runtime claim changes. Fulls18 preprocessing/encoding remain unrun.


### Genuine discovery-to-hierarchy integration (SCALING-171)

[Current integration](../experiments/distributed-discovery-hierarchy/report.md)
shares phase-bound integer count exchange,partitioned discovery and orbit routing
across2/4 workers,and feeds locally discovered roots into the same subgroup tree.
Allocation79541 passed genuine s4/s6 sampling/support and every
inter-level polynomial against native references;
complete four-worker Newton/restart/publication now passes79983;fullsize behavior remains unvalidated.
No fulls18 preprocessing/encoding table cell is promoted by these component tests.


Genuine four-worker discovery/root integration **passed79541**,genuine-01,
implementationb94139c after committed-source preflight17bf5e5.
[Receipt and resources](../experiments/distributed-discovery-hierarchy/submission.json).
GPU component results passed;fulls18 and four-worker complete generator/restart remain unrun.


### Four-worker complete pipeline prepared (SCALING-172)

79541 passed on two HPC H200s/two nodes/four CPU cores/8GiB requested RAM:
controller **2 min14.604 seconds**,normal composed arm **6.421 seconds**,maximum
logged RSS **377.020 MiB/worker**;ownedGPU discovery **0.064 MiB/worker**,root
products **0.048 MiB/worker**,excluding context/driver/transients;cache **14.718 MiB**.
These are small s4/s6 correctness measurements,not fulls18 capacity/runtime.

[Next complete pilot](../experiments/distributed-pipeline-four/report.md) reuses that
hierarchy and the existing reciprocal/identity/spectrum/streamed publication,with
rank-bound committed prefixes and a fresh MPI-process prefix8 restart. Shared
control gathering/reviewers preserve the two-worker path;bounded native support
replication is diagnostic only. Same allocation/caps;small complete four-worker GPU validation passed79983.
Then validate fullsize ownership/wide-slot execution and representative scaling
before final topology,headroom and full-run budget admission. Fulls18 table cells
remain unrun;the pending small complete pilot does not establish fulls18 readiness.


Complete four-worker cold/restart validation submitted as **79983**,pipeline-01,
implementatione48a26e after committed-source/native-input preflightbbc1f91.
[Receipt and resources](../experiments/distributed-pipeline-four/submission.json).
Same2HPCnodes/2H200/4CPU/8GiB request and existing caps;GPU results passed79983.
Fulls18 remains unrun;this is a genuine s4/s6 topology/correctness allocation.


### Complete four-worker pass;wide adapters next (SCALING-173)

[79983 passed](../experiments/distributed-pipeline-four/report.md),including cold
execution and fresh-process prefix8 resume. HPC:two H200s/two nodes,four CPU cores,
8GiB requested host RAM,four workers. Controller **2 min11.447 seconds**;cold
launcher arm **7.603 seconds**,resume **4.945 seconds**. Maximum pipeline-row RSS
**376.535 MiB cold /377.840 MiB resume per worker**;reported owned GPU **0.064
MiB/worker**,context/driver/transients additional;cache **15.541 MiB**. All completed
prefixes,full identity,spectrum,communication proofs and native canonical/generator
bytes passed;806 source/56 evidence/387 cache digests reconcile. Fulls18 cells remain unrun.

[Wide adapter validation](../experiments/distributed-pipeline-wide/report.md) keeps
one generic pipeline,32-bit base coefficients and size-dependent extension spectra.
It combines genuine four-worker cold/restart with bounded canonical field18
reciprocal/product/spectrum and explicit high-bit8-byte publication. These32/4096-point
field components are not the2^36-point s18 code grid or a fulls18 artifact. Same
allocation/caps;bounded wide adapter GPU validation passed80962. Scalable diagnostic-free ownership,full wide-domain
workflow,representative scaling/restart and separate full-run admission remain.


Wide adapter combined validation submitted as **80962**,wide-01,implementationee36990,
after committed-source/native-input preflight31c360d.
[Receipt and resources](../experiments/distributed-pipeline-wide/submission.json).
Same2HPCnodes/2H200/4CPU/8GiB request and existing caps;GPU validation passed80962.
Genuine complete cold/restart is s4/s6;field18 tail is bounded32/4096 points.
Fulls18 preprocessing remains unrun.


### Wide adapters passed;owned root execution (SCALING-174)

[80962 passed](../experiments/distributed-pipeline-wide/report.md):811 source/74
evidence/473 cache seals;36 commands. Controller **2min27.889 seconds**;cold
**6.097 seconds**,resume **4.585 seconds**,wide **6.090 seconds**. HPC:two H200s /
two nodes /four CPU cores /8GiB requested RAM,four workers. Maximum measurement-row
RSS **378.285/377.750/379.883MiB per worker** for cold/resume/wide;owned GPU
**0.064MiB/worker narrow /3.125MiB/worker wide**,context/driver/transients additional;
cache **16.001MiB**. All genuine prefixes/canonical bytes and bounded field18
scalar/native/64-bit stream checks passed. Fulls18 cells remain unrun.

[Owned-root integration](../experiments/distributed-owned-roots/report.md) shares
the generic pipeline executor and existing router/product/tree without loading
complete diagnostic support/native root arrays in genuine workers. Independent
native leaf/stage/final files remain controller validation,with complete checksum
provenance. Explicit experimental owned mode retains the diagnostic default;
owned-root GPU validation passed81079 at both field widths. Same admitted resources/caps.
Then finish scalable complete discovery/working-set/full wide-domain workflow,
representative scaling/restart and separate fulls18 admission. No fullsize runtime
or capacity claim follows from the small component pass.


Owned-root combined validation submitted as **81079**,owned-01,implementation638a273,
after committed-source/native-input preflight84fe1a7.
[Receipt and resources](../experiments/distributed-owned-roots/submission.json).
Same2HPCnodes/2H200/4CPU/8GiB request and existing caps;GPU validation passed81079.
Genuine cold/restart remains s4/s6;field18 root/tail cases are bounded components.
Fulls18 preprocessing remains unrun.


### Owned roots passed;restart storage (SCALING-175)

[81079 passed](../experiments/distributed-owned-roots/report.md):818 source/92
evidence/538 cache seals and42 commands. Controller **2min30.945 seconds**;cold
**6.301 seconds**,resume **5.052 seconds**,wide tail **6.355 seconds**,owned-wide
roots **3.766 seconds**. HPC:two nodes/two H200s/four CPU cores/8GiB requested RAM,
four workers. Maximum measurement-row RSS per worker:**376.793/376.992/379.969/
379.785MiB** by arm;owned GPU **0.064MiB narrow /3.125MiB wide tail /3.149MiB
owned-wide per worker**,context/driver/transients additional;cache **16.211MiB**.
All native leaves/stages/final outputs and complete narrow cold/restart artifacts
matched. Small owned-root correctness passes;fulls18 cells remain unrun.

The [next optional compact prefix layout](../experiments/distributed-prefix-checkpoints/report.md)
omits only verified disk padding and restores an admitted resident shard through
bounded I/O. Legacy padded data/default remain. CPU checking and large-file small-read
throughput are unmeasured;do not assume a cost-free runtime optimization. GPU arrays
and packed database encoding are unchanged. GPU cold/restart validation pending.

#### Restart payload projections

For the distributed pilot,assuming N=2^(2s),0<K<N/2 and retaining every completed
root/Newton prefix once,padded bytes are4N(2s+1);compact bytes are4(2N−1).
This is restart payload only,excluding metadata,code/support/generator/spectrum,
debug snapshots,profiles and temporary storage. It is not the historical s17
runner's storage inventory or full-run admission;K=N/2 uses a smaller schedule.

| s | Padded restart payload,projected | Compact restart payload,projected | Validation/admission |
|---|---|---|---|
| 16 | 528GiB | approximately32GiB | Not run with this distributed format;storage admission required |
| 17 | 2.1875TiB | approximately128GiB | Not run with this distributed format;storage admission required |
| 18 | 9.25TiB | approximately512GiB | Not run;full workflow and storage admission remain |
| 19 | 39TiB | approximately2TiB | Exploratory,not run/admitted |
| 20 | 164TiB | approximately8TiB | Exploratory,not run/admitted |

Shared retained bytes are independent of GPU count. Physical filesystem capacity
and agreed full-run budgets must cover all artifacts,not just these payloads.
Larger runtime/I/O/restart measurements and fulls18 resource/headroom/deadline
agreement remain required.


Compact checkpoint validation **passed81190**,prefix-01,implementation3dd1845,
producing revisione9d019a.
[Receipt and resources](../experiments/distributed-prefix-checkpoints/submission.json).
Same2HPCnodes/2H200/4CPU/8GiB request and existing caps. Exact small checkpoint
bytes and full cold/restart correctness passed;fullsize throughput remains unknown.
Legacy padded defaults remain. Fulls18 preprocessing remains unrun.


### Compact checkpoint measurements81190

Two HPC nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM.
All four arms passed independent native/artifact/MPI review in **2 min 49.562 s**
controller time. Genuine s4 cold checkpoint payload was **2,044 bytes versus
9,216 padded bytes**;s6 was **32,764 versus212,992 bytes**. Fresh-process prefix8
restart retained **2,020/32,740 bytes** for s4/s6 versus **7,168/180,224 padded bytes**,
with identical canonical results. This measures restart payload only.
Peak per-worker RSS stayed below **381 MiB**;owned GPU memory below **3.15 MiB/worker**
(context/driver additional). Padded default remains;fullsize compact runtime/total
cache feasibility remain unknown. The next gate is a larger genuine complete
workload within admitted resources,then fullsize working-set/ownership admission.
[Exact timings, resources and evidence](../experiments/distributed-prefix-checkpoints/report.md#verified-compact-restart-checkpoints-job81190-scaling-176).


Larger genuine s8/curve5 validation **81286** passed,implementation7395079,
readiness50f4867. N65,536/K7,381;complete cold and fresh-process prefix8 restart,
owned roots and explicit compact checkpoints. Same2HPCnodes/2H200/4CPU/8GiB
requested host RAM,existing caps. Measured controller2min22.513s;see outcome below.
[Runner,exact preflight,resources and limits](../experiments/distributed-pipeline-scaling/report.md).
Fulls18 remains unrun/unadmitted;no resident default or memory–time policy change.


### Larger genuine distributed workload81286

| Genuine case | Preprocessing cold arm | Fresh-process prefix8 restart | Measured memory | Feasibility/status |
|---|---|---|---|---|
| s=8,curve degree5,N65,536,K7,381 | **Passed;runtime11.146650 seconds** | **Passed;runtime8.002511 seconds**,identical outputs | Peak RSS **376.051/379.316 MiB per worker** cold/restart;owned GPU **0.244 MiB per worker** (driver/context additional) | Measured on **two HPC nodes/two H200s/four workers/four CPU cores/8 GiB total requested host RAM**;fits existing admitted caps |

Controller runtime2min22.513s includes builds,native setup,profiles and review.
Cache33.965MiB;cold checkpoint524284bytes versus4456448padded bytes. Complete
discovery,owned root hierarchy,Newton/full identity,spectrum,canonical publication
and fresh restart passed exact independent references. This is a genuine smaller
code,not fulls18 or an s18 cost estimate. [Full evidence](../experiments/distributed-pipeline-scaling/report.md#verified-larger-workload81286).

Next explicit memory option:stream resident GPU spectrum tiles through the same
canonical publisher,avoiding8N/r host bytes for wide fields (s18:512GiB aggregate,
32GiB/worker at16workers),with no additional transforms or output/transfer bytes
intended. GPU workspace unchanged;fullsize latency unknown,so resident remains
default. Sixteen workers are a capacity example,not implementation support or
resource admission.


### Optional GPU spectrum source: host allocation projection

[Prepared implementation and resource plan](../experiments/distributed-spectrum-publication/report.md)
replace the actual host spectrum shard with4096-word source tiles;GPU workspace
and output/transfer bytes are unchanged in the genuine path. GPU validation is
passed **81365**;resident remains default and fullsize latency is unknown.

| s | Host spectrum allocation avoided (aggregate) | Per worker with16 workers | Status/feasibility |
|---|---|---|---|
| 16 | 16 GiB | 1 GiB | Projection for s16;bounded s8/field18 source passed81365 |
| 17 | 128 GiB | 8 GiB | Projection;existing full resident s17 results remain authoritative |
| 18 | 512 GiB | 32 GiB | Projection;full workflow feasibility/admission remain unresolved |
| 19 | 2 TiB | 128 GiB | Exploratory projection;full workflow unrun |
| 20 | 8 TiB | 512 GiB | Exploratory projection;full workflow unrun |

These are isolated allocations,not total RSS/device/storage savings or16-worker
support/admission. Other host arrays,root ownership imbalance and fullsize logs/
checkpoint/artifact paths still require review. Wide validation retains small
native oracle arrays explicitly and separately tests high-bit GPU reads.


GPU spectrum-source validation **81365** passed,source-01,implementation
f6353af/matched baseline98fce69/readinessf1dcb47. Source genuine s8 cold/restart,
matched resident-default cold,and bounded canonical field18 GPU source/high-bit
checks. Same **two HPC nodes/two H200s/four CPU cores/8 GiB requested host RAM**,
existing guards/caps. Measured168.928s controller;see measured outcome below.
[Exact scope,source/native preflight and trade-offs](../experiments/distributed-spectrum-publication/report.md).
No fulls18 run or new topology admission;resident default retained.


### Spectrum source/resident measurements81365

| Genuine s8/curve5 mode | Status and runtime | Memory evidence | HPC resources |
|---|---|---|---|
| GPU source cold | Passed;10.713813 seconds | RSS379.027 MiB/worker;actual spectrum0resident words plus4096-word source tile | Two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM |
| GPU source fresh-prefix8 restart | Passed;7.749624 seconds | RSS379.039 MiB/worker;same bounded source allocation | Same allocation |
| GPU resident cold | Passed;10.346982 seconds | RSS379.180 MiB/worker;16,384resident actual spectrum words | Same allocation |

All canonical outputs match;each source downloads exactly16,384words/worker once.
Owned GPU0.244MiB/worker in genuine arms;context/driver additional. Source cold was
3.545%slower in this one fixed-order sample;not a fullsize runtime projection.
Bounded field18 source/high-bit checks also pass (5.793278s,RSS380.199MiB/worker,
owned GPU3.125MiB/worker). Total controller2min48.928s;cache37.413MiB.
[Exact evidence/limits](../experiments/distributed-spectrum-publication/report.md#verified-result81365).
Resident remains default;source remains explicit. Fulls18 is unrun. Next target
the full host discovery-result copy;phase-specific spectrum/discovery allocation
savings must not be summed as a simultaneous total-RSS saving.


### Optional discovery-result tiles: phase-specific allocation projection

[Prepared shared implementation/resource plan](../experiments/distributed-discovery-results/report.md)
consumes each GPU discovery result once into the existing local union bitmap.
Resident remains default;bounded validation passed **81454**,and fullsize latency remains unknown.

| s | Avoided discovery-result host allocation (aggregate) | Per worker with16 workers | Full-workflow status |
|---|---|---|---|
| 16 | 16 GiB per sample | 1 GiB | Allocation projection for new optional mode |
| 17 | 128 GiB per sample | 8 GiB | Projection;existing full resident results remain authoritative |
| 18 | 512 GiB per sample | 32 GiB | Projection;full workflow ownership/admission unresolved |
| 19 | 2 TiB per sample | 128 GiB | Exploratory,full workflow unrun |
| 20 | 8 TiB per sample | 512 GiB | Exploratory,full workflow unrun |

A≤4096-word source tile replaces that array;GPU workspace,base grid,monomial list,
bitmap,support and root arrays remain. These savings occur in a different phase
from spectrum output and must not be summed as total peak-RSS evidence. Sixteen
workers are illustrative,not a supported/admitted topology. Same two-node/two-H200
validation caps;fulls18 remains unrun.


Discovery-result tile validation **81454** passed,results-01,implementation
1d8e000/readiness77b9c8a. Genuine s8/curve5 cold/fresh-prefix8 restart and matched
resident-discovery cold,holding streamed spectrum output constant. Same **two HPC
nodes/two H200s/four CPU cores/8 GiB total requested host RAM** and prior caps;
measured2min50.037s controller;see outcome below.
[Source/native preflight,scope and trade-offs](../experiments/distributed-discovery-results/report.md).
Fulls18 remains unrun/unadmitted;no resident default or resource increase.


### Discovery-result measurements81454

| Genuine s8/curve5 mode | Status/runtime | Host memory and ownership | HPC allocation |
|---|---|---|---|
| Tiled discovery cold | Passed;11.256600 seconds | Peak RSS377.820 MiB/worker;0resident result words plus4096-word tile | Two nodes/two H200s/four workers/four CPU cores/8 GiB total requested host RAM |
| Fresh-process prefix8 restart | Passed;7.952252 seconds | Peak RSS378.926 MiB/worker;no discovery repeated | Same allocation |
| Resident discovery cold | Passed;10.994534 seconds | Peak RSS377.977 MiB/worker;16,384resident result words | Same allocation |

Both cold arms use streamed spectrum output and match all native/canonical/
restart checks. Each sample reads16,384words/worker once;two samples. Owned GPU
0.244MiB/worker,driver/context additional;controller2min50.037s;cache37.154MiB.
Tile mode was2.384%slower in one fixed-order matched sample,not a fullsize cost
estimate. [Exact evidence](../experiments/distributed-discovery-results/report.md#verified-result81454).
Resident remains default. Next audit remaining fullsize arrays/ownership and
prototype explicit Newton root-prefix input views;fulls18 remains unrun.


### Source-based fullsize audit and Newton prefix views

[Reproducible audit](../experiments/distributed-fullsize-audit/report.md) separates
array payloads,host inventories/GC uncertainty,root-domain imbalance,persistence
and current implementation limits. [Explicit prefix input prototype](../experiments/distributed-prefix-inputs/report.md)
avoids Newton's whole root clone/mask and uploads only selected coefficients,
using the same zero-reset GPU product. Dense remains default;bounded validation passed **81729**.

| s | Three-array GPU payload at16 cards | Payload-only minimum power-of-two cards at112 GiB/card | Root clone avoided per worker/step at16 workers | Full distributed feasibility |
|---|---|---|---|---|
| 16 | 3 GiB/card | 1 | 1 GiB | Projection;new distributed fullsize path unrun |
| 17 | 24 GiB/card | 4 | 4 GiB | New distributed path unrun;historical full resident results passed |
| 18 | 96 GiB/card | 16 | 16 GiB | Payload may fit;complete ownership/production/admission unresolved |
| 19 | 384 GiB/card | 64 | 64 GiB | Exploratory;current CUDA/Taylor limits stop this configuration |
| 20 | 1.5 TiB/card | 256 | 256 GiB | Hypothetical hardware,availability unverified;also beyond current64-rank bound |

Device tables/windows/context and host/native/GC memory are additional;these are
not run resources or full-feasibility claims. Prefix s18 Newton input uploads
project36 TiB→1,394.893 GiB,without reducing FFT/network work. Fullsize latency is
unknown. Current diagnostic serialization can create full temporary buffers;
P/G/identity plus spectrum debug files add1.25 TiB at s18. Current transpose proof
logging alone would emit440,401,920rows/worker. Native-reference/four-worker/pilot
limits,production persistence and local root counts remain separate gates.


Newton prefix-input validation **81729** passed,prefix-01,implementation
2f63e9d/readiness75524c2. Genuine s8/curve5 prefix cold/fresh restart,matched dense
input,and bounded field18 prefix/native/source checks. Other memory flags held
constant in matched arms. Same **two HPC nodes/two H200s/four CPU cores/8 GiB
total requested host RAM** and existing caps;estimate3–6min excluding queue.
[Exact preflight,scope and trade-offs](../experiments/distributed-prefix-inputs/report.md).
Dense default retained. Fulls18 remains unrun/unadmitted.


### Newton prefix-input measurements81729

| Genuine s8/curve5 mode | Status/runtime | Allocation/transfer evidence | HPC resources |
|---|---|---|---|
| Prefix input cold | Passed;11.328589 seconds |0root clone words;161,392Newton input upload words aggregate;RSS377.789 MiB/worker |Two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM |
| Prefix input fresh-prefix8 restart | Passed;8.994542 seconds |0root clone words;161,367upload words;RSS379.551 MiB/worker |Same allocation |
| Dense input cold | Passed;11.652888 seconds |16,384clone words/worker/step;2,097,152upload words aggregate;RSS380.027 MiB/worker |Same allocation |

All native/canonical/restart outputs match. Prefix cold was2.783%faster in one
fixed-order sample,not a fullsize speedup estimate. Owned GPU0.244MiB/worker,
context/driver extra. Bounded field18 prefix/native checks pass6.687571s with
RSS379.930MiB/worker and ownedGPU3.125MiB/worker. Controller3min11.051s;cache37.509MiB.
[Exact evidence](../experiments/distributed-prefix-inputs/report.md#verified-result81729).
Dense remains default. Next bounded proof/native recording must preserve all
runtime verification/resource guards;fulls18 production/admission remain unresolved.


### Bounded proof/native recording — passed81738

[Implementation and paired validation plan](../experiments/distributed-proof-recording/report.md)
retain all runtime descriptor/payload/peer/field/identity checks and per-proof RSS
guards while optionally storing fixed per-kind ordered digest/count/byte summaries.
Full/default behavior remains;trace mode additionally hashes full small baselines.
Native window totals replace optional per-window starts/completions while parent
pass/target progress and errors remain. GPU validation **passed81738**, all six paired arms.

This addresses the audit's recording growth,not GPU payload,fullsize rank support
or production persistence. Summary serialization/hashing overhead and fullsize
latency remain unknown. Six paired cold/restart/field18 arms use existing two-node/
two-H200 caps;no fulls18 or larger allocation is authorized.

### Proof-recording measurements81738

| Arm | Result/runtime | Peak host RSS | Aggregate rank stdout | Proof records checked |
|---|---|---|---|---|
| cold | Passed; 9.496726 seconds | 378.316 MiB/worker | 22.693 MiB | 31,574 |
| resume | Passed; 6.765242 seconds | 378.070 MiB/worker | 12.224 MiB | 17,090 |
| summary-cold | Passed; 9.415148 seconds | 378.484 MiB/worker | 2.792 MiB | 31,574 |
| summary-resume | Passed; 6.813566 seconds | 378.051 MiB/worker | 1.927 MiB | 17,090 |
| wide | Passed; 4.874992 seconds | 379.629 MiB/worker | 25.383 MiB | 40,012 |
| summary-wide | Passed; 4.944900 seconds | 380.543 MiB/worker | 1.061 MiB | 40,012 |

Cold summary replaces 31,574 per-packet rows with 16 summaries; restart replaces
17,090 with 12; wide replaces 40,012 with 12. Ordered digests, counts, semantic byte
lengths, all native Taylor window/coefficient totals and every output comparison
match. Runtime MPI/field/full-identity/canonical checks remained enabled.
Owned GPU peaks are 0.244 MiB/worker for s8 and 3.125 MiB/worker for bounded field18;
CUDA/MPI/context memory is additional. Controller runtime **3 min 19.204 seconds**;
retained cache **40.750 MiB** (exact inventory is sealed in raw results).

One paired cold summary was faster, while restart and wide summary were slower.
These single, fixed-order samples do not establish a runtime benefit or penalty
at full size; trace includes hashing that default full mode omits. Summary remains
explicit. Stdout totals include other unsuppressed stage/plan/metric events; native
stderr also remains, so these are not total retained-storage figures. No full s18
feasibility or runtime follows from this small validation.


[Exact methods/evidence](../experiments/distributed-proof-recording/report.md#verified-result81738). Next: explicit bounded diagnostic serialization, preserving all stored bytes and default behavior.


### Bounded diagnostic serialization — passed81744 partials plus81746

[Implementation/validation plan](../experiments/distributed-diagnostic-writes/report.md):
explicit bounded writes limit extra serialization to 64 KiB instead of a full
snapshot byte buffer. s18/16-worker base snapshot projects 16 GiB saved per writer;
source arrays, retained bytes and GPU workspace remain unchanged. Extra write calls
and full-size throughput/RSS are unmeasured; bulk remains default. CPU allocation
and four-arm cold/restart/bulk/wide validation use existing two-node/two-H200 caps.
No full s18 run or feasibility claim.

Diagnostic validation **81744 failed final controller review** after all GPU commands completed. Bounded cold/restart and matched bulk control independently passed; wide inherited bulk mode and does not validate bounded writes. [Exact failure and partial evidence](../experiments/distributed-diagnostic-writes/report.md#failure81744-and-targeted-correction). No full s18 admission.


| 81744 mode on HPC (two nodes/two H200s/four workers/four cores/8 GiB requested RAM) | Status | Runtime | Peak host RSS |
|---|---|---|---|
| s8 bounded diagnostic cold | Passed independently; overall job failed controller review | 9.135297 seconds | 376.750 MiB/worker |
| s8 bounded prefix8 restart | Passed independently | 6.804956 seconds | 379.473 MiB/worker |
| s8 bulk control | Passed; same canonical output | 9.338670 seconds | 378.961 MiB/worker |
| Bounded field18 workflow, actual bulk writes | Arithmetic passed; intended bounded-write check missing | 4.929941 seconds cumulative | 379.082 MiB/worker |

The failure was an inherited experiment option, corrected with explicit arm settings
and a regression that exercises the suite configuration before GPU work. Preserve
all partial results; retry only bounded wide under unchanged caps. Large-file
throughput/peak-memory savings and full s18 feasibility remain unmeasured.

Targeted bounded-wide retry **81746 passed**: [receipt](../experiments/distributed-diagnostic-writes/submission-wide-02.json). Same two HPC nodes/two H200s/four cores/8 GiB requested RAM and existing caps; only missing wide arm rerun, estimate2–4minutes excluding queue. 81746 passed; original81744 overall status remains failed with partial passes preserved.


### Completed bounded diagnostic composition81746

| Mode on HPC | Result | Runtime | Memory | Saved diagnostic evidence |
|---|---|---|---|---|
| Bounded field18 d5/d12 workflow | Passed81746 | 5.211485 seconds cumulative | Peak RSS379.434 MiB/worker;owned GPU3.125 MiB/worker |72 files/197,888 bytes;exact native/canonical matches |

Two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM, unchanged
caps;controller2min5.422s. [Methods/evidence](../experiments/distributed-diagnostic-writes/report.md#verified-result81746).
Combined with81744's separately labeled narrow partial passes, this completes the
planned bounded diagnostic codec validation. Bulk remains default; large-file
latency and peak-RSS savings remain unknown. Full s18 unrun/unadmitted.

### Coordinated restart support binding — passed81750

[Plan](../experiments/distributed-restart-binding/report.md): explicit one full-file
coordinator checksum plus complete per-checkpoint metadata agreement, retaining all
local shard/checksum/field/padding/ownership checks and immutable-cache assumptions.
At illustrative s18/16workers, support verification reads project3.795TiBaggregate
->121.446GiB, adding two small collective rounds. Memory is bounded; no extra stored
support copy or GPU work. Full-size latency remains unmeasured;local remains default.
Three-arm cold/coordinated-restart/local-restart comparison uses existing two-node/
two-H200 caps;this is not fullsize production support or admission.

Coordinated/local restart comparison **81750 passed**: [receipt](../experiments/distributed-restart-binding/submission.json). Three cold/restart/control arms,unchanged two-node/two-H200 allocation;all passed. Local default retained;fulls18 unrun/unadmitted.


### Coordinated restart measurements81750

| Genuine s8/curve5 mode on HPC | Result/runtime | Verification traffic,aggregate | Peak host RSS |
|---|---|---|---|
| Cold | Passed;9.149322 seconds |0 recovery hashes |379.121 MiB/worker |
| Coordinated fresh-prefix8 restart | Passed;6.792011 seconds |1 hash/29,524 bytes;two metadata agreements/worker |379.160 MiB/worker |
| Local fresh-prefix8 restart | Passed;6.661931 seconds |8 hashes/236,192 bytes |377.895 MiB/worker |

Same two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM.
Owned GPU0.244 MiB/worker;context/MPI/driver extra. Controller2min39.483s/cache37.252MiB.
[Methods and evidence](../experiments/distributed-restart-binding/report.md#verified-result81750).
One small cached coordinated sample was1.953%slower;local remains default. Logical
hash bytes are not physical disk metrics or full-size latency estimates. Full s18
remains unrun/unadmitted. Runtime rows precede deferred GPU close;launcher timing
includes it.

### Shared MPI/GPU runtime extraction — passed81763

[Plan](../experiments/distributed-runtime/report.md): unchanged validated transport
moves from test-only files into internal MPI runtime,with one error-returning GPU
session shared by narrow/wide workflows. Thin test adapters retain independent
native/canonical/proof observers. Ordinary CPU,CUDA-only,MPI-only and combined
builds stay explicit;no public production CLI/rank expansion/fulls18 claim. Resource
payloads/layouts/defaults unchanged;bounded cold/restart/wide integration uses the
existing two-node/two-H200 budget,not repeated verification-mode timing/sanitizers.

Shared-runtime integration **81763 passed**: [receipt](../experiments/distributed-runtime/submission.json). All three cold/coordinated-restart/wide arms passed under unchanged two-node/two-H200 caps. Internal runtime only;no production/full s18 claim.


### Shared runtime measurements81763

| Mode on HPC | Result/runtime | Peak host RSS | Owned GPU memory |
|---|---|---|---|
| Genuine s8/curve5 cold | Passed;9.105248 seconds |378.125 MiB/worker |0.244 MiB/worker |
| Fresh coordinated prefix8 restart | Passed;6.962673 seconds |379.320 MiB/worker |0.244 MiB/worker |
| Bounded field18 d5/d12 workflow | Passed;4.758988 seconds cumulative |381.219 MiB/worker |3.125 MiB/worker |

Two nodes/two H200s/four MPI workers/four CPU cores/8 GiB requested host RAM;
context/MPI/driver memory additional. Controller2min40.959s;cache34.725MiB.
[Exact methods/evidence](../experiments/distributed-runtime/report.md#verified-result81763).
No fullsize or production-entry claim. Rows precede deferred GPU close;launcher
includes it. Native/proof/canonical/checkpoint/diagnostic checks all passed.

### Shared reciprocal orchestration — passed81777

[Plan](../experiments/distributed-reciprocal/report.md): one backend-independent
error-returning Newton loop for cold/resumed/narrow/wide workflows, retaining the
existing square/product kernels, explicit dense/prefix semantics, counters and
checkpoint/reload sequence. Full code identity scan also moves to runtime; all
native/canonical observers remain independent. No additional intended coefficient
arrays/transforms/I/O or default change;full-size GC/latency/RSS remain unmeasured.
Same-budget cold/restart/dense-preservation/wide validation;not production entry,
larger-rank support or full s18 admission.

Shared reciprocal integration **81777 passed**: [receipt](../experiments/distributed-reciprocal/submission.json). All cold/restart/default-dense-preservation/wide arms passed under unchanged two-node/two-H200 caps. Full s18 remains unrun/unadmitted.


### Reciprocal runtime measurements81777

| Mode on HPC | Result/runtime | Peak host RSS |
|---|---|---|
| s8 prefix cold |Passed;10.020557 seconds |377.828 MiB/worker |
| Coordinated prefix8 restart |Passed;6.759730 seconds |378.211 MiB/worker |
| s8 dense cold control |Passed;9.592736 seconds |379.574 MiB/worker |
| Bounded field18 d5/d12 |Passed;4.999263 seconds cumulative |381.273 MiB/worker |

Two nodes/two H200s/four workers/four cores/8 GiB requested RAM. Owned GPU0.244MiB/
worker narrow,3.125MiB wide;driver/context/MPI extra. Controller3min4.002s/cache37.760MiB.
[Methods/evidence](../experiments/distributed-reciprocal/report.md#verified-result81777).
All native/canonical/proof/checkpoint/counter checks pass. No fullsize speed claim;
default dense preserved. Rows precede deferred GPU close;launcher includes it.

### Root runtime and actual-count admission — passed82085

[Plan](../experiments/distributed-root-runtime/report.md): local root construction
and owned subgroup lifecycle become internal runtime with error returns and optional
diagnostic observer. Actual leaf counts/rounded stage shards must fit explicit
single-shard bounds before GPU work;device allocation/RSS guards remain. This is
not whole-process resource admission or full s18 support. Existing-budget cold/
restart/owned-wide validation;no repeated unchanged dense/wide-tail/sanitizer arms.

Root-runtime integration **82085 passed**: [receipt](../experiments/distributed-root-runtime/submission.json). All cold/restart/owned-wide root arms passed under unchanged two-node/two-H200 caps. Single-shard admission is not whole-process/full s18 feasibility.


### Root runtime measurements82085

| Mode on HPC | Result/runtime | Peak host RSS | Owned GPU |
|---|---|---|---|
| s8 cold |Passed;9.349141 seconds |377.965 MiB/worker |0.244 MiB/worker |
| Coordinated prefix8 restart |Passed;6.174301 seconds |378.742 MiB/worker |0.244 MiB/worker |
| Bounded field18 owned roots |Passed;3.380598 seconds launcher elapsed |379.930 MiB/worker |3.149 MiB/worker |

Two nodes/two H200s/four workers/four cores/8 GiB requested RAM;context/MPI/driver
additional. Controller2min31.543s/cache34.675MiB. [Methods/evidence](../experiments/distributed-root-runtime/report.md#verified-result82085).
Cold/restart rows precede deferred GPU close;owned-root launcher includes cleanup.
Actual leaf/stage/native/proof/canonical/checkpoint checks pass. Post-run analyzer's
scratch-source path corrected with exact producing Git/hash binding;no GPU rerun.
No fullsize/whole-process admission claim.

### Shared discovery session — passed82096

[Plan](../experiments/distributed-discovery-runtime/report.md): reuse internal GPU
session for discovery allocation/transports/progress/error/close in both streamed
and default resident result modes. Sampling/randomness/union/archive order and
transform/proof identities remain;bounded harness still owns orchestration.
Same-budget cold/restart/resident-preservation validation;no new resource or
fulls18 support. Production phase admission/non-oracle entry/publication remain.

Discovery-session integration **82096 passed**: [receipt](../experiments/distributed-discovery-runtime/submission.json). All streamed cold/restart/resident-preservation arms passed under unchanged two-node/two-H200 caps. Full s18 remains unrun/unadmitted.


### Discovery session measurements82096

| Mode on HPC | Result/runtime | Peak host RSS |
|---|---|---|
| Genuine s8 streamed cold |Passed;8.816028 seconds |377.402 MiB/worker |
| Coordinated prefix8 restart |Passed;6.604928 seconds |377.555 MiB/worker |
| Resident-discovery cold control |Passed;9.292102 seconds |379.422 MiB/worker |

Two nodes/two H200s/four workers/four cores/8 GiB requested RAM;owned GPU0.244MiB/
worker,context/MPI/driver extra. Controller2min37.875s/cache37.485MiB.
[Methods/evidence](../experiments/distributed-discovery-runtime/report.md#verified-result82096).
All entropy/sample/support/native/proof/canonical checks passed. Single samples
are not fullsize performance evidence;resident default retained. Rows precede
deferred GPU close;launcher includes it.

### Explicit-policy discovery runtime — passed82097

[Plan](../experiments/distributed-discovery-loop/report.md): sampling/union/root-owner/
archive loop becomes internal error-returning runtime with explicit random reader,
CUDA/result policy and logical host-list limits. Deterministic seed stays in tests;
no native reference enters runtime. No intended extra arrays/FFT/transfers/files.
Single-list limits are not total RSS/admission. Same-budget cold/restart/resident
validation;full production entry/persistence/phase admission remain,fulls18 unrun.

Explicit-policy discovery validation **82097 passed**: [receipt](../experiments/distributed-discovery-loop/submission.json). All streamed cold/restart/resident-preservation arms passed under unchanged two-node/two-H200 caps. Fulls18 unrun/unadmitted.


### Discovery runtime measurements82097

| Mode on HPC | Result/runtime | Peak host RSS |
|---|---|---|
| Genuine s8 streamed cold |Passed;10.271066 seconds |377.656 MiB/worker |
| Coordinated prefix8 restart |Passed;6.841209 seconds |380.207 MiB/worker |
| Resident-discovery cold |Passed;9.283122 seconds |379.730 MiB/worker |

Two nodes/two H200s/four workers/four cores/8 GiB requested RAM;owned GPU0.244MiB/
worker,context/MPI/driver additional. Controller2min52.256s/cache37.534MiB.
[Methods/evidence](../experiments/distributed-discovery-loop/report.md#verified-result82097).
All entropy/native/proof/canonical checks pass;single samples are not fullsize
performance evidence. Rows precede deferred GPU close;launcher includes it.

### Reference-independent runtime workflow — verified execution82099; controller failed

[Plan](../experiments/distributed-workflow/report.md): compose runtime discovery,
local/subgroup roots,reciprocal,full identity,spectrum and canonical writing without
native reference inputs. Explicit caller randomness/resource/output/checkpoint
policy;test observers retain comparison data. Same-budget pilot cold/restart plus
assembled cold validates every output/checkpoint/proof against reviewed native data.
Assembled observer persistence differs from pilot durable manifests;timings are not
a direct speedup comparison. Production restart/atomic publication/command policy,
larger topology and full resource admission remain. Full s18 unrun/unadmitted.

Assembled runtime workflow validation **82099 failed controller review; all executed arms independently verified**: [receipt](../experiments/distributed-workflow/submission.json). Reviewed pilot cold/restart and reference-independent assembled cold all passed corrected independent review under unchanged two-node/two-H200 caps. Full s18 unrun/unadmitted.


### Recovered execution evidence82099

| Mode on HPC | Independent result/runtime | Peak host RSS |
|---|---|---|
| s8 reviewed pilot cold |Passed;8.897056 seconds |380.473 MiB/worker |
| Coordinated prefix8 restart |Passed;6.239122 seconds |378.480 MiB/worker |
| Reference-independent assembled cold |Passed corrected review;8.486219 seconds |378.082 MiB/worker |

Original job/controller status remains **failed**,not relabeled. Controller expected
an in-process reference event absent by design;explicit native metadata now supplies
the independent entropy/count reference without removing any checks. All76saved
comparison files,ordered proof summaries,native roots/identity/final bytes match.
[Exact failure and evidence](../experiments/distributed-workflow/report.md#recovered-validation82099).
Two nodes/two H200s/four workers/four cores/8 GiB requested RAM;owned GPU0.244MiB/
worker,context/MPI/driver extra. Controller151.588s/cache36.123MiB. No GPU rerun needed.
Timing boundaries/persistence differ;no direct speedup/fullsize inference.

### Final publication commit — passed82100

[Plan](../experiments/distributed-publication-commit/report.md): explicit staged
payloads,streaming hashes/exact byte counts,payload sync and last-written synced
manifest bind complete code/generator/raw-support output. No payload reread/copy;
extra SHA CPU and manifest/directory sync latency unmeasured. Direct observer mode
remains. Consumers verify manifest/hash before accepting outputs;no production
restart/authentication claim. Same-budget integration;fulls18 unrun/unadmitted.

Final publication-commit validation **82100 passed** under unchanged caps;
[verified evidence](../experiments/distributed-publication-commit/runs/publication-01/analysis.json).
Controller **2 min38.073 s**; cold **8.907 s**, restart **6.374 s**, committed assembled
**8.510 s**; assembled peak host RSS **378.824 MiB/worker**. All76comparison files,
proof/native checks and three manifest-bound payloads match. Timing boundaries differ;
no direct speedup claim. Fulls18 remains unrun/unadmitted.


### Durable assembled checkpoint integration — passed82101

[Plan/implementation](../experiments/distributed-workflow-checkpoints/report.md):
source-bound coordinated compact root/Newton commits, sealed support alias without
payload copying, fresh-process root/prefix8 file reload. No new GPU/host capacity
claim; extra metadata/control/sync latency unknown. Arithmetic restart integration
and larger parameter/topology/whole-job admission remain. Same HPC allocation:
2nodes/2H200/4cores/8GiB requested host RAM,4workers;limits in the report.

| Validation | Status | Runtime / memory |
|---|---|---|
| Final publication82100 | **Passed** | Controller **2 min38.073 s**; assembled launcher **8.510 s**; peak host RSS **378.824 MiB/worker**; owned GPU **0.244 MiB/worker** (context/driver extra); retained cache **36.220 MiB** |
| Durable assembled prefix commits82101 | **Passed**, including fresh-process file reload | Controller **2 min41.494 s**; assembled launcher **9.893 s**; peak host RSS **380.570 MiB/worker**; owned GPU **0.244 MiB/worker** (driver/context extra) |
| Assembled arithmetic resume82102 | **Passed** | Controller **3 min11.398 s**; restart launcher **8.297 s**; peak host RSS **381.387 MiB/worker**, owned GPU **0.244 MiB/worker** (driver/context extra) |
| Parameterized MPI command82181 | **Passed** | Controller **2 min54.239 s**; command restart **7.896 s**, peak host RSS **376.066 MiB/worker** |
| Host phase inventories /eight-worker82194 | **Passed** | Controller **2 min51.032 s**; cold **15.555 s**, restart **11.050 s**; logged peak RSS **374.309 /374.105 MiB/worker** |
| Discovery/root recovery and s10 representative case | **Submitted82199; results pending** | Same approved ceilings; estimated controller **4–9 min**; fullsize memory/runtime unknown |
| Full s18 preprocessing | **Not run / not admitted** | Capacity projections above are not measured costs or implementation support |

Durable assembled checkpoint validation **82101 passed** under unchanged caps;
[verified evidence](../experiments/distributed-workflow-checkpoints/runs/checkpoints-01/analysis.json).
Controller **2 min41.494 s**; cold **10.400 s**, pilot restart **7.532 s**, assembled
**9.893 s**; assembled peak host RSS **380.570 MiB/worker**. All17commit manifests,
68shards, source binding/alias, native/proof comparisons and fresh-process file
reload pass. Arithmetic restart integration remains next; fulls18 unrun/unadmitted.


### Next: assembled arithmetic resume

[Plan](../experiments/distributed-workflow-resume/report.md): shared cold/restarted
Newton/identity/spectrum/publication, coordinated source verification and no-copy
archive reuse. Same HPC budget2nodes/2H200/4cores/8GiB requested host RAM/4workers;
resource limits and timing boundaries in the report. No fullsize capacity change.
Restarted arithmetic82102 passed; larger topology/complete job admission
remain after this test.

Assembled arithmetic-resume validation **82102 passed**;
[verified evidence](../experiments/distributed-workflow-resume/runs/resume-01/analysis.json).
Controller **3 min11.398 s**, assembled cold **10.691 s** /restart **8.297 s**;
restart peak host RSS **381.387 MiB/worker**, owned GPU about **0.244 MiB/worker**
(context/driver extra). Remaining13steps/proof/native work/checkpoints/final bytes
match the pilot, with one support hash/no-copy archive. Fulls18 unrun/unadmitted.

Parameterized command [implementation/validation plan](../experiments/distributed-command/report.md); storage/shape admission and runtime guards do not establish fullsize host-memory feasibility.

Parameterized command validation **82181 passed**;
[verified evidence](../experiments/distributed-command/runs/command-01/analysis.json).
Controller **2 min54.239 s**, actual command cold **10.261 s** /restart **7.896 s**;
peak host RSS **375.211 /376.066 MiB per worker**, respectively. All arithmetic/
native/checkpoint/final comparisons match the pilot. Fulls18 unrun/unadmitted.


### Host phase inventory and next worker count

[Source-derived model and test budget](../experiments/distributed-host-admission/report.md).
Actual list capacities/root counts/rounded group domains determine named phase
bytes; command also requires caller `host_reserve_bytes` for unmodeled native/
runtime/GC/retained state. This is not measured fullsize peak RSS. At s18/16workers,
prefix Newton includes64GiB/worker base arrays (four arrays including product
output), dense80GiB, plus scratch/reserve. Earlier48GiB is input arrays only.
No new numerical allocation or memory–time policy change.

Eight-worker/two-H200 cold/restart is prepared, requesting8cores/16GiBhost RAM
versus current4cores/8GiB; resource agreement granted;82194passed. This tests worker
ownership/deeper root groups, not eight-GPU throughput. Fullsize host reserve,
node/GPU packing and complete s18 resources remain unvalidated/unadmitted.

Eight-worker validation **82194 passed**;
[verified evidence](../experiments/distributed-host-admission/runs/host-01/analysis.json).
Controller **2 min51.032 s**, cold **15.555 s** /restart **11.050 s**;
logged peak host RSS **374.309 /374.105 MiB per worker**. Largest named phase
**0.538 MiB**, separate from768MiB declared reserve. All native/logical outputs and
host/source/checkpoint checks pass. Fulls18 unrun/unadmitted.


### Larger representative case and earlier recovery boundary

[Plan/implementation](../experiments/distributed-discovery-restart/report.md):
actual s10,m2,t5,N1048576,K146641 on8workers/2H200, with completed discovery,
root-only and prefix8 recovery under the same8core/16GiB budgets. Reference is
retained verified native72839; no fullsize result inferred. Summary logs are an
explicit job option, preserving runtime checks/native-work/full artifact equality
but omitting independent full raw-proof replay for this larger shape. Recovery
from unfinished discovery/subtrees and fullsize reserve/node/storage/time admission
remain. No fulls18 job has been run or admitted.

Representative s10 recovery [82199 receipt](../experiments/distributed-discovery-restart/submission.json); no result/fullsize claim.
