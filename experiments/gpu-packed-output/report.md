# GPU output to a bounded packed sink

Status: GPU-038 implementation and compile checks complete; job74091 failed a small multi-chunk parity oracle; fix/retry runtime validation
is pending. This is the first of the four remaining gates authorized by the user:
GPU output sink, GPU host-backed transforms, CPU output sink, CPU disk-backed
transforms. Later dependent gates wait for this validation's completion ping.
None is silently treated as finished by a component benchmark.

## Direct packed output

[CUDA implementation](../../encoding/liftrs-cuda/matrix.cu),
[ABI](../../encoding/liftrs-cuda/gpu_api.h),
[low-level binding](../../encoding/liftrs-cuda/bridge.go),
[shared API](../../encoding/liftrs-shared/cuda_packed.go) and
[CPU-only stub](../../encoding/liftrs-shared/cuda_packed_unavailable.go) add an
explicit `EncodePackedTo(message, writer, bufferBytes)` method. Ordinary owning
Encode methods, public resident defaults and optional CPU builds are unchanged.
Older libraries fail clearly when optional packed symbols are missing. Current
compact/direct-row constructors reject the packed sink; this experiment preserves
fast matrix transforms and existing streamed-code/reuse layouts, s<=16.

After the full inverse, each coefficient maps to a unique canonical coordinate
through the verified primitive generator. Range-check before emitting bits. Write
its s-bit field directly into zeroed dead scratch, using atomic OR where adjacent
fields share a32-bit storage word, including cross-word fields. A joined parity
reduction writes coordinate zero without overwriting neighboring fields. Output
bits are canonical LSB-first, identical to the existing packed file format.
No full canonical uint32 array is assembled on device or allocated on the host.
The N-element coefficient arena and full transform scratch remain resident and
counted; after inverse, scratch holds packed output rather than another allocation.
This is not a host-backed transform or a VRAM reduction.

Prepare validates the product before publishing a readable device result. Copy
sequential bounded byte windows into the host buffer, write them, and hash emitted
bytes. Read state is reset after success/error; next owning or packed encoding
can reuse the engine. A failed writer can leave a partial sink; SHA256 is a full
result only on success. The caller controls file publication/durability. Calls
serialize through all writes, so the writer must not reenter the same encoder.
The destination may allocate its own buffers; bufferBytes bounds encoder-owned
staging only, not arbitrary writer memory or OS cache. No Go pointer is retained
by C/CUDA between calls. Code arrays for streamed constants stay immutable in Go.

At s16, selected streamed-code mode retains **32.005 GiB GPU workspace**. It
replaces16 GiB host canonical output with an8 MiB benchmark buffer; host code+
message principal storage falls38.845→22.845 GiB, plus staging/driver/runtime.
Packed transfer is8 instead of16 GiB and persistent output stays8 GiB. Atomic
packing, writer/hash and synchronization costs are unknown until measured. Do not
assume a speed improvement from fewer bytes. File-cache and verifier memory are
additional. Keep owning/resident defaults when memory fits; opt in with both
`--memory-mode streamed --output-mode packed` for this selected experiment.

NativeTotalNS counts GPU preparation plus downloads; ElapsedNS additionally
includes Go allocation, writer and hash work. Download time is already in native
D2H/total counters; do not add it twice. WriteNS and HashNS are separate CPU stages.
Caller File.Sync/Close timing is measured outside the API. A writer's digest
confirms emitted bytes, not a durable-storage commit.

## Reproduce, inputs and retention

[Runner](run.py), [launcher](run.sbatch), [analyzer](analyze.py), [self-check](check.py),
[small oracle/error tests](../../encoding/liftrs-shared/cuda_packed_test.go) and
[matched/full worker](../../encoding/liftrs-shared/gpu_packed_cached_test.go).

```bash
python3 -B experiments/gpu-packed-output/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/gpu-packed-output/slurm-%j.log experiments/gpu-packed-output/run.sbatch packed-02 --memory-mode streamed --output-mode packed
# After the user's completion ping:
python3 -B experiments/gpu-packed-output/analyze.py packed-02 --recorded-cache
```

Both CLI defaults remain resident/owning. Build with CUDA12.9.1 nvcc -O3 -lineinfo
-arch=native, GCC14.2.0, Go1.27.3/cgo. CPU-only and CUDA-only build checks are kept;
cpp,cuda tags enable comparable native clocks in the Linux validation harness.
Source checkpoint ce29557 plus the already validated dirty CPU-037 code precedes
this GPU change; unchanged CPU paths must match completed job73976. Record actual
revision, exact dirty-source snapshots, binary/source/input/evidence hashes,
commands, hardware/software/affinity, profiles, stage clocks and current/peak RSS.

