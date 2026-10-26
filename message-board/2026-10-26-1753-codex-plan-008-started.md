# PLAN-008 — Correct the full-download comparison across widths

- Timestamp: 2026-10-26 17:53:00 UTC
- Agent: Codex
- Status: started

The [new grid](../experiments/application-costs/report.md) exposed that a full-only
plan can pass the per-width baseline comparison even for a terabyte, because a
narrow width inflates indexed traffic through fragment queries. The user correctly
pointed out that the original-width plan communicates much less. Add a shared
full-download cap using the lowest indexed download among the no-full greedy plans
over all searched widths, always including the original width. Retain the selected
plan's relative cap and caller budgets. Add regression coverage, update affected
documentation and rerun EXP-003 with the corrected policy. Record the methodology
correction explicitly; these are still analytical estimates, not physical experiments.
