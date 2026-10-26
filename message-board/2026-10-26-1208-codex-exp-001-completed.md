# EXP-001 — All PIR configuration rates estimated

- Timestamp: 2026-10-26 12:07:13 UTC
- Agent: Codex
- Status: completed

## Implementation and documentation

Added [liftrs.GoodMonomial](../encoding/liftrs/monomial.go): an exact binary-carry
reformulation of the paper's monomial criterion, requiring `O(s*m*t)` time and
constant space rather than a field-sized residue table. Supports `s=2..20`,
`m=2..3` without field arithmetic or preprocessing. The original residue DP and
small encoder limits remain unchanged. Added independent [comparison tests](../encoding/liftrs/monomial_test.go)
and a public [example](../encoding/liftrs/example_test.go).

Added the [pir-rate command](../cmd/pir-rate/main.go), [usage](../cmd/pir-rate/README.md),
and [tests](../cmd/pir-rate/main_test.go). It samples reduced exponent tuples with
replacement and emits counts, seeds, confidence bounds, provenance, and sampling-only
timings. Fixed-seed `math/rand` is used only for this statistical experiment, not
for protocol randomness.

## Experiment and catalogue

The [script](../experiments/liftrs-rates/run.py), exact [input snapshot](../experiments/liftrs-rates/pir-configs.json),
[results](../experiments/liftrs-rates/results.json), and [report](../experiments/liftrs-rates/report.md)
record one million samples per configuration across all 13 target choices.
Rates range from 4.7565% (`s=8,t=8`) to 26.0249% (`s=20,t=5`). Two-sided Hoeffding
intervals allocate a total 1% error budget across 13 configurations, giving 99%
simultaneous coverage under the independent-sampling model (approximately 0.1983
percentage points half-width). The four matching best-precision paper values lie
inside their respective intervals. Paper values and new measurements remain
separate evidence.

Updated every [catalogue entry](../pirconfig/configs_catalogue.json) with the empirical
point capacity and linked uncertainty metadata, preserving previous paper evidence.
Updated [catalogue tests](../pirconfig/catalogue_test.go) and [example](../pirconfig/example_test.go)
to check the recorded counts and current estimated status. All capacities remain
estimated and target encoding remains unsupported.

Updated the [root README](../README.md), [catalogue README](../pirconfig/README.md),
[encoder package README](../encoding/liftrs/README.md), [planner document](../docs/parameter-planning.md),
[roadmap](../ROADMAP.md), [experiment index](../experiments/README.md), and
[agent repository map](../AGENTS.md). README review is complete. No parameter or
security profile changed; no layout evaluator or large-instance encoder was added.

## Verification

Using Go 1.27.3 with `GOCACHE=/tmp/sk-depir-go-cache`:

- Membership tests exhaustively compared carry and residue criteria for small
  fields/dimensions/degrees, checked randomized larger fixtures and target-field
  edge cases, validated inputs, and confirmed zero allocations on valid calls.
  Existing actual-curve and dense-encoding tests remain independent checks.
- Command tests verified fixed-seed reproducibility, classification against the
  dense basis, confidence calculations including family error allocation, and errors.
- Initial catalogue tests caught a stale example expecting `unknown`; its output
  was updated to `estimated`. Final targeted tests passed.
- `go test ./... -count=1`: all seven packages passed in isolated copy
  `/tmp/sk-depir-exp001.laNNws`; legacy CSV-writing tests did not alter the checkout.
- `go test -race ./encoding/liftrs ./cmd/pir-rate ./pirconfig -count=1` and
  `go vet ./encoding/liftrs ./cmd/pir-rate ./pirconfig`: passed.
- `/tmp/sk-depir-catalogue-schema-env/bin/python pirconfig/schema_test.py`:
  all three schema tests passed with the updated catalogue.
- Repeated the full one-million-sample command against the input snapshot and
  reproduced counts, rates, intervals, and seeds for every configuration exactly.
- Recorded source/input SHA-256 hashes, documentation links, `gofmt`, and
  `git diff --check`: verified.

The experiment report states base revision plus dirty-source hashes, command,
environment (hardware model/RAM unavailable in the sandbox), timing boundaries,
statistical assumptions, and limitations. Estimates do not establish exact
capacities, cryptographic security, or encoder feasibility. Next: implement layout
cost evaluation using these provisional capacities and exact tiny fixtures.
No commit or push was requested for this work item.
