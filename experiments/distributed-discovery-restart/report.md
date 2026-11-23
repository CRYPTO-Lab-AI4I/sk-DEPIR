# Discovery archive recovery and representative s10 validation

Status: **passed job82199**, recovery-01; independently reviewed saved evidence. SCALING-195 follows
[passed eight-worker82194](../distributed-host-admission/report.md#verified-outcome82194).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Results `runs/recovery-01/`; source/binaries/reusable data outside Git:
`.cache/liftrs/distributed-discovery-restart/recovery-01/`. Producing revision is
recorded before submission. All hardware/storage/time ceilings remain those of
the approved eight-worker validation; this is **not full s18 preprocessing**.

## Recovery at existing durable boundaries

The command already seals a complete discovery archive and source metadata before
root construction, but previously could resume only with a saved root and Newton
prefix. A later failure could therefore require repeating completed discovery.
[Cached discovery loader](../../encoding/liftrs-shared/distributed_archive_mpi.go)
now makes that immutable archive usable. `resume_directory` plus trusted
`source_sha256` and `known_prefix=0` selects it explicitly; ordinary cold discovery
remains unchanged. Source/spec/rank/count/format metadata must match before loading.

Only rank0 reads/hashes the whole archive. It checks exact EOF, strict global order
and reduced exponent range. Bounded65536-element tiles are partitioned by the
existing closed-orbit owner rule; paged count frames admit counts and reject
noncoordinator contributions/padding before transfers. Existing typed descriptor/
payload fingerprints protect each owner bucket. Each receiver checks owner/order/
list limits, then all workers agree source/support digests and global root count
before any GPU/root work. Recovery uses the same root-construction and numerical
tail as cold execution. Archive hashes are corruption checks, not authentication;
source provenance and immutability remain caller requirements.

`known_prefix=1` restores only the normalized root. P(0)=1 proves G's constant term
is1, so it initializes that term without requesting a nonexistent prefix1 file.
For `known_prefix>=2`, the existing saved Newton-prefix loader remains. No second
root/Newton/identity/spectrum algorithm is introduced. Both root-only and Newton
recovery use existing coordinator-only source verification and full identity check.
Same rank count is required; no resharding or failed-node recovery claim.

These modes recover **completed discovery, root or Newton stages**. Partial sampling
or unfinished local/subgroup root work is not checkpointed. Previously retained
archives/profiles are preserved. Final publication links the already verified source
archive without copying its payload. A corrupt/incomplete cache fails before GPU
allocation; incomplete output/reservations remain for diagnosis.

## Memory, time and I/O

Cached discovery omits good-monomial enumeration, sampling grids/bitmaps and discovery
FFTs, but reads/hashes the full archive once on rank0 and routes its indices. Three
bounded512KiB tile buffers total **1.5 MiB/worker**, plus bounded count/control state
and each admitted root list; append-capacity allowance is included in host admission.
It adds no retained support payload copy, numerical transform or forced GC.
At s18 the archive read is projected **121.446 GiB**; the actual source/hash/routing
latency and fullsize reserve remain unmeasured. This is explicit recovery from
available immutable setup, not an automatically selected slower memory mode.

## Representative size and reference provenance

Advance from s8's65536 grid values to **s10's1048576 values (16×)** on the same
8workers/2H200s. Exact K **146641**,m2,t5. Independently verified native cache72839
(source715bac2) already contains the complete online code,G and support. It is
retained at `.cache/liftrs/gpu-products/products-02/cpu-0/`; no expensive reference
preprocessing is rerun or old data removed. [Original measured methodology/results](../gpu-products/report.md)
report6.647s CPU preparation/9.756s all phase clocks and1.528s resident preparation/
4.644s all clocks at the earlier cutoff on one H200. Those timings are historical
budgeting evidence, not predictions of this distributed wrapper.

[Go header reader](export_header.go) validates canonical coordinates; the
[reference exporter](prepare_reference.py) checks all old cache payloads against
job72839's recorded seals. [Frozen reference](reference.py) includes complete code/
generator hashes, support indices widened from the old4-byte storage to the
command's explicit8-byte raw archive, and logical root/G-prefix hashes. Final G
prefixes equal completed Newton truncations by uniqueness of the inverse of
normalized P. Eight-rank root counts are independently derived from the old support.
The archive bootstrap itself consumes the new command's sealed archive, not an old
format guessed by filename. This preserves the distinction between cache codecs.

```sh
CGO_ENABLED=0 go run experiments/distributed-discovery-restart/export_header.go \
 .cache/liftrs/gpu-products/products-02/cpu-0/online.bin > /tmp/s10-header.json
python3 -B experiments/distributed-discovery-restart/prepare_reference.py \
 .cache/liftrs/gpu-products/products-02/cpu-0 /tmp/s10-header.json \
 experiments/distributed-discovery-restart/reference.py
```

## Validation and logging boundary

One allocation runs actual command **cold, cached-discovery restart, root-only
restart and prefix8 restart**. This checks each durable boundary at the larger
representative size without an additional unchanged s8/sanitizer suite. All compact
shard/commit/source/publication hashes and logical payloads must match frozen native
data; all modes verify the full identity and final code/G/archive bytes. Native
Taylor totals use actual eight-worker/group geometry. Host inventories, coordinator
hash counts, archive aliases and omitted discovery work are checked. Tile controls
and decoder tests cover pages beyond8workers, wrong contributors/padding, extent,
range and order across boundaries. Existing s8 saved evidence is reanalyzed after
generalizing the reviewer; default trace behavior stays strict.

This job explicitly selects `proof_log=summary`. S8 trace already reaches5.9MiB
for one rank; scaling raw rows unchanged risks exceeding the32MiB/file cap.
The shared summary recorder was validated against full traces in81738. Runtime
MPI descriptor/payload/field checks remain; ordered count/byte/SHA footers and native
work are saved, with no full raw-proof replay for this new shape. Summary and trace
perform the same digest work; summary omits raw-row writes. No new transforms or
extra hash passes are added by this diagnostic selection. Its fullsize timing is
unknown; do not confuse omitted logs with omitted correctness checks.

CPU/race tests, tagged builds, real suite-wiring/reference-size tests and summary
footer rejection precede submission. Initial local header export encountered a
read-only default compiler cache; rerunning with the authorized shared compiler
cache passed. No GPU allocation was used for that local tooling issue.

## Same admitted resources and shape preflight

**2 HPC nodes /2 H200 /8 CPU cores /16 GiB requested host RAM total**,8workers;
**32 MiB owned GPU/worker =128 MiB/card**, **1 GiB RSS/worker**,768MiB declared
runtime/native/GC allowance,zero CUDA-pinned; **128 MiB total new cache /2 GiB
scratch /32 MiB/file**, **720-second controller /15-minute Slurm cutoff**.
Eligible nodes exceptgnode03. Go1.27.3,GCC14.2,OpenMPI4.1.8,CUDA12.9/H200.
No ceiling increase. Field/array counts derive from the new shape, not raised byte
budgets. Eight workers share two cards; this is not eight-card bandwidth evidence.

Each full-domain shard has131072 coefficients; three owned GPU arrays are1.5MiB/
worker before tables/windows/context. Even a full-domain three-array local executor
would be12MiB, below the32MiB owned cap with auxiliary headroom; actual CUDA caps
remain enforced. Max root leaf count40877fits131072; rounded group caps are checked
from actual counts. The output code is4.559MiB; projected logical payloads plus
all21compact cold checkpoints/metadata stay below32MiB/arm. Four arms plus binaries
project below112MiB retained cache, below128MiB cap. RSS remains1GiB rather than a
tighter limit; s8 logged374MiB is not a measured s10 peak.

Estimated **4–9 minutes excluding queue**, using measured s8 controller171.032s,
retained s10 timings and source-derived size/log bounds; not a latency baseline.
Native/script/node preflight, progress, profiles, watchdogs and completed commits
remain. Controller includes builds/CPU tests/review/profiles/seals; launcher includes
profile flush/shutdown, command rows include persistence and close. No direct
speedup comparison with differently instrumented historical runs.

```sh
python3 -B experiments/distributed-discovery-restart/test_review.py
python3 -B experiments/distributed-discovery-restart/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-discovery-restart/slurm-%j.log \
 experiments/distributed-discovery-restart/run.sbatch recovery-01
# After terminal notice:
python3 -B experiments/distributed-discovery-restart/analyze.py recovery-01
```

Next: use representative measurements and source inventories to specify fullsize
host/native/GC reserve, root ownership/node/GPU packing, retained physical/logical
storage, deadline and remaining checkpoint granularity before fulls18 resource
agreement. **Full s18 remains unrun/unadmitted.**


The host inventory also corrects field-table accounting: the log has q entries
and the duplicated-power table2(q−1), totaling **12q−8 bytes**, not8q. A CPU test
compares the estimate with actual table lengths. This changes accounting only;
82194's logged inventories remain historical values for that producing revision.

### Frozen-source readiness

Implementation **b79c5c4** freezes **949 build inputs**; all **418** compiled code
hashes and selected inputs match the readiness tree. All **28 frozen reviewer
suites** pass. CPU/race, tagged command/reference builds and admission checks pass
without local GPU/MPI initialization. Native preflight matches **108014 historical
bytes** plus actual s8/curve5 references. The s10 header/cache payloads were separately
verified against native72839 seals, and saved82194 evidence still passes generalized
review. [Source checks](readiness/source-checks.json), [reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). Each-node executable/wrapper/
library/config/output checks and fresh frozen builds remain inside the allocation.

### Submission82199

Submitted from **0f02d72**, implementation **b79c5c4**. [Receipt](submission.json).
Results `runs/recovery-01/`, Slurm log `slurm-82199.log`, reusable cache above.
Same approved2nodes/2H200/8cores/16GiB and all existing byte/storage/time ceilings.
Estimated controller **4–9 minutes excluding queue**. Submission is not a pass;
completion notification registered locally, no agent polling. Fulls18 unrun/unadmitted.

## Verified outcome82199

[Independent analysis](runs/recovery-01/analysis.json), [saved evidence](runs/recovery-01/),
[Slurm log](slurm-82199.log). Producing revision0f02d72, implementationb79c5c4.
All **949 source /157 evidence /775 cache** seals reconcile. Controller runtime
**6 min34.547 s**; retained logical cache **96.118 MiB**. Same approved HPC allocation:
2 nodes/2 H200/8 cores/16 GiB requested host RAM,8 workers.

| s10 arm | Launcher runtime | Logged peak host RSS per worker | Maximum named host phase per worker |
|---|---|---|---|
| Cold | 71.482 seconds | 380.117 MiB | 4.184 MiB |
| Completed-discovery recovery | 70.117 seconds | 381.652 MiB | 3.563 MiB |
| Root-only recovery | 49.614 seconds | 382.391 MiB | 2.316 MiB |
| Prefix8 recovery | 43.343 seconds | 381.242 MiB | 2.316 MiB |

All canonical/native/logical payloads, source bindings, commit/shard hashes, archive
aliases, host admission and native-work checks passed. Cold/discovery/root arms each
retain21 commits/168 shards; prefix8 retains19/152. Summary-mode runtime checks and
full native/artifact equality passed; no independent full raw-proof transcript
replay for this larger shape. Launcher timings include profile flush/shutdown;
controller includes build/tests/review/sealing. Recovery timing is measured at s10
only and is not a fullsize speedup projection.

Next: [fullsize ownership/resource audit](../distributed-fullsize-plan/report.md),
including a read-only scan of retained s17 support before fulls18 admission.
