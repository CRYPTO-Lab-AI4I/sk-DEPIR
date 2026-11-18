# Two-node host-staged GPU exchange

Status: **retry 77281 passed** on HPC gnode04/gnode05, one H200 NVL per node.
Scheduler elapsed **34 seconds**; controller elapsed **33.504 seconds** including
builds, fixture generation, checks and profiles. Source revision `6931f6c`.
[Retry receipt](submission-retry-02.json), [raw results](runs/staged-02/results.json),
[verified analysis](runs/staged-02/analysis.json), [accounting](runs/staged-02/accounting.txt),
[Slurm log](slurm-77281.log). Review: SCALING-116.

First job 77241 failed during source snapshot, before build/GPU/MPI; the corrected
build-input selection and retry used the same approved budget.
[Original receipt](submission.json), [failed results](runs/staged-01/results.json),
[failure review](runs/staged-01/failure-review.json), [accounting](runs/staged-01/accounting.txt),
[Slurm log](slurm-77241.log). Historical failure evidence remains unchanged.
This is a component test, not distributed preprocessing/GPU-direct/full-size proof.

[CUDA/MPI probe](probe.cu), [runner](run.py), [rank/profile wrapper](rank.sh),
[completion reviewer](analyze.py), [review tests](test_review.py),
[Go metadata fixture](../../encoding/liftrs-shared/gpu_transport_fixture_test.go).

## Pipeline and independent verification

Each of two distinct HPC nodes has one MPI rank and one visible H200 GPU.
Slurm srun uses the installed pmix_v3 plugin matching OpenMPI ext3x, one GPU/task
and explicit CPU/GPU binding. Each rank verifies its Slurm/MPI identity and
exactly one visible device before allocating. This avoids forwarding a launch
node GPU ordinal onto a different node's allocation; runtime PMIx/GPU binding
still needs validation. Rank wrappers/binaries live in shared cache, not the
launch node's private /tmp.
The source GPU creates deterministic packed s17 extension-field patterns with
bits32/33 and distinct rank/position values, and multiplies them by a nonzero
field constant. The data moves through:

1. GPU producer using the maintained `liftrs_cuda::multiply` arithmetic.
2. Blocking device-to-pinned-host copy.
3. Direct MPI_Alltoall between pinned host buffers, exchanging owned halves.
4. Pinned-host-to-device copy on the receiving rank.
5. GPU multiplication by the inverse constant, checking every expected source
   rank and source coordinate, then synchronization/status readback.

CPU polynomial-basis multiplication independently checks **every producer value**;
complete post-consumer download is compared with every expected rank/index value.
A reverse ownership exchange restores and checks every original value. CPU checks
and reverse exchange are outside forward timing; no hash-only correctness gate.
A host-only exchange on the same pinned buffers is checked before every GPU pass.
No field arithmetic is copied into a new backend: GPU multiplication reuses
[the shared recurrence](../../encoding/liftrs-cuda/transform.cuh), while the CPU
oracle uses [the native polynomial reference](../../encoding/liftrs-cpp/transform.hpp).

The fixture uses existing Go canonical field identifiers/log/power tables,
serialized into a bounded1.5MiB local test format, not a new public cache schema.
The test validates field shape/counts/ranges, including log-index bounds, before
GPU allocation. Public
patterns and arithmetic are not constant-time; this is not secret-data or
production-security validation. Data is uint64 field storage; no new encoded
record layout or output packing claim is made.

This test checks host-staged GPU computation, rank ownership and transport. It
has **no global FFT, root construction, canonical code publication or distributed
checkpoint implementation**. CUDA-aware MPI/GPUDirect is not required or claimed.
Installed Open MPI4.1.8 and CUDA12.9 are optional server-tool requirements only;
ordinary CPU-client/library builds acquire neither dependency.

## Measurements and logging

