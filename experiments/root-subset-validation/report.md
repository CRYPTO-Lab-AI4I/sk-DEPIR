# Fully checked root-subset membership comparison

Status: **matched comparison77213 passed** on HPCgnode02 (4 H200 GPUs,
4 CPU cores,8GiB requested host RAM). Complete job8min11s; controller488.638s.
Producer `e061a08`. [Receipt](submission.json), [raw results](runs/compare-01/results.json),
[analysis](runs/compare-01/analysis.json), [measurements](runs/compare-01/comparison.json),
[accounting](runs/compare-01/accounting.txt), [Slurm log](slurm-77213.log).
Work SCALING-111/112. Ordinary preprocessing/cache/discovery defaults remain
bitmap; no full-support strategy or automatic slower fallback is adopted.

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
| Actual latency/RSS penalty or saving |Measured matched baseline below;77148 remains historical |Measured below for bounded subsets; full-support latency unknown |

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

Bounded timings and sampled process RSS are reviewed below. Fulls17/s18 pipeline
latency and memory remain unmeasured. GPU node communication/staging and separate
buffered persistence remain subsequent gates; no new job follows this review.

## Completed matched comparison (SCALING-112)

**601 source /51 evidence /164 original cache identities** reconcile. The single
explicit ignored-source snapshot from prerequisite77126 matches its producer
hash; all other code matches recorded Git objects. Every complete CPU coefficient
array, four independent scalar-root witnesses per fixture, degree/normalization,
persisted output and cross-mode/repeat digest agrees. All four mode/repeat processes
and192 normal rows (128 timed,64 warm-up), plus24 small sanitized rows are present.
Four CPU membership/root tests and race checks pass, including exhaustive65536
small-domain subsets, malformed inputs, wide-index memory/progress and both-mode
root merges. Both validator memchecks and inherited CLI memcheck report0errors
and0leakedbytes; larger synthetic cases are not sanitizer coverage. Ordinary
bitmap CLI/reference/restart/race/ASan/vet gates pass. No execution code changed
in this review.

### Matched operation times: four-sample means

All values are seconds; each mean combines two timed samples in each of two
fresh processes per mode, with bitmap/search/search/bitmap ordering. Input,
reference-validator setup, worker/merge arithmetic and profiling are matched.
Operation time includes partitioning, validation and complete merge. Whole-root
is a four-GPU executor;1/2/4 workers each use one GPU, with a four-GPU merge executor.

| Fixture | Validator | Whole-root | 1 worker | 2 workers | 4 workers |
|---|---|---:|---:|---:|---:|
| s8 genuine,7381 exponents |bitmap |0.019160 |0.019036 |0.015953 |0.016320 |
| same |sorted |0.019677 |0.018530 |0.016193 |0.016540 |
| s12 synthetic,65536 |bitmap |0.349136 |0.334073 |0.218549 |0.205891 |
| same |sorted |0.355211 |0.346894 |0.221009 |0.209219 |
| s17 synthetic,65536 |bitmap |0.696309 |2.319603 |1.578100 |1.480394 |
| same |sorted |0.552768 |0.506915 |0.325292 |0.304781 |
| s17 synthetic,262144 |bitmap |3.204251 |3.984055 |3.428800 |3.427332 |
| same |sorted |2.505141 |2.249716 |1.334992 |1.039159 |

On the largest sparse case, sorted four-worker time is **3.30 times faster by
the observed bitmap/sorted ratio** (3.427332/1.039159). Its four timed samples
range1.000781–1.080675s, versus3.039250–3.738064s bitmap. Comparing worker counts
with sorted validation held fixed gives **2.41 times the whole-root/four-worker
ratio** (2.505141/1.039159). This is a same-node root component including partition
and merge, not a complete PIR preprocessing or pure GPU-kernel speedup.

At s17/65536, four-worker sorted time is0.304781s versus1.480394s bitmap; sorted
whole-root/four-worker ratio is about1.81. At smaller s8/s12, sorting membership
provides no general elapsed-time benefit and is slightly slower in most paired
configurations. For s12/four-worker, means0.205891→0.209219s show about1.6% search
overhead, with overlapping sample ranges. These are small samples, not confidence
bounds or a validated dense-full-support crossover. No universal mode/worker-count
policy is inferred.

Largest-case four-worker mean subtimes:

| Phase | Bitmap | Sorted |
|---|---:|---:|
| Partition + global validation |0.740684s |0.062647s |
| Joined worker root wall time, including worker validation |2.370534s |0.667263s |
| Complete merge |0.316105s |0.309244s |
| Operation total |3.427332s |1.039159s |

The merge is now about30% of sorted operation time, and partitioning about6%.
Context setup remains separate and has cold-start differences; output
comparison/hash/persistence remains outside operation clocks. Aggregate end-of-call
owned GPU allocations for the largest four-worker case remain**43.102MiB** in
both modes (whole-root20.463MiB); no GPU array reduction is claimed.

### Host memory, command/profile scope

| Two fresh normal processes per validator | Bitmap | Sorted |
|---|---|---|
| Maximum sampled process-group RSS |4561.707 /4797.340MiB<br>4.455 /4.685GiB |519.672 /521.914MiB<br>0.5075 /0.5097GiB |
| Complete component command |78.629 /84.134s |46.465 /46.619s |

This is about**89% less sampled process RSS** in the matched sparse-subset
experiment. RSS includes runtime/CUDA contexts, fixtures, CPU references,
comparison/persistence and profiles; it is not a complete s17 database or a
per-worker allocation peak. Requested host RAM remains8GiB and guard6GiB, not
inferred minimum allocations. Scheduler MaxRSS4726632KiB is a separate sampler.
The CPU oracle uses sorted validation in both modes, unlike historical77148;
compare these new pairs, not historical peak RSS, for the strategy effect.

Representative retained [bitmap-0 CPU summary](runs/compare-01/bitmap-0-cpu-top.txt)
shows45.53CPU-s cumulative in membership validation (39.41% of115.53CPU-s),
while [sorted-0](runs/compare-01/sorted-0-cpu-top.txt) shows1.32CPU-s (2.00% of
65.88CPU-s). Memory clearing drops from21.02 to0.50 flatCPU-s. These are whole
component CPU profiles, include setup/oracles, and percentages overlap; they are
not GPU/wall fractions or sums. Flat cgo CPU samples stay about29s in each
representative process and do not separate native work from GPU waits.
All four original profiles and matching binary are retained; summaries were
read from them, not new execution/profiling runs.

Reusable reference/output/binary cache totals**60.852MiB** outside Git, plus
tracked profiles/logs/review evidence. No cache or prior raw evidence was deleted.

### Recommendation and limits

Use **explicit sorted validation for the measured sparse root-worker workloads**
in subsequent distributed prototypes: it preserves all checks and improves both
time and host memory here. Keep ordinary full-support preprocessing/cache loading
on the validated bitmap default; larger/dense-support search costs remain unknown
and the s8/s12 overhead reinforces workload-specific selection. Do not silently
trade memory for latency or add a field-degree-only mode switch.

The root-worker gate is now sufficiently measured to proceed to node-to-node
GPU staging, synchronization and deterministic rank ownership using the explicit
subset strategy, then separately controlled buffered persistent output. Fulls18
root distribution, communication, checkpoint/publication and complete runtime
still need implementation/measurements; no sixteen-GPU speedup, fullsize budget,
new automatic dispatch or production mode adoption follows from these numbers.
No new job was launched during review.
