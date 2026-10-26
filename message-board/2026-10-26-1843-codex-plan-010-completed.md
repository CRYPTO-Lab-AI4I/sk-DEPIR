# PLAN-010 — Protected baseline guarantees verified

- Timestamp: 2026-10-26 18:43:22 UTC
- Agent: Codex
- Status: completed

## Changes

[SearchTradeoffs](../planner/tradeoffs.go) now keeps successfully constructed feasible
greedy baselines and whole-download baseline separate from lossy exploration. Reduce
exploratory results to MaxFrontier, merge protected plans, then apply exact dominance
pruning only. The final set can exceed MaxFrontier. BaselineWitnesses records each
baseline's costs, logical width, mode and the index of an equal-or-better returned
alternative. This guarantees non-regression in storage and total communication for
those baselines, not strict improvement. Baselines stopped by the existing construction
limit are explicitly outside the guarantee; BaselineLimitReached remains visible.

Added RepresentativeIndices for display: retain every baseline witness and endpoints,
then sample extra options. Its size is a soft target and can be exceeded. Arbitrary
sampling was a separate source of lost guarantees in the earlier chat table.
Updated [tests](../planner/tradeoffs_test.go), [example](../planner/example_test.go),
[API documentation](../planner/README.md), and [planning contract](../docs/parameter-planning.md).
Read and updated root [README](../README.md) with the guarantee and output-size semantics.

## Experiment and verification

Updated the existing [tradeoff report](../experiments/tradeoff-costs/report.md),
[driver](../experiments/tradeoff-costs/main.go), [script](../experiments/tradeoff-costs/run.py),
[results](../experiments/tradeoff-costs/results.json), and
[CSV](../experiments/tradeoff-costs/summary.csv) together. The report explicitly records
the correction to final pruning and the regenerated results. The script checks today's
original-width greedy costs against the exact historical grid and verifies a displayed
witness against those historical values, rather than relying solely on internal metadata.
The baseline grid is included in recorded checksums. The driver independently validates
every protected baseline witness, display inclusion, costs, and nondominance.

All **24/24** original greedy cases have an equal-or-better witness in both results and
representative selections: **6 strictly improve** at least one cost, **18 match both**.
All results and selections reproduced exactly across two runs. All cases still reach
the boundary exploration limit, so this is not a global optimality claim. New rate
estimates or physical encoding/network measurements were not introduced.

Using `/tmp/sk-depir-go-toolchain/go/bin/go` and `/tmp/sk-depir-go-cache`:
`go test ./planner -count=1`, `go test -race ./planner -count=1`, targeted vet, and
`go test ./... -count=1` passed. Full suite used an isolated copy because legacy tests
write CSV fixtures. Regression tests exercise the old 1 GB / 1 KB failure, output
beyond the exploratory cap, display witnesses, malformed indices, and exact tiny
comparisons. `git diff --check` passed. No commit or push requested; changes remain local.
