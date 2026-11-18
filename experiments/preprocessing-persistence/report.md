# Buffered preprocessing publication and writer concurrency

Status: **bounded HPC job77352 passed**, reviewed as SCALING-120.
HPC gnode02, Intel Xeon6710E,2 allocated CPU cores/4 GiB requested RAM,0 GPUs;
NFS filesystem (`stat -f %S` reports1 MiB filesystem block size, not measured RPC size).
Scheduler elapsed **1 minute45 seconds**; controller elapsed **103.350 seconds**.
Exact source `f493cb977021661bb60f6d2d9a18d07f0540c14e`.
[Submission receipt](submission.json), [raw results](runs/io-01/results.json),
[verified analysis](runs/io-01/analysis.json), [accounting](runs/io-01/accounting.txt),
[Slurm progress log](slurm-77352.log).
[Actual frozen source check](readiness/snapshot.json):628 build files,2,972,651bytes;
all8 locally checked implementation-input hashes match. The separate completed
job review below verifies all source/evidence/cache seals.

[Runner](run.py), [scheduler script](run.sbatch), [completion reviewer](analyze.py),
[review tests](test_review.py),
[opt-in probe](../../cmd/pir-preprocess/persistence_probe_test.go),
[shared payload writer](../../cmd/pir-preprocess/artifacts.go),
[optional counters](../../cmd/pir-preprocess/payload_buffer.go),
[buffer/failure tests](../../cmd/pir-preprocess/payload_buffer_test.go).
This is the separate persistence gate after
[passed two-node GPU root/merge validation](../distributed-root/report.md).
No GPU computation is repeated or reserved for this I/O-only measurement.

## Shared implementation and correctness

The CLI's existing `writePayload` now delegates to the same implementation with
buffer size zero and no counters. That preserves the ordinary codec write sizes,
ordered checksums and format. Only the explicit experiment calls
`writePayloadBuffered` with an additional4 MiB buffer; there is no new CLI default
or public API. Codecs remain canonical little-endian; no padding or new record
layout is introduced. A producer error or buffer-flush error prevents fsync and
publication. Successful writes flush before file fsync/close, then use the existing
manifest fsync, directory fsync and atomic rename protocol. File-write timing/
counters are opt-in; normal writes do not collect per-call statistics.

The probe streams deterministic synthetic s=17 base-field words in64 KiB chunks.
It generates public xorshift32 words (seed=shard+1), masks them to17 bits and stores
uint32 little-endian. These are **not genuine root/code/checkpoint artifacts**.
Each custom `io-trial-XX-file-Y` stage uses the real publication/manifest/readback
helpers without claiming a mathematically verified preprocessing state. No
intermediate computation or cross-node exchange uses filesystem arrays.

Every trial writes the same four separate **24 MiB** payloads, **96 MiB total**.
One, two and four writer goroutines divide those four files deterministically:
`owner=shard % writers`. Exactly one goroutine owns each payload/manifest/stage;
all are threads within one process, never multiple processes writing one file.
Each writer finishes its assigned files serially. Different stage directories
share only parent-directory metadata synchronization. Readback occurs after all
writers finish, serially, with fresh manifest/payload opens and full SHA256 checks.
Expected shard hashes are generated before timing and must match every mode.

Compare direct64 KiB codec writes (buffer=0) with4 MiB coalescing, at each writer
count. Fixed order is `(direct,1),(buffered,1),(buffered,2),(direct,2),(direct,4),
(buffered,4)`, then its reverse:12 trials, two samples/configuration, no warm-up.
Total payload writes/readback are **1.125 GiB each**. There are48 publications,
not a single huge streamed file. Expected aggregate application file-write calls
are1,536/trial direct versus24/trial buffered, a64-fold call-count reduction.
Actual call sizes/counts/bytes are checked; these are `os.File.Write` observations,
not a syscall trace. Timing benefit is mixed in the completed run below and is not inferred from
call-count reduction. Two CPU cores and data-generation/hash work may limit this
probe; CPU profiles must inform interpretation of writer concurrency.

