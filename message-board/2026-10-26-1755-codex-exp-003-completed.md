# EXP-003 — Updated 24-case grid reproduced

- Timestamp: 2026-10-26 17:55:00 UTC
- Agent: Codex
- Status: completed

Added [application-costs/report.md](../experiments/application-costs/report.md), linking
its [run script](../experiments/application-costs/run.py), [Go driver](../experiments/application-costs/main.go),
[fixed catalogue snapshot](../experiments/application-costs/pir-configs.json),
[full results](../experiments/application-costs/results.json), and
[normalized CSV](../experiments/application-costs/summary.csv). Records include exact
bits, selected logical widths, fragment counts, full-piece sizes, capacity evidence,
source hashes, revision and environment, plus no-full alternatives for comparison.
All 24 scenarios were run twice and their plans/evaluations reproduced exactly.
The driver re-evaluates every plan and checks mixed-piece and original-width caps.

The initial run exposed the faulty full-only comparison. The user's correction was
applied in [PLAN-008](2026-10-26-1755-codex-plan-008-completed.md); the report explicitly
records this methodology change and contains the corrected results. The older unsliced
experiment remains separate. With default factor 1, only the 1 GB / 10 MB-record case
selects a full-only plan. Storage-first selection can still incur substantial indexed
communication; this is shown, not hidden, in the table and diagnostic alternatives.

Updated the [experiment index](../experiments/README.md) and root [README](../README.md)
with report links and the finding. Verification: `go test -race ./planner
./experiments/application-costs -count=1`, targeted vet, and whitespace checks passed.
Results are provisional logical calculations using estimated capacities, not measured
storage or network performance. No plaintext database was encoded and no network
service was contacted. Work remains local; no commit/push was requested for this run.
