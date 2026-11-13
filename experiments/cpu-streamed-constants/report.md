# Fast CPU encoding with file-backed constants

Status: CPU-037 validation job73976 completed; evidence reconciliation passes.
The user requested commit/push of the validated GPU work, then the CPU counterpart.
Checkpoint **ce29557** was committed and pushed first. This experiment selects
`--memory-mode file`; resident remains default and the compact CPU path stays.

## Implementation and memory–time choice

[Shared file encoder](../../encoding/liftrs-shared/cpu_file.go) opens the existing
LRSprobe1/2 file, checks counts/size/field ranges and mandatory trusted SHA256,
then prepares only small native transform metadata. It retains no resident P or
spectrum arrays. The digest establishes byte identity, not mathematical validity:
use code from verified preprocessing. Keep the opened backing inode immutable
until Close. Source metadata is checked before/after calls and short/read errors
fail the encode. Holding a descriptor survives pathname replacement; stat checks
are not a guarantee against hostile concurrent writes/restored metadata. Prevent
such writes externally; this prototype does not claim a sealed filesystem source.

[Native kernel](../../encoding/liftrs-cpp/c_api.cpp) and
[binding](../../encoding/liftrs-cpp/plan.go) add an explicit file execution ABI.
After clearing prefix padding, read P directly into the prefix arena in bounded
chunks; validate P[K], then clear that unused coefficient as required by the
existing prefix algorithm. Read all message coefficients as before. Preserve H's
first K values in place and clear only the tail before the full forward transform,
reusing the previously validated compact-prefix invariant without compact output.

Between full forward/inverse, full transpose scratch is dead. Borrow its first
256 MiB for successive spectrum reads and matching parallel products. Join each
product before overwriting the tile. The inverse reclaims scratch; canonical
assembly then writes it and transfers ownership to the caller. No new read tile,
P/spectrum allocation, saved H, compact passes or transforms are added. There are
still five fast blocked matrix transforms. All old resident/compact kernels and
ordinary builds remain usable without CUDA. The file path supports s<=16,2L<=N.

At s16,m2,t5,K918,672,421,N2^32,L2^31, principal arrays are **8N+4K bytes =35.422
GiB** (arena, full scratch/owning result, message), versus fast resident
12N+12K+4 =58.267 GiB modeled /58.351 GiB previously sampled RSS. Modeled reduction
22.845 GiB includes19.422 GiB constants plus3.422 GiB saved H. Overhead is additional;
arbitrarily retained extra results still add16 GiB each. After successful encode,
workspace owns only the16 GiB arena; next call lazily replaces detached scratch.

Every encoding logically reads P/spectrum **20,854,558,872 bytes (19.422 GiB)**
through78 chunks at s16:14 P +64 spectrum. These are read/range-check counters,
not physical disk traffic; warm pages can be served from RAM. Prefix read time
also appears in buffer timing and must not be double-counted. Initial whole-file
verification reads the20,854,558,905-byte source once, separate from later encoding.
It builds native metadata in the same Open phase; include that boundary when
comparing first-use totals. Workspace allocation is inside the first Encode,
with replacement scratch allocation in subsequent calls. Setup/encoding must
therefore be reported separately and together where relevant.

This trades resident constants for repeated I/O, checking and kernel file-cache
pressure. Measured latency/RSS/cache observations are recorded below; no unqualified total-memory saving is claimed. Logical
19.422 GiB at hypothetical1/3/8 GiB/s would mean19.42/6.47/2.43 s payload time;
these omit range checks, faults/copies and filesystem effects, not predictions.
Keep resident fast execution default when admitted memory fits. File-backed
streaming requires explicit selection even when a warm cache makes it inexpensive.

## Observe page cache; do not hide it outside RSS

[Cache observer](cache_memory.py) uses Linux mincore on PROT_NONE mappings of
tracked input files and newly written packed outputs. Mapping reserves virtual
addresses without reading/faulting file contents into observer RSS. Record resident
page counts every30 s and process-group RSS every2 s; inspect job-scoped cgroup v2
current/anon/file/peak where its path can be identified. Do not evict shared files,
drop node caches or force cold reads. First verification warms code; comparisons
are cached/warm filesystem observations, not cold-disk throughput measurements.

The controller guards both process RSS and **RSS plus tracked resident file pages**
under96 GiB. The sum is conservative/shared-cache inclusive, sampled and not an
exact total allocation: file pages may have other users, and compiler/library or
other untracked node cache is excluded. It is not uniquely charged memory. When
a job-specific cgroup is available, independently guard its memory.current at96
GiB; when unavailable, report that limitation rather than substituting a whole
user/node cgroup. Retain128 GiB request and32 GiB node headroom guard. Report page
cache and combined observations alongside private RSS; do not call35 GiB a total
host-memory cap. With19.422 GiB code fully cached, principal arrays plus just that
cache already approach54.845 GiB, before output cache and runtime overhead.

## Reproduce and retain

