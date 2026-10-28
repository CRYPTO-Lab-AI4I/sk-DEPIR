# CPU encoding at s=10 validated and reported

- Timestamp: 2026-10-28 13:15 UTC
- Work item: ENCODE-005
- Status: completed

## Implementation and evidence

Extended [orbit discovery/loading](../encoding/liftrs/orbit.go) and its
[packed field adapter](../encoding/liftrs/orbit_arithmetic.go) to m=2,s<=10,
N<=1048576. The measured catalogue case is s=10,t=6, polynomial 0x409 with exact
K=103556, established by complete enumeration and reached by degree discovery.
The previous Monte Carlo point K=103952 remains separately recorded.
The [artifact loader](../encoding/liftrs/orbit_artifact.go) retains JSON version 1;
its complete-artifact cap rises from 32 to 64 MiB to load the observed 35.74 MiB
paper solver. This cap is independent of publication/RAM budgets. The discovery
artifact cap remains 2 MiB. No encoding algorithm, coordinate order, information
set, mask/permutation behavior or application wrapper changed.

The [new experiment report](../experiments/encoder-scaling/report.md) links the
[script](../experiments/encoder-scaling/run.py),
[verification/memory driver](../experiments/encoder-scaling/main.go),
[driver test](../experiments/encoder-scaling/main_test.go),
[results](../experiments/encoder-scaling/results.json),
[table](../experiments/encoder-scaling/table.md) and
[public artifacts](../experiments/encoder-scaling/artifacts/).
Discovery/preparation/load/encoding/packing/output I/O run with explicit boundaries;
each stage/method has a fresh process. Separate encoding heap sampling records
loaded-cache/message baseline and observed increase, including output; it can miss
transient peaks and is not an exact workspace bound. The report preserves the
preflight observations and source hashes that motivated the cap increase.

## Measured results

Discovery: 1.447 s. Cyclic FFT: 26.475 s reusable preparation, 5.15 MiB artifact,
53.45 ms loading and 2.682 s median encoding. Paper: 23.971 s preparation,
35.74 MiB artifact, 165.34 ms loading and 10.708 s encoding. Preparation/encoding
process RSS: 74.28/55.88 MiB cyclic, 245.36/120.98 MiB paper. Separate sampled heap
increase: 10.40/34.63 MiB, including the 4 MiB unpacked output. Packed output is
1.25 MiB per codeword; it is not a GB/TB application encoding benchmark or grouped
storage measurement. No target-scale performance extrapolation is made.

Both methods produce identical complete outputs on timing seed 71 and verification
seed 7310. All 103556 systematic positions and 64 independently evaluated degree-six
curves pass. The [new fixture test](../encoding/liftrs/s10_checkpoint_test.go)
recounts the dimension and checks another message/16 curves using independent
carryless arithmetic. The [catalogue](../pirconfig/configs_catalogue.json) and
[tests](../pirconfig/catalogue_test.go) now record exact standalone s=10 support;
historical planning snapshots are unchanged. The existing
[timing driver](../experiments/tablefree-encoding/main.go) is reused; only its
support comment and [packing test](../experiments/tablefree-encoding/main_test.go)
coverage were extended to s=10.

## Documentation and verification

Updated [README](../README.md), [roadmap](../ROADMAP.md),
[encoder contract](../docs/encoder-contract.md),
[algorithm/scaling plan](../docs/encoding-algorithms.md),
[parameter planning](../docs/parameter-planning.md),
[package README](../encoding/liftrs/README.md),
[catalogue README](../pirconfig/README.md),
[experiment index](../experiments/README.md) and
[previous polynomial report's next-work section](../experiments/fast-polynomial/report.md).
The plan now prioritizes CPU scaling before format cleanup unless size/loading
obstructs measurement. README was checked against actual implemented support.

- Targeted tests for liftrs, pirconfig and both experiment drivers: passed.
- `go test -race` for those four packages: passed.
- `go vet` for those four packages: passed.
- `go test ./...` in an isolated repository copy: all packages passed; isolation
  avoids legacy CSV-writing tests changing the working tree.
- Final source/artifact hashes and preflight source hashes verified; updated
  documentation file links resolve; `git diff --check` passes.

Next prioritize the cyclic backend and profile the current s=10 pipeline to select
its next arithmetic/transform optimization before advancing s=11/12. JSON loading
is about 2% of one cyclic encoding and amortized across databases, so format
cleanup is not the immediate bottleneck. Larger encoders, grouped storage,
streaming and GPU support remain unimplemented. Changes are uncommitted.
