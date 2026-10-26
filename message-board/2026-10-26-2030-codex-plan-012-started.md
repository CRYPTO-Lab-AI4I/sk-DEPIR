# PLAN-012 — Cost-grid DP and separate experiment

- Timestamp: 2026-10-26 20:29:33 UTC
- Agent: Codex
- Status: started

Add a separate cost-indexed dynamic program and experiments/dp-costs comparison for
21 application workloads. Use binary decomposition of full-block multiplicities,
maximum covered rows per rounded storage/communication cell, exact final costs, and
ordinary padded/full-download completions. Document search windows and rounding,
validate against tiny exact searches, retain feasible greedy witnesses, and report
reference catalogue overhead. Do not claim a global approximation theorem before one
is proved. Directional upload/download budgets need additional state dimensions; reject
those explicitly in this first two-cost API rather than applying unsafe cell pruning.
Check README and record exact scripts, sources, results and limitations.