Small local checks cover equal bytes/digests with a non-aligned tail, propagation
of producer errors, rejected oversized buffers, no publication after interruption,
retry/reopen, corruption rejection and existing publication behavior. The scheduled
job additionally runs genuine small CLI recovery/restart and independent Go
arithmetic-reference checks. A separate race-enabled test binary repeats the full
12-case concurrency probe with128 KiB/file plus failure/equivalence unit checks;
its final512 KiB payload set and matching executable are retained. The Linux-only probe is opt-in via
`PIR_PERSISTENCE_OUTPUT`; ordinary tests skip it. CPU builds acquire no CUDA/MPI
dependency. This persistence test is not a new CPU preprocessing fallback.

## Memory, time, storage and measurement boundaries

| Change | Memory | Computation/transfer/I/O | Cost/status |
|---|---|---|---|
| Direct64 KiB codec writes | Existing bounded codec chunk; no outer buffer | Same canonical bytes, one producer hash and one full readback/hash | Measured baseline below |
|4 MiB outer buffer | Adds up to4 MiB per active writer,16 MiB at4 writers | Adds memory copies/coalescing; no extra transform, network message, payload bytes or readback pass | Measured effects below; experiment-only |
|1→2→4 writers | Separate file owners; buffers scale with active writers | Same four files/96 MiB per trial; concurrent hashing, file writes and metadata sync | Mixed measured effects below; no default selection |

Synthetic data generation and ordered write hashing are inside publication
clocks, along with file writes, buffer flush, fsync, close, manifest writes, directory
sync and rename. Each file records publication and readback latency; the trial
records joined publication wall time, serial full readback and their complete
wall time. Per-file values overlap across writers and must not be added to derive
trial wall time. `produce_flush_seconds` includes serialization, producer hashing,
file writes and final flush; `file_write_seconds` and `flush_seconds` are nested
subintervals, not additive costs. File fsync and close clocks are separate.
Profile/report/compilation and precomputed expected hashes are outside trial
clocks, inside complete job duration. Profiling/logging/counter overhead is included;
there is no unprofiled latency baseline or material default change.

This bounded workload fits host RAM. The test does not evict global caches, use
direct I/O, establish cold reads, saturate storage or infer terabyte-scale
throughput. Fresh filenames avoid overwriting earlier cases but do not guarantee
cold filesystem metadata/data. File fsync/rename guarantees depend on the shared
filesystem; no crash/power-failure durability claim is added. Multi-node writers
and a full distributed compute/checkpoint schedule remain unmeasured.

Timestamped JSON identifies trial, shard, phase, bytes/files completed and elapsed
time. Shared codecs emit bounded byte progress after30 seconds; each completed
file and readback is logged. A blocked write/fsync cannot be interrupted by a
Go byte deadline; controller/test/scheduler deadlines apply. Controller heartbeat
marks unknown remaining work and is not called progress. There is no unsupported
ETA. A CPU profile and matching executable are retained.

Process-group RSS and final retained payload-file cache residency are sampled
roughly every250 ms using the existing Linux `mincore` observer. Only the final
case is mapped, so observation cannot keep pruned transient-case inodes alive and
increase actual storage use. Mappings/descriptors close after each observation;
final retained residency is also sampled after worker exit. Earlier transient
payload residency is not measured separately. Sampled RSS+retained-payload cache
is a partial proxy, not total job memory. cgroup memory/current/peak/anonymous/file charges are recorded only when a
job-specific cgroup is available, otherwise unknown. Build caches/libraries and
other file cache are not counted by the payload-residency measure. Guards separate
2 GiB process-group RSS and3.5 GiB sampled cgroup charge beneath4 GiB requested RAM.
Sampling can miss short peaks; process high-water RSS is also logged by the probe.

Retained storage peaks at one96 MiB payload set plus executable/manifests/profiles,
under128 MiB. After each case is fully readback-verified and logged, only that
case's new, cheap synthetic payloads are deleted. All48 manifests, digests and
measurements remain, and the final complete96 MiB payload set is retained with
both matching normal/race binaries and the small race payload set. Earlier cases are intentionally not resumable preprocessing
caches. No useful code preprocessing, previous cache, failed partial publication
or profile is removed. The runner's final seals identify every retained file.

