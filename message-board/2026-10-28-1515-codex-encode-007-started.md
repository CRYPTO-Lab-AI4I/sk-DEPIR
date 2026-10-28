# Benchmark feasible larger cyclic configurations

- Timestamp: 2026-10-28 15:15 UTC
- Work item: ENCODE-007
- Status: started

User requests progressive benchmarking while estimated runtime is under one hour
and RAM is reasonable. Start with s=11, update resource projections after each
measured case, and cover the remaining catalogue with labelled extrapolations.
A clarification is pending on whether the hour limits first-use or per-database
cost; initial policy uses first-use time and 12 GiB RAM on the 48 GiB host.

Add `experiments/cyclic-scaling/` with a progressive runner and per-case streaming
results. Keep code-only discovery and preparation separate from per-database
encoding. Hold prepared constants in memory rather than serializing large caches;
format cleanup is not a prerequisite. Record exact capacities, sampled-curve and
systematic-position checks, packed streaming hashes, process/heap observations,
source hashes and forecast limitations. Resource limits and incomplete stages must
remain explicit. Update README, catalogue evidence, limits and plan only as justified
by completed runs; preserve earlier uncommitted profiling work and all measurements.
