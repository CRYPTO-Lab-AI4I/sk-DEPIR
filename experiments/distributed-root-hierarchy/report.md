# Hierarchical distributed root products

Status: **submitted as79376**, `hierarchy-01`, implementation `6785af2`.
[Receipt](submission.json), [scheduler log](slurm-79376.log). Integrated GPU results
pending; the controller records the actual producing revision and identities.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/checks.json).
Results will be retained in `runs/hierarchy-01/`; snapshots,binaries,shards and
profiles in `.cache/liftrs/distributed-root-hierarchy/hierarchy-01/` outside Git.
Follows [four-worker collective validation79239](../distributed-four-worker/report.md).

## One tree,shared arithmetic and transport

The previous assembled pilot merges two owned root polynomials once. The new
[shared schedule](../../encoding/liftrs-shared/root_shards.go) grows contiguous
worker groups2,4,...,64,with power-of-two rank counts explicitly admitted.
Sibling groups operate independently and concurrently. Each group performs one
product and returns coefficient shards; no intermediate complete polynomial is
sent to one parent owner. Domain size is the next power of two accommodating the
true product degree and at least one canonical matrix row/column per worker.
Scheduling depends on counts,budgets and worker topology,not a fixed s value.

[Operand routing](../../encoding/liftrs-shared/distributed_operands.go) now shares
one implementation for single-owner inputs and already-sharded child inputs.
Bounded source/target interval intersections move only actual coefficients,
validate canonical field values/padding and preserve sources. Local copies bypass
MPI. Group sizes and child transform domains may differ between levels; the
normal case exercises unequal sibling domains. No intermediate gather is added.
The original owner-to-shards wrapper retains its packet ordering and checks.

[Explicit MPI subgroup handles](../distributed-root/mpi/world.go) use `Split`
and communicator-local ranks for every existing data/control/collective method.
The codecs,tags,64-byte descriptors and proofs stay shared. Independent subtrees
can reuse operation IDs without matching each other's packets. `Counts` gathers
only at most64 integer metadata words. Child handles own staging/communicators,
borrow the root's locked OS thread and must close before parents. A child close
never finalizes MPI or unlocks the root thread. Errors abort the original job.

Example lifecycle on the existing initializing goroutine:

```go
group, err := world.Split(world.Rank/groupWidth, world.Rank)
// Check err; all group members use group-local Rank/Size and the same order.
// Route child shards, run the maintained distributed product, close its GPU.
err = group.Close() // collective group release; world remains usable
```

[GPU integration](../../encoding/liftrs-shared/root_hierarchy_gpu_test.go) reuses
the maintained resident product driver (same three FFTs,pointwise operation and
local Taylor access). A test-only result callback retains its verified owning
output for the following level; an explicit cap may lower subgroup-worker memory
without raising the physical-card budget. No CUDA arithmetic kernel changes.
Native independent roots and scalar witnesses remain bounded diagnostics.

## Workload and validation

Canonical s18 base coefficients occupy4-byte slots,extension values8-byte slots;
root indices exercise values above2^32 in the2^36-point field. These are synthetic
Frobenius-closed subsets,not genuine s18 PIR support discovery. Seed is the existing
fixture rule `109+s+rootCount`; balanced conjugate-pair partitioning is unchanged.

| Case | Leaf polynomial counts | First-level groups | Final group |
|---|---|---|---|
| 16 roots | 5,5,5,5 | Two concurrent2-worker products;16-value domains | One4-worker product;32-value domain |
| 4092 roots | 1025,1025,1023,1023 | Two concurrent2-worker products;4096/2048-value domains | One4-worker product;4096-value domain |

One normal arm runs both cases. One CUDA memcheck arm runs16 roots through both
levels. This covers new communicator isolation and inter-level GPU/host layout,
without another allocation for unchanged tiny primitives. Native references check
leaf,subtree and final polynomials; three scalar support-root witnesses check the
complete native polynomial. All resulting coefficients,normalization,degree and
padding are validated. Parent integer collectives work after group release.

The independent reviewer checks the exact group tree/child-domain geometry,
all three FFTs per merge,transpose/Taylor proofs,source interval/payload hashes,
all stage/final output bytes and group-close records. It reconstructs routing
fingerprints from actual preceding-stage output files (leaves for level1),so a
correct final artifact cannot conceal a broken inter-level handoff. The product
reviewer optionally admits a one-host subgroup; its ordinary two-host default is
unchanged. Proofs/checksums detect corruption; they are not authentication.

CPU/race tests use an independent characteristic-two binomial oracle for2/4/8
workers,unequal child domains,identity leaves/padding,source preservation and
bad bounds/results/errors. Existing owner-routing/tree tests still pass. MPI
closed-handle admission tests require no MPI initialization. Native preflight
checks exact s18 leaf counts/domains and native root construction before launch.
It caught a missing no-op fixture phase hook locally; the shared fixture now
supplies that sink when omitted. The protocol algorithm was not changed.

