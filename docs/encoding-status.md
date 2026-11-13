# Encoding scaling reference

The performance goal is **reasonable encoding performance through s=18 inclusive**
on this HPC system, retaining CPU and optional GPU paths. **s=19 and s=20 remain
exploratory targets**: keep their rows as algorithms, layouts and measurements
improve. This is the maintained capacity and results reference for the
[scaling plan](encoding-scaling.md), separate from TRL 5 acceptance.

Updated 2026-11-13; implementation checkpoint `026bbf1`. Standalone mathematical
encoding is validated through s=16. **s≥17 is not implemented**; wider arithmetic,
preprocessing, layouts and any multi-GPU execution must be implemented and
validated before those projections become demonstrated capabilities. “Reasonable”
has no agreed numerical latency threshold yet; select it using measured setup,
encoding, memory, I/O and deployment costs before accepting a scaling milestone.

## Reading the tables

All rows use m=2, t=5 and one field-element database. **R** means run on this HPC
system, **L** means a historical laptop run, and **N** means not run. `~` marks a
principal-array estimate, not measured peak memory. Memory is GiB unless stated;
runtimes are seconds. `?` means unknown. Each mode column gives **memory; time**.
Preparation excludes discovery, cache loading/saving and database encoding.
Ordinary encoding timings exclude output packing/storage unless marked **stored**;
those stored timings include packing, writing and hashing.

Measured CPU memory is sampled process RSS, except explicitly labeled small-case
array payloads. GPU memory is **owned device allocations / host RAM**; context,
driver and staging overhead are extra. Estimates exclude runtime overhead and
filesystem cache. Different revisions, core counts and machines mean these are
reference results, not controlled speedup comparisons. Detailed methodology,
profiles, source identities, scripts and results remain in the linked reports.

## Machine capacity and feasibility

Observed Slurm inventory: 17 GPU nodes, each with 128 CPU cores and 928610 MiB
scheduler RAM (about **907 GiB**); 16 nodes advertise four H200 GPUs and one three,
**67 GPUs total**. The measured H200 NVL has 143771 MiB (**140.40 GiB**) device RAM.
Inventory was read with `module load slurm/slurm/25.05` and
`sinfo -p gpuq -N -o '%N %c %m %G'`. Inventory is not a reservation or an experiment
budget; review current availability and agree budgets before larger submissions.

GPU counts below are **capacity floors**, dividing aggregate array bytes by one
card's memory and rounding up. They assume sharded arrays/constants rather than
replication. Communication, staging, runtime and allocation headroom can require
more GPUs. Multi-GPU execution is not implemented, and time cannot be estimated
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

## Preprocessing and CPU encoding

Reuse aliases buffers after their last use. Compact bounds transform scratch.
File streams code constants; packed file also writes packed output directly.
Disk keeps transform work arrays in files but still returns an owning output.
These are distinct modes; their savings cannot simply be added together.

| s | Preprocessing: host RAM; preparation time | Original CPU | Reuse CPU | Compact CPU | File CPU | Packed file CPU, stored | Disk CPU |
|---|---|---|---|---|---|---|---|
| 10 | R ≤0.24 whole-command RSS; CPU 6.65 / GPU 1.53 | R ~17.7 MiB arrays; 0.308–0.310 | R ~13.7 MiB arrays; 0.313–0.319 | N ~13.1 MiB; ? | R ~8.6 MiB arrays; 0.311–0.332 | R ~9.6 MiB arrays; 0.325–0.327 | R ~5.7 MiB private +8 MiB files; 2.860 first /1.136 reused |
| 14 | L 13.73; 2018 | R 5.594; 91.164 | R 3.595; 58.667–58.853 | R 2.526; 64.517–68.413 | R 2.214; 59.357–59.377 | N ~2.20; ? | N ~1.19 private +2 files; ? |
| 15 | L 20.55; 11767† | R 18.465; 655.937 | N ~14.42; ? | N ~9.87; ? | N ~8.81; ? | N ~8.82; ? | N ~4.81 private +8 files; ? |
| 16 | R 72.575; GPU 8415.469 | R 74.370; 1252.503 | R 58.351; 1247.173–1281.328 | R 39.409; 1289.759 | R 35.484; 1249.616 | N ~35.43; ? | N ~19.43 private +32 files; ? |
| 17 | N ≥~220.84 retained arrays, peak ?; ? | N ~491; ? | N ~427; ? | N ~287; ? | N ~270; ? | N ~270; ? | N ~78.4 private +256 files; ? |
| 18 | N peak ?; ? | N ≥1024 transform arrays; ? | N ≥1024 transform arrays; ? | N ≥1024 transform arrays; ? | N ≥1024 transform arrays; ? | N ≥1024 transform arrays; ? | N ~317 private +1024 files; ? |
| 19 | N peak ?; ? | N output alone 1024; ? | N output alone 1024; ? | N output alone 1024; ? | N output alone 1024; ? | N ≥4096 transform arrays; ? | N ~1279 private +4096 files; ? |
| 20 | N peak ?; ? | N output alone 4096; ? | N output alone 4096; ? | N output alone 4096; ? | N output alone 4096; ? | N ≥16384 transform arrays; ? | N ~5162 private +16384 files; ? |

