# Shared root/generator boundary validation

Status: **validation77126 passed** on HPC gnode13 (4H200,4CPU,8GiB requested
host RAM), complete job4min9s; controller247.575s.
[Receipt](submission.json), [raw results](runs/root-01/results.json),
[analysis](runs/root-01/analysis.json), [accounting](runs/root-01/accounting.txt),
[Slurm log](slurm-77126.log). Producer `9a1bbf7`.
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

## Completed validation and source reconciliation (SCALING-109)

Native root/export/recovery and race checks pass, along with whole small GPU
preparation/export and CUDA memcheck. Both the new export memcheck and inherited
CLI memcheck report0errors/0leakedbytes. All small canonical reference, cache/root
restart, ordinary CPU-only readback, race/ASan/vet and root-partition/publication
failure gates pass. Matching CPU profiles/native and GPU binaries are retained.
Maximum sampled command RSS587.914MiB includes build/test work, not a measured
root-lifetime saving. No parallel performance is established by this validation.

Reconciliation verifies589source identities,45evidence logs and75original cache
files. One source file was ignored by Git but copied into the producer snapshot;
it is unrelated to the executed workflow. The reviewer explicitly binds only
that named source to its exact producer SHA256 using a retained snapshot at
`.cache/liftrs/preprocessing-root-boundary/root-01/review-source-inspect.py`.
All remaining source identities match producer Git objects. The ignored original
is neither modified nor published; the preserved copy remains outside Git.
No execution result or sealed log was changed. This is an explicit source override,
not a relaxed default reviewer or a claim of an entirely clean producer snapshot.

```sh
python3 -B experiments/preprocessing-root-boundary/analyze.py root-01
```

Subsequent independent root-worker measurement is tracked in
[its experiment](../parallel-root-workers/report.md). It uses the same validated
root and product implementations. This review changes no canonical format, API,
execution default, admitted full-size workload or memory claim.