Normal sizes are1/8/32MiB input per rank, with one warm-up and three timed passes
per size. Memcheck uses32KiB per rank, one warm-up and one measured pass on both
nodes. Start/completion/error records carry UTC time, rank, phase and work count;
rank RSS is checked at those boundaries. Sizes are bounded; an individual stalled
copy/MPI/kernel has unknown remaining work, a command deadline and scheduler
cleanup, not a synthetic progress claim.

Record host-only exchange, GPU producer, D2H, MPI exchange, H2D, receiving-GPU
consumer/status, and complete forward pipeline separately. Forward pipeline
includes synchronization and the4-byte status download. An MPI barrier precedes
its clock; the timing reduction and full CPU/result checks follow it. All seven
reported phase/pipeline times are slowest-rank maxima. Those maxima may come from
different ranks, so do not add phase maxima to derive pipeline wall time. Both
rank records repeat the same reduced values and are not independent samples.
Setup includes allocation/table upload/CPU inverse generation, but first CUDA
context selection happens earlier and remains in command/job duration.

Each node sends **half its input remotely** per forward exchange; local MPI copies
are also performed. Divide input/2 by MPI time for per-node remote-payload rate.
Divide input/2 by pipeline time only for effective pipeline throughput, not link
bandwidth. D2H and H2D each copy the whole local input, plus an untimed full result
download for verification. No overlap or asynchronous benefit is promised.
Pinned staging differs from historical pageable-host test77127, so use the matched
host column here rather than a claimed cross-job speedup.

Native gprof is enabled for the component binary, with separate rank/run prefixes
that cannot collide across node PIDs. Matching binary/profiles and gprof reports
are retained. Profiling/logging overhead is included; gprof covers CPU oracle,
MPI/launch/control work, not GPU instructions or wall-time attribution. Go fixture
CPU profile and its binary are retained even if the phase is too short for useful
samples. There is no unprofiled latency baseline, saturation study, four-node
comparison, topology-tuned transport or selected-fabric instrumentation.

## Approved concrete resource plan

| Resource | Requested/guarded budget |
|---|---|
| HPC allocation |2 eligible GPU nodes; gnode03 excluded; no node pin |
| GPU allocation |1 H200/node,2 total |
| CPU allocation |2 cores/node,4 total;1 rank/node |
| Host requested RAM |4GiB/node,8GiB aggregate |
| Owned device cap |128MiB/card,256MiB aggregate; actual32MiB case about33.5MiB/card |
| Pinned staging |64MiB/rank maximum,128MiB aggregate |
| Rank process RSS guard |2GiB/rank; context/runtime additional to payloads but inside RSS |
| Retained artifacts |128MiB cap;32MiB per-file cap; metadata, binaries, profiles and bounded logs only |
| Build scratch |2GiB cap on launch node; source/build objects, no array workspace |
| Deadlines |280s each MPI command, kill-after10s;720s controller;900s Slurm |

The previous two-node approval covered **zero GPUs and a10-minute scheduler**;
the existing CUDA budget covers one node. The user explicitly approved this
plan's one GPU on each of two nodes and15-minute scheduler ceiling; the
[submission receipt](submission.json) records that agreement. It does not request16GPUs, a fulls18 allocation, large shared
writes or retained preprocessing files. Fixed payload sizes bound allocation;
phase RSS checks, command process-group shutdown and scheduler cutoff remain
active. Artifact/scratch size guards run between commands. A timed-out launcher
gets bounded TERM/KILL cleanup; scheduler cutoff remains the remote cleanup limit.

## Reproduction, provenance and readiness

```sh
python3 -B experiments/gpu-staged-transport/test_review.py
python3 -B experiments/gpu-staged-transport/run.py --check
# Only after approval:
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/gpu-staged-transport/slurm-%j.log \
  experiments/gpu-staged-transport/run.sbatch staged-02
# After completion ping only:
python3 -B experiments/gpu-staged-transport/analyze.py staged-02
```

