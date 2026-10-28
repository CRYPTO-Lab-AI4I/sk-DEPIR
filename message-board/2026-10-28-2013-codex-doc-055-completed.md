# Commodity hardware and local resource constraints recorded

- Timestamp: 2026-10-28 20:13 UTC
- Work item: DOC-055
- Status: completed

Updated [AGENTS.md](../AGENTS.md), [README](../README.md),
[optimization plan](../docs/encoding-algorithms.md#optimization-priority) and
[roadmap](../ROADMAP.md). Open implementation choices are constrained by locally
available benchmarking hardware and agreed budgets, plus commodity CPU/GPU
usability. Larger scales may require larger standard machines; keep hardware/
software requirements as generic as practical. Architecture-specific acceleration
must justify deployment costs and retain a practical generic path. Clearly label
projections for unavailable hardware; no specialized backend is chosen.

README was checked and updated. `git diff --check` and changed documentation
file-target checks pass. Guidance only: no code, caches, profiles or measured
results changed, so tests/benchmarks were not rerun. Previous uncommitted work
and the retained preprocessing cache are preserved. Changes remain uncommitted.
