# Bounded diagnostic serialization

Status: **81744 failed final controller review**; bounded cold/restart and bulk control independently passed. Wide ran in bulk mode; bounded wide validation remains missing. Implementation revision **04af166** (parent 7d774ed).
[Runner](run.py), [Slurm entry](run.sbatch), [analyzer](analyze.py),
[reviewer regression](test_review.py), [local readiness](readiness/).
Planned results: `runs/writes-01/`. Immutable source, binaries, native references
and reusable data remain outside Git in
`.cache/liftrs/distributed-diagnostic-writes/writes-01/`.
Exact implementation and producing revisions are recorded before submission.

## Scope and preserved behavior

[81738 passed](../distributed-proof-recording/report.md#verified-result81738):
all six cold/restart/wide trace/summary arms matched ordered evidence, independent
native Taylor totals and canonical outputs. The next recorded persistence gate
is the diagnostic writer's second full-size serialization array, identified in
the [allocation audit](../distributed-fullsize-audit/report.md).

The shared [diagnostic codec](../../encoding/liftrs-shared/diagnostic_write.go)
uses one reusable buffer of at most **64 KiB**, standard little-endian primitives
and exact write-length/error checks. Both uint32 and uint64 retain their complete
range. The [existing diagnostic saver](../../encoding/liftrs-shared/distributed_vectors_gpu_test.go)
selects this codec only with explicit `PIR_PIPELINE_DIAGNOSTIC_WRITE=bounded`.
Unset/`full` preserves bulk `binary.Write`; unknown values fail before file creation.
File names, ordering, width, contents, exclusive creation, permissions, sync and
close remain unchanged. This is an experimental environment option on the bounded
runner, not a new public production entry point.

The bounded path logs filename/stage, timestamp, completed/total serialized bytes,
elapsed time and RSS at start/end and roughly every 30 seconds of advancing writes.
Serialization completion precedes the unchanged fsync/close; it is not a durable
checkpoint commit. The ordinary per-worker RSS guard also runs at these updates.
Write errors remain visible and fail the arm. No heartbeat is called progress.

## Memory, time and storage trade-off

For a source array of B bytes, the extra serializer changes from B bytes to at
most 64 KiB. At the illustrative s18/16-worker layout, a base snapshot's temporary
buffer projects **16 GiB -> 64 KiB per writer** (a wide snapshot could be 32 GiB).
These are buffer inventories, not measured RSS bounds; source arrays, Go/native
runtime, GC and other phases remain. Existing compact checkpoints/canonical
streaming are unchanged. No diagnostic files are omitted, and no useful caches
or profiles are deleted. Total retained bytes and logical I/O are unchanged.

There are no extra transforms, device transfers or protocol computation. Writes
increase from one bulk payload to ceil(B/64 KiB) system calls (262,144 for 16 GiB),
with repeated encoding loops/progress checks. **Large-file throughput and peak RSS
are unmeasured**; no cost-free or full-size speedup claim follows. Default bulk mode
remains available whenever admitted memory fits. The bounded mode is an explicit
memory-constrained option; this validation measures its cost rather than silently
adopting it. It does not fix the audit's 1.25 TiB diagnostic retention, support
rehashing, production rank/lifecycle or full native-oracle dependence.

## Validation and measurement boundaries

Pure CPU tests compare every byte with independent `binary.Write` for both widths,
multiple chunks, empty arrays, high bits and maximum uint64. They check fixed write
bounds, actual progress totals, invalid modes, short writes and propagated errors.
The fixed-input serialization benchmark uses 1,048,576 words of each width,
`io.Discard`, three iterations, allocation reporting and no profiling; it isolates
serialization CPU/allocation cost, excluding disk, fsync, CUDA, MPI and progress.
See [benchmark output](readiness/serialization-benchmark.log). HPC login-node CPU
microbenchmarks cannot establish allocated-node disk throughput or peak RSS.

The GPU allocation contains four profiled arms:

1. Genuine s8/curve5 bounded writes, cold discovery through canonical publication.
2. Fresh-process prefix8 restart from that verified cold cache.
3. Matched cold bulk-write control with every other execution flag unchanged.
4. Existing bounded field18 product/reciprocal/spectrum/high-bit checks with bounded writes.

Both codecs use full trace plus ordered proof summaries (the validated trace mode),
so the maintained independent packet reviewers remain usable directly. No proof
rows are synthesized. Every runtime descriptor/payload/peer/field/identity check
remains. Compare canonical artifacts, diagnostic bytes against independent native
references, restart state and expected native Taylor windows. Separately validate
logged file sizes/completed totals and reject failures/incomplete progress.
Identical arithmetic needs no new sanitizer-only arm; prior clean arithmetic
coverage is retained, not claimed as new coverage of this composition.

Matched arm runtime includes setup, CPU/GPU transforms, MPI, serialization,
checkpoint/hash/sync/reload, canonical publication and validation. Controller adds
builds, native export, tests, independent review and profiles. Fixed arm order and
one sample are not a statistical performance claim. Profiles/completed results
persist before later work.

## Resources and reproduction

Unchanged admitted HPC resources: **two nodes/two H200s/four CPU cores/8 GiB
requested host RAM total**; four MPI workers/two per card; **64 MiB owned GPU per
worker = 128 MiB/card**, **1 GiB RSS guard per worker**, zero CUDA-pinned buffers;
**128 MiB retained cache/2 GiB scratch/32 MiB per file**, **720-second controller /
15-minute Slurm limit**. Eligible nodes excluding gnode03. Context/MPI/driver memory
is additional to owned bytes. No resource increase or full s18 admission.

Estimate **3–6 minutes excluding queue**, based on 81729's 191-second/four-arm and
81738's 199-second/six-arm controllers. Retained cache is projected below 64 MiB;
full traces below 12 MiB per rank file, within the existing file limit. Full-size
runtime remains unknown. CPU/native/source checks precede submission; the admitted
allocation rebuilds committed source and checks input/library/wrapper/cwd/access
on every node before expensive phases. HPC Go1.27.3, GCC14.2, OpenMPI4.1.8 and
CUDA12.9/H200; ordinary Go builds still require no CUDA.

```sh
python3 -B experiments/distributed-diagnostic-writes/test_review.py
python3 -B experiments/distributed-diagnostic-writes/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-diagnostic-writes/slurm-%j.log \
 experiments/distributed-diagnostic-writes/run.sbatch writes-01
# After a terminal notice:
python3 -B experiments/distributed-diagnostic-writes/analyze.py writes-01
```

After verification, update the primary scaling reference and continue production
parameter/rank/lifecycle ownership, non-oracle execution, bounded retention and
coordinated restart binding; admit real phase counts/domains before any full s18 job.

### Local serialization measurement

HPC login CPU: Intel Xeon Silver 4516Y+, linux/amd64, Go1.27.3. Three iterations
per fixed input, no profiler. This is allocation/serialization evidence only:

| Input | Bulk CPU time | Bounded CPU time | Bulk allocated bytes/op | Bounded allocated bytes/op |
|---|---|---|---|---|
| 1,048,576 uint32 words (4 MiB) | 3.485 milliseconds | 1.195 milliseconds | 4,194,304 | 65,536 |
| 1,048,576 uint64 words (8 MiB) | 3.957 milliseconds | 1.622 milliseconds | 8,388,608 | 65,541 |

The bounded wide result includes about five bytes/op of amortized runtime overhead
in addition to its 64 KiB buffer. These samples do not establish disk throughput,
peak RSS or full-size speedup. [Exact outputs/commands/source hashes](readiness/checks.json).

```sh
CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
 .cache/liftrs/toolchains/go1.27.3-linux-amd64/go/bin/go test \
 ./encoding/liftrs-shared -run '^TestDiagnosticWrite$' \
 -bench '^BenchmarkDiagnosticWrite$' -benchtime=3x -count=1 -timeout=20s
```

Initial use of a newer codec API failed the module's Go1.21 compatibility check;
the final implementation uses existing endian primitives without raising the
supported Go version. No GPU allocation was used to discover or resolve that issue.

### Committed-source readiness

Implementation **04af166** was frozen into **857 build inputs**. All **385**
Go/C/C++/CUDA source hashes match the compiled readiness tree, all **15 frozen
Python reviewer suites** passed, and CPU-only native preflight reproduced
**108014 historical reference bytes** plus actual s8/curve5 inputs and owned-root
references. [Source checks](readiness/source-checks.json),
[frozen reviewers](readiness/frozen-reviewers.json),
[native preflight](readiness/native-preflight.log). No local CUDA/MPI initialization.
The allocation rebuilds the committed snapshot and preflights every node.

### Submission

Job **81744** submitted from **c30ef0d**, unchanged resources above.
[Receipt](submission.json); results `runs/writes-01/`, Slurm log `slurm-81744.log`.
No scheduler polling or GPU result claim follows submission.

## Failure81744 and targeted correction

The allocation completed every GPU command and saved profiles/artifacts, then the
final Python reviewer failed. The shared launcher carries arm environment overrides
forward. The bulk control set `PIR_PIPELINE_DIAGNOSTIC_WRITE=full`; the following wide
arm omitted an explicit reset and therefore also ran bulk. The reviewer correctly
rejected its absent bounded-write progress. This was an avoidable experiment
configuration error, not a GPU arithmetic, resource-guard or encoding failure.

[Original raw results](runs/writes-01/results.json) remain **failed**, unchanged;
[Slurm traceback](slurm-81744.log), [independent partial analysis](runs/writes-01/analysis.json).
All **857 source / 96 evidence / 496 cache** seals reconcile to producing revision
**c30ef0d**. Every recorded command exited zero without a resource guard failure.
Independent packet/digest/native/canonical/restart checks pass for all four arms;
the historical wide result establishes **bulk arithmetic only**, not bounded writes.

| Arm | Independent result | Runtime | Peak host RSS |
|---|---|---|---|
| Bounded s8 cold | Passed; 32 diagnostic files / 989,708 bytes logged exactly | 9.135297 seconds | 376.750 MiB/worker |
| Bounded s8 prefix8 restart | Passed; 16 files / 845,480 bytes logged exactly | 6.804956 seconds | 379.473 MiB/worker |
| Matched s8 bulk control | Passed; canonical artifacts identical to bounded cold | 9.338670 seconds | 378.961 MiB/worker |
| Field18 wide, actual bulk | Arithmetic passed; bounded write validation missing | 4.929941 seconds cumulative | 379.082 MiB/worker |

Same admitted HPC resources; owned GPU peaks unchanged (0.244 MiB/worker s8,
3.125 MiB/worker wide). Controller **2 min 55.735 seconds**, retained cache
**37.621 MiB**. Single cold comparison is not a full-size throughput claim.
Zero logged files for bulk means bounded instrumentation was off, not zero files
written or retained. All native/canonical diagnostic files remain available.

`comparison_arms()` now explicitly selects bounded mode for wide. A focused test
executes the actual suite configuration and simulates successive environment
updates, so inheriting bulk mode fails before an allocation. The targeted retry
configuration has its own regression. Existing progress validation still rejects
missing, incomplete or failed bounded writes. Go/CUDA arithmetic and writer code
are unchanged; no repeated cold/restart or sanitizer run is needed.

The [targeted runner](run-wide.py) and [Slurm entry](run-wide.sbatch) execute only the
missing bounded wide arm, with the same maintained node/path preflight, resource
caps, CPU checks, profiles and independent native/digest/byte verification. It
rebuilds committed source and creates a separate cache/results directory; the
failed run and its passing phases remain intact. No larger budget is requested.
Estimate **2–4 minutes excluding queue**, with the existing 720-second controller /
15-minute scheduler limits and all memory/storage caps unchanged. It does not
claim full s18 preprocessing or wide full-domain scaling.

```sh
python3 -B experiments/distributed-diagnostic-writes/test_review.py
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/distributed-diagnostic-writes/slurm-%j.log \
 experiments/distributed-diagnostic-writes/run-wide.sbatch wide-02
# After a terminal notice:
python3 -B experiments/distributed-diagnostic-writes/analyze.py wide-02
```

The corrected analyzer recognizes only the exact immutable 81744 producing revision
as this known partial failure, retains its failed status, and labels its wide mode
as bulk. It never infers bounded success from missing instrumentation. A successful
targeted retry must provide complete bounded logs and pass all independent checks.

### Targeted retry readiness

Correction implementation **845c3eb** is frozen and passes all three focused
regressions, including actual suite environment updates and the one-arm retry.
The **385 Go/C/C++/CUDA sources** and prior compiled readiness binary are unchanged;
prior native compatibility evidence remains applicable. [Frozen checks](correction/source-checks.json),
[test output](correction/frozen-review-tests.log). The GPU allocation still performs
its normal fresh committed-source build and every-node path/library preflight.
No extra local GPU/MPI execution or repeated arithmetic calibration was needed.
