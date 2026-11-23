# Shared local and distributed root runtime

Status: **submitted job82085**, roots-01; integration results pending. Implementation **07872d5** (parent df409fa).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Planned results `runs/roots-01/`; immutable source, binaries and reusable caches:
`.cache/liftrs/distributed-root-runtime/roots-01/` (outside Git).
Exact implementation and producing revisions are recorded before submission.

## Scope and invariants

[81777 passed](../distributed-reciprocal/report.md#verified-result81777) all four
reciprocal cold/restart/dense/wide arms. The next recorded integration step is
production root ownership and actual-count phase admission.

[Local root runtime](../../encoding/liftrs-shared/distributed_roots.go) borrows
an explicit field/domain and engine, reusing the existing normalized factor/product
algorithm. Empty ownership returns [1]. Fixture methods now delegate to that same
function; genuine main-pipeline root arithmetic no longer requires a fixture object.
[Distributed root runtime](../../encoding/liftrs-shared/distributed_roots_cuda_mpi.go)
reuses the same reduction, subgroup mapping, checked operand transport and GPU
merge executor. It returns errors and accepts an optional diagnostic save observer,
without testing objects, full support or native-reference inputs.

The runtime admits actual leaf lengths and every rounded subgroup domain against
an explicit maximum **single base-coefficient shard** before stage allocation.
Main cold execution also checks these counts before local leaf GPU work. Existing
GPU owned-memory admission and RSS guards remain. Admission logs expose actual
leaf words, largest owned stage shard, cap, domain and stage count for independent
review. This cap is not aggregate process memory: simultaneous arrays, support,
lookup tables, GPU workspace, runtime/GC and storage remain separate admission work.
Metadata-only tests cover s18-sized counts without allocating those arrays.

GPU storage closes before each successful subgroup communicator closes. Any failure
once a subgroup exists logs its stage and aborts the containing world, preserving
the existing collective failure policy. Returned errors before subgroup creation
require the caller to abort peers. Save callbacks are diagnostic observations,
not checkpoint commits. The test adapter preserves every existing filename and
native comparison; no caches or profiles are deleted. No wire codec, CUDA kernel,
field representation, product/reduction algorithm or persistence format changes.
No new public production entry, rank expansion or full s18 support is claimed.

## Memory/time and measurement boundaries

No new intended coefficient arrays, transforms, transfers or stored outputs.
Count admission uses O(ranks) small metadata; diagnostics remain bounded/explicit.
This is runtime consolidation and earlier rejection, not a memory-saving strategy
or speedup claim. Actual whole-process peaks/liveness and fullsize latency remain
unmeasured. Dense/resident/local defaults remain unchanged.

Local tests check actual unequal leaf counts, rounded subgroup shards, rejection
at insufficient caps, 64-bit metadata, empty ownership, missing engines, wrong
field/domain, runtime slot/ownership errors and observer error propagation before
MPI/CUDA initialization. Existing independent native fixtures validate local root
arithmetic. The integration has three profiled arms: genuine s8 cold, fresh
coordinated prefix8 restart, and existing bounded field18 **owned root merges**.
This targets the changed root lifecycle; no repeat of unchanged reciprocal dense
or wide-tail/sanitizer comparisons is needed. Captured stage bytes, route/payload
proofs, native Taylor totals, canonical artifacts, recovery agreement and diagnostic
progress remain independently checked. Actual suite tests clear restart state for
owned-wide and retain the bounded writer explicitly.

Per-case pipeline timings precede deferred GPU close; launcher/controller includes
close, builds, CPU tests, native export, independent review and profile processing.
Root-stage completion includes GPU close and precedes diagnostic save/group close.
Preserve these boundaries; single samples are not statistical speed claims.
All long phases retain existing timestamped pass/stage/progress/resource logging;
completed results/profiles persist before subsequent review.

## Resources and reproduction

Unchanged admitted HPC resources: **two nodes/two H200s/four CPU cores/8 GiB
requested host RAM total**, four workers/two per card; **64 MiB owned GPU/worker =
128 MiB/card**, **1 GiB RSS/worker**, zero CUDA-pinned; **128 MiB retained cache /
2 GiB scratch /32 MiB per file**, **720-second controller /15-minute Slurm limit**,
eligible nodes excluding gnode03. The bounded diagnostic cap gives the caller's
8,388,608-word base-shard admission (32 MiB), not new fullsize permission.
CUDA context/MPI/driver memory is additional. HPC Go1.27.3, GCC14.2/OpenMPI4.1.8,
CUDA12.9/H200; ordinary CPU builds remain independent of CUDA/MPI.

Estimate **3–6 minutes excluding queue**, based on preceding three/four-arm
controllers; cache projected below64 MiB, files below32 MiB. Historical owned-wide
reference data is copied and checksum-verified before execution, with shared-path
checks on every node. Exact source/native preflight precedes submission; allocation
rebuilds committed source and checks executable/wrapper/library/input/cwd/access.
No resource increase, unchanged sanitizer or fulls18 job.

```sh
python3 -B experiments/distributed-root-runtime/test_review.py
python3 -B experiments/distributed-root-runtime/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-root-runtime/slurm-%j.log \
 experiments/distributed-root-runtime/run.sbatch roots-01
# After terminal notice:
python3 -B experiments/distributed-root-runtime/analyze.py roots-01
```

After validation update the main reference and continue production discovery/phase
ownership, non-oracle end-to-end entry, whole-workflow resource admission and bounded
retention/publication. Full s18 remains unrun/unadmitted.

### Immutable-source readiness

Implementation **07872d5** is frozen into **886 build inputs**;
all **396** Go/C/C++/CUDA hashes match the compiled readiness source.
All **19 frozen reviewer suites** pass. CPU-only native preflight reproduced
**108014 historical reference bytes** plus actual s8/curve5 inputs/root geometry.
[Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
Fresh committed-source build and each-node path/environment preflight remain in the job.

### Submission82085

Job **82085** submitted from **ac59ea6**, implementation **07872d5**.
[Receipt](submission.json);results `runs/roots-01/`,Slurm log `slurm-82085.log`.
Unchanged resources above;estimate3–6minutes excluding queue. No polling or GPU result claim.
