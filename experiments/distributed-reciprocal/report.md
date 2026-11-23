# Shared sharded reciprocal orchestration

Status: **passed job81777**, all four reciprocal-01 arms independently verified. Implementation **46404c6** (parent9922a8a).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[arm regression](test_review.py), [readiness](readiness/).
Planned results `runs/reciprocal-01/`; immutable source, binaries, native references
and reusable cache remain outside Git at
`.cache/liftrs/distributed-reciprocal/reciprocal-01/`.
Implementation/producing revisions are recorded before submission.

## Runtime scope

[Shared runtime81763 passed](../distributed-runtime/report.md#verified-result81763)
all cold/restart/bounded-wide independent checks. The next recorded integration
step is generator construction without test-framework or oracle dependencies.

[Shared reciprocal orchestration](../../encoding/liftrs-shared/distributed_reciprocal.go)
now runs the existing characteristic-two Newton schedule for both genuine code
construction and the bounded wide adapter. The plan describes domain, actual root
coefficient count, target/known inverse prefixes, rank ownership and an admitted
single-shard word limit. It uses supplied square/product/checkpoint operations;
no native reference, GPU type, testing object or field convention enters the loop.

For normalized P(0)=1 and an initial inverse prefix G, the existing refinement
G_new=P*G^2 doubles the correct prefix. The shared loop keeps the existing source
counts, nonaliasing bounds, dense mask/prefix view choice, operation names and
checkpoint/reload order. Backend upload counters and root-clone inventories remain
actual callback measurements. Completed product counters do not certify that the
following durable checkpoint committed, or count partial failed transfers.
The caller retains kernel field/padding/ownership validation, transport proofs,
actual persistence and resource guards. Returned errors propagate; failed results
are not published as a completed generator.

The same runtime file also exposes an internal full sharded identity checker for
P*G=1+X^(N-1), moved from the main test's identical scan. This checks the complete
result, not merely the inverse prefix. Main code construction still performs its
existing full product and mandatory identity check. The wide synthetic polynomial
retains independent coefficient recurrence and full convolution checks; it is not
silently treated as a genuine full code. Native/canonical observers remain outside
the runtime. Discovery, provenance, full phase admission and prepared-code publication
are not established by refinement or this algebraic identity alone.

## Memory, time and default behavior

No new transforms, square/product transfers, stored files or intentionally retained
coefficient arrays are introduced. Dense mode still allocates and masks one full
root clone per step; explicit prefix mode borrows the unchanged logical root prefix.
Dense remains the default. Square/product workspaces, checkpoint/reload allocation
and final identity scan are unchanged. The single-shard limit is not a peak-RSS
bound: Go compiler liveness, GC, callback/native buffers and overlapping arrays
remain relevant. Full-size memory/runtime after refactoring is unmeasured.

This is code consolidation, not a speedup claim or a new memory-saving mode. Actual
resource/latency observations must be reported before changing capacity guidance.
Neither callback injection nor placing the loop in runtime creates a production
entry point, permits larger parameters/ranks or admits a full s18 allocation.

## Validation and measurement

CPU references independently check 1/(1+X) for cold, warm and already-complete
prefixes in both modes, including non-power-of-two final targets, immutable source
views, dense-tail masking, clone and reported upload counters. Failures from square,
product and checkpoint operations stop refinement; wrong callback extents and
invalid ownership/shape/prefix limits reject. Full-identity tests reject corruption
in every rank's shard and incorrect lengths. Existing scalar/native tests remain.

One allocation contains four profiled arms:

1. Genuine s8/curve5 prefix-input cold pipeline.
2. Fresh coordinated prefix8 restart from that verified cold cache.
3. Genuine s8 **dense-input cold preservation control**, because that branch moved too.
4. Existing bounded field18 d5/d12 workflow using the same reciprocal loop.

Both input modes must match every native root/stage/prefix/generator/identity/
spectrum/canonical comparison and checkpoint artifact. Existing independent input
upload/clone accounting, packet/digest reviews, restart support frames, native Taylor
window totals and diagnostic byte/progress checks remain. The actual suite regression
resets restart state and explicitly restores prefix mode after dense. No unchanged
sanitizer arm or repeated local-vs-coordinator comparison is added.

Per-case runtime rows precede deferred GPU closure, as before; launcher/controller
time includes close. Timings include verification/persistence/publication through
those boundaries. Controller additionally includes builds, native export, CPU/race
checks, independent review and profile processing. One sample is not a statistical
refactoring-regression study or full-size estimate. Completed phases/profiles persist
before later review. Logs retain the existing timestamped phase/pass identifiers,
completed work and periodic resource observations; logging policy is unchanged.

## Resources and reproduction

Unchanged admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB
requested host RAM total**, four workers/two per card; **64 MiB owned GPU/worker =
128 MiB/card**, **1 GiB RSS/worker**, zero CUDA-pinned; **128 MiB retained cache /
2 GiB scratch /32 MiB per file**, **720-second controller /15-minute Slurm limit**,
eligible nodes excluding gnode03. Context/MPI/driver memory is additional.
HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9/H200. Ordinary Go builds need no CUDA/MPI.

Estimate **3–6 minutes excluding queue**, based on prior three/four-arm controllers;
cache projected below64 MiB, trace files below12 MiB per rank. The dense preservation
arm adds about10seconds measured GPU workflow plus review/profile work, comfortably
inside the existing deadline. No increased resource budget is requested. Committed
source/native-reference preflight precedes submission; fresh builds and required
executable/wrapper/library/input/cwd/access checks run on every node before GPU phases.

```sh
python3 -B experiments/distributed-reciprocal/test_review.py
python3 -B experiments/distributed-reciprocal/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-reciprocal/slurm-%j.log \
 experiments/distributed-reciprocal/run.sbatch reciprocal-01
# After terminal notice:
python3 -B experiments/distributed-reciprocal/analyze.py reciprocal-01
```

After validation update the main scaling reference and continue production discovery/
root/phase ownership and parameter admission, non-oracle end-to-end entry, bounded
retention and coordinated final publication. Full s18 remains unrun/unadmitted.

### Immutable-source readiness

Implementation **46404c6** is frozen into **878 build inputs**;all **392**
Go/C/C++/CUDA source hashes match the compiled readiness tree. All **18 frozen
reviewer suites** pass. CPU-only native preflight reproduces **108014 historical
reference bytes** and validates actual s8/curve5 root/leaf/stage/canonical inputs.
[Source checks](readiness/source-checks.json), [frozen reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
The admitted allocation rebuilds committed source and checks every node before GPU phases.

### Submission81777

Job **81777** submitted from **b808243**, implementation **46404c6**.
[Receipt](submission.json);results `runs/reciprocal-01/`,Slurm log `slurm-81777.log`.
Unchanged resources above;estimate3–6minutes excluding queue. No polling or GPU result claim.

## Verified result81777

Receipt/revision and **878 source /96 evidence /496 cache files** reconcile to
producing revision **b808243**. [Analysis](runs/reciprocal-01/analysis.json),
[raw results](runs/reciprocal-01/results.json), [Slurm log](slurm-81777.log).
All independent native/prefix/full-identity/proof/canonical checks, upload/clone
counters, checkpoint/reload sequence, restart frames and diagnostic totals passed.

| Mode | Runtime | Peak host RSS |
|---|---|---|
| Genuine s8 prefix cold |10.020557 seconds |377.828 MiB/worker |
| Fresh coordinated prefix8 restart |6.759730 seconds |378.211 MiB/worker |
| Dense cold preservation control |9.592736 seconds |379.574 MiB/worker |
| Bounded field18 d5/d12 |4.999263 seconds cumulative |381.273 MiB/worker |

Same admitted HPC two nodes/two H200s/four workers/four cores/8 GiB requested RAM.
Owned GPU0.244 MiB/worker narrow,3.125 MiB/worker wide;context/MPI/driver additional.
Controller **3 min4.002 seconds**, cache **37.760 MiB**. All modes passed;one paired
sample is not fullsize speedup evidence. Default dense remains. Runtime rows precede
deferred GPU close;launcher/controller includes it.

Next extract local root construction and owned subgroup lifecycle into reusable
runtime with actual-count shard admission, retaining native observers. Production
discovery/phase ownership, non-oracle entry, complete resource admission and final
retention/publication remain. Full s18 remains unrun/unadmitted.
