# Shared root/generator boundary validation

Status: **job77126 submitted**, native/CUDA/recovery validation pending.
[Receipt](submission.json), [results](runs/root-01/), [Slurm log](slurm-77126.log).
Prepared source revision `9a1bbf7`; actual revision/file identities are sealed by
the runner. No monitoring after submission.
Work SCALING-106. No full s18 support, distributed worker or memory saving claimed.

[Runner](run.py), [reviewer](analyze.py), [Slurm entry](run.sbatch),
[partition/boundary test](../../encoding/liftrs-shared/preprocessing_root_test.go),
[shared implementation](../../encoding/liftrs-shared/preprocessing_core.go).

The existing preparation procedure now calls a normalized-root helper followed
by a reciprocal/full-identity helper. Both CPU and GPU paths retain the same
factor builder, product dispatch, root normalization, recovery source binding,
immutable saved roots and complete identity check. Input recovery is captured
before callbacks, preserving its original resume boundary. Generator preparation
needs K and P rather than the support index slice. The public API, canonical
representation and default execution schedule are unchanged. Support checksums
are captured as small metadata instead of referenced by the root-save closure.

This enables independent Frobenius-closed subsets to use the same root helper;
their normalized products merge into the complete normalized P. The new test
partitions complete conjugate orbits into1/2/4 subsets in s3/s4/s6, compares their
merged product with whole-root preparation, checks caller immutability and checks
that root-publication failure prevents generator work. This is sequential
arithmetic validation, not parallel or cross-node performance. No asynchronous
worker implementation or forced GC is introduced; CLI-owned support/recovery
release before Newton remains future work. No measured RAM/latency saving follows.

The bounded runner reuses the existing four-H200 controller and all small complete
CLI/regression/reference/race/ASan/vet/restart/memcheck gates. It additionally runs
shared native recovery/export/root tests, race checks, one/four-card GPU small
preparation/export and one-card export memcheck. Go CPU profiles are retained for
native and GPU test phases; logs have phase/work progress and automatic guards.
Timings include test/profile/logging overhead and are not latency benchmarks.

Existing admitted HPC budget: one node,4H200,4CPU cores,8GiB requested host RAM,
6GiB RSS guard,4GiB Go target,1GiB/card aggregate4GiB owned-device cap,2GiB new
retained cache,4GiB scratch;10min/command,15min controller,20min scheduler.
New shared GPU tests cap128MiB/card; inherited CLI regressions retain1GiB/card.
No node pin; gnode03 excluded. This does not admit any multi-node resources.

```sh
python3 -B experiments/preprocessing-root-boundary/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/preprocessing-root-boundary/slurm-%j.log \
  experiments/preprocessing-root-boundary/run.sbatch root-01
```

The runner seals its exact source revision/file hashes and all evidence/cache
files in `runs/root-01/results.json`. Logs/profiles stay in that directory;
reference caches and binaries remain outside Git at
`.cache/liftrs/preprocessing-root-boundary/root-01/`. After submission, await the
user completion ping before reviewing. Full-size caches and prior profiles stay
untouched. Local new root test passes in0.046s; full gates are not yet passed.
