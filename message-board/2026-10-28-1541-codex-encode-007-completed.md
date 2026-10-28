# Resource-gated cyclic benchmarks through s=12 completed

- Timestamp: 2026-10-28 15:41 UTC
- Work item: ENCODE-007
- Status: completed

## Implementation and policy

The user confirmed the one-hour cutoff applies to discovery + reusable
preprocessing + one database encoding. RAM budget is 12 GiB on the 48 GB host,
with a 1.5 forecast safety factor. The [progressive runner](../experiments/cyclic-scaling/run.py)
and [policy](../experiments/cyclic-scaling/policy.json) update projections from the
largest completed case, enforce a one-hour whole-case timeout and use a Go memory
guard/soft GC limit. [The driver](../experiments/cyclic-scaling/main.go) separates
exact discovery, reusable cyclic preparation, repeated encoding and validation/I/O.
All data-dependent code is unchanged; prepared caches remain in memory instead of
serializing large JSON. Existing 2 MiB/64 MiB loader limits remain unchanged.

The [orbit bound](../encoding/liftrs/orbit.go),
[field-adapter documentation](../encoding/liftrs/orbit_arithmetic.go) and
[invalid-spec test](../encoding/liftrs/orbit_test.go) now permit resource-limited
s<=13 experiments, N<=67108864. Accepted specifications and completed benchmark
support are distinct: measured catalogue evidence stops at s=12.

## Results and evidence

The [report](../experiments/cyclic-scaling/report.md),
[results](../experiments/cyclic-scaling/results.json),
[table](../experiments/cyclic-scaling/table.md),
[s=11 events](../experiments/cyclic-scaling/s11-events.jsonl),
[s=12 events](../experiments/cyclic-scaling/s12-events.jsonl) and
[driver tests](../experiments/cyclic-scaling/main_test.go) preserve exact commands,
source hashes, workload/guard boundaries, host details and limitations.

- s=11,m=2,t=6: exact K=455434,N=4194304. Discovery 15.07 s (two samples),
  preprocessing 145.26 s, median encoding 14.60 s, first use 2.92 min,
  whole-case peak RSS 0.26 GiB.
- s=12,m=2,t=6: exact K=1965739,N=16777216. Discovery 34.98 s (one sample),
  preprocessing 677.19 s, median encoding 67.76 s, first use 13.00 min,
  peak RSS 1.19 GiB.
- All systematic symbols, sampled public coordinate mappings, 16 independent
  actual-curve sums and packed sync/readback checks passed for both cases.
- Actual case wall time including three encodings and verification/I/O was
  204.82 s / 914.88 s. No runtime or memory guard triggered.
- s=13 was not admitted: conservative first-use forecast 171.7 min, RAM 10.86 GiB.
  Time alone exceeds the user policy. s>=14 also exceeds the RAM policy. The
  report does not claim excluded fields are impossible or that distant model
  outputs represent an executable larger backend.

These add to existing s=8..10 measurements, covering all admitted catalogue fields
through s=12. No CPU-specific, GPU, external-memory or large paper-encoder run was
added. Heap samples, cumulative allocations and whole-process RSS are labelled
separately; the memory sampler/guard is included in phase timings. Large-cache
JSON size/loading and isolated phase RSS were not measured.

## Documentation, catalogue and verification

Promoted only the measured s=11/12 entries in the
[catalogue](../pirconfig/configs_catalogue.json), with exact capacity, fixed field
polynomials and in-memory cyclic evidence. Historical Monte Carlo capacities and
cost-grid snapshots remain. [Catalogue tests](../pirconfig/catalogue_test.go)
check the recorded dimensions/verification metadata.

Updated [README](../README.md), [roadmap](../ROADMAP.md),
[encoder contract](../docs/encoder-contract.md),
[algorithm plan](../docs/encoding-algorithms.md),
[parameter planning](../docs/parameter-planning.md),
[package README](../encoding/liftrs/README.md),
[catalogue README](../pirconfig/README.md),
[experiment index](../experiments/README.md) and
[s=10 profile context](../experiments/encoder-profile-s10/report.md).
README was checked; it distinguishes the experimental bound from measured support.
Previous uncommitted ENCODE-006 profiling work is preserved.

- Targeted tests for liftrs, pirconfig and cyclic-scaling: passed.
- `go test -race` and `go vet` for those packages: passed.
- `go test ./...` in an isolated repository copy: all packages passed; isolation
  prevents legacy CSV-writing tests altering working-tree data.
- Driver tests cover independent arithmetic through s=20 and bounded-buffer packed
  output through s=13 against the bit definition.
- Source/seed/event hashes verify. Every forecast reproduces from its recorded
  anchor, policy and c06375c catalogue input; policy matches the user choice.
- Documentation file links and `git diff --check`: passed.

The requested baseline is complete. Use this scaling boundary and existing CPU
profiles to select arithmetic/transform improvements, then repeat the workload
before admitting larger cases. Changes are uncommitted.
