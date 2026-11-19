# Rank-owned support and GPU polynomial merge tree

Status: **bounded validation job77704 submitted**, results not reviewed (SCALING-121).
Implementation `6c7d28f`. [Submission receipt](submission.json), result/profile
location `runs/tree-01/`, progress `slurm-77704.log`.
[Actual frozen-input check](readiness/snapshot.json):635files,3,022,873bytes;
all11 checked implementation-input hashes match. No GPU/MPI execution is claimed
by this snapshot check.
[Runner](run.py), [scheduler script](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py),
[shared routing](../../encoding/liftrs-shared/root_partitions.go),
[shared tree](../../encoding/liftrs-shared/root_tree.go),
[CPU-only tests](../../encoding/liftrs-shared/root_tree_test.go),
[GPU/MPI test](../../encoding/liftrs-shared/root_tree_gpu_test.go),
[optional MPI binding](../distributed-root/mpi/world.go).
This replaces the replicated test execution of
[77285](../distributed-root/report.md) with uniquely owned shards/subtrees.
No full s=18 or four-GPU speedup is claimed.

## Ownership and shared implementation

The coordinator validates the sorted source with every existing range/order/
Frobenius check, then routes complete orbits to deterministic least-loaded owners.
Assignment/ties match the previous greedy partitioner. Packets contain at most
16,384 uint64 indices (128 KiB); lazy per-owner buffers add at most512 KiB for
four ranks, rather than a second full K-index coordinator partition array.
Non-coordinator ranks receive only their shard, sort it and use the same explicit
sorted-membership root builder. They do not construct a full source support or
CPU reference. The old in-memory partition wrapper uses this same routing logic,
retaining ordinary bitmap defaults and sorted copies. The wrapper also has bounded
packet overhead; historical timings are not relabeled as measurements of new code.

Rank0 still retains the validated full source and CPU oracle in this bounded
experiment. It is not distributed discovery, a streaming source-file decoder or
proof that discovery/validation costs scale. Routing retains the existing O(ranks)
greedy owner search per orbit; measure it before selecting a full workflow.
Each receiver keeps its owned support/root. Empty shards use polynomial1.

The deterministic tree handles1..64 logical owners, including odd counts. Senders
hand their complete normalized polynomial to the next owner and leave the reduction;
receivers merge once per peer. Rank0 alone owns the final product. With four ranks:
1→0 and3→2, then2→0: exactly3 sends and3 products, no duplicated final merge.
The pure-Go schedule/transport interface is independent of field degree, MPI and
GPU kernels. Multiplication delegates to the existing field/product helper.
MPI is opt-in,2..64 ranks, initializing `MPI_THREAD_FUNNELED` on a locked OS thread.
The original two-rank `Peer` operation remains explicitly two-rank; typed packet/
polynomial send/receive methods carry64-bit totals and exact peer/message lengths.
Payloads use bounded chunks, without extra staging copies, so MPI's per-call
32-bit count does not cap the admitted total. Default chunks are2^20 words;
normal validation explicitly uses65,536 words and memcheck1,024 words to exercise
boundaries/tails in both uint64 support and uint32 coefficients. More chunks add
message/progress overhead of unknown cost; these validation settings are not a
production speed recommendation. Native chunk counters flush at30-second intervals
and completion; the reviewer reconciles them with actual transfer extents.
Errors abort the communicator; no rank is left waiting after a local test failure.

GPU maximum elements follow each owner's largest admitted subtree, instead of
reserving a final-product workspace everywhere. Owned-rank bounds are4/1/2/1 for
ranks0/1/2/3; transform maxima are power-of-two ceilings with balanced-shard
headroom. One explicit local-layout executor per rank,64 MiB owned cap each; executors
remain reusable across passes even after a rank hands off polynomial ownership. Caller references/GC timing
can extend array lifetimes; measure actual memory rather than treating logical
ownership as immediate reclamation. None of this makes the last full-size product
fit one GPU: distributed transforms and recruitment/scheduling of idle GPU workers
for large products remain required before full s=18 integration. The callback is
an ownership boundary, not an implemented production worker-service protocol.

## Validation, logging, timing and retention

Inputs reuse existing canonical-field fixtures and seed rules: genuine s=8 support
(t=5,seed109), synthetic closed s=17 subsets of65,536/262,144 exponents with indices
above2^32. Only rank0 computes a complete native CPU polynomial and four independent
extension-field scalar witnesses; it checks every final coefficient. Ordinary CPU/
client builds require no MPI/CUDA, and no production CLI/encoding default changes.

Normal samples are−1 warm-up,0/1 timed and2 artifact-validation. The last pass
retains per-rank support/initial roots, every sent/received polynomial, every owned
merged state, and the sole final/reference. It uses validation-only filesystem
writes; **no transfer or arithmetic reads those files**. Timing medians exclude
that artifact pass. Timed passes include routing, source/local checks, local roots,
blocking tree sends/receives, GPU/host work, metadata/digest logs and full final CPU
comparison. Setup/reference construction and collective maxima are outside pipeline
clocks. Phase maxima can come from different ranks and are not additive. Exchange
and tree clocks include rank skew, not isolated network bandwidth.

