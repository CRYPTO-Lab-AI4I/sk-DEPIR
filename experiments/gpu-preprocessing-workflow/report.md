# Persistent complete GPU preprocessing workflow

This experiment validates the [server CLI](../../cmd/pir-preprocess/README.md),
not another transform tile. It uses existing shared discovery, preparation,
identity and serialization APIs; no second implementation of their mathematics.
Full s17 cache/latency and resource admission remain unvalidated.

## Scripts, source and outputs

[Controller](run.py), [Slurm launcher](run.sbatch), [completion analyzer](analyze.py),
[CLI](../../cmd/pir-preprocess/main.go), [atomic artifact I/O](../../cmd/pir-preprocess/artifacts.go),
[publication/bounds tests](../../cmd/pir-preprocess/artifacts_test.go),
[independent reference/restart tests](../../cmd/pir-preprocess/reference_test.go).
[Local check commands/source hashes](local-checks.json) and saved test logs document
short pre-submission checks; full CUDA/race/ASan/vet validation is pending.
Submitted source revision and exact copied file hashes are sealed in each
`runs/<name>/results.json`; binaries/plugin and complete caches/profiles persist in
`.cache/liftrs/gpu-preprocessing-workflow/<name>/`, outside Git. Command logs and
flushed progress/resource events remain beside results. No large output is in Git.

```sh
module load slurm/slurm/25.05 gcc/14.2.0 cuda12.9/toolkit/12.9.1
sbatch --output=experiments/gpu-preprocessing-workflow/slurm-%j.log \
  experiments/gpu-preprocessing-workflow/run.sbatch validation-01
# Only after the user's completion ping:
python3 -B experiments/gpu-preprocessing-workflow/analyze.py validation-01
```

## Complete-path checks and measurement boundaries

One node, four H200s;4CPU,8 GiB requested host RAM,6 GiB sampled process-group RSS,
4 GiB Go soft heap limit,4 GiB aggregate owned GPU cap (1 GiB/card),2 GiB retained
new cache/4 GiB scratch. Worker600 s/controller900 s/Slurm1200 s deadlines. These
reuse the approved bounded four-card resources; no full-size budget is inferred.
Guards sample RSS/storage and emit30-second resource heartbeats labeled unknown
remaining work. Workers independently emit actual phase and completed/total byte/
native work progress. No agent polling/monitoring occurs before completion ping.

Compile CUDA-free and native/CUDA CLI variants. Run publication/partial-retention,
bounds/16-card rejection, specification/source-binding/corruption checks, and a
fresh/root-restart/native complete-code comparison against independent Go at
s4,t3. Race/ASan/vet cover the new command and actual root-restart logic. These
short unit fixtures use temporary directories; their profiles are transient, not
performance measurements. The executable workflow below retains reusable caches
and a unique CPU profile from every invocation, including failed/interrupted runs.

For s8,m2,t5, run peer and local layouts on one and four cards. Each case executes
in separate processes: fresh complete discovery/preparation/cache publication,
verified completed-cache restart without GPU setup, root-checkpoint restart
without discovery, and final readback with a CUDA-free executable. To reproduce
a root-only view while retaining all useful artifacts, rename the first final
`code/` to `code-first/`; do not delete it. Support/root manifests must remain
unchanged, final canonical SHA256 must agree. Compare every GPU result with the
independent-Go s8 canonical fixture retained and hash-verified by job74635; record
its path, exact digest and source revision. This is existing code-only public data,
not a database/key or an unverified downloaded fixture.

A fresh s4,t3 four-card local executable run also passes through CUDA memcheck
with full leak reporting required. Previously validated kernel/protocol behavior
is reused; no new full-domain performance or security claim is made.

CLI events separate setup/discovery/checkpoint publication, preparation/complete
P·G identity, final canonical/generator writes and readback. Fresh preparation
includes root checkpoint write/readback inside its callback; root-restart
preparation does not. Hashing, sync and serialization are included in persistent
workflow timings; they must not replace kernel-only/unprofiled preparation
baselines. CPU profiling/logging stays enabled; its overhead is unquantified.

## Cache and resource consequences

Atomic stage directories publish payload+manifest together after file/directory
sync. Failed hidden partial directories remain for diagnosis and count against
storage guards. Completed stages are never overwritten. An advisory exclusive
lock prevents cooperating writers sharing a cache. Support/root use existing
source-bound codecs. The final P/spectrum and G stream without a second resident
code copy. Every fresh payload is hashed during writing and read back once;
support/root reuse their codec digests instead of hashing writes twice.

A completed-cache restart streams all four files and does not instantiate an
arithmetic backend. A root restart loads support and root arrays, skips their
construction, and still verifies the full polynomial identity. Interrupted
sampling is not certified and is restarted; root restart granularity is the
complete normalized polynomial, not each internal product-tree level. Binary/
plugin SHA256, source revision, field/code identifiers and cache bindings persist.
Manifests are trusted local provenance, not signatures or independent proofs.

