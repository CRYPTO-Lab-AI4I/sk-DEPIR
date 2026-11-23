# Coordinated support verification on checkpoint restart

Status: **passed job81750**, all three binding-01 arms independently verified. Implementation **946a2f6** (parent2227ccc).
[Runner](run.py), [Slurm entry](run.sbatch), [independent analyzer](analyze.py),
[reviewer regressions](test_review.py), [readiness](readiness/).
Planned results `runs/binding-01/`; immutable source, binaries, verified native
references and reusable caches remain outside Git under
`.cache/liftrs/distributed-restart-binding/binding-01/`.
Implementation/producing revisions are recorded before submission.

## Why this step

[Bounded diagnostic validation81746 passed](../distributed-diagnostic-writes/report.md#verified-result81746),
completing the missing wide check alongside81744's retained cold/restart/control
partial passes. The next recorded persistence issue is restart support binding:
`loadRankPrefix` hashes the entire global support file for each checkpoint in each
worker. Root plus Newton-prefix recovery therefore reads 2R copies for R workers.
This is verification traffic, separate from loading local coefficient shards.

## Shared recovery and correctness boundaries

[The existing loader](../../encoding/liftrs-shared/distributed_recovery.go) now
shares metadata/shard admission with a caller-supplied support verifier. Ordinary
`loadRankPrefix` still performs local full-file hashing. The pipeline's explicit
`PIR_PIPELINE_RESTART_SUPPORT=coordinator` selects
[coordinated binding](../../encoding/liftrs-shared/distributed_support_binding.go).
Unset/`local` retains two local hashes per worker; invalid modes fail admission.
No new public API, production parameters or full-size readiness is claimed.

On first recovery rank0 hashes the **complete admitted support file**, then all
workers exchange fixed 64-byte control frames using the existing MPI primitive.
Each frame binds the full checkpoint metadata, absolute shared directory, label
and a versioned context. Metadata includes every shard checksum, code shape,
ownership count, stage, prefix, layout and support digest. Rank0 contributes its
actual support digest; all other workers must contribute zero digest bytes.
Every worker requires identical metadata and the expected coordinator digest.
The two fixed-size agreement captures are also saved and independently checked
against the persisted commit metadata with a separate Python codec; shared literal
Go/Python fixtures bind typed field order, exact integers and string escaping.
The second checkpoint performs a fresh metadata agreement while reusing only the
already verified support digest, scoped to that immutable directory/label/session.
A different support digest, disagreement, invalid ownership or transport error
rejects the load. Failed agreements are not cached. Collective callers abort peers
on error using the existing lifecycle; no later arithmetic starts on a partial load.

Every worker still parses/adopts its own admitted metadata, verifies its complete
local shard hash, exact EOF, field range, omitted-zero padding, root normalization
and cross-record code/support binding. The support file is assumed immutable and
trusted during recovery, as in existing checksum-based caches. These hashes detect
corruption and disagreement; they do not authenticate an untrusted manifest, MPI
peer or file modified after verification. Coordinated mode is unsuitable if ranks
see different mutable files at the same shared path. It does not reshard checkpoints
or change their format, canonical representation or supported rank count.

No checksum sampling is introduced. The complete global support checksum is still
compared with both checkpoint records; per-worker repetition is replaced by explicit
collective evidence. No protocol or GPU arithmetic changes. CPU-only builds remain
CUDA-free, and client encoding backends are unchanged.

## Memory, communication, I/O and time trade-off

At illustrative full s18/16 workers, the support payload is about **121.446 GiB**.
Two local verifications per worker project **3.795 TiB aggregate reads**, versus
**121.446 GiB** for one coordinated verification: 32-fold fewer verification bytes.
The actual allocated storage and coefficient-shard reads do not change. No new
persistent support copy is created; retained experiment profiles/logs remain.
This projection is not full-size feasibility or measured restart latency.

Hashing uses a bounded 64 KiB buffer. Coordinated mode uses that buffer only on
rank0 and a control-frame array of 64R bytes per worker (1 KiB at R=16, at most
4 KiB under the current 64-rank transport limit). Each worker also temporarily
serializes bounded checkpoint metadata. Two agreement rounds exchange 2R(R-1)
64-byte payloads: **30 KiB aggregate at R=16**, excluding MPI overhead. Control
messages use host memory and add no GPU transforms, CUDA allocations or device
transfers. Filesystem caches, MPI buffers and runtime headroom are additional.

The new mode serializes full support hashing on one worker and adds two collective
rounds. It can reduce redundant I/O but loses parallel readers; **large-file latency
is unmeasured**, especially with filesystem caching. Local mode remains default.
No slower memory mode is silently enabled, and unknown cost is not treated as free.

## Validation and methodology

Pure CPU tests retain local recovery's malformed metadata, ownership, digest,
field/padding/EOF and compact-layout coverage. New tests cover one hash across two
checkpoint agreements, every coordinator/peer ownership role, inconsistent metadata,
incorrect coordinator digest, non-coordinator digest injection, support corruption,
read admission, bad mode and cross-record mismatch. A trusted external support
verifier cannot bypass a corrupted coefficient shard. SHA256 metadata encoding
binds full-width integers without floating conversion.

A three-arm profiled allocation uses genuine s8/curve5 data:

1. Cold discovery/root/Newton/full-identity/spectrum/canonical preparation.
2. Fresh-process prefix8 restart using coordinated support binding.
3. Matched fresh-process prefix8 restart from the same cold cache using default
   local verification. The local override is explicit and configuration-tested.

Both recovery arms retain the same bounded diagnostic writes, prefix inputs,
streamed discovery/spectrum, compact checkpoints, full trace and ordered proof
summaries. Verify actual support-hash passes/bytes and agreement counts per rank:
coordinator aggregate 1 pass versus local aggregate 8 passes, two agreements per
worker versus zero. All shards/native root/generator/canonical artifacts, independent
packet/digest reviews, process freshness, full identity and diagnostic totals must
match. Run no unchanged wide arithmetic or sanitizer-only arm. Completed phase
results/profiles persist before later checks. One fixed-order sample is not a
statistical or full-size speedup claim.

Hash progress saves timestamped checkpoint identifiers, completed/total bytes,
elapsed time and RSS at start/end and roughly every30 seconds of actual reads.
This is serialized-byte progress, not a heartbeat. Final events report measured
passes/bytes and agreements. Restart timing includes all verification, load,
CPU/GPU transforms, MPI, persistence, canonical publication and correctness checks;
controller timing also includes builds, native export, tests, review and profiles.

## Resources and reproduction

Same admitted HPC allocation: **two nodes/two H200s/four CPU cores/8 GiB requested
host RAM total**, four MPI workers/two per card; **64 MiB owned GPU/worker =128 MiB/card**,
**1 GiB RSS/worker**, zero CUDA-pinned; **128 MiB retained cache/2 GiB scratch/32 MiB
per file**, **720-second controller/15-minute Slurm limit**, eligible nodes excluding
gnode03. No new budget or full s18 admission. CUDA/MPI/context memory remains
additional to owned bytes. HPC Go1.27.3, GCC14.2/OpenMPI4.1.8, CUDA12.9/H200.

Estimate **3–6 minutes excluding queue** based on preceding three/four-arm
controllers. Cache projected below64 MiB, trace files below12 MiB per worker.
Fresh committed-source build and each-node executable/wrapper/library/input/cwd
preflight precede GPU phases. CPU/race/native/reference and frozen reviewer checks
precede submission; do not just raise caps or use full-size estimates as admission.

```sh
python3 -B experiments/distributed-restart-binding/test_review.py
python3 -B experiments/distributed-restart-binding/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-restart-binding/slurm-%j.log \
 experiments/distributed-restart-binding/run.sbatch binding-01
# After a terminal notice:
python3 -B experiments/distributed-restart-binding/analyze.py binding-01
```

After validation update the primary reference and continue production parameter/
rank/lifecycle ownership, non-oracle execution, bounded retention and actual phase
count/domain/resource admission. Full s18 remains unrun and unadmitted.

### Immutable-source readiness

Implementation **946a2f6** is frozen into **865 build inputs**; all **387**
Go/C/C++/CUDA source hashes match the compiled readiness tree. All **16 frozen
reviewer suites** and the focused CPU/race/tagged-native tests pass, including the
independent typed-metadata golden and corrupt-shard regression. The frozen native
preflight reproduces **108014 historical reference bytes** and validates actual
s8/curve5 root/leaf/stage/canonical inputs without initializing CUDA or MPI.
[Source checks](readiness/source-checks.json), [frozen reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). The admitted allocation still
rebuilds committed source and checks paths/environment on every node before GPU work.

### Submission81750

Job **81750** submitted from **15f3ab1**, implementation **946a2f6**.
[Receipt](submission.json); results `runs/binding-01/`, Slurm log `slurm-81750.log`.
Unchanged resources above; estimate3–6minutes excluding queue. No polling or GPU result claim.

## Verified result81750

All **865 source /79 evidence /376 cache files** reconcile to producing revision
**15f3ab1** and the submitted receipt. [Analysis](runs/binding-01/analysis.json),
[raw results](runs/binding-01/results.json), [Slurm log](slurm-81750.log),
[receipt](submission.json). Every captured metadata/control frame matches the
independent Python codec and actual persisted root/Newton manifests. Native Taylor
counts,proof digests,local shard/canonical/generator/identity/spectrum comparisons,
fresh-process recovery and diagnostic byte/progress totals passed.

| Mode on genuine s8/curve5 | Status/runtime | Support checksum work,aggregate | Peak host RSS |
|---|---|---|---|
| Cold | Passed;9.149322 seconds | No recovery hashes |379.121 MiB/worker |
| Coordinated prefix8 restart | Passed;6.792011 seconds |1 hash/29,524 bytes;2 metadata agreements per worker |379.160 MiB/worker |
| Local prefix8 restart | Passed;6.661931 seconds |8 hashes/236,192 bytes |377.895 MiB/worker |

HPC two nodes/two H200s/four workers/four CPU cores/8 GiB requested host RAM total,
with the unchanged caps above. Owned GPU0.244 MiB/worker,context/MPI/driver extra.
Controller **2 min39.483 seconds**, retained cache **37.252 MiB**. These runtime
rows precede deferred GPU closure;launcher/controller timing includes it. The small
support file was cached;logical bytes verified are not physical disk-read metrics.
Coordinated restart was **1.953% slower** in one fixed-order sample. This demonstrates
complete checksum/metadata agreement and eightfold less small-case verification
traffic,not a full-size speedup. Local remains default;coordinator selection explicit.

Next extract validated MPI transport and session ownership from test-only helpers
into reusable internal runtime,retaining independent native/test observers. Then
complete production parameter/rank/lifecycle,non-oracle entry,bounded retention,
restart/publication ownership and real phase count/domain/resource admission.
Full s18 remains unrun/unadmitted;its coordinated restart latency is unknown.