Results for retry: `runs/packed-02/results.json`, phases, profiles/logs; allocation log
`slurm-JOB_ID.log`. Retain executables/library, small fixtures and packed output
under `.cache/liftrs/gpu-packed-output/packed-02/` outside Git. Read-only saved s6/
s14/s16 code inputs are reused, no preprocessing rerun or old cache deletion.
s16 source is20,854,558,905 bytes with SHA256
`f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874` from verified
preparation job72875. s14 source is `.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`.
Its existing verified manifest is used by the loader. Checksums preserve byte
identity; mathematical validity still comes from verified preprocessing.

The benchmark retains **one output per workload** to respect16 GiB new-cache
budget. After verifying repetition1, repetition2 intentionally overwrites only
this run's same-input packed output. All hashes/phase metadata/profiles persist;
no preprocessing or previous experiments' outputs are removed. A failure leaves
partial output and earlier verification evidence, not a published success. This
is an experiment retention policy, not an atomic-publication guarantee of the API.

## Validation and changed measurement boundary

Small s2/3/4/6/8 cases compare every emitted byte with independent Go words/CPU
packing. Resident/default and supported streamed-code shapes exercise cross-byte/
word packing, atomic collisions, zero/nonzero repeats, input immutability, bounded
buffers, writer errors/short writes, recovery, old ABI, mixed owning/packed calls
and Close. Existing scalar oracles, stream chunk-boundary tests,80 matrix
regressions, CPU/CUDA-only/combined/CGO-disabled suites, Go race and vet remain.
Run both packed and old scalar cases under CUDA memcheck with leak checking.
Race/vet do not prove GPU race freedom; bit-field atomic operations and full-word
oracles address the new global-memory writes. Code/resource caps do not change.

Matched s14 compares **owning+CPU pack/write/hash** versus **direct GPU packed
sink**, two repetitions each, same code mode/input/allocation and one packed file
per encoding. The comparator uses the earlier CPU packing algorithm, but now
packs every repetition so output counts match. This deliberately differs from
old workers that packed only their final owning result. PipelineNS includes
encoding, packing, writes and hashing; File.Sync/Close is separate. CUDA stages
and CPU pack-write time are also reported. Never compare new whole sink latency
with old encoding-only latency as a speedup. Fixed order/two samples still have
cache/allocation bias. Full s16 comparison to old owning jobs is historical.

After s6 smoke, two full s16 calls verify fresh all-K systematic positions and16
independent sampled curves through a **read-only mapped packed file**, separately
profiled from encoding. Reconstruct canonical uint32 SHA256 through a bounded64
KiB block, compare with retained CPU/GPU digests, recompute the packed-file hash
with an8 MiB buffer, then1027 ReadAt values compare with mapped decoding. There
is no full verification uint32 array. The verifier may fault8 GiB packed pages
into RSS at s16; that is additional validation memory, not encoder memory.
Every mapping is closed before the next overwrite. Record phase-current RSS as
well as command peak; do not use cumulative RSS as the encoding-phase peak.

Track code/output file residency without eviction via the existing Linux cache
observer. RSS+tracked-cache is conservative and can double-count verifier-mapped
pages; cgroup charging and shared global residency are different observations.
No cold-cache or total-host-RAM bound is claimed. Preserve all CPU profiles.

## Admission and pending evidence

One GPU16 CPU cores,128 GiB requested host RAM/96 GiB RSS,64 GiB Go soft;
selected streamed mode40 GiB device cap (+2 GiB initial free headroom),4 GiB
matched s14,512 MiB small cases. RSS+tracked-cache and identifiable batch-task
cgroup guards96 GiB, initial host128 GiB available/minimum32 GiB headroom. Cache/
scratch16 GiB each, initially20 GiB free. Slurm2h/controller110m/full worker90m;
small/memcheck cases have5–10m limits. Estimate20–45 minutes whole validation,
unknown atomic/sink costs. Submit and yield without monitoring until user ping.