The reviewer checks exact disjoint source coverage/orbit closure, per-rank allocation
bounds, lack of non-coordinator full support/reference, every tree step and handoff,
complete sent/received bytes, state-chain identities and final CPU bytes. Small
s=8 repeats warm-up/artifact passes under CUDA memcheck on all four processes;
all must report zero errors/leaks. Pure-Go tests cover odd trees, empty shards,
1/2/3/4/5/7/8/16/64 ranks,64-bit admitted extents without large allocation, failure/bounds, and pre-refactor per-owner golden hashes.
Scheduled race tests exercise the in-memory concurrent ownership network.

Timestamped phase/case/sample/stride/peer records identify active work. Native
CPU/GPU progress stays enabled; packet/root loops emit bounded30-second updates
and completed/total counts. Unknown local routing targets are labeled as bounds,
not ETA/progress percentages. Per-rank process high-water RSS is logged with a
1 GiB guard. Idle MPI waits do not count as increasing work. Profile all four
normal/sanitized processes and retain matching binary/library plus summaries.
CPU profiles include coordinator reference/host work and MPI/native frames; they
are not GPU instruction or pure kernel-time measurements.

This is **four logical ranks sharing two physical H200s**, two contexts/card.
Context scheduling/contention and reduced CPU workers differ from the previous
2-rank experiment. It validates two ownership levels, not four-GPU throughput or
speedup versus77285. There is no fixed two-GPU design ceiling; a one-card-per-rank
or larger allocation requires separate resource admission and actual measurements.

## Same-budget admission and commands

| Resource | Allocation/guard |
|---|---|
| HPC nodes/cards | 2 nodes,1 H200 each; gpuq, excluding gnode03 |
| Logical ranks | 4 total,2 per node/card;1 CPU/rank,2 CPU/node |
| Requested host RAM | 4 GiB/node,8 GiB total |
| Rank RSS / Go target | 1 GiB guard /512 MiB soft target; not total node peak |
| Owned device payload | 64 MiB/rank,at most128 MiB/card; driver/context extra |
| Pinned staging | None added |
| New retained cache / local scratch / file cap | 128 MiB /2 GiB /32 MiB |
| Test / step / controller / Slurm deadlines | 220s /240s /720s /900s |

Counts, cores, host request and card-owned caps stay within the existing approved
2-node/2-GPU envelope; this changes logical ownership, not physical GPU count.
Two contexts/card may increase RSS and runtime; both remain guarded/unmeasured.
It adds messages/merge synchronization, avoids non-owner full support/reference
copies and duplicated final arithmetic; no measured memory/time saving yet.
Source is frozen from tracked Git build blobs. Go1.27.3/GCC14.2/CUDA12.9/OpenMPI4.1.8
modules and exact build commands are recorded by the runner. Slurm25.05's installed
`srun` manual documents `single:2`; the launcher explicitly requests one GPU/node,
two tasks/node, block rank placement and task sharing. It does not request a GPU per task.

```sh
python3 -B experiments/root-merge-tree/test_review.py
python3 -B experiments/root-merge-tree/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/root-merge-tree/slurm-%j.log \
  experiments/root-merge-tree/run.sbatch tree-01
# After user completion ping only:
python3 -B experiments/root-merge-tree/analyze.py tree-01
```

Results/profiles/logs: `runs/tree-01/`, `slurm-JOBID.log`. Reusable public fixture/
shard/root/state files and matching binaries remain outside Git at
`.cache/liftrs/root-merge-tree/tree-01/`, sealed in results.json; regenerate with
same revision/runner and a fresh run name. No existing cache/profile is deleted.
Local small tests/build checks are correctness evidence, not GPU/MPI measurements.

Next after this gate: finish distributed discovery/source ownership, large-product
worker coordination/transforms and generator/Newton/publication integration,
retaining existing atomic/checksum helpers and every verification check. Larger
hardware/I/O budgets and full s=18 costs need separate audit/admission. Keep the
fast validated resident path and CPU clients available; no slower memory-saving
mode is silently selected.

Local readiness: CPU-only routing/tree/partition tests pass, all four pre-refactor
CPU polynomial hashes match, tagged CUDA/MPI binary compiles, and three reviewer
regressions pass. [Checks/source hashes](readiness/checks.json),
[shared-test log](readiness/shared-tests.log), [build log](readiness/build.log),
[reference log](readiness/fixture-regression.log),
[review tests](readiness/reviewer-tests.log). Prior distributed/staged/shared-review
regressions also pass. Login-node OpenMPI compiler wrappers emit sandbox socket
warnings while exiting zero; no local MPI initialization or GPU run is claimed.
