# Published baseline and checked disk job

- Timestamp: 2026-11-13T23:15:18.893151+00:00
- Agent: Codex
- Work: SCALING-060
- Status: completed

Committed/pushed verified job74412 evidence and current documentation as44c8a9c
before checking74253, as requested. `git diff --cached --check` reports existing
raw Slurm job.log trailing whitespace; preserved verbatim to retain sealed hashes.
No source/document whitespace changes were needed.

Loaded slurm/slurm/25.05 and made one user-authorized squeue/sacct/scontrol
check of74253: RUNNING ongnode17, elapsed7:19:05,48h limit,32CPU/128GiB RAM
request, zero restarts; no final outcome. Saved the
[snapshot](../experiments/encoding-table/2026-11-13-2315-job-74253.json) and linked it in the
[report](../experiments/encoding-table/report.md). No live logs/processes or
partial results inspected. Other unreviewed jobs/artifacts unchanged.

s17 implementation: shared32/64-bit arithmetic/transforms; checked wide native
interfaces/matrix scheduling; width-aware compatible code data with caps/hash
validation; complete wide CPU encoding schedule passes bounded small real-code,
synthetic boundary, ownership/race/sanitizer/regression checks. Still missing
shared higher-level sequence (dedup next), genuine verified s17 preprocessing,
representative wider calibration and full CPU run/admitted budgets; wider GPU/
multi-GPU execution remains unimplemented. The427GiB CPU principal-array model
is an estimate, not measured peak/latency/feasibility.

README checked: current bounded-gate/support descriptions match; no edit needed
for a scheduler snapshot. No new tests/jobs or refactor needed for this publication
and read-only status request. No continued monitoring; await completion ping.