The runner freezes tracked build inputs from exact Git blobs into bounded scratch,
including source/header/assembly/tool files, Go module files and core JSON
assets. Historical experiment artifacts/reports/run logs/profiles are excluded.
Per-file32MiB and total128MiB snapshot-size checks use Git metadata before reading
contents; the existing per-file/storage limits are not increased. It records the revision,
all copied source identities and generated field/binary/profile/evidence hashes.
Ignored/racing worktree files are not compiled or silently admitted. No source
review override is required. Source snapshot parsing and rank/timing/cap checks
have unit tests. [Original prepared-source readiness](local-checks.json) records source
identities and checks; it is not GPU execution evidence. Source/build scratch is kept on failure for forensics, distinct
from reusable data; no computational exchange array is stored there.

Retry runtime records/logs/profiles go to `runs/staged-02/`; matching binary, field fixture
and native profiles remain outside Git at
`.cache/liftrs/gpu-staged-transport/staged-02/`. No prior caches or profiles are
modified/deleted. Compilation/link succeeds with installed NVCC/MPI; local tool
query emits the known sandbox socket-permission warning. GPU/MPI execution has
not occurred. Four reviewer tests pass, including Git-only snapshot and missing/
same-node/cap/timing rejection. Go fixture compile/skip passes; actual metadata
fixture generation/schema checks also pass (1572880bytes, SHA256
`43ade37b8e8ce2cb8bf1e3a91dc89a320f8f98ba4cf4e08ddc9f935ca626f189`).

After this gate passes, integrate independent normalized-root subsets across
ranks using existing root/product helpers and explicit sorted validation for the
measured sparse cases. Include complete polynomial transfer, joined merge and
CPU/canonical checks in elapsed time before selecting the next node/GPU count.
Buffered persistent output remains separate. No complete s18 runtime estimate,
capacity demonstration or default change follows from preparing this test.

## First-run source snapshot failure and corrected retry (SCALING-115)

77241 FAILED1:0 in5s on HPCgnode04/gnode05. No command record was started: the
runner failed while writing its copied source tree. Its exclusion-based filter
selected a historical37479757-byte JSON experiment artifact outside a `runs`
directory; that exceeded the32MiB per-file guard. The artifact was never an input
to this probe. This is a runner-selection bug, not a measured GPU/MPI failure.

The replacement selects build extensions, go.mod/go.sum and core JSON assets,
explicitly excluding experiment data artifacts. Git object sizes are checked
before reading any content; oversized real source or total build snapshots reject
with a named diagnostic. File, host, device, scratch and scheduler caps remain
unchanged. Regression tests include the historical-sized artifact, required core
JSON, and per-file/total-size rejection before content reads. Four tests pass.

[Actual snapshot readiness](snapshot-check.json) records613 tracked build inputs,
2908958bytes (about2.77MiB), exclusion of the historical artifact, successful
shutdown-helper import, CPU fixture compilation/generation and profiled CUDA
compilation/MPI-cudart linking **from that snapshot**. The snapshot used recorded
Git build inputs26bb006 with the corrected selection implementation identified
by SHA256. No GPU/MPI execution occurred in this local check. This corrects the
original tests, which exercised a synthetic snapshot but not a real build tree.

The failed source snapshot is incomplete and has no completed source seal; the
failure review does not claim otherwise. Original result/Slurm log hashes, Git
artifact size, accounting and traceback are retained; failed scratch/caches are
not deleted. Retry uses fresh `staged-02` names, same approval/resource budget,
same kernels/validation/timing boundaries and public input fixture. Record its
new source revision/job ID; hand off without monitoring after submission.

## Completed GPU exchange review (SCALING-116)

The completion reviewer verified **613 frozen Git source files, 16 evidence
files and 9 retained cache files**, without source overrides. All 14 commands
exited zero. All 24 normal rank measurements (warm-up plus three passes at each
size on both ranks) and four small memcheck measurements passed complete checks:
canonical Go field metadata, independent CPU polynomial arithmetic, receiving-GPU
ownership/inverse, full CPU result comparison, and reverse ownership exchange.
Both sanitizer processes reported zero errors and zero leaked bytes.

