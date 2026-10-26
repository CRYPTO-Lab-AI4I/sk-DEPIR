# PLAN-012 — Cost-grid DP and analogous table verified

- Timestamp: 2026-10-26 20:41:32 UTC
- Agent: Codex
- Status: completed

## Implementation

Added independent [SearchDPCosts](../planner/dp.go), leaving SearchTradeoffs and greedy
APIs unchanged. Binary full-block groups represent all multiplicities without count or
row-state sampling. DP cells retain maximum row coverage at a fixed item stage and
rounded storage/communication coordinates; costs and reconstruction remain exact bits.
Overshooting groups can terminate with fewer blocks and padding, preserving coverage
dominance. Full residual downloads and padded completions are costed exactly. Repeated
runs are coalesced, and final plans use the existing evaluator and protected witnesses.

Library defaults: 64 cells per axis, 32 exploratory completed points, 256 baseline
blocks. Explicit axes accept 1..512 cells. Default storage window is twice the storage
of the least-communication one-code/greedy seed. Communication windows grow by factor 4
from seed communication through application size; a single database-size rounding scale
was too coarse near low-traffic plans. Explicit window overrides are supported.
Windows, units, binary item counts and transition counters are reported. Per-item
floor rounding has a path accounting bound, not a global Pareto approximation theorem.
One-code and protected seeds are considered outside DP windows. Separate upload/download
budgets are explicitly rejected in this two-cost API; exact storage limits are supported.

## Experiment and results

Added separate [experiments/dp-costs](../experiments/dp-costs/report.md), including the
[script](../experiments/dp-costs/run.py), [driver](../experiments/dp-costs/main.go),
[catalogue snapshot](../experiments/dp-costs/pir-configs.json),
[complete results](../experiments/dp-costs/results.json), [CSV](../experiments/dp-costs/summary.csv),
and generated [analogous options table](../experiments/dp-costs/options.md) with nearest
catalogue overhead. The experiment uses 128 cells per axis; the report distinguishes
this from library defaults and records the exact windows, source hashes, revision,
environment, commands, units and limitations. Earlier experiment results were not replaced.

All 21 workloads have at least 5,000 records and retain equal-or-better original greedy
witnesses (4 strict, 17 equal), including displayed options. Minimum communication
matches the batched solver in 20/21 cases, and is about 1.8% higher for 1 TB / 1 KB
(301.162 MB versus 295.797 MB). Previous solver candidates are not silently backfilled.
Some intermediate tradeoffs differ. These are provisional model calculations, not
new capacity measurements, encoder support, production security or physical I/O claims.

## Documentation and verification

Updated [planner API guide](../planner/README.md), [planning contract](../docs/parameter-planning.md),
[roadmap](../ROADMAP.md), [agent map](../AGENTS.md), [experiment index](../experiments/README.md),
and runnable [example](../planner/example_test.go). Read/updated root [README](../README.md)
for the independent DP and its limits. [DP tests](../planner/dp_test.go) compare exact-unit
grids against independent exhaustive tiny enumeration, verify rounding/determinism,
actual plan evaluation, witness/display guarantees, malformed limits and overflow.

Passed planner tests/race tests, targeted vet, and full go test ./... -count=1 in an
isolated copy because legacy RM tests write CSVs. Two final experiment runs reproduced
all scenarios, comparison counts and generated tables exactly. Driver/script independently
validate costs, nondominance, historical greedy values and protected/display witnesses.
Relative documentation links and git diff --check passed. Compiler:
/tmp/sk-depir-go-toolchain/go/bin/go; cache: /tmp/sk-depir-go-cache.
No global optimum/FPTAS claim is made. Changes remain local; no commit/push requested.