[Runner](run.py), [Slurm launcher](run.sbatch), [analyzer](analyze.py),
[self-check](check.py), [Go correctness tests](../../encoding/liftrs-shared/cpu_file_test.go),
[cached worker](../../encoding/liftrs-shared/cpu_file_cached_test.go) and
[native sanitizer fixture](../../encoding/liftrs-cpp/online_test.cpp).

```bash
python3 -B experiments/cpu-streamed-constants/check.py
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-streamed-constants/slurm-%j.log experiments/cpu-streamed-constants/run.sbatch file-01 --memory-mode file
# Only after the user's completion ping:
python3 -B experiments/cpu-streamed-constants/analyze.py file-01 --recorded-cache
```

CPU-only32 cores/no GPU, Go1.27.3/GCC14.2.0/cgo/liftrs_cpp. Runner records actual
revision and exact dirty-source snapshots, source/binary/input/evidence/cache
hashes, OS/CPU/affinity/compiler versions, commands, clocks, profiles, RSS and
file-cache/cgroup samples. Results: `runs/file-01/results.json`, phase JSON,
profiles and logs; allocation log `slurm-JOB_ID.log`. Profiles stay enabled.

Use the same verified s16 input from preparation job72875, read-only:
`.cache/liftrs/gpu-preparation/s16-m2-t5/online.bin`,20,854,558,905 bytes,
SHA256 `f33dc66cb3b616697866a298ad04dd3d4853fb98bf840f3c81976955bc895874`.
s14 uses `.cache/liftrs/parallel-s14/s14-m2-t5/online.bin`,1,277,647,153 bytes,
SHA256 `88d5ff39c5df8bdd4aa0f85fc19833ae673df95365b4fd862bb83936e5a0adf8`.
No discovery/preprocessing rerun, input copying or cache deletion. Retain binaries,
small fixtures and one8 GiB packed result under
`.cache/liftrs/cpu-streamed-constants/file-01/` outside Git. Existing backing file
adds no new persistent copy; output remains8 GiB. Cache residency is additional
RAM, not extra disk storage. Recorded-cache analysis checks completed controller
hash plus metadata/length instead of another8 GiB payload rehash.

## Validation and limits

Small file-backed words match the Go reference, nonzero/zero/repeated messages,
retained results/Close, input immutability, bad inputs/recovery, old/wide count
headers, hash/limit/normalization/field errors, native short reads and scratch
alias rejection. Test tiles1/257/N-1/N cover exact/full/tail scheduling; tests also
check opened inode identity across pathname replacement and backing truncation.
Native s9 K1 independent constant-word fixture covers tiles257/65537 (parallel products
and partial tails) at1/4
workers, bytes/chunks, truncated reads, parallel native transforms/products and
assembly under ASAN/UBSAN and TSAN. Go race covers binding/lifetime checks;
normal/compact regression suites, Go-only build and vet run in the job.

Matched retained s14 resident then file-backed, two encodes each on one allocation,
uses the same seed32-worker workload and checks full canonical/packed agreement.
Two fixed-order warm/cached samples do not establish cold I/O or a latency
distribution. Run s6 smoke, then **one full s16 file encode**, all systematic
values,16 sampled independent curves,1027 packed readbacks, exact bytes/chunks/
workspace ownership and full retained CPU/GPU hashes. Repeated ownership is tested
at small/s14 sizes; full s16 comparison against prior resident is historical.
The helper's file-load hook returns thin private metadata for existing independent
checks; it is not a public resident plan with missing arrays.

Budget unchanged128 GiB request/noGPU32CPUs, file RSS48 GiB guard/Go44 GiB soft,
96 GiB RSS+tracked-cache/job-cgroup guards where applicable,2h allocation/110m
controller/90m full worker,16 GiB new cache/scratch guards,20 GiB initially free.
Resident CLI default uses prior68 GiB RSS/64 GiB Go limits. Estimate40–75 minutes
whole validation, uncertain filesystem/runtime cost. Automatic guards/profiles
remain active; submit and yield without monitoring until completion ping.

Pre-submission Go file/normal/compact ownership/reference tests pass (1.608 s),
optimized native s9 file/compact fixture passes. Python admission/packed/model/
read-count/ownership and non-faulting cache observer checks, AST/bash/whitespace
checks pass. Broader race/sanitizer/full RSS/cache/time checks subsequently pass below.
No s17 arithmetic/artifact/preprocessing, cryptographic/application or TRL5 gate
closes; no automatic selection or larger budget is introduced.

Job73976 was reviewed after the user completion ping; allocation log
`slurm-73976.log`. No new validation job follows this review.


## Completed validation

[Manifest](runs/file-01/results.json), [analysis](runs/file-01/analysis.json),
[full phases](runs/file-01/encoding/phases.json),
[cache/cgroup samples](runs/file-01/progress.jsonl),
[ASAN/UBSAN](runs/file-01/sanitizer.log), [TSAN](runs/file-01/tsan.log) and
[final accounting](runs/file-01/accounting.log) pass review. Reconciliation resolves
358 exact source identities (ce29557 plus dirty-source snapshots), checks sealed
cache/evidence digests and compares full canonical/packed hashes with retained
CPU/GPU references. The controller already hashed the8 GiB output; recorded-cache
analysis uses its digest plus metadata/length, avoiding another full bulk reread.
Accounting is additional post-completion evidence outside the controller seal.

