# Fast full and low polynomial products validated

- Timestamp: 2026-10-28 12:04 UTC
- Work item: ENCODE-004
- Status: completed

## Implementation

Added the [hybrid product backend](../encoding/liftrs/orbit_product.go) and
[tests/kernel benchmarks](../encoding/liftrs/orbit_product_test.go).
[Variable-size subspace transforms](../encoding/liftrs/orbit_fft.go) multiply large
polynomials without requiring a full-field transform for every product. Products
beyond one field domain use overlap-add; small products retain Karatsuba. Low
products prune irrelevant recursive work, and unequal operands use balanced blocks
below the FFT cutoff. FFT padding always covers the full local convolution to
avoid aliasing high coefficients into the prefix.

[Orbit setup and both encoders](../encoding/liftrs/orbit.go) reuse immutable
transform constants through the [field adapter](../encoding/liftrs/orbit_arithmetic.go).
[Newton inversion and division](../encoding/liftrs/orbit_polynomial.go) now request
low products where appropriate. Database-dependent buffers remain per call.
No public API, artifact version, serialized bytes, code parameters, supported
encoder size or information-set mapping changed.

## Results and documentation

The [new experiment report](../experiments/fast-polynomial/report.md) links the
[script](../experiments/fast-polynomial/run.py),
[results](../experiments/fast-polynomial/results.json),
[kernel timings](../experiments/fast-polynomial/kernels.txt),
[comparison table](../experiments/fast-polynomial/table.md) and the retained
initial trial with exact source snapshots. The script builds baseline `7e5a8ac`
and the modified implementation on the same machine, separately measures reusable
preparation and database encoding, and compares artifact/output hashes.

At s=9, paper encoding improves 4143.95→1544.79 ms (2.68×), and cyclic FFT
encoding 490.45→430.43 ms (1.14×). Preparation improves 8522.68→3311.10 ms for
paper and 3904.11→3409.64 ms for cyclic. Allocated bytes per call fall substantially.
At s=8, paper encoding improves 1.32× and cyclic encoding is effectively unchanged;
cyclic preparation regresses about 15% (385.53→442.82 ms). This limitation and
conservative crossover choices are explicit. These are short local measurements,
not larger-instance projections or evidence of bounded-RAM feasibility.

Updated [README](../README.md), [roadmap](../ROADMAP.md),
[encoder contract](../docs/encoder-contract.md),
[algorithm/scaling plan](../docs/encoding-algorithms.md),
[package documentation](../encoding/liftrs/README.md) and
[experiment index](../experiments/README.md). Appended successor links to the
[table-free checkpoint](../experiments/tablefree-encoding/report.md) and
[CPU-profile report](../experiments/encoder-profile/report.md); their measurements
are unchanged. README was checked against the current bounded implementation.

## Verification and remaining work

- `go test ./encoding/liftrs -count=1`: passed.
- `go test -race ./encoding/liftrs ./pirconfig`: passed.
- `go vet ./encoding/liftrs ./pirconfig`: passed.
- `go test ./...` in an isolated repository copy: all packages passed; isolation
  avoids legacy tests modifying working-tree CSV data.
- Old/new s=8 and s=9 prepared artifacts are byte-identical; full canonical and
  packed output hashes and every systematic symbol agree.
- Independent convolution, subspace Horner/round-trip, division/inversion,
  domain-boundary, unequal-operand and forced-FFT dense-pipeline checks pass.
- Final and initial source/input/kernel hashes verified; updated documentation
  file links resolve; `git diff --check` passes.

This completes the requested bounded fast polynomial-product step. Compact public
artifacts and moderate-instance CPU scaling remain next in the plan; field-kernel
improvements can be measured separately. Specialized truncated FFTs, new CPU
profiles, streaming, GPU and s>=10 encoding were not added. Changes are uncommitted.
