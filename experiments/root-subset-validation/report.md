# Fully checked root-subset membership comparison

Status: explicit sorted validator implemented; matched GPU/worker measurements
pending. Work SCALING-111. Ordinary preprocessing/cache/discovery defaults remain
bitmap; no full-support strategy or slower fallback is adopted.

[Runner](run.py), [Slurm entry](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [membership implementation](../../encoding/liftrs-shared/preprocessing_core.go),
[CPU equivalence/memory tests](../../encoding/liftrs-shared/support_validation_test.go),
[shared GPU driver](../../encoding/liftrs-shared/root_workers_gpu_test.go).

## Method and explicit trade-off

The shared validator still checks cardinality, nonnegative in-range exponents,
strict sorted/distinct order and every q-Frobenius partner. A bitmap and binary
search differ only in membership representation; the shared order/range scan
finishes before any search. Caller indices are borrowed immutably. No sampling,
probabilistic membership, validation shortcut, cached proof or copied arithmetic
is introduced. Both modes retain the same progress stage names/counts.

| Cost | Existing bitmap | Explicit sorted membership |
|---|---|---|
| Temporary membership storage, s17 |2GiB/invocation |No N-sized array or new index copy |
| Temporary membership storage, s18 projection |8GiB/invocation |No N-sized array or new index copy |
| Membership work |O(K) bit lookups, plus domain-sized allocation/initialization |O(K log K) binary searches in existing sorted input |
| New GPU work/transfers or canonical file bytes |none |none |
| Actual latency/RSS penalty or saving |Measured baseline77148 is historical; new matched baseline runs below |Unknown until paired job completes |

Removing the bitmap does not remove root coefficients, the existing K-index
partition copy, metadata, CUDA contexts, profiles or runtime allocation overhead.
It avoids2GiB per validation at s17; four partial-root validators can otherwise
request8GiB aggregate bitmap backing. Allocation sizes are not measured RSS.
The binary searches can be slower for dense supports; retain the existing faster
validated full-support path when it fits. Only this explicit experiment selects
sorted membership. Any later production choice requires measurements and an
explicit option for a slower memory-saving mode when memory is available.

The private root and partition helpers accept an explicit membership choice and
share all remaining logic. Existing wrappers select bitmap. Ordinary Go/CPU
clients gain no CUDA/MPI dependency; public APIs, CLI, code representation,
root/recovery identity and default schedule are unchanged. Sorted root validation
is selected on the optional test binary with `-root-validation=sorted`;
`-root-validation=bitmap` is the default. Unknown names reject.

## Correctness and measurement boundaries

CPU tests exhaust all65536 subsets of a16-point domain, comparing bitmap/search
acceptance and error strings, including the excluded final index. Further negative
cases cover duplicates, unsorted/negative/out-of-range values and missing partners.
Wide-index tests use an s17 domain, include indices above2^32, verify immutability
and progress, and bound total sorted-validation allocation below1MiB. Root subset
partition/merge tests exercise both membership choices and verify complete P;
publication failure still prevents generator work. These local checks pass.

Reuse the [root-worker driver and configurations](../parallel-root-workers/report.md):
one genuine s8 support and synthetic s12/s17 subsets, whole-root/four-GPU schedule
and1/2/4 independent one-GPU workers with complete merge. Each process retains a
warm-up plus two timed samples/configuration. Repeat in **bitmap/search/search/
bitmap order**, with two processes per mode and four timed samples per case and
worker count. Each fresh process has its own CPU profile, raw output and sampled
RSS; cold constructor setup remains separate from operation time. Fixed case/
configuration order inside each process and four samples still do not establish
fullsize scaling/confidence limits.

### Deliberate change from historical77148

Both new variants explicitly use **sorted validation for CPU oracle construction**
(`-root-reference-validation=sorted`), so the common untimed CPU reference does
not allocate its own N-sized bitmap and contaminate the memory comparison.
The oracle still uses independent native CPU root factors/products, complete
coefficient equality and four scalar extension-field root evaluations. Its
validator is independently tested against exhaustive bitmap results. Fixture
seeds, support sets, cutoff4096, worker/merge arithmetic and call clocks match.
Both variants use this same reference setup; historical77148 used bitmap for its
CPU references. Thus compare the new paired baseline and variant, not historical
whole-command RSS/profile as an isolated strategy effect. All row/log records
identify both validation and reference-validation modes.

All persisted actual outputs and CPU reference digests must agree across both
strategies and repeats. The parent worker reviewer checks warm/timed coverage,
partition balance, complete file hashes and owned-device caps. Genuine-s8 CUDA
memcheck repeats both choices; larger synthetic cases remain unsanitized. The
existing guarded controller additionally runs complete native/Go/GPU CLI,
restart/corruption/readback/race/ASan/vet/memcheck regressions with default bitmap.
CPU profiles contain host arithmetic and opaque cgo waits, not GPU instruction
traces. Profiling/logging remain enabled and included in measured call times.
Setup, partition, joined worker and merge clocks stay separate; CPU/scalar checks
and output publication are outside call clocks and inside command time.

## Resources and reproduction

Use the existing admitted HPC budget: one node,4H200,4CPU cores,8GiB requested
host RAM,6GiB RSS guard,4GiB Go target,4GiB aggregate owned-device cap,2GiB new
retained storage,4GiB scratch. New executors cap128MiB/card; merge+workers at most
1GiB summed caps. Context/runtime overhead is additional. Ten minutes/command,
fifteen-minute controller and twenty-minute scheduler limits; gnode03 excluded,
no node pin. No resource increase, fullsize job, shared-filesystem array workspace
or persistence-concurrency load is requested. Repeated small output artifacts are
kept within the existing retained budget; earlier caches/profiles are untouched.

```sh
python3 -B experiments/root-subset-validation/test_review.py
python3 -B experiments/root-subset-validation/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/root-subset-validation/slurm-%j.log \
  experiments/root-subset-validation/run.sbatch compare-01
# After completion ping only:
python3 -B experiments/root-subset-validation/analyze.py compare-01
```

Runner seals exact Git/copied source/evidence/cache identities in
`runs/compare-01/results.json`. `comparison.json` retains all measurements;
bitmap/search profiles/logs are named by mode/repeat. The matching binary, CPU
polynomials and actual outputs remain outside Git at
`.cache/liftrs/root-subset-validation/compare-01/`. The inherited reviewer retains
only the single named ignored-source override from root-boundary77126, bound to
its exact preserved producer hash. No blanket dirty-source acceptance.

No timings or whole-process RSS savings are claimed before review. Next is
interpretation of root-worker scaling using the fully checked strategy costs,
then GPU node communication/staging and separate buffered persistence. Submit,
report ID/results and hand off without monitoring.