## Memory–time and storage trade-offs

No production default is changed or new memory-saving mode selected. This is an
explicit experimental hierarchy; its total runtime advantage is **unknown**.

| Item | Cost / ownership |
|---|---|
| Device working arrays | Three8-byte arrays per group,about24D bytes aggregate for group domain D;metadata,transfer and Taylor buffers additional. Each stage chooses its own D and closes device workspace before the next stage. |
| Host intermediate data | Current child shard,two routed parent input shards and one base-coefficient result shard can overlap,plus driver capture/oracle buffers. Each principal output is admitted; library/runtime/GC and diagnostic arrays are additional. |
| Transfers / computation | One maintained product (two forward FFTs,one inverse) per merge. Independent sibling groups run concurrently. Coefficients return to host and are redistributed/uploaded for the next group; communicator creation and per-stage GPU/table setup are additional. No whole polynomial is gathered. |
| Metadata / transport | About3 MiB lookup tables per worker at s18; parent8 KiB and child8 KiB MPI-owned staging can overlap. CUDA contexts/driver/MPI internals are additional. |
| Persistence | Leaf and each-stage actual/native reference outputs and profiles are retained for bounded validation. No production intermediate checkpoint format is introduced yet. Projected cache below32 MiB; exact bytes measured by controller. |
| Runtime | Grouping can improve parallelism and avoid a single-executor final merge; extra setup/host transfers can add latency. Complete setup/routing/native/GPU/I/O costs must be measured before selecting a production strategy. |

Sizing smaller root subproducts does not establish a lower worst-case generator/
identity/spectrum requirement. The fulls18 device/host capacity table remains a
projection; it must not be replaced with these bounded root-component sizes.
The experiment still replicates small support/native reference arrays for its
independent oracle. The reusable tree/routing code itself consumes only local
leaf/shard data and bounded global count metadata.

## Unchanged physical resources and timing boundaries

**Two HPC nodes /two H200s /four CPU cores /8 GiB requested host RAM**,four MPI
workers,two workers/card,one CPU/worker. **64 MiB owned GPU/worker =128 MiB/card
aggregate**,**1 GiB RSS/worker**,zero CUDA-pinned buffers,**128 MiB retained cache /
2 GiB scratch /32 MiB file**,**720-second controller /15-minute scheduler**.
Exclude gnode03;other eligible nodes permitted. No resource increase.

79239 completed a larger four-arm suite in2 min27.292 seconds. Here normal and
small sanitizer each run three subgroup products per case;maximum4096 values,
with native preflight bounding setup. Estimate **3–6 minutes excluding queue**;
keep240-second launcher/220-second test deadlines and720-second total cutoff.
This is an estimate for this bounded validation,not fulls18 preprocessing time.

Whole-case times include native references,leaf construction,communicator/GPU
setup,routing,logging/proof capture,all stages and persistence. Product rows keep
the maintained kernel/capture boundaries and report setup separately. Profiled
normal and sanitized timings remain distinct. Progress logs flush timestamped
leaf/stage/pass/operand events and completed/total work; native/MPI chunk logs
include global/group-local rank and group size,and remain bounded and retained.
Completed arms/profiles persist before later work.
Submit and yield without monitoring until the user's completion ping.

```sh
python3 -B experiments/distributed-root-hierarchy/test_review.py
python3 -B experiments/distributed-root-hierarchy/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-root-hierarchy/slurm-%j.log \
 experiments/distributed-root-hierarchy/run.sbatch hierarchy-01
# After completion ping:
python3 -B experiments/distributed-root-hierarchy/analyze.py hierarchy-01
```

Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. The ordinary library/client build
still needs neither MPI nor CUDA. Local compiler socket warnings are retained
separately from successful build/test results. Before submission,create the real
committed-source snapshot,check driver/wrapper paths and run reviewer tests from
it. In-job preflight checks shared paths on all workers; no runtime JSON input
outside the frozen build is needed.

Full genuine discovery/Newton/checkpoint/publication integration beyond two
workers,larger representative scaling and fulls18 remain unvalidated. Odd worker
counts are not admitted by this hierarchy; the original reference tree retains
its odd-count support. No node-loss/elastic restart claim is made.


[Actual committed-snapshot preflight](readiness/snapshot-checks.json) passed on
6785af2:798 source files /3,675,323 bytes,required MPI/driver/wrapper/runner paths
present,and reviewer suite passed from the snapshot. Submission6bbd8f5 adds only
that evidence. No local GPU/MPI initialization;subgroup integration awaits79376.