Projected fulls17 retained files, using the existing estimated K: support~29 GiB,
root~14 GiB, canonical code~142 GiB, G~50 GiB, total~235 GiB before profiles,
metadata, failed stages and headroom. Fresh execution writes and reads these
payloads; runtime cost and storage throughput are unknown. This explicitly adds
persistence/readback work, not a cost-free memory optimization. Exact K and full
peak concurrent host arrays still need admission. Current2 GiB validation cache
cap prevents this controller from being a fulls17 runner; it only runs fixed s8/s4.
The generic CLI accepts s17 only with separately explicit large array/payload caps.

## Eight or sixteen GPUs

The user authorizes considering8/16 cards if necessary. Current engine and this
command use same-node peer pointers; observed nodes have four H200s. A16-ordinal
list is rejected, and eight cards across two nodes cannot be addressed as one
CUDA context. Cross-node collective/transpose orchestration is still required;
no MPI/NCCL dependency is introduced speculatively. Compare full-workflow benefit
before choosing that implementation, retaining CPU/Go correctness references.

For the384 GiB s17 product-array model, four/eight/sixteen evenly sharded cards
would hold96/48/24 GiB each. Only the four-card same-node path is implemented and
boundedly validated. The larger counts are projections, with unknown communication,
host distribution, replication, staging/headroom and latency. User willingness to
use more GPUs is not a numeric host/storage/deadline admission. No new full job is
submitted until runner validation and a reviewed complete resource plan.

## Bounded submission

Job75194 is submitted, sourceec8b9dc ([submission and exact caps](submission.json)).
Results/logs: `runs/validation-01/`, `slurm-75194.log`; reusable cache/binaries/
profiles: `.cache/liftrs/gpu-preprocessing-workflow/validation-01/`. sbatch returned
successfully. No monitoring/polling/results inspection until user completion ping.
No fulls17 run or increased host/storage/device/time budget is admitted.

## Job75194 failure and bounded retry (SCALING-079)

Job75194 **fails**, not a successful long preprocessing run. [Forensic analysis](runs/validation-01/failure-analysis.json)
reconciles507 source,22 evidence,20 retained-cache identities; preserve [results](runs/validation-01/results.json)
and the original [Slurm log](slurm-75194.log). CUDA-free/native unit tests, Go-oracle
reference/publication/restarts, race, ASan, CUDA builds and vet all pass. The full
one-GPU peer s8 workflow passes fresh, completed-cache, root restart and CUDA-free
readback; both canonical outputs match the independent Go fixture and support/root
manifests remain unchanged. These timings are not new scaling measurements.

The next case, four-GPU peer fresh on **gnode03**, completes setup and exact
monomial classification, then stalls at discovery base-grid's first
`peer-transpose` (65536 elements). No support/root/code stage is published for
that case. Child emits termination at about600 command seconds, but the prior
shared `stop_group` waits without a timeout after SIGKILL. The controller therefore
remains blocked until near1200 seconds. Slurm then reports repeated NVML
`GPU requires reset`; accounting records FAILED/1:0,20min4s. This is GPU failure
plus a shutdown guard defect, not evidence that s8 preparation needs20minutes.
The four-card/local cases and final CUDA memcheck never execute. Whole validation
is incomplete, so no fulls17 readiness claim follows.

[One-time metadata observation](2026-11-15-job-75194-observation.json) records
completed-job accounting and node status. IDLE does not certify GPU health.
Previous four-card validation/calibration passed on gnode02. Logs do not establish
whether this fault originated in application, driver, fabric or hardware; do not
claim a proven node fault. No reset, administrative change or external message is
attempted. Node03's four-card behavior remains unresolved even if a retry passes.

The controller now uses local [bounded shutdown](run.py),5seconds after TERM then
5seconds after KILL, recording unreaped workers and deferring allocation cleanup
to Slurm rather than waiting indefinitely. Scheduler/driver cleanup is outside
this guarantee. [Controller tests](test_controller.py), [saved outcomes](retry-controller-tests.log),
[check record](retry-checks.json) cover a real owned worker and a simulated
unkillable process; CUDA probe compilation/Python/diff checks pass. Original
experiment helpers and historical results are unchanged.

A new [peer probe](peer-probe.cu) runs before expensive Go builds: all12 directional
CUDA runtime peer copies and the same 256x256 transpose kernel, checking every
word, with only512KiB aggregate explicitly owned GPU buffers. It has60seconds,
uses actual advancing device/link logs and performs no protocol/discovery/field
arithmetic or GPU reset. It is diagnostic, not a profiled latency baseline. If it
fails, stop before the full CLI suite. New success analysis requires both probe
stages and all original workflow/sanitizer tests.

Retry `validation-02` uses **gnode02 explicitly**, same4H200/4GiB,4CPU/8GiBhost,
6GiBRSS,2GiBcache/4GiBscratch and600/900/1200s guards. New compiler/preflight logs,
caches/profiles and exact source hashes are separate; original partial data is
retained. Command: `sbatch --nodelist=gnode02 --output=experiments/gpu-preprocessing-workflow/slurm-%j.log experiments/gpu-preprocessing-workflow/run.sbatch validation-02`.
No budget increase or fulls17 job. Submit and hand off without polling; review only
after the next completion ping. Peer probe runtime and retry results remain pending.
