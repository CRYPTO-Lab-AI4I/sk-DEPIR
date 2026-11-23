# Explicit-policy partitioned discovery runtime

Status: **submitted job82097**, loop-01; GPU integration results pending. Implementation **5423bc7** (parent48bbd15).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[configuration regression](test_review.py), [readiness](readiness/).
Planned results `runs/loop-01/`; immutable source/binaries/reusable caches outside
Git: `.cache/liftrs/distributed-discovery-loop/loop-01/`.
Implementation/producing revisions are recorded before submission.

## Runtime scope and correctness

[82096 passed](../distributed-discovery-runtime/report.md#verified-result82096)
shared-session streamed cold/restart/resident validation. The remaining sampling
and ownership loop now moves from tests into
[internal discovery runtime](../../encoding/liftrs-shared/distributed_discovery_runtime.go).
A thin test adapter supplies the existing deterministic seed109 and job limits.
The runtime has no deterministic RNG default or native-reference input.

Callers supply an explicit random reader on rank0, CUDA library/device/owned-memory
budget, result mode and logical limits for the local grid, good-monomial list,
routed root list and optional gathered support archive. Production callers must
select a suitable secure random source; the experiment's math/rand reader remains
only reproducibility policy, not a cryptographic randomness claim. Non-coordinator
workers need no random reader. A streamed archive sink retains bounded transfers;
a nil sink explicitly retains the diagnostic gathered archive subject to its cap.
Ordinary/resident result mode remains default;streaming remains explicit.

Existing monomial enumeration, global-good entropy byte order, coefficient masking,
base-grid/inverse transforms, union/stopping/sample limit, closed-orbit ownership,
root sorting/validation and canonical archive order remain unchanged. Count frames
still bind phase, sequence, rank and bounded integer total; wire frames retain their
exact descriptor and before/after payload fingerprints. New argument checks reject
invalid wire ownership/extent and admission before communication. Independent tests
still review every packet, entropy hash and native/canonical result.

All former test failures become returned errors. GPU close errors propagate;
callers must abort peers on any error and never publish a partial result. The bounded
test wrapper retains the established collective-abort policy. No test-framework,
fixed seed, environment lookup or output-directory policy enters discovery runtime.
Native GoodMonomialRange remains the existing CPU component in the delegated GPU
workflow;this extraction does not claim GPU-only execution or fastest fullsize
monomial enumeration. CUDA/server requirements remain explicit feature tags.

## Admission, memory/time and remaining gates

Grid/list bounds are individual logical lengths, not complete peak memory budgets.
Checks occur before the grid allocation and before good/root/archive appends exceed
limits. Go slice growth/capacity, simultaneous good/grid/bitmap/raw/routed support,
field tables, GC, callback buffers and native/MPI/driver memory are additional.
No whole-process host/device/storage feasibility follows from these checks.
The existing owned-device budget and per-worker RSS/file/cache guards remain.

No new intended arrays, transforms, transfers or retained files are introduced;
file formats, field slots and scalar/CUDA arithmetic remain. Fixed control/data
checks and callback/error propagation do not justify a fullsize speed or memory
claim. Resident/dense/local defaults and explicit memory-saving options remain.
Fullsize latency/GC/RSS are unmeasured.

Root, reciprocal and discovery primitives are now reusable runtime components, but
there is still no complete production command independent of the test framework
and full native fixtures. Parameterized workflow/phase admission, durable ownership,
bounded retention, canonical final publication and representative larger-rank
validation remain. Full s18 is unrun/unadmitted;hardware projections are not support.

## Validation and measurement

Pure tagged tests check missing rank0 randomness, non-coordinator reader omission,
shape/rank/slot/result-mode/device/list limits, count-frame phase/sequence/ownership/
reserved bytes/overflow and wire self/interval rejection without CUDA or MPI work.
Existing discovery sampling/union and session poisoned-state tests remain.

One profiled same-budget allocation runs genuine s8/curve5 streamed cold,
fresh coordinated prefix8 restart and resident-discovery preservation from the
same runtime loop. Review exact entropy/sample/support/union/owner counts, root
stage files, every proof route/digest/native Taylor total, checkpoint metadata,
full identity and canonical output bytes, and diagnostic progress. Both cold modes
must match artifacts. Existing configuration regression exercises the actual suite
and clears inherited restart state for the resident control. No unchanged wide,
dense or sanitizer-only arm is needed.

Per-case rows precede deferred GPU close;launcher/controller includes it. Timing
includes setup, communication, verification/persistence/hash/sync/reload/publication;
controller also adds source builds, native export, CPU tests, independent review and
profiles. One sample is not statistical performance evidence. Existing timestamped
phase/pass/completed-work/resource logs and completed phase persistence remain.
Fullsize logging/admission policy still requires assembled-workflow review.

## Resources and reproduction

Same admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**, four workers/two per card; **64 MiB owned GPU/worker=128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache/2 GiB scratch/32 MiB
per file**, **720-second controller/15-minute Slurm limit**,eligible nodes excluding
gnode03. CUDA context/MPI/driver memory additional. HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,
CUDA12.9/H200. Runtime server build uses liftrs_cpp,liftrs_cuda,liftrs_mpi and cgo;
ordinary CPU client builds remain independent of CUDA/MPI.

Estimate **3–6 minutes excluding queue**, based on preceding three-arm controllers.
Cache projected below64 MiB,logs below12 MiB/rank;no increased resources. Frozen
source/native preflight precedes submission;fresh builds and node-by-node required
library/executable/wrapper/input/cwd/access checks precede GPU work. All reusable
references/caches/profiles remain retained.

```sh
python3 -B experiments/distributed-discovery-loop/test_review.py
python3 -B experiments/distributed-discovery-loop/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-loop/slurm-%j.log \
 experiments/distributed-discovery-loop/run.sbatch loop-01
# After terminal notice:
python3 -B experiments/distributed-discovery-loop/analyze.py loop-01
```

After verification update the primary scaling reference and assemble production
parameter/phase ownership and non-oracle workflow with bounded persistence/final
publication. Admit representative topology and full job resources separately.

### Immutable-source readiness

Implementation **5423bc7** is frozen into **896 build inputs**;
all **398** Go/C/C++/CUDA hashes match the compiled readiness tree.
All **21 frozen reviewer suites** pass. CPU-only native preflight reproduces
**108014 historical reference bytes** plus actual s8/curve5 inputs/root geometry.
[Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
Fresh committed-source build and each-node path/environment preflight remain in the job.

### Submission82097

Job **82097** submitted from **c820151**, implementation **5423bc7**.
[Receipt](submission.json);results `runs/loop-01/`,Slurm log `slurm-82097.log`.
Unchanged resources above;estimate3–6minutes excluding queue. No polling or GPU result claim.