Pre-submission CUDA sm90 compile-only and Go admission/lifecycle/matrix-description/
packed-accessor/CPU comparator checks pass. Compile artifact/source provenance is
`.cache/liftrs/gpu-packed-output/build-check/`. Python admission/bounded sink/
transfer checks, AST, shell syntax and whitespace checks pass. Runtime packing,
memcheck, RSS/cache and time remain pending. Later GPU backed transforms and both
CPU routes remain open; s17/arithmetic/preprocessing/cryptographic/TRL5 gates do
not close here. No slower default or larger resource budget is introduced.

Initial submission with untyped `--gres=gpu:1` was rejected before job creation.
Partition GRES inspection reports h200; launcher now requests `gpu:h200:1` with
the same count/budgets. This is experiment placement, not a library prerequisite.

Submitted job74091; manifest `runs/packed-01/results.json`, allocation log
`slurm-74091.log`. No monitoring until user completion ping.


## Initial failure and correction

[Job74091 manifest](runs/packed-01/results.json),
[failure log](runs/packed-01/packed-tests.log),
[verified failed evidence](runs/packed-01/failure-review.json) and
[accounting](runs/packed-01/accounting.log) retain the failed run unchanged.
It passed scalar/stream-boundary/matrix/build checks, then failed s8 resident
packed bytes in the small oracle. Matched/full-size workloads never started;
there are no new s16 time/memory results. Failed-source/evidence/cache hashes
reconcile366 exact source identities; this does not turn a failure into a pass.

The packed coordinate-zero kernel reduced chunk parity in shared sums, but wrote
thread0's original local sum instead of sums[0]. It agrees only in single-chunk
cases, explaining why smaller tests passed. The fix writes the completed shared
reduction, matching the established owning-output kernel. Added an independent
s8 K1 constant-word fixture spanning16 coefficient chunks: the first4096-value
chunk has parity0, while the last4095-value chunk gives nonzero global parity.
Values1/255/0/17 run in resident and streamed layouts, with full-byte expectations,
normal and memcheck. This regression does not depend on code discovery. Oracle
failures now identify repeat/first differing byte and counters; test failures
also close their engines through deferred cleanup.

Fixed CUDA compile-only and static Go/admission/accessor checks pass; fixed binary
provenance is `.cache/liftrs/gpu-packed-output/build-check/fixed-provenance.json`.
Retry uses fresh `runs/packed-02/` and `.cache/liftrs/gpu-packed-output/packed-02/`,
preserving the failed library/profiles/fixtures and old reference caches. Same
explicit H200/time/host/device budgets. Do not proceed with GPU host-backed
transforms until the packed sink passes its runtime gate. Both CPU routes remain
queued under the original request, not abandoned.

Correction retry submitted as job74094; active manifest
`runs/packed-02/results.json`, allocation log `slurm-74094.log`. No monitoring
until completion ping. GPU host-backed and CPU gates remain dependent/open.


## Completed retry74094

[Results](runs/packed-02/results.json), [analysis](runs/packed-02/analysis.json),
[accounting](runs/packed-02/accounting.log) pass:366 exact sources, sealed evidence/
cache hashes, new parity regression and both memchecks, all systematic positions,
16 independent curves, complete canonical/packed CPU/GPU hashes and1027 readbacks
on each of two full s16 outputs. Failed74091 evidence remains unchanged. No cache
or profile deletion. Job completed0:0 on gnode17/H200 in30m46s within all guards.

GPU ownership remains32.005 GiB. Encoding current-RSS samples are about23.01 GiB,
versus earlier39.03 GiB owning output (historical). Separate mapped verification
raises current RSS to31.014 GiB; tracked file/cache inclusive peak is60.395 GiB,
not a23 GiB total-memory cap. Whole sink pipelines take66.227/65.485 s, plus
Sync/close1.943/2.267 s. Native preparation+download57.449/57.550 s, atomic packed
assembly0.399 s, packed download0.843/0.816 s for8 GiB, writer4.597/3.752 s and
hash4.177/4.178 s. Code uploads/kernel stages remain separately recorded. New
pipeline includes output production; comparison to old encoding-only time is invalid.

Matched s14 same-code output pipelines: owning+CPU pack/write/hash8.000/8.199 s,
direct packed5.827/5.804 s, about27–29% shorter in these two fixed-order samples.
Both produce/verify one packed file per call; this is not a general latency claim.
Load318.2 s and independent full checks395.7/393.8 s stay separate and vary with
cache/node conditions. Owning/resident defaults remain unchanged despite this
measured end-to-end benefit. This completes GPU output gate1; host-backed GPU
transforms and both CPU routes remain open and proceed next.