Measured environment: two HPC nodes, gnode04 and gnode05, each with one H200 NVL,
2 allocated CPU cores and 4 GiB requested host RAM. CUDA runtime 12.9, driver API
13.0, OpenMPI 4.1.8; the runner records compiler/module details and exact commands.
Peak rank RSS was **421.016 MiB** (rank 0) and **411.207 MiB** (rank 1), from
process high-water marks in normal measurements, not aggregate job memory.
Largest owned device allocation was **33.500 MiB/card**; pinned host buffers were
**64 MiB/rank**. Small memcheck peak rank RSS was **396.801 MiB**. Guards and
allocations stayed within the approved resource plan. Scheduler step MaxRSS is
launcher accounting and must not replace these rank measurements.

Medians below use three timed passes per size, after one warm-up. Each timing
is a slowest-rank maximum, deduplicated across rank records. Units are explicit;
per-phase maxima need not add to pipeline time. These are profiled/instrumented
latencies in fixed size order, not an unprofiled saturation or confidence study.

| Input per rank | Remote payload per node | Host-only exchange | GPU producer | Device→host | Staged MPI exchange | Host→device | GPU consumer | Complete forward pipeline | MPI remote payload per node | Effective pipeline payload per node |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 MiB | 0.5 MiB | 0.0817 ms | 0.0209 ms | 0.1595 ms | 0.1797 ms | 0.1595 ms | 0.0348 ms | 0.4518 ms | 2.717 GiB/s | 1.081 GiB/s |
| 8 MiB | 4 MiB | 0.5849 ms | 0.0418 ms | 1.1857 ms | 1.5642 ms | 1.1812 ms | 0.0641 ms | 3.0172 ms | 2.497 GiB/s | 1.295 GiB/s |
| 32 MiB | 16 MiB | 2.2691 ms | 0.0929 ms | 4.6929 ms | 6.0844 ms | 4.6827 ms | 0.1535 ms | 11.8274 ms | 2.568 GiB/s | 1.321 GiB/s |

Staging, synchronization and exchange dominate this small arithmetic component;
its producer/consumer are field multiplications, not representative root/FFT
workloads. The matched host-only comparison is the valid baseline here. No
full-size multiplier, four-node bandwidth or GPU-direct capability follows.
Forward timing excludes the independent CPU/reference checks, reverse exchange
and extra verification download; the 34-second job duration includes them and
build/setup overhead. See the measurement boundaries above.

Native profiles for both normal ranks and both sanitizer ranks, the matching
binary and metadata fixture remain under
`.cache/liftrs/gpu-staged-transport/staged-02/` (**8.283 MiB** of sealed cache
files, outside Git). [Go CPU profile](runs/staged-02/fixture.pprof) and
[rank 0](runs/staged-02/native-profile-0.log)/[rank 1](runs/staged-02/native-profile-1.log)
gprof summaries are retained. gprof includes independent CPU reference work and
does not measure GPU instructions; short fixture profiling may yield few samples.

The first completion-review invocation exposed a summary-schema mismatch:
the shared reviewer expected command-group RSS, which this producer never
sampled. The reviewer now accepts an explicit set of collected summary fields;
its default remains strict and all source/evidence/cache, exit-status and guard
checks remain active. Raw producer results were not rewritten and no zero RSS
was invented. Command-group RSS remains unknown. The added regression verifies
both strict default rejection and explicit omission; all five review tests pass.

Verification commands: `python3 -B experiments/gpu-staged-transport/test_review.py`
and `python3 -B experiments/gpu-staged-transport/analyze.py staged-02`.
Next: integrate actual normalized-root subsets and complete polynomial
transfer/merge across ranks, retaining the shared helpers and independent
canonical/CPU checks. Buffered persistence remains a separate gate. No new job
or resource expansion was submitted in this review; full s=18 costs remain unknown.