† Historical s15 preparation has pause/clock complications; it is not a clean
latency baseline or an HPC forecast. Discovery was additional: about 1.1 s at
s10, 255 s at s14, 1131 s at s15, and 10597.545 s at s16. The s16 CPU-only
preparation attempt was stopped after about nine hours; it did not complete.
GPU-assisted preparation completed with a 48 GiB device-array model under a
64 GiB device cap; 48 GiB is not a measured device peak. s10 GPU preparation used
one GPU under a 512 MiB cap; its device peak was not isolated.

CPU measurements use 16 cores at s10/s15 and 32 at s14/s16. Small-case payloads
omit runtime/metadata. File-backed RSS does not include all filesystem cache:
s16 file CPU reached **61.820 GiB RSS plus tracked cache** across the workflow,
versus 35.484 GiB peak RSS. Disk-backed latency can vary substantially with
filesystem/cache state; its full encoder is validated only at s10. File and disk
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

Each entry gives **device / host memory; encoding time; GPU count**.
Stream uses the fast computation layout while streaming code constants. Packed
stream also writes packed output directly. Host-backed keeps full arithmetic
arrays in host RAM and transfers tiles; it is a slower memory alternative.

| s | Original GPU | Reuse GPU | Compact GPU | Stream GPU | Packed stream GPU, stored | Host-backed GPU |
|---|---|---|---|---|---|---|
| 10 | R 16.6/9.1 MiB arrays; 0.104; 1 | R 12.6/9.1 MiB arrays; 0.103; 1 | R 10.6/9.1 MiB arrays; 0.108–0.109; 1 | N ~8/9.1 MiB; ?; 1 | N ~8/6.4 MiB; ?; 1 | R 0.534/13.4 MiB arrays; 1.606–1.617; 1 |
| 14 | N ~4.19/2.38; ?; 1 | R 3.190/~2.38; 5.176–5.179; 1 | N ~2.19/2.38; ?; 1 | R 2.000/~2.38; 5.467–5.527; 1 | R 2.000/~1.39; 5.804–5.827; 1 | N <2 MiB/~3.38 GiB; ?; 1 |
| 15 | N ~16.81/9.62; ?; 1 | N ~12.81/9.62; ?; 1 | N ~8.81/9.62; ?; 1 | N ~8/9.62; ?; 1 | N ~8/5.62; ?; 1 | N <2 MiB/~13.62 GiB; ?; 1 |
| 16 | R 67.423/39.024; 60.216–85.954; 1 | R 51.427/39.031; 56.975–57.418; 1 | R 35.429/39.028; 93.332–101.526; 1 | R 32.005/39.029; 58.851–63.487; 1 | R 32.005/23.014; 65.485–66.227; 1 | N <2 MiB/~54.85 GiB; ?; 1 |
| 17 | N ~526.42/220.84; ?; 4 | N ~398.42/220.84; ?; 3 | N ~270.42/220.84; ?; 2 | N ~256/220.84; ?; 2 | N ~256/156.84; ?; 2 | N few MiB/~348.84 GiB; ?; 1 |
| 18 | N ~2109/889; ?; 16 | N ~1597/889; ?; 12 | N ~1085/889; ?; 8 | N ~1024/889; ?; 8 | N ~1024/633; ?; 8 | N few MiB/~1401 GiB; ?; 1 device, host exceeds node |
| 19 | N ~8447/3581; ?; 61 | N ~6399/3581; ?; 46 | N ~4351/3581; ?; 31 | N ~4096/3581; ?; 30 | N ~4096/2557; ?; 30 | N few MiB/~5629 GiB; ?; 1 device, host exceeds node |
| 20 | N ~33834/14420; ?; 241 | N ~33834/14420; ?; 241* | N ~25642/14420; ?; 183* | N ~24576/14420; ?; 176* | N ~24576/10324; ?; 176* | N few MiB/~30804 GiB; ?; 1 device, host exceeds node |

Host estimates in multi-GPU rows are **aggregate**, not per node; distributing
arrays or avoiding replication requires an explicit host as well as device design.
At s18 the ~889 GiB host model leaves little room on one node; distribute it or
reduce it rather than assuming the payload fits with runtime headroom.

* At s20, the estimated K makes the prefix workspace exceed the current N-sized
arena. These entries allow an expanded arena; the current modes reject that
shape. A redesign could change the counts. Host-backed s20 likewise needs a
larger arena, included in its host estimate. All s≥17 rows assume a future port.

The packed s16 host peak is encoding-only; verification reaches 31.014 GiB RSS,
and RSS plus tracked cache reaches 60.395 GiB. Host-backed full encoding is
validated at s10 only; tiny tiles substantially increase transfers and runtime.
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

All entries are **not run** and exclude device/runtime headroom. They model three
resident extension-field arrays, not a distributed implementation. A streamed or
external-memory preprocessing design may change these requirements. Preparation
latency and peak host memory remain unknown at s≥17.

| s | Aggregate device arrays | GPU capacity floor | Host/machine feasibility |
|---|---|---|---|
| 17 | ~384 GiB | 3 | Retained host arrays ≥~221 GiB; peak unknown; potential single-node route after port |
| 18 | ~1536 GiB | 11 | At least 3 four-GPU nodes for this layout; host distribution and peak unknown |
| 19 | ~6144 GiB | 44 | At least 11 four-GPU nodes; host distribution and peak unknown |
| 20 | ~24576 GiB | 176 | Exceeds observed 67-GPU inventory; needs a different layout or more hardware |

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
