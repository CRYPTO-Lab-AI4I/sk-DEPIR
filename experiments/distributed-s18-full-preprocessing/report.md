# First full distributed s18 preprocessing trial

SCALING-197. **Prepared; not admitted or submitted.**
[Budget](budget.json), [runner](run.py), [Slurm entry](run.sbatch),
[each-node preflight](node.sh), [no-GPU checks](test_plan.py), [check result](checks.json).
This runs the existing shared server command, not another arithmetic implementation.
CPU and optional GPU client encoding are unchanged.

## Why eight nodes for sixteen GPUs

[Retained s17 ownership82200](../distributed-fullsize-plan/runs/owners-01/analysis.json)
verified3,876,566,203 support indices in143.040 seconds, with14.785 MiB peak RSS.
For16 workers, the largest list has482,718,852 indices, **1.992 times the average**.
It occupies3.597 GiB as raw indices;32 workers have a2.261-times maximum/average
ratio. These are s17 measurements, not s18 bounds.

The s18 setup below admits **4,294,967,295 roots per worker**, leaving the extra
polynomial coefficient within the4,294,967,296-word shard cap. This is over four
uniform-average lists and over twice the largest list projected by scaling the
s17 counts by K18/K17 (about2.030 billion). That projection is not a proof.
Actual counts and rounded hierarchy domains are checked before root GPU work.
Discovery is sealed before root admission; if the unexpected distribution exceeds
this cap, the completed archive survives for an explicitly redesigned retry.

With worst admitted append capacities, discovery has at most about **176.516 GiB
of named host arrays per worker**, including scratch: grid, bitmap, good-list
capacity, raw indices, routed-root capacity and bounded result tiles. Reserve
**192 GiB/worker** separately for native/MPI/CUDA context, Go runtime/allocator/GC
and prior allocations. Named plus reserve is below **368.516 GiB/worker**, leaving
another31.484 GiB before the **400 GiB RSS guard**. This is a source-derived
inventory with an explicit allowance, not a measured peak bound.

Four such workers cannot safely share a906.8 GiB configured-memory node. Use
**two workers/two H200s per node across eight nodes**. Two RSS caps total800 GiB;
request864 GiB/node, leaving64 GiB of allocation headroom. A no-GPU each-node
preflight also checks832 GiB actually available before work, node-local scratch,
shared binary/library/config/output paths, dynamic linkage and command policy.
The allocation has64 active CPU cores total,4 per worker. This avoids a memory
strategy that would silently trade runtime for fewer nodes. Eight-node network
performance is unmeasured; more nodes can increase communication costs.

GPU principal arrays are **96 GiB/card**, with a **112 GiB owned-device cap**,
plus physical H200 capacity outside that cap for driver/context headroom. Local
root domains and every rounded subgroup shard fit the same word cap when all
admitted leaf counts fit: subgroup count is at most(group size × shard cap),
whose power-of-two ceiling is no larger than that product. Lazy local executors
are closed before shared full-domain sessions; no duplicate full executor is
intentionally retained. Full-size allocations remain unmeasured.

## Selected modes and costs

- **Resident GPU arithmetic**, existing prefix-input Newton:64 GiB/worker in
  base host arrays. No dense root clone or extra transforms; prefix inputs avoid
  uploading known zero padding. No fullsize speedup estimate is claimed.
- **Resident host spectrum**: the faster validated default is retained because
  its extra32 GiB/worker fits. Its named final-phase inventory is about80 GiB,
  below the worst discovery inventory. The streamed spectrum alternative saved
  that32 GiB but was3.545% slower in one matched small-case sample; its fullsize
  penalty is unknown, so it is not selected here.
- **Explicit streamed discovery results/archive**: resident archive allowance
  would add about243 GiB/worker under the current conservative inventory and
  exceed the400 GiB RSS policy. Streamed mode downloads bounded tiles instead;
  fullsize runtime/overlap penalty is unknown. Total canonical output is unchanged.
- **Summary proof diagnostics** retain runtime descriptor/payload checks, ordered
  summary digests and full identity verification; raw proof rows are omitted.
  Summary versus trace has already been checked on bounded cases; independent
  full raw-transcript replay is not claimed for this full-size shape.

No library default changes. These are explicit job configuration values, not
automatic fallbacks based on s or backend.

## Storage, logging and recovery

