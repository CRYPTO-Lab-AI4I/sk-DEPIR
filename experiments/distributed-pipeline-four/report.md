# Four-worker complete distributed preprocessing pilot

Status: prepared; GPU/MPI cold/restart validation pending. Parent revisionf153e4f;
actual implementation/producing revision will be recorded with the committed
snapshot and sealed controller results. [Runner](run.py), [Slurm entry](run.sbatch),
[analyzer](analyze.py), [reviewer tests](test_review.py), [readiness](readiness/).
Future results use `runs/pipeline-01/`,with reusable source/binaries/shards/profiles
in `.cache/liftrs/distributed-pipeline-four/pipeline-01/` outside Git.

## Scope and shared implementation

The [genuine discovery/root gate79541](../distributed-discovery-hierarchy/report.md)
passed on four workers sharing two H200s. This gate carries those roots through
the existing Newton reciprocal,complete `P*G` identity,spectrum,coordinated prefix
commits and streamed canonical/generator output,then starts new MPI processes
from completed prefix8. It validates the assembled topology,not fulls18 or node-loss
recovery. The original two-worker pilot/default and ordinary CUDA-free client
build remain available.

The maintained [pipeline](../../encoding/liftrs-shared/distributed_pipeline_gpu_test.go)
selects four workers explicitly with `PIR_PIPELINE_RANKS=4`; it requires the existing
distributed discovery,streamed publication and retained-reference options. The
same hierarchy helper/product driver serves synthetic s18,genuine support and this
complete pilot. No alternate polynomial algorithm or CUDA kernel is introduced.
The hierarchy returns compact coefficient shards;the shared operand router expands
these into full-domain ownership before the existing reciprocal executor starts.
No whole-polynomial gather or extra transform is added. A constant identity input
reuses the existing two-operand router;its unused second output is temporary host
workspace,explicitly counted below.

[Bounded MPI control gathering](../distributed-root/mpi/world.go) handles up to64
fixed64-byte records,ordered by rank pairs. Discovery retains phase/sequence/rank
checks; support binding and completed-prefix digest lists share the same transport.
All ranks save/sync their shards before committing one ordered hash list at rank0,
then a barrier prevents reuse before commit. The trusted manifest binds s,K,N,
rank count,prefix,stage and canonical support digest. Same-topology fresh-process
resume verifies this binding,shard checksums,canonical field values and padding;
it does not change rank count or recover a failed node.

## Inputs and checks

Genuine cases use `Spec{s,2,2}` (two evaluation coordinates,curve degree2),seed109:
s4,N256,K97 and s6,N4096,K1910. Native preflight already measured one discovery
sample for each. Cold execution performs discovery and hierarchical root construction;
resume loads root/prefix8 and skips both. The [retained native input manifest](../distributed-coordinator-stream/reference-inputs.json)
points to verified reference files,whose hashes/size are checked before GPU work
and copied into admitted shared cache for all workers. Per-node preflight verifies
readable inputs,executable/library/wrapper/output paths and environment.

The complete native canonical code and generator remain independent file oracles.
Final output hashes must match the previously retained s4/s6 canonical digests;
the full polynomial identity is checked on every owned coefficient. Every saved
Newton prefix and completed-root commit is checked against native reference bytes,
not only the final output. The same reviewers verify full-domain transpose/Taylor,
Frobenius-square routing,compact-root expansion and streamed publication packets.
The hierarchy reviewer additionally checks all inter-level source bytes and native
leaf/subtree/final polynomials. Discovery sampling,entropy and support ownership
checks stay intact. PID/host records prove a fresh launch; resume bytes and final
artifacts must match cold outputs. Missing/corrupt fourth-rank manifests and shards
are focused reviewer regressions. Checksums detect corruption,not malicious rewrites.

For bounded hierarchy diagnostics only,each worker loads at most1910 support indices
from independently retained native input (about14.9 KiB of principal index slots).
These references never supply discovery samples or owned GPU root inputs;actual
owned roots come from discovery. This explicit test mode adds reference/setup work
and is not the fullsize memory strategy. Retained file oracles/streamed publication
remain enabled;no full generator/spectrum reference is materialized on every rank.

