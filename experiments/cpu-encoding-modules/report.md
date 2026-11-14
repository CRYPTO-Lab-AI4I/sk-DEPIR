# CPU encoding modularity and regression validation

Refactor the complete native CPU implementation while preserving public APIs,
mathematical order, canonical representation, mode defaults and principal-array
accounting. [Implementation guide](../../docs/cpu-encoding.md), [controller](run.py),
[launcher](run.sbatch), [analyzer](analyze.py), [direct helper checks](check.cpp),
[cross-mode/lifetime tests and benchmark](../../encoding/liftrs-shared/cpu_modules_test.go).
Initial validation used job74419, source checkpoint `c3544fd`;
[submission](submission.json). No speedup, full-size equivalence or s17 result is claimed.
The run manifest records exact revision, source hashes, commands, hardware and
resource policy. Results will be under `runs/modules-01/results.json`.

## Changes and scope

One templated cyclic sequence now serves resident32-bit, wide64-bit and disk
adapters; existing compact, streamed-code and packed output modes use it too.
Pointwise products, generator/power/fixed-product helpers, scatter/parity and
little-endian file I/O are shared. Output layouts stay separate, with the same
ownership transfers, chunk sizes and allocation models. Native/Go bindings are
split by responsibility; repeated metadata setup and distinct-plan locking are
shared. Resident/streamed code loading reuse the width-aware verifier without
another payload pass or full-array copy. The independent Go/dense/hybrid
references remain separate for correctness. Resident legacy loading also shares
the decoder, replaying only its small header and retaining the decoded arrays.

Review found two robustness issues: resident workspace creation after Close or
on an uninitialized encoder could panic/allocate; it now returns nil/error.
Compact parallel16-bit outputs can share an original32-bit storage location;
use atomic character stores as in the wide output path. No extra data pass or
allocation is added. The runtime effect is unmeasured until the comparisons below.
A full-width stream verifier shortcut retains the old s16 optimization: every
uint32 is a valid extension value, but the whole payload is still read/hashed.
Digest parsing now uses common byte comparison, accepting either hexadecimal
letter case; byte identity and mandatory digest checks remain. Errors may have
more general shared messages; no typed error contract is changed.

C ABI fragments remain included once by c_api.cpp/wide_api.cpp, retaining existing
standalone compiler commands. clang-format18.1.8 was used temporarily from `/tmp`
for readability; [.clang-format](../../encoding/liftrs-cpp/.clang-format) records
style and adds no build/runtime dependency. Go build tags retain their exact
`//go:build` spelling. Historical evidence/source hashes are not rewritten.
No preprocessing/GPU algorithms, parameters, security contracts or running jobs
are changed. All earlier caches/profiles are retained.

## Reproduce and resource admission

```sh
python3 -B experiments/cpu-encoding-modules/run.py --check
module load slurm/slurm/25.05
sbatch --output=experiments/cpu-encoding-modules/slurm-%j.log \
  experiments/cpu-encoding-modules/run.sbatch modules-01
```

Use a fresh run name. One node,4CPU/8GiB request,6GiB sampled RSS guard,4GiB Go
soft limit,2GiB retained cache/4GiB scratch, no GPU. Worker600s/controller900s/
Slurm1200s deadlines match prior bounded validation. Go1.27.3/GCC14.2, generic
C++17 flags, no hardware-specific instructions or new package dependency.
All commands execute from copied source, with test binaries launched from their
package directories. Matched binaries, real small verified preprocessing and
native checker executables stay in `.cache/liftrs/cpu-encoding-modules/modules-01/`;
logs/profiles/results remain in `runs/modules-01/`. No new large preprocessing.
Submit and hand off without monitoring. Prior job74253 is untouched.

## Correctness and performance checks

Run Go-only and liftrs_cpp shared suites, encoding/PIR/configuration packages,
vet, targeted Go race, Go ASan with C++ UBSan, and direct native optimized and
ASan/UBSan helper/online/matrix/Taylor/wide schedule/wide matrix drivers. Real
s4/6/8 fixtures compare every resident/hybrid/wide/compact/file/disk result with
Go, including zero/invalid/recovered messages, retained-output ownership and
packed bytes. Existing wide fixtures add dense code membership, all systematic
positions, P·G identities and width/range/cap/count checks. Both25/33-byte legacy
headers run through the common file verifier; the legacy resident loader still
rejects the new wide format before allocation. The helpers driver exercises4096
high-bit products, non-chunk-aligned scatter,16389 unaligned/multi-block file
words, short input/range/overflow and periodic progress output without sleeping.
Go race checks do not instrument native C++ worker memory; native ASan/UBSan are
additional checks, not a proof of race freedom.

