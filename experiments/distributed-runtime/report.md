# Shared MPI/GPU preprocessing runtime

Status: **passed job81763**, all three runtime-01 arms independently verified. Implementation **c4685d4** (parent830433c).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[configuration regression](test_review.py), [readiness](readiness/).
Planned results `runs/runtime-01/`. Immutable source, binaries, native references
and reusable caches stay outside Git in `.cache/liftrs/distributed-runtime/runtime-01/`.
Implementation/producing revisions are recorded before submission.

## Purpose and current boundary

[Coordinated support verification81750 passed](../distributed-restart-binding/report.md#verified-result81750),
with actual complete checkpoint/frame/native/canonical agreement. It read29,524bytes
once versus236,192bytes/eight hashes for local recovery. Small cached restart was
6.792011seconds coordinated versus6.661931seconds local;local remains default.
The next recorded production step is reusable runtime/lifecycle ownership.

Until this extraction, the validated MPI adapters and pipeline session wiring were
compiled only by tests. The internal runtime now contains:

- [Operand transport](../../encoding/liftrs-shared/distributed_operands_mpi.go),
  [transpose transport](../../encoding/liftrs-shared/distributed_transpose_mpi.go) and
  [Taylor transport](../../encoding/liftrs-shared/distributed_taylor_mpi.go): unchanged
  descriptors, routing, pre/post payload fingerprints and peer comparisons.
- [GPU session](../../encoding/liftrs-shared/distributed_session_cuda_mpi.go): the
  shared existing GPU workspace, operation identities, progress/proof recording,
  dense/prefix products and resident/streamed spectrum operations, returning errors.
- [Storage transfer helper](../../encoding/liftrs-shared/distributed_storage_cuda.go):
  previously available only through a test file even though runtime kernels used it.

[Thin test adapters](../../encoding/liftrs-shared/pipeline_executor_gpu_test.go)
convert returned errors to test failures; arithmetic/transport wiring is no longer
duplicated there. The [bounded wide workflow](../../encoding/liftrs-shared/pipeline_wide_gpu_test.go)
now uses this same session, removing its duplicate workspace/packet/pass setup.
A small observer adapter preserves its established `tail-call` event vocabulary;
the independent native/oracle computations remain in tests. The owned root wrapper
now receives only its domain bound instead of an otherwise unused fixture object.

The internal session owns GPU storage and explicit close, not the MPI communicator.
Session operations reject uninitialized, closed or failed state; an operation error
poisons the session. Collective callers must abort peers on failure. Close is safe
after error and repeated calls. Low-level one-shot subgroup execution retains its
existing close-before-communicator-release order. Ordinary builds do not include
these CUDA/MPI files. MPI-only transports need `liftrs_mpi`;GPU-only builds need
`liftrs_cuda`;the combined session needs both and cgo, without `liftrs_cpp` or a
`testing.T` dependency. The existing experiment-scoped MPI bridge is retained;
this is not a stable public production API or a relocated MPI package.

Admission now checks communicator metadata, field degree/dimension and the existing
32-bit extension slots through s16 /64-bit above s16 before allocation. This follows
the current CUDA implementation, not a new representation. Canonical base
coefficients, packed database output and every successful operation remain unchanged.
No protocol, scalar arithmetic, CUDA kernels, FFT schedule, transfer layout or
checkpoint format is changed. CPU and optional GPU client paths remain available.

## Validation and limitations

Exact moved transport bodies are checked against their prior revision in
[transport provenance](readiness/transport-moves.json). Pure tagged tests exercise
invalid construction and slot-width selection, error propagation/poisoning, rejected
reuse and repeated close without opening CUDA or MPI. Build regressions cover:
ordinary CPU with cgo disabled;CUDA-only with GCC and no MPI feature;MPI-only;
combined CUDA/MPI;and the full independent native test binary.

Runtime compilation exposed the old test-only `storageTransfer` dependency. During
that local move an existing storage file was briefly overwritten;the compiler
caught the absent routines,and the original implementation was restored with only
the helper appended. Final source checks preserve all original storage routines.
These issues were resolved before any GPU allocation. No failed build is reported
as a successful GPU measurement.

One bounded profiled allocation validates genuine s8 cold discovery through output,
fresh coordinated prefix8 restart,and the existing small field18 d5/d12 workflow
through the shared session. All independent native leaves/stages/root/generator/
identity/spectrum/canonical comparisons,packet/digest proof checks,actual native
Taylor counts,checkpoint support frames and diagnostic byte/progress totals remain.
Configuration tests exercise the real arm setup,including clearing restart state
for wide and retaining its explicit bounded codec. Do not repeat the already
measured local-vs-coordinator comparison or unchanged sanitizer arm.

This is a maintainability/runtime extraction,not a speedup claim or production entry.
Timing still includes setup,CPU/GPU work,communication,verification,persistence,
checkpoint/hash/sync/reload and publication. Controller also includes builds,
native export,tests,independent review and profiles. Wide completion events retain
their operation meaning through the observer adapter;per-case runtime rows still precede
deferred GPU closure,as before;launcher/controller times include closure and all validation. Extra progress during long wide operations now
uses the same bounded30-second policy as the main pipeline. Single timings are not
a statistical refactoring-regression study or full-size cost estimate.

Array payloads,retained files and GPU allocation budgets are unchanged. Session
state adds a few flags and bounded progress metadata;no memory/time storage mode is
silently selected. Verified resident/dense/local defaults remain available.
No full s18 preprocessing,larger rank count,non-oracle CLI,production lifecycle or
final retention/admission claim follows from placing these internal pieces in `.go`.
Those remain subsequent integration work;test harness limits stay enforced.

## Resources and reproduction

Same admitted HPC resources: **two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**, four workers/two per card; **64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache/2 GiB scratch/32 MiB
per file**, **720-second controller/15-minute Slurm limit**,eligible nodes excluding
gnode03. Context/driver/MPI memory is additional to owned bytes. No increased budget.
HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200.

Estimate **3–6 minutes excluding queue**,based on prior three-arm controllers.
Cache projected below64 MiB;trace logs below12 MiB per worker. Fresh committed-source
build and executable/library/wrapper/input/cwd checks on every node precede GPU work.
CPU/native/frozen reviewer preflight occurs before submission;profiles and completed
phase outputs are retained even if later review fails. Logging remains timestamped
and bounded with phase/pass identifiers and actual completed work/resource samples.

```sh
python3 -B experiments/distributed-runtime/test_review.py
python3 -B experiments/distributed-runtime/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-runtime/slurm-%j.log \
 experiments/distributed-runtime/run.sbatch runtime-01
# After terminal notice:
python3 -B experiments/distributed-runtime/analyze.py runtime-01
```

After validation,update the main scaling reference and continue parameterized
production ownership/phase admission,non-oracle execution,bounded retention and
restart/publication integration. Full s18 resources remain separately unadmitted.

### Immutable-source readiness

Implementation **c4685d4** is frozen into **872 build inputs**;all **390**
Go/C/C++/CUDA source hashes match the compiled readiness tree. All **17 frozen
reviewer suites** pass. CPU-only native preflight reproduces **108014 historical
reference bytes** and validates actual s8/curve5 leaf/stage/root/canonical inputs.
[Source checks](readiness/source-checks.json), [frozen reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
The admitted allocation rebuilds this committed source and preflights every node.

### Submission81763

Job **81763** submitted from **742dbc7**, implementation **c4685d4**.
[Receipt](submission.json);results `runs/runtime-01/`,Slurm log `slurm-81763.log`.
Unchanged resources above;estimate3–6minutes excluding queue. No polling or GPU result claim.

## Verified result81763

All **872 source /82 evidence /368 cache files** reconcile to producing revision
**742dbc7** and the submitted receipt. [Analysis](runs/runtime-01/analysis.json),
[raw results](runs/runtime-01/results.json), [Slurm log](slurm-81763.log),
[receipt](submission.json). All independent native/canonical/packet/digest checks,
actual Taylor window totals, checkpoint recovery metadata frames and bounded
diagnostic byte/progress checks passed. Optional build variants and pure session
admission/terminal-state tests also passed in the allocation.

| Mode | Result/runtime | Peak host RSS | Owned GPU memory |
|---|---|---|---|
| Genuine s8/curve5 cold | Passed;9.105248 seconds |378.125 MiB/worker |0.244 MiB/worker |
| Fresh coordinated prefix8 restart | Passed;6.962673 seconds |379.320 MiB/worker |0.244 MiB/worker |
| Bounded field18 d5/d12 shared session | Passed;4.758988 seconds cumulative |381.219 MiB/worker |3.125 MiB/worker |

HPC two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM,under
unchanged caps above;CUDA/MPI/context memory additional. Controller **2 min40.959
seconds**;retained cache **34.725 MiB**. Runtime rows precede deferred GPU close;
launcher/controller times include it. These small cases establish shared runtime
integration,not full-size speedup,peak-memory bounds or a public production entry.

Next consolidate sharded generator refinement and full identity verification into
reusable runtime while preserving independent test/native observers, then finish
production ownership/phase admission,non-oracle entry and bounded publication.
Full s18 remains unrun/unadmitted.