## Resource and memory–time accounting

Unchanged allocation: **two HPC nodes,two H200s,four CPU cores,8 GiB total requested
host RAM**,four workers/two per card;**64 MiB owned device/worker=128 MiB/card**,
**1 GiB RSS/worker**,zero CUDA-pinned;**128 MiB retained cache,2 GiB scratch,32 MiB
per file**,**720-second controller/15-minute scheduler**. Eligible nodes allowed,
gnode03 excluded. Context/driver/MPI internals are additional to owned-device arrays.
No new experiment resources are requested.

Root groups close GPU/communicator resources before the next stage and before
full-domain executor allocation. Expanding compact roots keeps the compact input
plus two full-domain host output shards temporarily:about `4D/r + 8N/r` bytes,
where D is compact domain and r is rank count;the second output is discarded.
This adds bounded host routing/copy/network work,not transforms. Native diagnostic
support/root buffers,Go/runtime/GC and transport storage are additional. Control
metadata is at most4 KiB/rank,with quadratic pair traffic; at four ranks it is256
bytes/rank per gather. Production defaults and faster validated paths are unchanged;
no slower memory-saving mode is silently enabled. Fullsize cost remains unknown.

Estimate **3–5 minutes excluding queue**,including build,preflight,proof review,
profiles and both launches,based on79541's2 min14.6 seconds and earlier two-worker
cold/restart runs. Per-launch220-second test/240-second launcher bounds and overall
720-second cutoff remain;small complete arms are expected well below those limits.
Prior two-worker cold stdout was about1.45 MiB/rank; projected four-worker log volume
is below8 MiB/rank per arm,leaving headroom under32 MiB/file. Retained cache is
projected below32 MiB;exact size is measured. No fulls18 runtime is inferred.

Both cold/resume arms retain CPU profiles and completed evidence before later work.
Timings include reference loading/native checks,discovery,setup,communication,
transforms,proof logging,checkpoint sync/reload and publication;controller runtime
also includes builds and profile summaries. Existing flushed timestamped phase,
sample,tree/transform/pass/tile and resource logs remain;diagnostic logging is
included,with no isolated overhead measurement. No CUDA arithmetic changes: retain
prior clean genuine/hierarchy memchecks instead of another sanitizer-only arm.

## Reproduction and local checks

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. Local checks initialize neither MPI
nor CUDA:compile all interfaces;run genuine native shapes,closed-control admission,
focused ownership/expansion/restart/publication race checks and shared reviewers.
Replay already verified two-worker cold/resume evidence through the generalized
reviewers without scheduler queries or changing saved results. Freeze committed
source and verify all referenced scripts/wrappers/native input manifest before launch.

```sh
python3 -B experiments/distributed-pipeline-four/test_review.py
python3 -B experiments/distributed-pipeline-four/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-pipeline-four/slurm-%j.log \
 experiments/distributed-pipeline-four/run.sbatch pipeline-01
# After completion notice:
python3 -B experiments/distributed-pipeline-four/analyze.py pipeline-01
```

[Tagged build](readiness/build.log), [native shapes](readiness/native.log),
[MPI admission](readiness/mpi-control.log), [race checks](readiness/race.log),
[legacy evidence replay](readiness/legacy-replay.log) and reviewer logs are retained.
GPU/MPI composition remains pending. Fulls18 still needs scalable ownership without
these bounded replicated diagnostics,canonical wide-slot complete-pipeline admission,
representative scaling,resource/headroom review and separate full-run agreement.


[Committed-snapshot preflight](readiness/snapshot-checks.json) passed on **e48a26e**:
806 build inputs/3,703,423 bytes;required wrapper/MPI/pipeline and
retained-native manifest paths exist. All six affected reviewer entry points passed
from frozen source;69,746 retained input bytes were independently copied and
checksum-verified using that snapshot's exact helper. Subsequent readiness commit
changes documentation/evidence only.