Request a **2 TiB logical output/checkpoint cap**, **2.25 TiB total new retained
cap** including binaries/logs/profiles, **8 GiB head-node scratch**, and768 GiB
maximum individual payload file. Require2.5 TiB initially free and keep128 GiB
free. Model: about1.37 TiB physical /1.49 TiB logical for canonical code, generator,
support and all compact prefixes (support/archive names are hardlinks). Storage
is shared; no complete grid or support is exchanged through files between ranks.
All useful caches and profiles remain outside Git, with manifests/results in Git.

The job writes final payload hashes and full polynomial identity evidence, then
independently streams each final payload once to reconcile its hash/size with the
publication manifest. This is about **889 GiB of extra read I/O**, no retained
payload copy. It is included in controller time and logged as a separate phase.
The source archive scan's roughly207 MiB/s inclusive read/hash/count rate suggests
about73 minutes for that read volume, but is not a filesystem bandwidth guarantee.
Checkpoint writes and fsyncs are included in preprocessing, not excluded costs.

Saved progress includes native transform/pass/tile identifiers and completed/total
work, timestamps and resource observations. Controller observations are labeled as
such, not forward progress. CPU profiles are mandatory per rank and preserved.
Successful discovery, root and Newton-prefix commits survive later failure; root
subtrees and partial sampling are not restartable. No claim of failed-node recovery
or automatic resharding. A new resume invocation must bind verified source metadata
and use a fresh output namespace; this first-trial runner starts cold only.

## Deadlines and uncertainty

The live gpuq partition limit was checked on2026-11-23: **100 hours**. Sixteen
eligible H200 node types are configured with4 cards and928610 MiB RAM each; capacity
is not current availability or guaranteed account/QoS admission. gnode03 remains
excluded. No other jobs were queried.

Use **94-hour worker /99-hour controller /100-hour scheduler** cutoffs. This is
five hours after the worker cutoff for verification/flush, plus scheduler margin;
these cutoffs are not a runtime estimate. The prior single-node s17 resident
preprocessing measured about10 hours, dominated by an8.57-hour root product; s18
has4 times the grid and a new distributed execution schedule, so extrapolating
that time by GPU count is unreliable. Full-size runtime and network throughput are
**unknown**; allow several days and use logged completed-work counters for an ETA
once meaningful progress is available. Do not promise completion within100 hours.
Completed durable stages are retained if a cutoff or guard is reached.

## Preflight and validation boundary

Before full arrays, the same allocation runs **one genuine s8 cold case on all
16 workers/eight nodes**, using only32 MiB GPU/worker and1 GiB RSS/worker. Internal
full identity and independent saved native canonical hashes must match. This
checks actual rank/node mapping,16-worker collectives, publication and profiles
inside the admitted full allocation; no separate calibration queue round trip.

The prepared policy/configs passed the actual retained MPI command's no-GPU
`--check-config` for s8 and s18. Six unsafe-budget cases were rejected, shell syntax
passed and unadmitted launch is rejected before directories/allocation work.
No live GPU, MPI or fullsize allocation was tested locally; topology preflight
remains inside the proposed job. Source is frozen from the submitted Git revision
using the existing bounded source-snapshot helper; every build-input hash is saved.

```sh
module load gcc/14.2.0 openmpi/gcc/64/4.1.8 cuda12.9/toolkit/12.9.1
python3 -B experiments/distributed-s18-full-preprocessing/test_plan.py
python3 -B experiments/distributed-s18-full-preprocessing/run.py --check
# Only after explicit agreement: record admitted=true in budget.json and commit.
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-s18-full-preprocessing/slurm-%j.log \
 experiments/distributed-s18-full-preprocessing/run.sbatch full-01
```

Expected results: `runs/full-01/`; reusable data/profiles linked from metadata under
`.cache/liftrs/distributed-s18-full-preprocessing/full-01/`. No job submitted.
AGENTS.md resource-admission rules require agreement: this exceeds current small
validation ceilings in nodes, GPUs, CPU, host/device RAM, storage and time.

### Prepared-source checkpoint

Prepared implementation **ff99e37**. [Frozen-source path check](source-checks.json)
selected959 build inputs and verified the command, rank/node wrappers, native s8
reference and CUDA translation units against the committed tree. No MPI/GPU was
initialized for this check. Resource agreement is pending; admitted=false and
no fulljob submitted. Partition capacity was queried without inspecting jobs.
