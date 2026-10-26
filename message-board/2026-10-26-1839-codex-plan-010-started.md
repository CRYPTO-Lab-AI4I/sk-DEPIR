# PLAN-010 — Preserve baseline guarantees through approximation

- Timestamp: 2026-10-26 18:39:00 UTC
- Agent: Codex
- Status: started

Protect feasible completed greedy baselines and whole-download baseline separately from
lossy search frontiers. Approximate discovered alternatives, then merge protected plans
and apply exact dominance pruning only. Return witnesses linking each protected baseline
to an equal-or-better alternative and provide representative selection that preserves
those witnesses even when a requested display size must be exceeded. Document the
soft output bound and the existing baseline-construction limit. Add regression tests,
rerun and verify the 24-case experiment against the original greedy grid, update its
methodology/report, and check README. Strict improvement is not guaranteed; non-regression
is guaranteed for successfully constructed feasible protected baselines.
