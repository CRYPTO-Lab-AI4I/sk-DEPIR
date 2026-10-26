# PLAN-011 — Batched search and updated grid verified

- Timestamp: 2026-10-26 20:23:23 UTC
- Agent: Codex
- Status: completed

## Search and representation

Replaced row-boundary admission in [SearchTradeoffs](../planner/tradeoffs.go) with
capacity-ordered configuration stages. [Batch helpers](../planner/batches.go) choose
repetition counts and preserve small codes for residuals without exploring permutations
of disjoint full blocks. The final padded block may use any configuration; suffix
full downloads remain ordinary alternatives. Protected baseline/display guarantees
remain intact. Existing single-choice greedy APIs retain their behavior.

Defaults: MaxFrontier=32; MaxStates=256 live partial plans per stage; MaxCounts=8
counts per state/configuration; MaxBaselineBlocks=256. Enumerate all counts when they
fit; otherwise sample zero, one, fill counts and powers of two near both endpoints.
Stage reduction samples remaining-row keys and cost endpoints. StateLimitReached,
CountChoicesTrimmed and StatesExplored replace boundary-admission options/statistics;
frontier and baseline flags remain. These approximations have no global or multiplicative
guarantee. The earlier prototype MaxBoundaries/boundary fields were removed from the
current API and experiment driver, not retained as misleading aliases.

Added RowBlock.Repeat for equal-height consecutive encoded blocks; zero means one.
The [evaluator](../planner/evaluate.go) validates per-block capacity and costs repetitions
in aggregate, without per-block allocation. A shorter padded final block is separate;
full-download blocks cannot repeat. Each physical database retains separate coordinate
identity, masks and full traffic. Updated [expanded fixtures](../planner/greedy_test.go)
and added [batch tests](../planner/batches_test.go). Runtime wrappers and durable stored
manifests are not extended by this planning representation.

## Evidence and documentation

Updated [API guide](../planner/README.md), [planning](../docs/parameter-planning.md),
[representation](../docs/data-representation.md), [roadmap](../ROADMAP.md), and
[example/tests](../planner/example_test.go). Read and updated root [README](../README.md)
for capacity-ordered batching, compact repetition and state/count flags.

Reran the 21 qualifying workloads in the [tradeoff report](../experiments/tradeoff-costs/report.md).
Updated [driver](../experiments/tradeoff-costs/main.go), [script](../experiments/tradeoff-costs/run.py),
[JSON](../experiments/tradeoff-costs/results.json), [CSV](../experiments/tradeoff-costs/summary.csv),
and newly generated [two-option table](../experiments/tradeoff-costs/options.md), including
nearest catalogue capacity and N/K reference overhead. The script now reproduces that
table too. Archived [previous boundary-search results](../experiments/tradeoff-costs/boundary-search-results.json)
retain their source/environment metadata and are checksum-recorded by the new experiment.
The report explicitly identifies this algorithm/measurement-method change.

All 21 original greedy witnesses remain equal-or-better (4 strict, 17 equal). Minimum
communication endpoints did not improve; some intermediate choices changed, including
100 GB / 1 MB moving from 10.394x / 22.929 GB to 10.309x / 21.843 GB. All runs still
encounter stage-state and count approximation; eliminating the boundary cutoff does
not establish optimality or substantially close every catalogue-overhead gap.

## Verification

Passed focused planner tests, race tests, targeted vet, and full `go test ./... -count=1`
in an isolated temporary copy because legacy RM tests write example CSVs. Tiny exhaustive
comparisons pass without truncation. Tests additionally compare compact/expanded repeated
runs, validate billion-block aggregate costs, count sampling, final padding, budgets,
invalid repetitions and deterministic output. Two full experiment runs reproduced all
21 scenarios and generated options exactly; driver/script validate costs, nondominance,
original historical greedy values and display witnesses. `git diff --check` passed.
Compiler: /tmp/sk-depir-go-toolchain/go/bin/go; cache: /tmp/sk-depir-go-cache.
No new encoding backend, security claim, physical performance measurement or server
computation was added. Changes remain local; no commit or push requested.
