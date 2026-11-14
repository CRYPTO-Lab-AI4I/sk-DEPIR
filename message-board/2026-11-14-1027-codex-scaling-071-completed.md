# Preprocessing GPU count is a performance choice

- Timestamp: 2026-11-14T10:27:39.756284+00:00
- Agent: Codex
- Work: SCALING-071
- Status: completed

Recorded the user's standing instruction in [AGENTS.md](../AGENTS.md): no fixed
two-GPU preprocessing ceiling; prioritize complete server latency, additional
cards and multiple nodes when useful, with explicit numeric budgets retained.
CPU client/reference roles stay distinct. [Plan](../docs/encoding-scaling.md),
[preprocessing guide](../docs/preprocessing.md),
[reference](../docs/encoding-status.md), README and
[experiment report](../experiments/gpu-s17-preprocessing/report.md) now distinguish
initial two-card validation, capacity floors and fastest measured device count.

Prepared the existing runner for1/2/4 available same-node GPUs. The tile test
compares every power-of-two prefix within one allocation (1/2/4 for four cards),
with the same dense input and hashes. Analyzer handles those sets and checks
window allocations against each run's actual per-card cap. Existing1/2-card
output formats remain compatible. The aggregate512MiB small-test cap is unchanged;
four cards mean128MiB/card, and the wide-window payload fits. No production
kernel, code representation, method timing or memory strategy changes.

[Checks](../experiments/gpu-s17-preprocessing/gpu-count-checks.json): runner
--check --devices4, deadline/window-cap assertions,1/2/4 comparison-set checks,
Python AST, Go formatting and diff checks pass. Actual four-card topology/runtime
remains unvalidated; no new performance claim or source capability beyond prior
prototype. No long test/new job was launched;74453/74454 and other pending results
remain unmonitored/uninspected. Existing budgets are not silently enlarged.
README checked and updated. Next review completed validation on user ping, then
benchmark larger representative GPU work and extend beyond one node if useful;
do not treat either two or four GPUs as the overall performance ceiling.
