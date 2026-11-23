# Shared discovery GPU session

Status: prepared; integration not yet submitted. Parent revision e8695dd.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[configuration test](test_review.py), [readiness](readiness/).
Planned results `runs/discovery-01/`; immutable source, binaries, profiles and
reusable data outside Git: `.cache/liftrs/distributed-discovery-runtime/discovery-01/`.
Exact implementation/producing revisions are recorded before submission.

## Scope and preserved behavior

[Root runtime82085 passed](../distributed-root-runtime/report.md#verified-result82085).
The next recorded discovery integration removes duplicate GPU session setup from
[the bounded discovery driver](../../encoding/liftrs-shared/distributed_discovery_gpu_test.go).
It now uses [the shared session](../../encoding/liftrs-shared/distributed_session_cuda_mpi.go)
for allocation, transpose/Taylor transport, operation identities, progress, errors
and close. A shared `Discovery` method selects resident results or an explicit tile
consumer through the existing executor. Any failed operation poisons the session;
failed/closed reuse rejects before GPU work. Successful close checks remain visible.

Sample generation and global entropy order, local good-monomial enumeration,
union bitmap and stopping criterion, support/orbit routing, archive order and
canonical checks stay unchanged. The session tracks base/inverse operation starts
and pass zero exactly as before, preserving every per-packet descriptor/fingerprint
and independent route/transform oracle. No new executor-call event is synthesized.
Progress uses the same bounded shared policy with pass/target completion. Native
progress and resource guards remain active. No GPU kernels, transform schedule,
field/slot layout, saved file format or default backend changes.

This is a partial discovery-runtime integration. Sampling/ownership/persistence
orchestration still resides in the bounded test driver, whose fixed seed and native
reference requirements are not a production randomness/API contract. Neither this
extraction nor earlier root/reciprocal runtime creates a full production pipeline.
Complete phase/parameter admission, non-oracle end-to-end entry, bounded retention
and coordinated final publication remain. Full s18 is unrun/unadmitted.

## Memory, runtime and validation

GPU workspace, host result ownership, grid/bitmap storage, transfers and retained
bytes are unchanged. Stream mode still uses a bounded result tile; resident remains
the default when memory permits. There is no new memory/time mode or claimed speedup.
Fullsize latency, compiler liveness/GC and peak RSS remain unmeasured.

Pure tagged tests reject invalid/failed discovery sessions without opening CUDA or
MPI; existing discovery sampling/union tests stay independent. Integration has
three profiled genuine s8/curve5 arms: streamed discovery cold; fresh coordinated
prefix8 restart; resident discovery cold preservation (because both result branches
now use the shared method). All other flags stay fixed. Existing reviewers compare
entropy bytes, sample count, union/support/ownership, every proof route/digest,
native root/generator/full identity/spectrum/canonical bytes, checkpoint agreement
and diagnostic progress. Both cold artifacts must match. Actual suite tests clear
restart state for the resident control. No unchanged wide/root-only/dense/sanitizer
repetition is needed.

Per-case runtime precedes deferred GPU close; launcher/controller includes close.
Timing includes setup, communication, verification, persistence/hash/sync/reload
and canonical publication. Controller also includes builds, native export, CPU
checks, independent review and profile processing. One sample is not a statistical
regression study or fullsize cost estimate. Progress retains timestamped phases,
actual completed work and resource observations; completed results/profiles persist.

## Resources and reproduction

Same admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**, four workers/two per card; **64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS/worker**, zero CUDA-pinned; **128 MiB retained cache/2 GiB scratch/32 MiB
per file**, **720-second controller/15-minute Slurm limit**, eligible nodes excluding
gnode03. CUDA context/MPI/driver memory is additional. Go1.27.3,GCC14.2/OpenMPI4.1.8,
CUDA12.9/H200. Ordinary CPU builds require no CUDA/MPI.

Estimate **3–6 minutes excluding queue** based on previous three-arm controllers;
cache below64 MiB and logs below12 MiB per rank projected. No new resource budget.
Exact frozen-source/native-reference preflight precedes submission; fresh builds
and executable/wrapper/library/input/cwd/access checks on every node precede GPU
phases. Existing native sources/caches/profiles remain intact.

```sh
python3 -B experiments/distributed-discovery-runtime/test_review.py
python3 -B experiments/distributed-discovery-runtime/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-runtime/slurm-%j.log \
 experiments/distributed-discovery-runtime/run.sbatch discovery-01
# After terminal notice:
python3 -B experiments/distributed-discovery-runtime/analyze.py discovery-01
```

After verification update the primary reference and continue production discovery
sampling/ownership and phase admission, non-oracle end-to-end entry and bounded
retention/publication. Fulls18 resources remain separately unadmitted.
