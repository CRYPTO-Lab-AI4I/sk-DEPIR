# Independent GPU root workers, including merge

Status: **bounded comparison77148 passed** on HPC gnode02 (4 H200 GPUs,
4 CPU cores,8 GiB requested host RAM), complete job5min22s, controller321.241s.
Producer `00360cc`. [Receipt](submission.json), [raw results](runs/roots-01/results.json),
[analysis](runs/roots-01/analysis.json), [measurements](runs/roots-01/workers.json),
[accounting](runs/roots-01/accounting.txt), [Slurm log](slurm-77148.log).
Work SCALING-109. [Prerequisite root-boundary validation77126 passed](../preprocessing-root-boundary/report.md#completed-validation-and-source-reconciliation-scaling-109).
This is a same-node scheduling component, not distributed preprocessing support.

[Runner](run.py), [Slurm entry](run.sbatch), [reviewer](analyze.py),
[review tests](test_review.py), [opt-in driver](../../encoding/liftrs-shared/root_workers_gpu_test.go),
[partition helper](../../encoding/liftrs-shared/root_partitions.go).

## Design and reuse

The shared normalized-root procedure, existing factor-product builder and existing
CPU/GPU product dispatch are reused. The internal product adapter now requires
only Product, permitting an independent executor to use it without implementing
whole-grid discovery/export. No arithmetic recurrence, public API, CLI/default
mode or CPU-client dependency changes. The only new production helper is private,
explicit and unused by default preparation: deterministic partitioning of a
sorted Frobenius-closed support. Each conjugate pair stays on one worker, so each
worker constructs a base-field polynomial using the same root helper. Assign
orbit degrees to the least-loaded worker and sort private output lists; balancing
is within two exponents. Normalized partial products merge exactly into P.

One K-index copy is allocated by partitioning and included in operation timing.
This small in-memory helper is **not** the sharded-support representation needed
for fulls18; do not introduce that extra121GiB host copy into a full run implicitly.
The caller's support is immutable; every worker has its own executor and errors,
while field lookup tables are shared read-only. The current prototype is threads
on one HPC node, not ranks across nodes. Independent root computation/merge
correctness must pass before adding GPU/rank transfers.

## Cases and measurement boundaries

| Fixture | Input type |
|---|---|
| s8,m2,t5 | Genuine complete discovered PIR support, seed109 |
| s12,65536 exponents | Synthetic deterministic Frobenius-closed root subset |
| s17,65536 exponents | Synthetic deterministic closed subset with indices above2^32 |
| s17,262144 exponents | Larger synthetic closed subset with indices above2^32 |

Synthetic sets use seed109+s+count and are not certified s12/s17 PIR supports.
No fulls18 support/grid/cache is created. A complete native CPU root polynomial
is retained for every case. Four independent extension-field scalar Horner root
witnesses plus degree/normalization checks supplement complete CPU/GPU coefficient
comparisons. All final outputs are persisted/hash-checked against the reference;
no performance result can pass from a hash alone or a partial merge.

Compare four fixed configurations:

- `independent_workers=0`: original whole-root schedule on a four-GPU executor.
- `1`, `2`, `4`: that many concurrent root workers, one GPU/executor per worker,
  followed by a balanced merge using a four-GPU executor. Small product merges
  still use the shared host cutoff4096. A one-worker case has no actual merge.

The merge executor is retained separately from worker executors; its setup and
owned allocations are counted even when unused. This permits attribution but is
not an optimized deployment-memory claim. CPU reference uses four native workers;
Go GOMAXPROCS4 and total requested CPU allocation stay4 for all configurations.
Whole-root host work may use fewer cores than independent roots; that scheduling
change is intentional, not a claim of pure GPU kernel acceleration.

Record constructor setup separately, one warm-up (`sample=-1`), then two profiled
samples. Operation time includes deterministic partitioning/global validation,
per-worker validation/root-factor/normalization work, goroutine launch/join and
complete merge. Partition, slow joined worker wall time and merge are separate
subtimes. Input discovery/generation, CPU oracle, full output comparison and
hash/file publication are outside the operation clock but inside command runtime.
Timing includes Go CPU profiling and progress overhead; fixed configuration order
and two samples preclude strong scaling confidence claims. Scalar checks are
outside operation clocks. No native/GPU instruction profile is collected; the Go
CPU profile includes host work and opaque cgo calls and cannot separate GPU waits.

Root-factor batching, root stages and completed exponent counts identify workers;
shared native arithmetic logs identify transform/pass work. Updates flush and
throttle to30s for long host stages. No ETA is inferred without measured progress.
Actual artifact profiles and binaries are retained, with no filesystem intermediate
exchange. Persisted small references/results are verification artifacts, not
array scratch used by root computation. Large buffered checkpoint/publication
measurement remains a separate gate after computation/transport integration.

## Resources, checks and evidence

Reuse the admitted one-node **4H200 /4CPU /8GiB requested host RAM /6GiB RSS guard /
4GiB Go target /4GiB aggregate owned-device /2GiB new retained /4GiB scratch**
budget. Every component executor has a128MiB per-card cap; four worker executors
plus the four-card merge executor total at most1GiB of owned caps, below the
4GiB budget. Context/runtime overhead is additional and remains in host guards.
Owned GPU allocations are reported after calls, not measured physical peaks.
The native reference cap is64MiB. Per command10min, controller15min,
Slurm20min. Exclude gnode03; no node pin or multi-node GPU allocation.

The controller reuses complete small CLI/reference/restart/race/ASan/vet/memcheck
gates. The extra suite builds a retained test binary, runs partition race checks,
normal measurements with CPU profile, then genuine-s8 CUDA memcheck/leak checks
across all configurations. Larger synthetic cases are normal-execution checks,
not sanitizer coverage. Original/full-size caches and profiles stay untouched.

```sh
python3 -B experiments/parallel-root-workers/test_review.py
python3 -B experiments/parallel-root-workers/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/parallel-root-workers/slurm-%j.log \
  experiments/parallel-root-workers/run.sbatch roots-01
# Only after completion ping:
python3 -B experiments/parallel-root-workers/analyze.py roots-01
```

Source revision/per-file identities and evidence/cache hashes go to
`runs/roots-01/results.json`; measurements to `workers.json` and `workers.log`,
CPU profile to `workers.pprof`. Reference/output polynomials and binary remain
outside Git at `.cache/liftrs/parallel-root-workers/roots-01/`. Review retains
strict source matching except the single explicitly preserved ignored-source
snapshot documented in prerequisite77126; that snapshot must match the producer
seal, never a blanket dirty-tree allowance. GPU component measurements and correctness are now reviewed below; no fullsize
time/memory improvement or production worker selection is established.
Local root partition tests pass and the optional GPU test compiles/skips without
its explicit output environment. Review tests reject missing/imbalanced/corrupt
results. Submit and hand off without monitoring.

## Completed worker comparison (SCALING-110)

All four fixtures pass complete native CPU polynomial equality, four scalar
extension-field root evaluations per fixture, normalized constant/degree checks
and persisted-output SHA256 equality. All four configurations (whole-root,
1/2/4 independent workers), warm-up and both timed samples are present. Small
**genuine s8** memcheck covers all configurations and reports0errors/0leakedbytes;
large synthetic s12/s17 cases are not sanitizer coverage. Partition race/admission
checks and inherited whole-workflow CPU/reference/CLI/restart/race/ASan/vet and
memcheck gates pass. No producer algorithm changed during this review.

**596 source /43 evidence /99 original cache identities** reconcile, including
one explicitly named exact ignored-source snapshot described in prerequisite77126.
All other source identities match producer Git objects. Original ignored source
and original logs are unchanged; caches and profiles stay retained outside Git.

### Warm profiled elapsed time, partition and complete merge included

Cells are arithmetic means of two measured samples, excluding sample−1. Units
are seconds. Config0 is whole-root on four GPUs. Config1/2/4 uses that many
one-GPU workers and the retained four-GPU merge executor, as described above.
Construction/setup is excluded from these operation clocks and reported below.

| Field / root input | Whole-root,4 GPUs | 1 worker | 2 workers | 4 workers |
|---|---:|---:|---:|---:|
| s8 genuine support,7381 exponents |0.018061 s |0.019742 s |0.015707 s |0.017343 s |
| s12 synthetic,65536 exponents |0.350214 s |0.335428 s |0.217454 s |0.212747 s |
| s17 synthetic,65536 exponents |1.762294 s |1.819699 s |1.551633 s |1.931126 s |
| s17 synthetic,262144 exponents |3.431855 s |3.992141 s |3.708884 s |2.944695 s |

The largest case shows **1.165 times the observed whole-root/4-worker ratio**
(about14.2% less elapsed time), including partitioning and merge. Its two samples
are3.523821/3.339889s whole-root versus3.063059/2.826332s four-worker. At s12,
four-worker mean is about1.646 times faster by this ratio. At s17/65536,
four workers are **slower** than whole-root; two workers have the best observed
mean. No fixed worker count wins across these fixtures, and two fixed-order
profiled samples do not establish confidence bounds or a fulls17/s18 speedup.

Largest-case mean decomposition:

| Configuration | Partition/validation | Joined worker wall time | Complete merge | Operation total | End-of-call owned GPU allocations, aggregate |
|---|---:|---:|---:|---:|---:|
| Whole-root,4 GPUs |not separate |3.431855 s |none |3.431855 s |20.463 MiB |
| 1 worker |0.715685 s |3.276449 s |about1 microsecond |3.992141 s |21.380 MiB |
| 2 workers |0.803777 s |2.764505 s |0.140595 s |3.708884 s |36.689 MiB |
| 4 workers |0.617872 s |2.017420 s |0.309395 s |2.944695 s |43.102 MiB |

The four-worker root phase alone improves substantially, but partition/validation
and merge consume about0.927s. Operation totals also contain small orchestration
costs, so rounded subtimes need not sum exactly. Owned allocation observations
are not physical GPU peaks and exclude contexts/runtime. Separate constructor
setup for the largest case is about0.109/0.109/0.111/0.187s for0/1/2/4workers.
The very first genuine-case whole-root setup takes4.723s, while later setup is
usually about0.1s after GPU/runtime initialization. Thus setup comparisons are
not equivalent cold starts; do not claim one-shot end-to-end speedup from them.

### Validation allocation dominates this small-subset measurement

The [retained Go CPU profile](runs/roots-01/workers.pprof) and its
[top summary](runs/roots-01/cpu-top.txt) show43.11% cumulative CPU samples in
`validateExponentSupport`,20.45% flat in memory clearing and23.93% flat in
`runtime.cgocall`. These percentages overlap and must not be added. The profile
covers the whole component, including references, setup and scalar checks, not
just the timed worker rows; CPU samples are not wall-time or GPU-time fractions.
Native/GPU execution and waits remain opaque inside cgo.

[Validation](../../encoding/liftrs-shared/preprocessing_core.go) builds a
[newExponentSet](../../encoding/liftrs-shared/exponent_set.go) bitmap over all N−1
possible exponents. Its requested backing is
`8 * ceil((N−1)/64)` bytes: **2 GiB at s17**, independent of whether a subset has
65536,262144 or billions of exponents. Global partition validation allocates one;
each worker root validation allocates another. Four worker validations can
therefore overlap **8 GiB of bitmap allocations** at s17, before other state;
that is allocation payload, not measured resident physical memory. Global dead
bitmaps/GC retention are additional. At s18 it projects **8 GiB/invocation or
32 GiB for four simultaneous workers**. This stage overhead is separate from the
K-index partition copy and must not be silently added as permanently live Newton
storage. It limits this prototype's current representativeness for sparse roots.

Maximum sampled process-group RSS is **5064.574 MiB (4.946 GiB)**, within the6GiB
guard; scheduler batch MaxRSS is3409044KiB (about3.25GiB), from a different sampler.
Neither is per-worker bitmap size or algorithm-only memory. Normal component
command83.476s includes fixtures/references/scalar checks/comparisons/persistence;
small memcheck6.244s. Total reusable cache **37.767 MiB** remains outside Git,
including polynomials and binary. No profiles or useful prior caches were removed.

### Decision and next measurement

Independent workers are correct and sometimes improve operation time, but the
current results **do not justify committing to sixteen GPUs** or using a uniform
parallel speedup in fulls18 estimates. Keep the validated resident default.

Before interpreting root-worker scaling, measure a controlled alternative for
**subset validation** that preserves sortedness, range and complete Frobenius
membership checks while avoiding an N-sized bitmap for a small subset. A sorted
membership search needs no new N-sized storage but adds O(K log K) searches;
an O(K) membership table uses extra subset-sized storage and setup. Their actual
latency trade-offs are unmeasured, especially for dense full supports, so no such
mode is adopted here. Retain the faster dense validator for full supports when
it fits unless measurements justify changing it; any slower memory-saving choice
must remain explicit. Reusing a proof from deterministic partitioning requires
an explicit trusted ownership boundary, not silently skipping validation.

Record the extra validator bitmap in host/resource planning and repeat matched
root-worker measurements after that controlled comparison, with adequate host
headroom and the same complete CPU/scalar/hash checks. Then proceed to node-to-node
GPU staging/synchronization/rank ownership, followed separately by buffered
persistent output and bounded writer concurrency. No new job, larger resource
budget or validation shortcut follows from this review.