## Resource admission and reproduction

This reduces the existing approved two-node GPU envelope; no increase:

| Resource | Allocation/guard |
|---|---|
| HPC | 1 node on gpuq,2 CPU cores,4 GiB requested RAM,0 GPUs |
| Process/group RSS / job cgroup charge | 2 GiB /3.5 GiB when available |
| New retained cache / local build scratch | 128 MiB /2 GiB |
| Per-file cap | 32 MiB; each synthetic payload24 MiB |
| Test / worker / controller / Slurm deadlines | 520s /540s /720s /900s |

Modules are Slurm25.05 and GCC14.2; Go1.27.3 is the retained HPC toolchain. The
runner freezes tracked Git build inputs and records its exact source revision/
hashes, host/CPU and filesystem type/block-size diagnostics. It builds a native
CPU-tagged CLI test binary for the small correctness references, not for CPU
preprocessing benchmarks. No CUDA/MPI library or GPU is used.

```sh
python3 -B experiments/preprocessing-persistence/test_review.py
python3 -B experiments/preprocessing-persistence/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/preprocessing-persistence/slurm-%j.log \
  experiments/preprocessing-persistence/run.sbatch io-01
# Only after the user's completion ping:
python3 -B experiments/preprocessing-persistence/analyze.py io-01
```

Results, profiles and logs go to `runs/io-01/`; progress/Slurm output to
`slurm-JOBID.log`. Public synthetic payloads/manifests and matching binary remain
outside Git at `.cache/liftrs/preprocessing-persistence/io-01/`. Regenerate with
the same committed runner and a fresh name. Local small smoke/readiness logs are
in [readiness checks/source hashes](readiness/checks.json),
[build log](readiness/build.log), [correctness log](readiness/correctness.log) and
[small smoke log](readiness/smoke.log); they exercise128 KiB/files and are correctness checks, not HPC
shared-storage performance measurements.

Use measured full publication/readback and memory costs to decide which buffer/
writer settings warrant integration into the distributed producer. The faster
validated resident compute path remains default. A larger sustained I/O or
multi-node trial needs its own concrete resource review before full s=18
publication/runtime estimates or a production-default change.

## Completed persistence review (SCALING-120)

Completion verification passed **628 frozen source files,14 evidence files and
106 retained cache files**, without overrides or rewriting producer results.
All11 commands exited zero, with no resource/deadline guard. Normal12-trial and
small race12-trial arrays reconcile: fixed mode order, distinct logical file
owners, repeated per-shard digests, retained manifests, final payload hashes,
byte counts and exact application write-call histograms. Buffered tails, producer
errors, failed publication/retry, corrupt readback rejection, genuine small CLI
recovery/restart and independent Go-reference checks pass. The race-enabled binary
also executes the concurrent1/2/4-writer cases, not only serial unit tests.

All cases below write96 MiB across four24 MiB files. Medians use **two samples**
in mirrored order, no warm-up. Publication includes generation/hash, payload writes,
flush/fsync/close, manifest/directory sync and atomic rename. Readback is serial,
includes fresh opens and full checksums, and may use filesystem cache. CPU profile,
file-call counters and job-side resource observation are enabled. No controlled
unprofiled baseline, cold-cache or sustained-storage claim follows.

| Write mode | Writers | Publication | Full readback | Publication + readback | Payload file-write calls per trial |
|---|---|---|---|---|---|
| Direct64 KiB codec writes | 1 | 0.336030 s | 1.296711 s | 1.632754 s | 1,536 |
| Direct64 KiB codec writes | 2 | 0.268352 s | 1.042680 s | 1.311052 s | 1,536 |
| Direct64 KiB codec writes | 4 | 0.198815 s | 1.284555 s | 1.483389 s | 1,536 |
|4 MiB coalescing | 1 | 0.346113 s | 0.895899 s | 1.242028 s | 24 |
|4 MiB coalescing | 2 | 0.212306 s | 1.391488 s | 1.603806 s | 24 |
|4 MiB coalescing | 4 | 0.184447 s | 1.459786 s | 1.644246 s | 24 |

