# Reference-independent distributed preprocessing workflow

Status: prepared; GPU integration not yet submitted. Implementation **9a2edd5** (parent9af53a7).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Planned results `runs/workflow-01/`; immutable source, binaries and reusable data
outside Git: `.cache/liftrs/distributed-workflow/workflow-01/`.
Implementation/producing revisions are recorded before submission.

## Assembled runtime and boundaries

[Discovery runtime82097 passed](../distributed-discovery-loop/report.md#verified-result82097)
all streamed cold/restart/resident checks. The next recorded integration is a
complete mathematical workflow without native reference inputs.

[prepareRankWorkflow](../../encoding/liftrs-shared/distributed_workflow_cuda_mpi.go)
composes the existing runtime discovery, local normalized root construction,
owned subgroup merge, full-domain root redistribution, Newton reciprocal,
complete P*G=1+X^(N-1) verification, spectrum transform and canonical section writers.
It has no testing framework, native reference path, fixed seed or environment policy.
The test wrapper supplies those experiment policies outside the runtime. Discovery
and native GoodMonomialRange still use the measured CPU/GPU mix;no GPU-only claim.

Callers supply code parameters, explicit discovery RNG/device/list policy, a local
shard cap, dense/prefix input and resident/streamed spectrum selection. Dense and
resident remain defaults; this validation explicitly selects the previously
validated prefix/stream options. Local arrays and actual rounded root domains
are admitted before their relevant phases;existing device-cap/field/ownership/
transport/full-identity checks remain. Single-array/list admission is not total
host/device/storage feasibility, GC/liveness bounds or authorization to run s18.

Archive/output writers and optional support/root/identity observers and checkpoint
callbacks belong to the caller. The runtime does not create, sync, close, rename
or publish caller files. A checkpoint callback may return a verified reload;
shape is rechecked,while integrity/durable coordinated commit is the callback's
responsibility. The test observer writes/syncs compact intermediate arrays without
claiming a production checkpoint manifest or restart protocol. Final code contains
the same header/root/spectrum bytes; generator and raw support archive remain
separate sinks. A full production support archive/manifest policy is not supplied
by the raw test archive. Partial output must never be treated as committed.

Collective callers must abort peers on any error. Engines close between discovery,
leaf construction, root merge groups and final workspace;the final workspace closes
before successful workflow return/barrier. Error returns and closed/failed session
checks remain;no arithmetic oracle is needed by the workflow. Independent test/native
observers remain mandatory for this bounded validation. No public CLI/API contract,
production restart or fullsize parameter support is established here.

## Memory, time and retention

The workflow reuses the same kernels, operation identities, transform sequence,
square/product inputs and canonical writers. Spectrum transformation still precedes
root/generator/spectrum publication as in the reviewed pilot. No additional intended
transforms, transfers, persistent copies or coefficient arrays are introduced.
Compiler/GC lifetime differences after composition can change peak RSS;large-case
latency and memory are unmeasured. Diagnostics are explicit observer policy rather
than built-in runtime files;this experiment retains all comparison arrays/checkpoints,
CPU profiles and raw evidence. No useful historical data is deleted.

The assembled arm omits test-only native reads and coordinated checkpoint manifests;
it has real observer writes but is **not timing-equivalent** to the existing pilot.
Do not interpret a timing difference as a backend speedup or cost-free replacement
for durable persistence. All numerical checks and transport proofs remain. Future
production retention/restart/atomic publication policy must be measured separately.

## Validation

CPU-only tests reject missing sinks and insufficient shard admission before any
CUDA/MPI call. Existing discovery/root/session/reciprocal correctness/error tests
remain. Reviewer regressions reject missing identity, wrong sample/work counts and
inherited restart settings. A local replay using verified82097cold files and an
explicit synthetic workflow-metric fixture checks reviewer plumbing only;it is
not evidence that the assembled GPU arm ran or passed.

One profiled allocation contains three arms:

1. Existing independently reviewed genuine s8/curve5 cold pilot.
2. Its fresh coordinated prefix8 restart, preserving prior checks/composition.
3. A separate **assembled runtime cold workflow**,whose API receives no reference.

The controller independently checks assembled discovery entropy/sample/union/
support/owner routing, native root leaf/stage/final bytes, every compact root/Newton
checkpoint and full identity shard against the fully reviewed pilot. Ordered per-kind
proof digests/counts/bytes must match the pilot exactly (same transport arithmetic,
ordering and operation identities). Actual native Taylor totals must match plans.
The final code/generator must be byte-identical to both pilot and independent native
export. Raw archive indices must equal sorted complete support. Both root/subgroup
and final metadata/work totals are checked. Ordinary diagnostic progress/file sizes
remain required. The native-count reviewer recognizes the actual `workflow-result`
domain event as well as pilot/wide events;it does not invent a domain estimate.

This reuses one reviewed baseline,not a new separate calibration or unchanged
sanitizer/wide/dense comparison. Complete phase results/profiles persist before
later review. Existing timestamped phase/pass/completed-work/resource logging
remains,with bounded local-root progress in the assembled wrapper.

Pilot per-case timing precedes deferred GPU close. Assembled timing includes runtime
close and test observer fsync,while launcher/controller includes process shutdown.
Both include their actual verification/communication/I/O through those boundaries.
Controller also includes builds, native export, CPU tests, independent review and
profiles. Different checkpoint/native-observer costs preclude a direct speedup
claim;single samples are not fullsize performance projections.

## Resources and reproduction

Unchanged admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB
requested host RAM total**,four workers/two per card; **64 MiB owned GPU/worker=
128 MiB/card**, **1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache/
2 GiB scratch/32 MiB per file**, **720-second controller/15-minute Slurm limit**,
eligible nodes excluding gnode03. Context/driver/MPI memory additional.
HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Server runtime uses feature tags;
ordinary CPU client builds remain CUDA/MPI-free.

Estimate **3–6 minutes excluding queue**,based on preceding three-arm controllers.
Assembled workload uses the same measured s8 shape/work counts;cache projected
below64 MiB,logs below12 MiB/rank. No larger resources/fullsize job requested.
Frozen-source/native preflight precedes submission;fresh builds and each-node
executable/wrapper/library/input/cwd checks precede GPU phases.

```sh
python3 -B experiments/distributed-workflow/test_review.py
python3 -B experiments/distributed-workflow/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-workflow/slurm-%j.log \
 experiments/distributed-workflow/run.sbatch workflow-01
# After terminal notice:
python3 -B experiments/distributed-workflow/analyze.py workflow-01
```

After verification update the maintained reference and integrate real production
checkpoint/restart/atomic publication and parameterized command policy,then admit
representative larger topology and complete s18 resources. Full s18 remains unrun.

### Immutable-source readiness

Implementation **9a2edd5** is frozen into **902 build inputs**;
all **400** Go/C/C++/CUDA hashes match compiled readiness source.
All **22 frozen reviewer suites** pass. CPU-only native preflight reproduces
**108014 historical reference bytes** plus actual s8/curve5 inputs/root geometry.
[Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
Fresh committed-source build and each-node path/environment preflight remain in the job.
