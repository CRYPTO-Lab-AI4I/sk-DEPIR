# Durable assembled workflow checkpoints

Status: **passed job82101**, checkpoints-01; all three arms and fresh-process loader verified. SCALING-191 follows
[passed publication82100](../distributed-publication-commit/report.md#verified-outcome82100).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Results: `runs/checkpoints-01/`; retained source/binary/data outside Git:
`.cache/liftrs/distributed-workflow-checkpoints/checkpoints-01/`.
Producing revision is recorded in the receipt before submission.

## Implementation and boundary

The reference-independent numerical workflow calls an optional discovery-ready
sink after its real code header/sample count are known and before root arithmetic.
The coordinator seals/syncs the already streamed raw support archive and hard-links
it to `s8-support.bin` for the existing loader. No payload copying or reread is
needed. Further archive writes are rejected. The alias survives the final archive
rename and is usable even if final publication never commits. This requires a
shared POSIX filesystem with same-filesystem hard links and directory fsync.
The job tests those filesystem operations before GPU phases.

[Checkpoint writer](../../encoding/liftrs-shared/distributed_checkpoint_mpi.go)
agrees code/source metadata and coordinator digest through the existing bounded
control transport. It writes a synced source JSON binding the full canonical
header (including curve/field/tower), rank count, actual discovery sample count,
raw archive format/byte count/hash. Each root/Newton manifest binds that source
JSON hash in addition to its existing shape/support/shard hashes. Before saving
a stage, workers agree its metadata; then each writes/syncs its compact shard,
collects all checksums, and the coordinator syncs directory/commit metadata,
renames the marker and syncs the directory. A barrier precedes success events.
The caller aborts peers on failure. No success is inferred from a shard alone.
Incomplete output remains; exclusive filenames prevent overwriting old evidence.
A final directory-sync failure remains an error even if the marker is visible.

[Source metadata](../../encoding/liftrs-shared/distributed_checkpoint_source.go)
and [existing loader](../../encoding/liftrs-shared/distributed_recovery.go) enforce
bounded metadata, canonical shape and binding before allocating checkpoint values.
Existing legacy manifests omit the optional `source_sha256` and remain compatible.
Support hash, shard digest, exact EOF, field range, zero padding and root normalization
checks remain. The support alias is explicitly a raw `uint64-exponents-le` archive,
not the public canonical support cache. Checksums are corruption checks, not
producer authentication. Directory ownership and trusted immutable provenance are
required; experiment seals bind the producing source separately.

`PIR_WORKFLOW_CHECKPOINT=commit` explicitly enables this policy in the bounded
adapter and requires final publication commit mode. Unset preserves the previous
observer-only checkpoint writer. Numerical kernels and client encoding are unchanged;
ordinary CPU builds still do not need CUDA/MPI. This is not a public preprocessing
command or assembled arithmetic-resume path yet.

## Memory–time and storage trade-off

No added numerical arrays, transforms or GPU transfers. Existing compact shard
payload bytes and fsyncs remain; one archive hard link shares payload blocks, with
no duplicate support allocation or cold-path hash reread. Extra storage is one
bounded source JSON (at most64KiB) and one small manifest per checkpoint; control
frames and JSON are bounded metadata. At the validation shape there are17stages
(root plus16Newton prefixes) and68rank shards. Added operations are archive seal/
directory sync, source metadata persistence, two small control exchanges per stage,
and commit-file/directory syncs. **Their fullsize latency is unmeasured**. This
explicit persistence policy is not a claim of free memory or runtime improvement.
The faster validated path/default is unchanged.

Logical inventory guards count both archive and support alias names although
physical payload blocks are shared. Fresh-process loader validation deliberately
hashes the tiny support file once per load; scalable coordinated verification and
arithmetic resume will reuse the existing validated binding path in the next step.
No per-worker fullsize support read should be silently selected as the fastest path.

## Validation

Local CPU/race tests cover source binding, changed source rejection, exclusive
metadata/no overwrite, existing compact/legacy loader checks, archive alias identity
and post-seal write rejection. Tagged compilation/admission tests require no GPU
or MPI initialization. Python tests exercise the actual suite's explicit mode and
reject changed source digests/shard bytes through independent file review.

One admitted job keeps the reviewed genuine s8/curve5 pilot cold/restart and runs
the assembled runtime with durable source-bound commits. It retains comparisons of
all76intermediate support/root/Newton/identity files, ordered proof summaries,
native roots/Taylor counts, discovery entropy/support ownership and canonical
code/generator/archive bytes. It additionally verifies17commit manifests, all68
shard sizes/hashes, every rank's commit event, full source/header binding and same
inode for the support alias. A fresh process loads all four root and prefix8 shards
through the runtime loader, with a saved CPU profile; this is **file reload**, not
arithmetic resume. No unchanged wide/dense/sanitizer repeat.

Pilot timing rows precede deferred GPU close; assembled rows include runtime close,
observer commits and final sync. Launcher includes shutdown; controller includes
builds, tests, native exports, review, profiles and evidence/cache hashing. These
boundaries prevent a direct speedup comparison. Timestamped phase/progress logs
and completed profiles/data persist even if a later phase fails.

## Resources and reproduction

Same admitted HPC budget: **2 nodes /2 H200 /4 CPU cores /8 GiB requested host RAM
total**,4MPI workers/two per card; **64 MiB owned GPU/worker =128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache /2 GiB scratch /
32 MiB per file**, **720-second controller /15-minute Slurm cutoff**; any eligible
nodes except gnode03. Context/driver/MPI memory extra. Go1.27.3, GCC14.2,
OpenMPI4.1.8, CUDA12.9/H200. No budget increase/fullsize admission.

Estimated controller **3–6 minutes excluding queue**, based on82100's measured
158.073seconds plus small metadata/reload work; projected cache below64MiB.
Exact frozen source and native input preflight precedes submission, and existing
each-node environment/path checks remain before expensive GPU phases.

```sh
python3 -B experiments/distributed-workflow-checkpoints/test_review.py
python3 -B experiments/distributed-workflow-checkpoints/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-workflow-checkpoints/slurm-%j.log \
 experiments/distributed-workflow-checkpoints/run.sbatch checkpoints-01
# After terminal notice:
python3 -B experiments/distributed-workflow-checkpoints/analyze.py checkpoints-01
```

Next: assembled arithmetic resume with coordinated source verification, parameterized
command and full phase/working-set admission, then representative larger topology
and complete fulls18 budget approval. **Full s18 preprocessing remains unrun.**

### Frozen-source readiness

Implementation **ce295e2** is frozen into **915 build inputs**; all **405**
Go/C/C++/CUDA code hashes match the compiled readiness tree. All **24 frozen
reviewer suites** pass. CPU/race and tagged CPU admission tests pass without local
CUDA/MPI initialization. CPU native preflight reproduces **108014 historical bytes**
and the actual s8/curve5 native/root input geometry. [Source checks](readiness/source-checks.json),
[reviewers](readiness/frozen-reviewers.json), [native preflight](readiness/native-preflight.log).

An initial manual preflight passed both native exports but used the frozen source
subdirectory as provenance root while generated files were its siblings, causing
its final relative-path inventory to fail. [Original log](readiness/native-preflight-path-error.log)
is retained. The corrected invocation uses the repository root containing both
paths and passes. This was a local inventory-command error; no GPU allocation
was consumed and the job's existing source/cache layout is unchanged.

### Submission82101

Submitted from **28c37c7**, implementation **ce295e2**. [Receipt](submission.json).
Results `runs/checkpoints-01/`, Slurm log `slurm-82101.log`, retained cache above.
Same admitted limits; estimate **3–6 minutes excluding queue**. Submission is not
validation success. No agent polling; hand off until the terminal notice.

### Verified outcome82101

Producing revision **28c37c7**. [Analysis](runs/checkpoints-01/analysis.json),
[raw results](runs/checkpoints-01/results.json), [Slurm log](slurm-82101.log).
Receipt and all **915 source /87 evidence /390 cache** file seals reconcile.
Controller runtime **2 min41.494 s**, retained logical cache **36.387 MiB**.
HPC allocation above:2nodes/2H200/4cores/8GiB requested RAM,4workers.

| Arm | Launcher runtime | Peak host RSS per worker | Outcome |
|---|---|---|---|
| Pilot cold | 10.400 s | 377.277 MiB | Passed |
| Pilot fresh-process restart | 7.532 s | 379.125 MiB | Passed |
| Assembled durable commits | 9.893 s | 380.570 MiB | Passed |

Owned GPU about **0.244 MiB/worker**, excluding driver/context/MPI. All17commit
manifests/68rank shards,source header/hash/raw archive alias,76baseline comparison
files, ordered proof summaries, native roots/Taylor work, entropy/count/ownership
and all three final payloads match. Fresh-process loader passed all four root and
prefix8 shards; its CPU profile is retained. This is file reload, not arithmetic
restart. Controller and launcher boundaries are described above; no direct speedup
or fullsize peak/runtime inference. Next: [assembled arithmetic resume](../distributed-workflow-resume/report.md).
