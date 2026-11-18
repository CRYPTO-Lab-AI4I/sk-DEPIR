# Independent GPU root workers, including merge

Status: **job77148 submitted**; GPU performance/correctness validation pending.
[Receipt](submission.json), [results](runs/roots-01/), [Slurm log](slurm-77148.log).
Prepared source `00360cc`; the controller seals actual revision and copied sources.
No monitoring after submission.
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
seal, never a blanket dirty-tree allowance. GPU measurements and correctness are
pending; no expected speedup or fullsize time/memory improvement is asserted.
Local root partition tests pass and the optional GPU test compiles/skips without
its explicit output environment. Review tests reject missing/imbalanced/corrupt
results. Submit and hand off without monitoring.
