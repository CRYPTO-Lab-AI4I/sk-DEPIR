# ENCODE-002 — Table-free arithmetic and s=9 encoding verified

- Timestamp: 2026-10-28T11:00:00Z
- Agent: Codex
- Status: completed

## Implementation

Added public [BinaryField / QuadraticField](../encoding/liftrs-go/binary_field.go)
arithmetic for every s=2..20: checked irreducible polynomials, portable carryless
multiplication/reduction, powers, inverse and trace. Quadratic elements have two
uint32 components, with a fixed-size tower description; no arithmetic table grows
with q or q². Existing field polynomials are retained; degree 9 uses 0x211.
Methods requiring canonical inputs and non-constant-time behavior are documented.
A [runnable example](../encoding/liftrs-go/binary_field_example_test.go) uses s=20.

Replaced [orbit arithmetic tables](../encoding/liftrs-go/orbit_arithmetic.go) with a
bounded adapter to those operations, preserving coordinate enumeration, tower delta
and first primitive generator. [Orbit preprocessing/loading](../encoding/liftrs-go/orbit.go)
now accepts m=2,s=2..9,N<=262144. Sampling uses four bytes per good-monomial
coefficient above s=8 so it remains uniform; earlier fields keep their one-byte
sampling convention. Full-field arithmetic tables are gone, but transform and
polynomial buffers remain in memory. The full encoder is not extended to s=10..20.
Fast truncated multiplication, GPU and external-memory work are not implemented.

## Correctness and reusable evidence

[Independent arithmetic tests](../encoding/liftrs-go/binary_field_test.go) exhaustively
compare base multiplication through s=8 and extension multiplication through s=4;
randomized independent multiplication, inverse and Frobenius checks cover all s
through 20. Tests reproduce all prior paper/FFT output hashes, regenerate the s=8
prepared artifact bytes identically, and check s=9 systematic symbols plus eight
sampled degree-7 curve constraints using separate arithmetic. Updated the
[invalid-spec test](../encoding/liftrs-go/orbit_test.go) to reject s=10 instead of s=9.

The [new experiment report](../experiments/tablefree-encoding/report.md),
[script](../experiments/tablefree-encoding/run.py), [driver](../experiments/tablefree-encoding/main.go),
[packing test](../experiments/tablefree-encoding/main_test.go),
[results](../experiments/tablefree-encoding/results.json), [table](../experiments/tablefree-encoding/table.md)
and [public artifacts](../experiments/tablefree-encoding/artifacts/) keep discovery,
per-method preparation and encoding in fresh separate processes. Complete artifacts
are loaded once per encoding process. Preparation, loading, packing, output sync,
cached readback, allocations and process peak RSS are recorded separately.
Both methods agree on every canonical/packed output and all systematic symbols.

At s=9,t=7, exact K=17200,N=262144. Local CPU median encoding times were 4.114 s
paper versus 0.508 s cyclic FFT; cyclic preparation 3.918 s plus shared discovery
0.264 s, artifact 1204.1 KiB, load 14.51 ms, encoding-process peak RSS 17.98 MiB.
The latter includes loading and output buffers, not only encoding workspace.
Table-free scalar arithmetic is slower than prior table-backed small-instance
arithmetic; the established benefit is bounded arithmetic state, not a throughput
win. Timings are three local samples and cannot establish scaling or hardware
optimality. CPU/memory details were captured; diskutil device inspection failed
under sandbox DiskManagement restrictions and is explicitly reported.

[Current catalogue](../pirconfig/configs_catalogue.json) records exact s=9 capacity,
polynomial and standalone support with separate encoding evidence. Prior Monte Carlo
K=17118 remains evidence, and all historical planning snapshots are unchanged.
[Catalogue tests](../pirconfig/catalogue_test.go) link exact dimensions to saved
agreement results. No integrated application encoding or security qualification
is implied. The other 11 target configurations remain estimated/candidate.

## Documentation and verification

Updated [package README](../encoding/liftrs-go/README.md), [root README](../README.md),
[encoder contract](../docs/encoder-contract.md), [algorithm plan](../docs/encoding-algorithms.md),
[parameter design](../docs/parameter-planning.md), [roadmap](../ROADMAP.md),
[catalogue README](../pirconfig/README.md), [experiment index](../experiments/README.md),
and historical [encoder](../experiments/encoder-comparison/report.md) /
[rate](../experiments/liftrs-rates/report.md) reports. README checked and changed for
new arithmetic/API limits, exact evidence and current performance source.

- Targeted liftrs/catalogue/packing tests: passed.
- `go test -race ./encoding/liftrs ./pirconfig ./experiments/tablefree-encoding`: passed.
- Targeted `go vet` on those packages: passed.
- Full `go test ./...` in an isolated copy: passed (legacy CSV writes stay isolated).
- Experiment completed twice; deterministic dimensions, artifacts and output hashes
  agree. Final source/artifact hashes and relative documentation links verified.
- `git diff --check`: passed. No commit/push requested this turn.

Next: fast truncated multiplication for H=(P*M) mod X^K and preprocessing polynomial
kernels, followed by compact artifacts and further CPU scaling checkpoints. Later
external-memory/GPU decisions remain subject to the documented user review gates.
