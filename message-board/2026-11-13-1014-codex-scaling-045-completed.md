# Encoding scaling reference established

- Timestamp: 2026-11-13T10:14:33.828331+00:00
- Agent: Codex
- Work: SCALING-045
- Status: completed

Added [the maintained encoding reference](../docs/encoding-status.md), with preprocessing and CPU/GPU mode tables, measured/unrun status, memory/time boundaries, capacity and multi-GPU floors. The performance objective is reasonable encoding through s18 inclusive; s19–20 remain exploratory. Current execution is still s≤16. The s20 host-backed projection includes the expanded prefix arena (~30804 GiB host arrays); counts are capacity floors, not implementations or job admission.

Changed [README](../README.md) for a prominent entry point, [the scaling plan](../docs/encoding-scaling.md) for current objective/status links and multi-GPU scope, and [AGENTS.md](../AGENTS.md) to require ongoing table upkeep. Historical reports remain intact. Numerical latency acceptance remains to be agreed; no resource budgets or TRL5 criteria changed.

Verification: Python checked table structure, s17–20 presence, local link targets and GPU capacity arithmetic; manually checked array projections, source reports and measurement boundaries. git diff --check passes. README checked and updated. Documentation only; no runtime tests or jobs needed/submitted. All source links point to existing experiment reports containing reproduction scripts/results.
