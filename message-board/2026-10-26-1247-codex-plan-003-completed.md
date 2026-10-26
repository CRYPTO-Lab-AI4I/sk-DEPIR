# PLAN-003 — Supplied-plan evaluator implemented

- Timestamp: 2026-10-26 12:47:11 UTC
- Agent: Codex
- Status: completed

Added [planner.Evaluate](../planner/evaluate.go), [tests](../planner/evaluate_test.go),
a runnable [example](../planner/example_test.go), and [usage documentation](../planner/README.md).
The API takes a rectangular plan, configuration catalogue, and optional security/bit
budgets. It validates canonical coordinates, exact coverage, capacity, group ordering
and compatibility, and requested catalogue security profiles. Logical costs count
full storage and separate position uploads/downloads for every database, including
dummies and repetitions. All integer arithmetic is checked for overflow.

Coverage uses pairwise rectangle checks plus total area, independent of application
matrix size. This is quadratic in rectangle count, explicitly documented rather than
presented as a scalable partition-search algorithm. Group members reference the same
configuration ID; alias normalization is not implemented. The evaluator checks row
compatibility but neither selects permutations nor handles keys.

Estimated capacities are allowed for provisional plans; unknown capacities are
rejected. Unsupported encoders also make plans provisional. Results copy used
capacity/security evidence, report the minimum security profile, and distinguish
logical costs from byte alignment or transport overheads. Resource measurements
are informational and have no effect. No preprocessing, encoding, query generation,
or partition search runs inside evaluation.

Updated [planner design](../docs/parameter-planning.md), [data representation](../docs/data-representation.md),
[catalogue usage](../pirconfig/README.md), [roadmap](../ROADMAP.md),
[agent repository map](../AGENTS.md), and the [root README](../README.md).
README review is complete, including removal of stale cost-evaluator status and
repeated planner-status prose. No catalogue parameters or experiment results changed.

## Verification

Using Go 1.27.3 and `GOCACHE=/tmp/sk-depir-go-cache`:

- `go test ./planner -count=1`: passed, including the public example.
- `go test ./... -count=1`: all eight packages passed in isolated copy
  `/tmp/sk-depir-plan003.xkeiCC`; legacy CSV-writing tests did not alter the checkout.
- `go test -race ./planner -count=1` and `go vet ./planner`: passed.
- Hand fixture: two GF(4) databases encode three 4-bit records with 64 logical
  storage bits, 1,048 upload bits, 524 download bits, and storage rate 0.1875.
  Actual application-reference storage and read counts agree; retrieval recovers
  every original record. Byte-range response alignment is excluded from logical D.
- Tests cover mixed fields, exact/violated budgets, security/sample requirements,
  general partitions, shared-query row compatibility, gaps/overlaps (including equal
  total area), unknown/estimated capacities, evidence ownership, resource-cost
  independence, concurrency, and arithmetic overflow.
- A two-block `s=20` plan checks terabyte-scale application coordinates against
  measured capacity estimates without constructing or encoding that database.
  This is an arithmetic unit test, not large-scale encoder validation or a benchmark.
- `gofmt`, `git diff --check`, updated Markdown links, and existing experiment
  computational-source hashes: verified. Schema tests and rate experiments were
  not repeated because their formats, sources, and results did not change.

Next: implement partition generation/comparison using this evaluator. Target-size
encoder support and stored-manifest persistence remain separate work. No commit or
push was requested for this work item.
