# Host phase admission and eight-worker validation

Status: **prepared; awaiting resource agreement; no job submitted**. SCALING-194
follows [passed actual command82181](../distributed-command/report.md#verified-outcome82181).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Planned results `runs/host-01/`; reusable data/source/binaries outside Git:
`.cache/liftrs/distributed-host-admission/host-01/`. Producing revision will be
recorded in the receipt before submission. This is eight workers sharing **two**
H200s, not eight GPUs or a fullsize throughput benchmark.

## Host admission model

[Inventory implementation](../../encoding/liftrs-shared/distributed_host.go)
separates named phase allocations from an explicit caller allowance for native/MPI/
CUDA-context/runtime memory, predecessor allocations, Go slice growth and GC/allocator
retention. New command configurations must supply `host_reserve_bytes`; this is an
experimental policy schema addition. Old frozen JSON remains evidence for its
producing revision. Admission requires named bytes + declared reserve ≤RSS cap;
the actual RSS watchdog and Slurm hard limits remain independent.

| Phase | Named inventory and input |
|---|---|
| Good enumeration | Twice the configured good-list length, plus membership block; checked before enumeration |
| Discovery | Actual good-list capacity, base grid/bitmap, bounded raw support, twice route/archive list caps, selected resident result or stream tile |
| Local roots | Actual support-list capacity, six coefficient equivalents of actual leaf length, bounded factor batch |
| Root groups | Actual leaf counts determine every rounded subgroup domain; six shard coefficient equivalents per group |
| Root reshape | Actual compact shard plus two full-domain routing outputs, retained support capacity |
| Newton | P,G,G² and product output; dense mode additionally includes its root clone |
| Identity | P,G and full product output |
| Spectrum/publication | Three base shards plus resident extension result only when selected |
| Checkpoint reload | Two resident restored shards before GPU allocation |

Transform phases include two bounded Taylor host windows, transfer tiles, field
log/power tables and codec/control scratch. Six root coefficient equivalents are
a conservative named builder/input/output/normalization envelope, not a proven
complete host peak. GPU memory is still admitted by actual owned-allocation caps.
Local/group domains use real counts, not a balanced K/rank assumption. All future
Newton/identity/spectrum plans are admitted before starting Newton, preventing a
late rejection after expensive refinement. No numerical algorithm, transform,
transfer, forced GC or slower memory mode is added.

Each accepted inventory logs component bytes, named total, reserve, RSS cap and
ordinary observed RSS. **The reserve is a caller allowance, not a measured upper
bound.** Fullsize GC/native headroom still requires representative evidence.
These plans do not turn tiny-case RSS into a fullsize feasibility claim. A source-
derived s18/16-worker prefix Newton inventory now explicitly includes its product
output: **64 GiB base arrays/worker**, versus **80 GiB dense**, plus named scratch
and reserve. The earlier **48 GiB** figure describes prefix *input arrays only*;
it was never a total peak. No extra allocation was introduced by this accounting.

## Reproducible validation

CPU/race tests check component sums, actual rounded group domains, missing reserve,
exact-budget acceptance and one-byte-over rejection. The s18 unit checks byte
units/counts without allocating large arrays. Tagged command/reference builds
and CPU-only native input preflight require no MPI/GPU initialization. Python tests
capture actual eight-worker launch/resource arguments and exercise the generalized
native-work reviewer. Existing four-worker reviewer behavior stays unchanged.

The proposed allocation runs only actual command cold/fresh-prefix8 restart at
s8/curve5,N65536,K7381,seed109. This combines host-observer validation with the
necessary next worker-count step; no separate unchanged four-worker/dense/sanitizer
suite. Eight ranks exercise another hierarchical root level and more transport
peers. Live descriptor/payload/field/identity checks remain, full trace summaries
are reconciled against their rows, and native Taylor totals use actual eight-rank
windows/group geometry. Proof hashes are not compared directly across four and
eight ranks because packet ownership/shapes change.

[reference.py](reference.py) freezes logical canonical and checkpoint hashes from
independently verified four-worker82181 (sourcea9fb7c2), already compared with native
code/generator/root data. New eight-shard concatenations must reproduce each logical
root/Newton payload byte-for-byte, as well as final code/generator/raw support.
Archive ownership independently determines expected eight-rank root counts.
Source/spec/rank/sample/commit/shard hashes, exact compact lengths, alias identity,
coordinator-only restart support hashing and all host-phase inventories are checked.
The command omits test-only diagnostic copies, as in82181.

Regenerate the small frozen reference only from the trusted independently verified
baseline (the exporter checks bytes, not mathematical correctness/authentication):

```sh
python3 -B experiments/distributed-command/analyze.py command-01
python3 -B experiments/distributed-host-admission/prepare_reference.py \
 .cache/liftrs/distributed-command/command-01 \
 experiments/distributed-host-admission/reference.py
```

## Proposed resources — approval required

| Resource | Current admitted validation | Proposed eight-worker test |
|---|---|---|
| HPC nodes /H200 GPUs | 2 /2 | 2 /2, unchanged |
| MPI workers /CPU cores total | 4 /4 | **8 /8** (4 cores/node) |
| Requested host RAM total | 8 GiB | **16 GiB** (8 GiB/node) |
| Peak RSS guard per worker | 1 GiB | 1 GiB, unchanged |
| Declared host runtime/native/GC reserve | Previously unspecified | 768 MiB/worker; an allowance, not allocated memory |
| Owned GPU cap | 64 MiB/worker,128 MiB/card | 32 MiB/worker,128 MiB/card; aggregate unchanged |
| Retained cache /scratch /per-file cap | 128 MiB /2 GiB /32 MiB | Unchanged |
| Controller /Slurm cutoff | 720 s /15 min | Unchanged |

Per-rank RSS stays at1GiB, comfortably above82181's measured375–376MiB, rather
than tightening guards to fit more workers. Extra requested RAM leaves controller/
build/native/MPI headroom. Zero CUDA-pinned allocation; context/driver memory extra.
Eligible GPU nodes exceptgnode03. Go1.27.3,GCC14.2,OpenMPI4.1.8,CUDA12.9/H200.

Estimated **3–7 minutes excluding queue**, based on174.239seconds for82181's four
arms plus build work; this has two arms/more ranks, not a throughput prediction.
Projected retained cache below96MiB; per-rank logs/profiles stay under32MiB guard.
One-second watchdogs, phase/native progress, profiles, checkpoint sync and fail-fast
launch remain. Frozen-source/native preflight is completed before requesting
submission; each-node binary/wrapper/library/config/output checks precede GPU work.

```sh
python3 -B experiments/distributed-host-admission/test_review.py
python3 -B experiments/distributed-host-admission/run.py --check
# Only after resource agreement:
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-host-admission/slurm-%j.log \
 experiments/distributed-host-admission/run.sbatch host-01
# After terminal notice:
python3 -B experiments/distributed-host-admission/analyze.py host-01
```

Logging/admission/one-second resource-walk costs are not assumed free. Controller
timing includes builds, CPU tests, profile processing, review and evidence/cache
seals; command rows include close and persistence, launcher includes profile flush.
No direct comparison with the old pilot's earlier timing boundary. Completed caches/
profiles remain on later failure. Full s18 remains **unrun/unadmitted**; next use
representative measurements to select fullsize host reserve, node/GPU layout and
complete storage/time/headroom budget. This eight-worker correctness test alone
cannot establish multi-card bandwidth or fullsize GC/RSS behavior.

### Frozen-source readiness

Implementation **aff14a1** freezes **939 build inputs**; all **414** compiled code
hashes and selected inputs match the readiness tree. All **27 frozen reviewer
suites** pass. Ordinary CPU/race, tagged command/reference builds and admission
checks pass without local GPU/MPI initialization. Native preflight reproduces
**108014 historical bytes** and actual s8/curve5 reference/root geometry.
Saved82181evidence also passes the generalized reviewer's default four-worker path.
[Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No job submitted; CPU/RAM
resource agreement is the remaining requirement before this prepared test launches.
