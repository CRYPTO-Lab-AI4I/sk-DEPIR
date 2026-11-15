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
