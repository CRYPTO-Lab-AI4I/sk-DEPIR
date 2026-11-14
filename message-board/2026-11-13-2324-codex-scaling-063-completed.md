# Server preprocessing target recorded

- Timestamp: 2026-11-13T23:24:21.377737+00:00
- Agent: Codex
- Work: SCALING-063
- Status: completed

Existing experimental s16 preprocessing is hybrid: GPU eligible resident root
products/reciprocal/identity products and spectrum transform; Go host root
arithmetic/small products/overlap-add/persistence; small transforms use native
CPU. Actual run reused saved support, so discovery was not GPU-demonstrated.
[Measured preparation](../experiments/gpu-preparation/report.md#completed-actual-preparation)
2h20m15.469s excludes original support discovery; s17 preparation remains unported.
No new performance measurement or current-job check.

Updated AGENTS.md, root README, docs/encoder-contract.md and scaling plan:
delegated preprocessing targets fastest measured server workflow, with no local
or CPU-only fallback requirement; GPU/multi-GPU may be required by the separate
preprocessing toolchain. Keep useful independent CPU/Go references/caches,
reviewability, complete measured cost and existing budgets. Client CPU encoding
and optional GPU requirement is unchanged. No fixed Go/CPU boundary is imposed.
README checked/updated with target and link to actual hybrid implementation.
`git diff --check` passes; documentation-only clarification needs no tests/jobs.
Earlier staged logging/status records are preserved.
