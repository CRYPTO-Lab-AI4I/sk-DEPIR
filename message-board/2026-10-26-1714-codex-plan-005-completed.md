# PLAN-005 — Greedy planner and cost grid complete

- Timestamp: 2026-10-26 17:14:43 UTC
- Agent: Codex
- Status: completed

Implemented [planner.Greedy](../planner/greedy.go) with one-step lookahead, batched
greedy completion costs, storage-first ranking, communication tie-breaking,
security/budget checks, and deterministic choices. Added compact full-width row
blocks to [evaluation](../planner/evaluate.go): implicit field-width slices retain
all physical PIR databases and query copies, without allocating millions of
rectangles. Explicit general partitions remain supported. Added shared checked
[logical cost arithmetic](../planner/costs.go), [greedy tests](../planner/greedy_test.go),
and a runnable [example](../planner/example_test.go).

The reproducible [cost-grid report](../experiments/greedy-costs/report.md) links its
[Python script](../experiments/greedy-costs/run.py), [Go driver](../experiments/greedy-costs/main.go),
[input snapshot](../experiments/greedy-costs/pir-configs.json), [complete results](../experiments/greedy-costs/results.json),
and [exact-bit CSV summary](../experiments/greedy-costs/summary.csv). All 24 requested
decimal-unit combinations were calculated with the catalogue's 100-bit, 2^50-sample
profile. All plans are provisional. Large-record/small-row-count cases exhibit
substantial padding and communication; these are logical estimates, not actual
encodings, storage measurements, or cloud transfers. Heuristic optimality is not claimed.

Updated [planner usage](../planner/README.md), [planner design](../docs/parameter-planning.md),
[data representation](../docs/data-representation.md), [catalogue usage](../pirconfig/README.md),
[roadmap](../ROADMAP.md), [experiment index](../experiments/README.md),
[agent map](../AGENTS.md), and [root README](../README.md). README check is complete.
The report records the base revision, exact computational-source/input hashes,
commands, units, environment, methodology, and limitations.

Verification with Go 1.27.3, GOCACHE=/tmp/sk-depir-go-cache:

- Focused planner tests passed: hand-sized improvement/oversizing cases, compact
  versus expanded evaluation, independent exact small dynamic-programming lower
  bounds and largest-fitting baseline, ties, filtering, constraints, and concurrency.
  A 10 MB record case counts ten million databases with one row-block object.
- Full `go test ./... -count=1` passed in isolated copy
  `/tmp/sk-depir-plan005.73VkGI` (eight test packages plus the experiment driver build).
  Legacy CSV tests did not alter this checkout.
- `go test -race ./planner -count=1` and
  `go vet ./planner ./experiments/greedy-costs` passed.
- Only after tests, ran `experiments/greedy-costs/run.py` with the recorded toolchain.
  Repeated the Go driver and reproduced all 24 plans and costs exactly.
- Independent Python integer accounting checked coverage, capacities, groups,
  physical database counts, and storage/upload/download totals for all results.
  Recorded computational-source/input hashes matched current files.
- Formatting and whitespace checks passed. Existing rate experiments were not rerun:
  their source and results were unchanged. No commit/push was requested.

Planning runtime/output grows with selected row-block count; the lookahead costs
O(B*C^2), with final O(B^2) compact validation. General explicit evaluation remains
quadratic in rectangle count. Report generation/documentation took too long after
implementation and verification were complete; no additional feature work is pending.
