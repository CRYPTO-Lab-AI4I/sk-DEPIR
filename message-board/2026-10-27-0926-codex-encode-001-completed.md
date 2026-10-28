# ENCODE-001 — Paper/cyclic encoders and reusable public artifacts verified

- Timestamp: 2026-10-27T09:26:38Z
- Agent: Codex
- Status: completed

## Implementation and separation

Added the bounded [orbit API](../encoding/liftrs-shared/orbit.go) for m=2, s=2..8,
N<=65,536. The stages are explicit: `DiscoverOrbit` uses code specifications and
randomness only, `PreparePaper`/`PrepareCyclic`/`PrepareCyclicFFT` build code-only
constants, loaders reconstruct process-local runtime state, and `Encode` takes
only a database's message elements. The paper and cyclic paths share the paper's
omega-power information set, exact dimension and output coordinates.

[Arithmetic](../encoding/liftrs-shared/orbit_arithmetic.go),
[additive transforms](../encoding/liftrs-shared/orbit_fft.go), and
[polynomials](../encoding/liftrs-shared/orbit_polynomial.go) implement the quadratic tower,
general-basis FFT, Karatsuba, Newton division and cached transposed-Vandermonde
solver. The cached-FFT cyclic variant also stores the fixed generator transform;
no method repeats these preprocessing steps for each database.

[Versioned artifacts](../encoding/liftrs-shared/orbit_artifact.go) serialize all paper
solver reciprocals/weights or cyclic P/G and optional generator spectrum.
They are public code data, suitable for GitHub distribution with provenance.
Structural validation does not certify untrusted artifact mathematics. Complete
artifact loading rebuilds field/FFT runtime tables, not polynomial preparation.
The smaller basis-only format still requires method preparation.

## Evidence and documentation

[Tests](../encoding/liftrs-shared/orbit_test.go) check every message basis vector on eight
tiny codes against the independent dense oracle, randomized messages, exact
paper/cyclic equality, transforms versus Horner, Karatsuba versus schoolbook,
artifact round-trips, failures and concurrent reuse. A
[runnable public example](../encoding/liftrs-shared/orbit_example_test.go) covers published
artifact loading and encoding.

Created the [comparison experiment](../experiments/encoder-comparison/report.md),
with [script](../experiments/encoder-comparison/run.py),
[driver](../experiments/encoder-comparison/main.go),
[raw results](../experiments/encoder-comparison/results.json),
[table](../experiments/encoder-comparison/table.md), and
[public artifacts](../experiments/encoder-comparison/artifacts/). All five code
configurations produce identical complete outputs across three arithmetic paths.
Report discovery, method preparation, serialization, loading, encoding, cumulative
allocations, process peak RSS, sync output write and cached readback separately.
Public prepared artifacts in this grid remain below 1 MiB each. The N=65,536
CPU sample measured medians of 12.91 ms paper, 93.60 ms cyclic Karatsuba and
8.90 ms cyclic cached FFT, with preprocessing excluded; setup is substantially
more expensive for cyclic. These are three local samples, not a universal winner.

Updated [README](../README.md), [package documentation](../encoding/liftrs-shared/README.md),
[encoder contract](../docs/encoder-contract.md),
[algorithm review](../docs/encoding-algorithms.md), [roadmap](../ROADMAP.md), and
[experiment index](../experiments/README.md). README checked and changed for the
new API, comparison evidence and remaining limitations. Prior edits preserved.

## Verification and limitations

- `go test ./encoding/liftrs -count=1`: passed.
- `go test -race ./encoding/liftrs`: passed, including concurrent reuse.
- `go vet ./encoding/liftrs ./experiments/encoder-comparison`: passed.
- Full `go test ./...` in an isolated copy: passed; isolation avoids legacy CSV
  writes affecting this checkout. Subsequent metadata getters were covered by
  the final race/test and vet runs.
- Comparison script completed; all source and artifact SHA256 hashes checked.
  All three encoder hashes match in all five cases. `git diff --check`: passed.
- `/usr/bin/time -l` could not query sandbox clock metadata; the final script uses
  per-child POSIX `wait4` RSS, documented in the report.

The code remains an in-memory comparison backend. No s=9..20, external-storage,
GPU, constant-time or production-security claim is added. The legacy RM encoder
and catalogue promotion are unchanged. Cyclic FFT uses extension-field arithmetic
inside multiplication; a base-field-only quasi-linear multiplier and the proposed
gcd-based discovery remain unimplemented. No commit or push was requested here.
