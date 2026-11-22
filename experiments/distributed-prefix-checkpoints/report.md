# Compact restart prefix checkpoints

Status: prepared; GPU cold/restart integration pending. Parent revision525888f.
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer tests](test_review.py), [readiness](readiness/). Exact producing revision
will be sealed by the committed snapshot/controller. Results use `runs/prefix-01/`;
reusable source,binaries,profiles and shards remain outside Git under
`.cache/liftrs/distributed-prefix-checkpoints/prefix-01/`.

## Why this is the next fullsize boundary

[Owned-root execution81079](../distributed-owned-roots/report.md) passed complete
genuine cold/restart and narrow/wide root comparisons. Restart persistence still
writes a complete N/r coefficient shard for every completed reciprocal prefix,
including its known zero padding. Repeating those full files would be an avoidable
fullsize storage/I/O cost.

The optional `PIR_PIPELINE_CHECKPOINT_LAYOUT=prefix` stores only the intersection
of each rank's ownership band with the completed global prefix. Base coefficients
retain4-byte little-endian slots. This is zero-padding omission,not bit packing
or compression of actual coefficient values. Final canonical code/database encoding
and GPU arithmetic/layout are unchanged. The existing padded layout stays default.

[Checkpoint metadata and codec](../../encoding/liftrs-shared/distributed_recovery.go)
add an optional `layout: "prefix"`;missing layout remains legacy padded format.
The existing s,K,N,prefix,stage,rank count,support digest and ordered payload hashes
bind the payload's meaning. Exact stored length is derived from these fields;
no redundant extent list is trusted. A padding-only rank writes a real empty file
whose digest is SHA-256 of empty data. It still participates in the coordinated
save/sync/commit/barrier,so missing rank output is never mistaken for intentional
omission. Unknown layouts,truncated/extra bytes,wrong rank/shape/bounds,corrupt
hashes,out-of-field values or nonzero omitted padding are rejected.

Compact restore streams through64KiB into one admitted zero-initialized resident
shard,with exact EOF/hash/range checks. The omitted tail stays zero. This avoids
a second full file-sized restore buffer. Legacy padded reads retain their previous
bulk-I/O path and default behavior;in-process reload shares the same decode checks
without adding another scan of global support. Fresh-process support verification
is unchanged. Checksums detect corruption,not malicious manifest rewrites.

## Storage projection and runtime trade-off

Assume N=2^(2s),0<K<N/2,the existing doubling schedule starts at1,and every completed
root/prefix is retained once. There are2s Newton prefixes plus the root. Padded
payload is `4N(2s+1)` bytes. Complete power-of-two prefixes sum to N−2;adding the
final N−K prefix and K+1 root gives2N−1 coefficients. Compact payload is therefore
`4(2N−1)` bytes,independent of K under this strict assumption. Metadata and all other
artifacts are additional. K=N/2 has a smaller exact schedule;derive actual sizes
from its metadata instead of applying the strict-inequality formula.

| s | Padded restart payload projection | Compact restart payload projection |
|---|---|---|
| 16 | 528GiB | approximately32GiB |
| 17 | 2.1875TiB | approximately128GiB |
| 18 | 9.25TiB | approximately512GiB |
| 19 | 39TiB | approximately2TiB |
| 20 | 164TiB | approximately8TiB |

These are mathematical payload estimates for the distributed pilot format,not
measurements or full-run storage admission. They do not describe the historical
s17 preprocessing runner's checkpoint implementation. Retained code/support,
generator/spectrum,debug snapshots,root-stage outputs,profiles,temporary data and
filesystem overhead are excluded. Current validation still keeps diagnostic
padded P/G/identity files;its complete cache is not simply this payload formula.
Nothing is deleted or overwritten for storage savings.

