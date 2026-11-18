# Two-node host-staged GPU exchange

Status: compiled/prepared, **not submitted**; explicit two-node GPU budget
agreement pending. Work SCALING-113. This is a communication/computation component,
not distributed Lift-RS preprocessing, GPU-direct support or fulls18 performance.

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

## Concrete resource plan (requires agreement)

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
the existing CUDA budget covers one node. This plan adds one GPU on each of two
nodes and a15-minute scheduler ceiling, so AGENTS.md budget agreement is required
before submission. It does not request16GPUs, a fulls18 allocation, large shared
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
  experiments/gpu-staged-transport/run.sbatch staged-01
# After completion ping only:
python3 -B experiments/gpu-staged-transport/analyze.py staged-01
```

The runner freezes tracked build inputs from exact Git blobs into bounded scratch,
excluding docs/message-board/historical run logs/profiles. It records the revision,
all copied source identities and generated field/binary/profile/evidence hashes.
Ignored/racing worktree files are not compiled or silently admitted. No source
review override is required. Source snapshot parsing and rank/timing/cap checks
have unit tests. [Local readiness evidence](local-checks.json) records source
identities and checks; it is not GPU execution evidence. Source/build scratch is kept on failure for forensics, distinct
from reusable data; no computational exchange array is stored there.

Runtime records/logs/profiles go to `runs/staged-01/`; matching binary, field fixture
and native profiles remain outside Git at
`.cache/liftrs/gpu-staged-transport/staged-01/`. No prior caches or profiles are
modified/deleted. Compilation/link succeeds with installed NVCC/MPI; local tool
query emits the known sandbox socket-permission warning. GPU/MPI execution has
not occurred. Three reviewer tests pass, including Git-only snapshot and missing/
same-node/cap/timing rejection. Go fixture compile/skip passes; actual metadata
fixture generation/schema checks also pass (1572880bytes, SHA256
`43ade37b8e8ce2cb8bf1e3a91dc89a320f8f98ba4cf4e08ddc9f935ca626f189`).

After this gate passes, integrate independent normalized-root subsets across
ranks using existing root/product helpers and explicit sorted validation for the
measured sparse cases. Include complete polynomial transfer, joined merge and
CPU/canonical checks in elapsed time before selecting the next node/GPU count.
Buffered persistent output remains separate. No complete s18 runtime estimate,
capacity demonstration or default change follows from preparing this test.
