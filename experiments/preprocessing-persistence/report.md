# Buffered preprocessing publication and writer concurrency

Status: **bounded HPC job77352 submitted**, results not reviewed (SCALING-119).
Implementation `f493cb9`. [Submission receipt](submission.json); result/profile
location `runs/io-01/`, progress `slurm-77352.log`.
[Actual frozen source check](readiness/snapshot.json):628 build files,2,972,651bytes;
all8 locally checked implementation-input hashes match. No workload execution
is claimed by the snapshot check.
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
not a syscall trace. Timing benefit remains unmeasured and is not inferred from
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
| Direct64 KiB codec writes | Existing bounded codec chunk; no outer buffer | Same canonical bytes, one producer hash and one full readback/hash | Baseline; runtime pending |
|4 MiB outer buffer | Adds up to4 MiB per active writer,16 MiB at4 writers | Adds memory copies/coalescing; no extra transform, network message, payload bytes or readback pass | Runtime/flush/fsync effects pending; experiment-only |
|1→2→4 writers | Separate file owners; buffers scale with active writers | Same four files/96 MiB per trial; concurrent hashing, file writes and metadata sync | Speed or slowdown pending; no default selection |

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