Runtime costs are explicit:compact saving checks the omitted tail for zeros and
hashes/writes only stored values;restore uses64KiB reads instead of a bulk file
buffer. Disk traffic and restore auxiliary memory decrease,but extra CPU checking
and small-read syscall/remote-filesystem throughput costs are **unmeasured at
fullsize**. Compact mode is explicitly selected for this experiment;no slower
memory-saving production default is adopted. Keep padded mode when its measured
cost/available resources favor it. This experiment validates correctness/storage,
not a large-file runtime speedup. Three-array GPU memory and resident coefficient
ownership are unchanged.

## Validation and resources

Reuse the [owned-root four-arm suite](../distributed-owned-roots/run.py):genuine
s4/s6 compact cold execution and fresh-process prefix8 restart,plus existing
bounded field18 adapter and owned-root regressions. All native leaf/stage/final,
canonical/generator,completed-prefix,full identity,spectrum and MPI packet checks
remain. Independent Python checkpoint expansion verifies each exact stored band,
hash and full canonical prefix against native bytes. Source/resume payloads and
final artifacts must match;PIDs/hosts prove a new process. Controller records exact
checkpoint bytes,padded equivalents,zero-length rank files and savings separately
from complete cache size.

Expected cold checkpoint payloads:s4 **2,044 bytes versus9,216 padded bytes**;
s6 **32,764 bytes versus212,992 padded bytes**. Expected resumed root/prefix8-suffix
payloads:s4 **2,020 versus7,168 bytes**;s6 **32,740 versus180,224 bytes**. These are
predictions from the measured K values and schedule;actual sizes will be checked.

Same admitted HPC allocation:**two nodes/two H200s/four CPU cores/8GiB total host
RAM requested**,four workers/two per card;**64MiB owned GPU/worker=128MiB/card**,
**1GiB RSS/worker**,zero CUDA-pinned;**128MiB retained cache,2GiB scratch,32MiB/file**;
**720-second controller/15-minute Slurm cutoff**. Eligible nodes allowed,gnode03
excluded. Context/driver/MPI internals are additional to owned memory. No resource
increase or GPU-count design ceiling is implied. Root/native input checks reuse
verified112,460-byte stage references and69,746-byte code inputs.

Estimate **3–6 minutes excluding queue**,based on81079's2min30.945s for these four
small arms. New codec latency is unmeasured. Cache projected below32MiB /logs below
8MiB per rank per arm,inside existing caps. Completed arms/profiles persist before
later work. Timings include save/hash/sync/reload,reference/packet validation,
setup/communication and publication;controller adds build/preflight/profile review.
No CUDA arithmetic changes or repeated sanitizer-only arm;prior clean coverage is
retained. Existing timestamped phase/prefix/group/pass/tile and resource logs remain;
proof logging is included,with isolated logging overhead unmeasured.

## Local checks and reproduction

HPC Go1.27.3,GCC14.2/OpenMPI4.1.8,CUDA12.9,H200. Ordinary CPU codec tests require
neither CUDA nor MPI. Round trips cover partial/empty rank payloads,zero-width
ownership,bounds,unknown layout,hash corruption,extra bytes,field range and omitted
nonzero padding. Independent reviewer tests compare compact and legacy artifacts.
Local tests caught an incorrect zero-width fixture and a digest-type compile error;
both were corrected before allocation. Their initial logs are retained separately
from final passing readiness. Replay81079's complete saved evidence through the
backward-compatible reviewers. Freeze committed source and verify exact code,
wrapper/input manifests and both retained reference sets before submission.

```sh
python3 -B experiments/distributed-prefix-checkpoints/test_review.py
python3 -B experiments/distributed-prefix-checkpoints/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-prefix-checkpoints/slurm-%j.log \
 experiments/distributed-prefix-checkpoints/run.sbatch prefix-01
# After completion notice:
python3 -B experiments/distributed-prefix-checkpoints/analyze.py prefix-01
```

[CPU](readiness/cpu.log), [race](readiness/race.log), [build](readiness/build.log),
[reviewer](readiness/review.log). GPU integration pending. This closes only a restart
payload/storage boundary;fulls18 still needs complete discovery/working-set admission,
full wide-domain execution/artifacts,representative larger runtime/I/O/restart
measurements and separate resource/storage/headroom/deadline agreement.
