# Matched subset-validator job77213 submitted

- Timestamp: 2026-11-18T13:36:12.116277+00:00
- Agent: Codex
- Work: SCALING-111
- Status: Handoff

Prepared source e061a08 committed. Submitted **77213**, existing1node/4H200/4CPU/
8GiBhost/6GiBRSS/4GiBGo/4GiBowned/2GiBcache/4GiBscratch/20min scheduler budget.
[Receipt](../experiments/root-subset-validation/submission.json),
[report](../experiments/root-subset-validation/report.md),
[results](../experiments/root-subset-validation/runs/compare-01/),
[log](../experiments/root-subset-validation/slurm-77213.log).
Bitmap/search/search/bitmap order; shared CPU reference validation sorted in both
modes, explicit test flags, warm+2timed samples/config/process, full CPU/scalar/
file-hash equivalence, per-process profiles and both small CUDA memchecks.
No worker/default/fullsupport mode adoption. GPU timings and RSS remain pending.

Actual main reference and active plan updated with pending ID. README/code docs
and local checks recorded in prepared message; no further README change needed
for receipt. Whitespace check passes. No scheduler/process polling, log following,
waiting or results inspection after submission. Await completion ping before
review or dependent work. No resource increase, shared-filesystem intermediate
workspace or new fullsize claim.
