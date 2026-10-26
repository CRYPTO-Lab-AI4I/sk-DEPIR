# PLAN-009 — Bounded tradeoff search implemented and verified

- Timestamp: 2026-10-26 18:26:24 UTC
- Agent: Codex
- Status: completed

## Implementation

Added [SearchTradeoffs](../planner/tradeoffs.go), returning validated nondominated
application plans rather than a single storage-first winner. It explores the existing
logical-width family, capacity-sized full-width encoded row blocks (a padded final
block), and full-download suffixes. Whole-application download is an ordinary candidate;
this API does not use the earlier full-download ratio exceptions. Existing greedy
APIs retain their behavior.

Partial pruning compares plans only at the same boundary. Disjoint row blocks and
traffic-independent grouping make future transitions identical there. Use storage and
total traffic normally, or storage/upload/download when directional budgets require
those dimensions. Final alternatives are nondominated by storage and total traffic.

Defaults: 32 retained frontier points per boundary and returned alternatives; 256
admitted boundaries per width. Overflowing frontiers preserve storage/communication
endpoints and geometric communication-bucket representatives. Boundary admission is
separately bounded, since a frontier bound alone does not bound distinct row positions.
No multiplicative approximation guarantee is claimed. Frontier, boundary and baseline
truncation have separate flags, plus processed-boundary counts. Greedy baselines are
seeded before exploration; their construction also has a 256-block default limit to
avoid unbounded work. The [greedy helper](../planner/greedy.go) now supports that private
limit; existing callers remain unlimited. ErrNoPlan wording now covers both searches.

Added `Evaluation.CommunicationRate` in [evaluation](../planner/evaluate.go) and
[application evaluation](../planner/application.go): **original record bits / (upload +
download per application lookup)**. Higher is better; full download of n records gives
1/n. This supersedes the inverse definition discussed earlier.

## Documentation and verification

Updated [API guide](../planner/README.md), [examples](../planner/example_test.go),
[planning](../docs/parameter-planning.md), [representation](../docs/data-representation.md),
[catalogue guide](../pirconfig/README.md), [roadmap](../ROADMAP.md), and
[agent map](../AGENTS.md). Read/updated root [README](../README.md) with the API,
rate convention and evidence links. Encoding/I/O responsibilities are unchanged.

[Tests](../planner/tradeoffs_test.go) compare unrestricted tiny searches with independent
exhaustive complete-plan enumeration, including budgets. They check approximation
endpoints, work/output bounds, deterministic catalogue ordering, whole-download
comparison, overflow, rate convention, and terabyte-scale coordinates. A tiny-code
billion-row input verifies bounded baseline construction rather than allocating huge
block lists. No production encoding or security claim is made.

Commands passed using `/tmp/sk-depir-go-toolchain/go/bin/go` and
`GOCACHE=/tmp/sk-depir-go-cache`: `go test ./planner -count=1`,
`go test -race ./planner -count=1`, targeted `go vet`, and `go test ./... -count=1`
in an isolated temporary copy because legacy tests write CSV fixtures. Changed
Markdown relative file links resolve; `git diff --check` passed.

The [new tradeoff report](../experiments/tradeoff-costs/report.md) records all 24 cases.
All hit the boundary cap, so the explored alternatives are approximate. Search can
miss superior or feasible layouts; flags must accompany results. No automatic balanced
recommendation, arbitrary rearrangement, durable manifest or new encoding backend was
added. Changes remain local; no publication was requested.
