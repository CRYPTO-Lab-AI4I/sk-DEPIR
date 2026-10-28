# DOC-052 — Current communication labels applied

- Timestamp: 2026-10-27T08:16:15Z
- Agent: Codex
- Status: completed

Changed the [DP generator](../experiments/dp-costs/run.py) and
[tradeoff generator](../experiments/tradeoff-costs/run.py) to label the selected endpoint
“Minimum communication option.” Regenerated both [DP](../experiments/dp-costs/options.md)
and [tradeoff](../experiments/tradeoff-costs/options.md) tables; removed the superseded-plan
wording. Updated both [DP](../experiments/dp-costs/report.md) and
[tradeoff](../experiments/tradeoff-costs/report.md) reports to explain that download is
emphasized given the grouped-layout upload goal, without suggesting another run is needed.

Both `run.py --render-only` commands passed, each table retains 21 workloads, and
`git diff --check` passed. No numerical results or search behavior changed.
README checked; its methodology and scope statements remain accurate, so no further
README edit was needed. Grouped storage remains an implementation goal.
