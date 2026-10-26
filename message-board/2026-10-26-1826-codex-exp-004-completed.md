# EXP-004 — Tradeoffs for all 24 specifications reproduced

- Timestamp: 2026-10-26 18:26:24 UTC
- Agent: Codex
- Status: completed

Added a separate [report](../experiments/tradeoff-costs/report.md) linking its
[script](../experiments/tradeoff-costs/run.py), [driver](../experiments/tradeoff-costs/main.go),
[fixed input](../experiments/tradeoff-costs/pir-configs.json),
[complete results](../experiments/tradeoff-costs/results.json), and
[per-alternative CSV](../experiments/tradeoff-costs/summary.csv). The sources, revision,
environment, explicit limits, units, rate convention, and limitations are recorded.
These are model calculations, not physical encoding or network measurements.

All 24 cases ran twice and reproduced alternatives, exact costs, and truncation flags.
The driver independently re-evaluates every plan and rejects duplicate/dominated
alternatives, including encoded plans worse than whole-application download. Default
limits are 32 frontier points and 256 boundaries per width. All 24 hit the boundary
cap; none hit the baseline block cap. Results are approximate explored tradeoffs.
For the 1 GB cases with 1 MB or 10 MB records, only the full-download plan survives;
this is not a proof that the unrestricted construction cannot do better.

Checked and updated the root [README](../README.md) and
[experiment index](../experiments/README.md). Historical single-choice grids were not
rerun or replaced. Full repository tests passed in an isolated copy, targeted race
and vet passed, and whitespace/link checks passed. Future comparisons must preserve
the search limits and rate convention or explicitly report changes. No production
security or large-instance encoding support is inferred. New work remains local.