Performance uses an exact baseline CPU tree from `671bf5e` with the identical
new benchmark test file copied into both trees. Read-only s10,m2,t5 input:
`.cache/liftrs/gpu-products/products-02/cpu-0/online.bin`, SHA256
`3ab552f2ea40fe7b8ee39e2aabde2852f8fc536e9945fb7237fee188f357877e`.
Its preparation/provenance is in [GPU products](../gpu-products/report.md).
Each binary independently checks full words/packed hashes against Go. Compare
resident/hybrid/wide/compact/file/disk/packed modes plus resident loading and
file opening with4 workers,3 timed
iterations per sample,3 samples per implementation, alternating binary order.
Warmup and setup/verification are excluded; workspace replacement/allocation,
complete encoding, and packed writer/hash work are included. Packing writes to
a bounded synchronous discard sink, not durable storage. OS/NFS cache remains
uncontrolled; these are small same-node regressions, not fulls16 guarantees.
A separate one-iteration profiled run retains CPU attribution; it is excluded
from latency comparisons. No timing improvement is required from deduplication;
material regressions require investigation before adoption.

## Progress and measurement boundaries

The controller saves command starts/completions/failures and30s resource
heartbeats to `progress.jsonl` and Slurm output. Heartbeats explicitly say
remaining work is unknown. A targeted correctness run enables
`LIFTRS_CPU_PROGRESS=1`: native JSON logs include phase start/completion/error,
completed/total encoding operations and periodic completed native blocks/tiles.
The runner asserts actual complete/failure records, while the helper driver
checks advancing periodic units with a zero interval solely for that test.
Example schema: `{"phase":"full-forward","work":"parallel-blocks",
"event":"progress","completed":512,"total":4096,"elapsed_ms":30001}`
with an additional UTC timestamp. This example is illustrative, not a measured run.
All comparison timings disable logging/profiling. Instrumentation adds no domain
array/transfer/full data pass; its enabled cost is not claimed to be zero.
Principal buffers/backing stay as documented; allocator/thread/logging metadata,
retained previous results and file cache are extra.


After the user's completion ping, reconcile the sealed source/evidence/cache
hashes and baseline CPU sources, inspect all checks and compare median ratios:

```sh
python3 -B experiments/cpu-encoding-modules/analyze.py modules-01
```

Ratios above1.2 are flagged for review, not automatically explained away as
noise. Native metadata/file-cache costs are outside Go allocation counters;
fulls16/s17 timing remains a separate gate.

## Initial validation failure and fixture correction

[Job74419 results](runs/modules-01/results.json) failed in the targeted module
test. [Forensic reconciliation](runs/modules-01/failure-analysis.json) verifies
460 source,13 evidence and14 cache identities. Go regressions passed. The s4,t2
code has K97,N256,L256; the all-mode test incorrectly expected file/disk/compact
execution despite their documented2L≤N bound. No production bound is changed.

Use s4,t3 (K37,L128) for all-mode comparisons, plus an explicit dense-prefix
admission test covering rejected file/disk/compact and valid resident/wide Go
comparisons. The targeted selector includes that test. Retry `modules-02` keeps
all previous budgets and benchmark boundaries, independently comparing the same
baseline against the corrected new tree. Race/sanitizers and matched timings
remain pending; no performance equivalence is claimed from the failed attempt.

Retry job74448 is submitted ([record](retry-submission.json)), source
`2a2095e0b287a7751c9399d464285ea7a32aa02a`. No monitoring; results in `runs/modules-02/` await a completion ping.

## Completed corrected validation (SCALING-072)

Job74448 on gnode02 passes [reconciled validation](runs/modules-02/analysis.json):
479 source,56 evidence,31 retained-cache and141 baseline-source identities.
[Exact commands, timings and provenance](runs/modules-02/results.json) come from
[the runner](run.py); [Slurm log](slurm-74448.log) and all logs/profiles remain beside
these results. Go/native suites, vet, race, ASan/UBSan and direct kernels pass.

Matched, unlogged s10 encoding median new/old ratios range1.010–1.028 across
resident, compact, file, disk, packed, hybrid and wide modes. Setup flags exceed
the1.2 review threshold: resident loading5.826→8.387 ms (1.440×), file opening
7.714→9.789 ms (1.269×). These are unresolved regressions, not evidence of equal
performance or full-size slowdown. Keep setup profiling as a follow-up; GPU
server preprocessing remains the active scaling priority. Sampled command RSS
peaks at about0.982 GiB including builds; this is not isolated encoder RAM.
No fulls16/s17 rerun, memory-mode change or preprocessing admission follows.