The64-fold application-call reduction is verified, but **elapsed benefit is not
consistent**. Single-writer publication is slightly slower buffered;2/4-writer
publication medians improve, while complete readback-inclusive medians do not
follow the same ordering. Individual publication trials range0.150–0.389s;
readback0.863–1.889s. With only two samples/configuration and small cached files,
these results do not select a buffer or concurrency default or prove a full-size
speedup. The existing unbuffered producer/default remains unchanged.4 MiB buffers
add up to16 MiB at4 active writers plus copying; there are no added transforms,
network transfers, payload bytes or readback passes.

Full readback exceeds publication in every trial and is the largest measured
component. Investigate coalesced verification reads during pipeline integration
while preserving every checksum, shape and source-binding check. This is a
candidate, not a measured speedup or justification to omit verification. The
[CPU profile](runs/io-01/persistence.pprof) and
[summary](runs/io-01/profile.log) cover18.24s wall time/4.96s sampled CPU: syscall
frames58.67% and SHA25627.22% of sampled CPU, with generation/encoding/memory copy
work also visible. CPU sample proportions are not wall-time attribution and do
not isolate network/storage latency. Job-time build/profile costs are separate:
normal build33.891s, race build19.076s, full persistence command18.308s and profile
reporting21.134s, all included in the105s scheduler duration.

| Memory/storage measure | Observed value |
|---|---|
| Persistence command sampled process-group peak RSS | 41.137 MiB |
| Probe process high-water RSS at measurement events | 40.418 MiB |
| Final retained payload residency | 96.000 MiB |
| Sampled concurrent RSS + retained-payload cache proxy | 109.609 MiB |
| Persistence-phase max sampled cgroup current charge | 362.695 MiB |
| Persistence-phase max sampled cgroup file charge | 297.734 MiB |
| Whole job max sampled cgroup current charge | 706.316 MiB |
| Job cgroup high-water `memory.peak` | 718.566 MiB |
| Whole job max sampled process-group RSS (compilation) | 587.570 MiB |
| Sealed retained cache,106 files | 111.946 MiB |

Process high-water RSS and sampled process-group RSS are different observations;
they are not interchangeable with cgroup memory. File charge includes build,
binary and other cache as well as payload data. RSS+retained-payload residency
omits other cache and earlier transient payload residency; do not add its
independent peaks or present it as total memory. Cgroup `memory.peak` covers the
whole job, including earlier compilation, not a persistence-only peak. During
persistence, sampled current charge stays below363 MiB, with substantial headroom
under3.5 GiB cgroup and2 GiB process-group guards/4 GiB requested host RAM. Fast
diagnostic commands sometimes finish between RSS samples; recorded zeros are
missed observations, not zero memory use. Scheduler MaxRSS is separate accounting.

The final96 MiB normal payload set,512 KiB small race set, all96 manifests and
matching normal/race executables remain at
`.cache/liftrs/preprocessing-persistence/io-01/`, with seals in raw results.
Only prior verified synthetic payloads from this probe were pruned as planned.
No expensive code preprocessing/profile/older cache was deleted. Raw logs,
[resource observations](runs/io-01/resources.jsonl), profile and bytewise source/
evidence/cache provenance are retained. Three persistence reviewer regressions
and existing five staged/three shared-review regressions pass; completion analyzer,
Python/shell syntax, local doc links and edited-text whitespace checks pass.
No new job or production-mode change was made in this review.

The bounded persistence gate is complete. **Next: scalable support/rank ownership
and a multi-level GPU polynomial merge**, then full generator/Newton/transform
integration and rank-owned checkpoint/final publication. Integrate the existing
atomic/checksum helpers and explicit writer ownership; keep compute exchanges in
RAM. Carry the readback/copy/fsync observations into those complete workflow
measurements. Fulls18 memory/runtime, multi-node persistence, optimal GPU/node
count and any production buffer/concurrency choice remain unvalidated. Any larger
or multi-node I/O trial needs explicit resource admission; this result is not
permission to expand the128 MiB retained-storage cap.