The full s16 encode checks all918,672,421 systematic values,16 sampled independent
curves and1027 packed readbacks. Small retained-output/header/checksum/range/
truncation/path-replacement/chunk/alias/recovery/Close tests, normal/compact native
regressions, CPU/Go-only suites, race/vet and native ASAN/UBSAN/TSAN all pass.
No kernel correction was required by validation. Profiles, verified backing inputs,
binaries and packed output remain retained; no caches were evicted or deleted.

Hardware gnode17,32 CPU cores, no GPU; GCC14.2.0/Go1.27.3. Job completed exit0:0
in **36m45s**, slightly below the conservative40–75-minute forecast, below2h limit
with no guard tripped. Full cached worker took1533.1 s; allocation also includes
builds/sanitizers/matched runs/profiles. These are mathematical-encoder timings,
not cryptographic/application processing or TRL5 acceptance.

| Full s16 metric | Result |
|---|---:|
| Principal arrays, model | 35.422 GiB |
| Peak sampled process-group RSS | **35.484 GiB** |
| Previous resident peak sampled RSS, historical | 58.351 GiB |
| Maximum tracked resident file pages, sampled | 26.336 GiB |
| Maximum sampled RSS + tracked file pages | **61.820 GiB** |
| Maximum sampled charged cgroup memory.current | 47.917 GiB |
| Highest recorded cgroup memory.peak | 47.919 GiB |

Process RSS decreases22.868 GiB (39.19%) versus prior resident, consistent with
removing constants/H. That is **not** a39% reduction in total RAM. Tracked cache
includes s6/s14/s16 input files and current packed outputs/temporaries; cached
pages can be shared or charged elsewhere. The observed cgroup is this batch task
subtree (`.../job_73976/step_batch/user/task_0`), containing the controller/children;
it is not whole-node memory. Its charged file pages differ from globally resident
tracked-file pages, so47.917 and61.820 GiB are different observations and must not
be substituted for each other. Full resident cache/cgroup accounting was not
measured alongside this s16 run, preventing a matched total-memory savings claim.

During encode/verification, tracked file pages reach20.620 GiB and process RSS
35.484 GiB, about56.104 GiB combined. Packing raises the recorded sum to61.820 GiB.
Residency is sampled about every30 s, RSS about every2 s plus observer overhead;
short peaks/final output-cache growth may be missed. The sum includes a cumulative
command RSS peak and shared tracked pages, rather than an exact instantaneous
unique-memory total. Other node/cache users are excluded. Both sampled limits
stay below the96 GiB inclusive/charged guards and48 GiB RSS guard. Slurm MaxRSS
is37,230,472 KiB. Initial verification warms the code; these results do not measure
forced cold-disk latency or establish a35 GiB total-RAM deployment.

| Full s16 phase | Seconds |
|---|---:|
| Open/verify entire file and native metadata | 37.181 |
| Online setup hook (already prepared) | 0.000003 |
| Encoding, including first work-array allocation | **1249.616 (20m49.6s)** |
| Independent verification | 178.754 |
| Pack/write8 GiB | 55.768 |
| Packed readback | 2.415 |

Encoding reads exactly20,854,558,872 logical code bytes through78 chunks. Read/
range-check time8.963 s is already included in whole encoding; prefix reads also
overlap buffer9.126 s, so do not add counters twice. Columns/transposes455.630 s,
rows549.781 s and Taylor197.185 s dominate. First-use opening/native metadata and
encoding have a different phase boundary from resident setup; compare their sums
where relevant. Logical reads may be served from page cache, not physical disk.
The new full encode is within the earlier resident1247.173–1281.328 s range on
separate allocations; this is historical evidence, not a speedup/equivalence claim.

Matched s14, same allocation, fixed resident-then-file order:

| Mode | First encode, s | Second encode, s | Peak sampled RSS, GiB |
|---|---:|---:|---:|
| Resident | 58.8530 | 58.6672 | 3.595 |
| File-backed | 59.3773 | 59.3573 | 2.214 |
| Encoding penalty | 0.89% | 1.18% | — |

File reads1,277,647,128 bytes through5 chunks in0.643/0.649 s. Both full hashes
match. Sampled RSS plus common tracked background/input/output cache decreases
20.385→19.009 GiB; this inclusive comparison includes shared s16 input residency
even during s14 runs. Two fixed-order cached samples support a small observed
penalty, not a cold-storage latency distribution. Keep resident default when it
fits and file streaming an explicit optional capacity choice.

This completes the selected CPU counterpart. Both CPU and GPU streamed constants
now have s16 evidence. The CPU fast two-array layout would still need about270.42
GiB for its two work arrays and message at s17 with uint64 extension slots, before
result-conversion copies, cache and other overhead. Combining bounded columns/one arena with file-backed constants
could lower that model, but requires new wider arithmetic/artifact/ownership work
and cache-aware admission; the earlier144.67 GiB scenario is not an implementation
or guaranteed total RAM. s17 and application/cryptographic/TRL5 gates remain open.
