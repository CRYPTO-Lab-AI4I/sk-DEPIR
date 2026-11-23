# Assembled arithmetic resume

Status: **passed job82102**, resume-01; all four arms independently verified. SCALING-192 follows
[passed durable checkpoint82101](../distributed-workflow-checkpoints/report.md#verified-outcome82101).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/).
Results: `runs/resume-01/`; source/binaries/reusable data outside Git:
`.cache/liftrs/distributed-workflow-resume/resume-01/`.
Implementation and producing revision are recorded before submission.

## Runtime and recovery boundary

[Resume runtime](../../encoding/liftrs-shared/distributed_workflow_resume_cuda_mpi.go)
accepts immutable checkpoint directory/label, selected completed prefix and a
trusted source metadata digest. It verifies the complete source header/curve/field,
rank count and prefix extent against requested code parameters. Existing compact
loaders retain bounded metadata, exact EOF, shard checksums, field range, omitted
zero padding and root normalization checks. The existing coordinated support
binding hashes the complete archive once on rank0 and agrees each complete root/
Newton record on all workers; both records must bind the requested source digest.
All this occurs before allocating GPU state. There are no native oracle inputs.
Checksums detect corruption; callers still need trusted immutable source provenance.

Cold and restarted entries use one shared
[Newton/full-identity/spectrum/publication tail](../../encoding/liftrs-shared/distributed_workflow_cuda_mpi.go).
Resume skips discovery and local/hierarchical roots, restores resident root/G shards,
continues after the selected prefix and recomputes the full polynomial identity.
Remaining operation identifiers match the earlier pilot restart. There is no second
numerical implementation. The caller aborts peers on error and owns phase admission,
retention and publication policy. This is not yet a parameterized public command.

The bounded adapter selects `PIR_WORKFLOW_RESUME_FROM` explicitly (prefix8 at s8/
curve5 for this comparison). It obtains the digest from the just-produced, reviewed
cold cache in the same allocation; production must supply trusted saved provenance.
Its cold default and old pilot/ordinary CPU-client builds remain unchanged.
Completed-prefix restart does not recover partial discovery/root construction,
reshard to a different worker count or prove node-failure recovery.

## Archive reuse and memory–time trade-off

[Publication reuse](../../encoding/liftrs-shared/distributed_publication_files.go)
links the already fully verified support file into the fresh output's archive
staging name, replacing only this new writer's empty staging inode. It opens that
archive read-only, prohibits writes, retains its verified streamed digest and
creates a support alias for new checkpoints. Final publication syncs all three
payloads and commits the manifest last as before. This requires same-filesystem
hard links/directory sync and immutable old input throughout the lifetime of both
names. Failed reuse poisons the writer and leaves evidence; no useful cache is
deleted. Hash values are trusted only after the recovery loader's complete check.

No support payload copy or extra cold-path reread; no new numerical arrays,
transforms or GPU transfers. Restart reads two admitted resident shards and one
support archive, then computes only remaining steps. New prefix commits and code/
generator publication still write their outputs; alias names consume metadata,
not duplicate payload blocks. Logical inventory limits count each name again.
Compared with old local verification this explicit assembled recovery uses one
coordinator hash rather than two hashes per rank; measured fullsize hash/storage/
communication time remains unknown. No fullsize latency/RSS saving is claimed.
Slower optional memory policies and existing faster defaults are unchanged.

## Validation and timing boundaries

CPU/race tests cover existing source/compact/legacy loaders, archive no-copy/read-only
behavior, committed hashes and failed reuse extent. Tagged admission checks reject
invalid recovery digest/slot width and insufficient host-shard budget before source
reads. CPU-only native export retains historical bytes and actual s8 reference/root
geometry. Python regressions capture actual nested suite wiring, explicit cold/
resumed environment transitions, and reject incorrect resumed completion counters.

One same-budget allocation runs four arms: reviewed pilot cold/restart, assembled
cold with durable commits, and assembled fresh-process prefix8 arithmetic restart.
Cold remains compared with every prior native/support/root/proof/output check.
Resume must match every remaining proof digest/count/byte with pilot resume, all
native Taylor work, remaining root/Newton/identity files and final native/pilot
canonical outputs; it must perform13remaining Newton steps,161367aggregate upload
words and no root clones. Verify15commit manifests/60shards, source/header binding,
coordinator-only support hashing, no discovery/root work, and same inode across old
support and resumed archive. The existing fresh CPU file reload/profile remains.
No unchanged wide/dense/sanitizer repeat.

Pilot rows precede deferred GPU close; assembled rows include close, checkpoint/
final publication sync. Launcher includes process shutdown, controller includes
builds/native exports/tests/review/profile processing/evidence seals. No direct
speedup inference from these different observer boundaries. Timestamped progress,
completed artifacts and CPU profiles persist on later failure.

## Resources and reproduction

Unchanged admitted HPC allocation: **2 nodes /2 H200 /4 CPU cores /8 GiB requested
host RAM total**,4MPI workers/two per card; **64 MiB owned GPU/worker =128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned; **128 MiB retained cache /2 GiB scratch /
32 MiB per file**, **720-second controller /15-minute Slurm**, any eligible nodes
except gnode03. Context/driver/MPI extra. Go1.27.3, GCC14.2, OpenMPI4.1.8,
CUDA12.9/H200. Same bounded shape s8/2dimensions/curve5,N65536,K7381,seed109,
2discovery samples; seed is test policy only. No fullsize/resource increase.

Estimated controller **3–6 minutes excluding queue**, based on82101's measured
161.494seconds plus one short resume arm and review; projected cache below80MiB.
Frozen-source/native preflight before submission and each-node path/environment
checks before GPU work. No calibration-only job when these measurements suffice.

```sh
python3 -B experiments/distributed-workflow-resume/test_review.py
python3 -B experiments/distributed-workflow-resume/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-workflow-resume/slurm-%j.log \
 experiments/distributed-workflow-resume/run.sbatch resume-01
# After terminal notice:
python3 -B experiments/distributed-workflow-resume/analyze.py resume-01
```

Next: parameterized launch command with complete phase memory/retained-storage/
restart admission, representative larger topology validation and full s18 resource
agreement. **Full s18 preprocessing remains unrun/unadmitted.**

### Frozen-source readiness

Implementation **9006b3b** freezes **920 build inputs**; all **406** compiled code
hashes and every selected build input match the working tree. All **25 frozen
reviewer suites** pass. Ordinary CGO-disabled CPU tests, CPU/race tests and tagged
admission/recovery/publication/reciprocal tests pass without GPU/MPI initialization.
CPU-native preflight verifies **108014 historical bytes** and actual s8/curve5
native/root inputs. [Source checks](readiness/source-checks.json),
[reviewers](readiness/frozen-reviewers.json), [native preflight](readiness/native-preflight.log).
Fresh source builds and each-node environment/path preflight remain inside the job.

### Submission82102

Submitted from **ed99872**, implementation **9006b3b**. [Receipt](submission.json).
Results `runs/resume-01/`, Slurm log `slurm-82102.log`, retained cache above.
Same admitted limits; estimated controller **3–6 minutes excluding queue**.
Submission is not a pass; no agent polling until the terminal notice.

### Verified outcome82102

Producing revision **ed99872**. [Analysis](runs/resume-01/analysis.json),
[raw results](runs/resume-01/results.json), [Slurm log](slurm-82102.log).
All **920 source /105 evidence /476 cache** seals reconcile. Controller
**3 min11.398 s**; retained logical cache **37.851 MiB**. HPC allocation above.

| Arm | Launcher runtime | Peak host RSS per worker | Outcome |
|---|---|---|---|
| Pilot cold | 11.055 s | 381.492 MiB | Passed |
| Pilot restart | 7.692 s | 380.402 MiB | Passed |
| Assembled cold | 10.691 s | 381.152 MiB | Passed |
| Assembled fresh prefix8 restart | 8.297 s | 381.387 MiB | Passed |

Restart completed13Newton steps/161367aggregate upload words without root clones,
matching pilot restart's ordered arithmetic proofs/native Taylor work, all remaining
compact checkpoints/identity/final bytes. Its15commit manifests/60shards and source
bindings pass; support was hashed once by rank0 and reused by hard link. Cold retains
all17commit/68shard and76baseline file checks. Owned GPU about0.244MiB/worker,
excluding driver/context/MPI. Fresh CPU file reload also passed. Timing boundaries
differ; no direct speedup or fullsize inference. Next: [parameterized command](../distributed-command/report.md).
