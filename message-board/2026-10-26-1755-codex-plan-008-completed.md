# PLAN-008 — Shared download reference corrected

- Timestamp: 2026-10-26 17:55:00 UTC
- Agent: Codex
- Status: completed

Corrected [GreedyApplication](../planner/application.go): before enabling full pieces,
compute no-full greedy plans for every searched width, always including the original
width. Use the lowest indexed download as a shared reference, with security requirements
but without cost budgets. Full pieces must satisfy this factor-scaled cap in addition
to their selected encoded-part cap, optional absolute cap, and ordinary budgets.
A width can no longer inflate its own baseline to qualify a large full download.
The factor remains configurable, default 1; a pure full download remains possible when
actually comparable with all these references.

Added a [regression test](../planner/application_test.go) comparing alternative widths
against original-width download, for multiple factors and storage-budget settings.
Updated [planner documentation](../planner/README.md), [planning contract](../docs/parameter-planning.md),
and root [README](../README.md). The README check also links corrected experiment results.
The prior PLAN-007 message records the superseded policy; this message corrects it.

Verification: planner tests and race tests passed; `go vet ./planner
./experiments/application-costs` and `git diff --check` passed. The first regression
fixture was unnecessarily large and was reduced while preserving the failure case;
no runtime optimization was introduced. New [24-case results](../experiments/application-costs/report.md)
reproduce exactly and reject full-only plans for every 1 TB scenario. Selection still
prioritizes storage: indexed fragment traffic can exceed the original database size.
No new execution backend or global optimality guarantee is claimed. Changes remain local.
