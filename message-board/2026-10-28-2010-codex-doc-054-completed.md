# Encoding optimization choices explicitly remain open

- Timestamp: 2026-10-28 20:10 UTC
- Work item: DOC-054
- Status: completed

Updated [AGENTS.md](../AGENTS.md), [README](../README.md),
[roadmap](../ROADMAP.md) and the
[optimization priority](../docs/encoding-algorithms.md#optimization-priority):
fast encoding is the objective; language/compiler, algorithms, CPU/GPU backend,
layout, memory strategy and parallelization are not settled. Evaluate alternatives
using reproducible kernel and full-stage measurements while preserving correctness,
PIR functionality/security requirements and standard-setup usability.

Removed blanket optimization deferrals and step-order restrictions that conflicted
with the user's instruction. TRL 6 deployment acceptance remains separate; neither
a hardware investigation nor an implementation change claims that maturity level.
The concrete next comparison is equivalent batched Go/C++ kernels before substantial
Go-specific tuning, with Rust and other implementations still eligible. No backend
is chosen and no encoder behavior, cache data or results changed.

README was checked and updated. Documentation file targets and `git diff --check`
pass. No tests or benchmarks were rerun for this guidance-only change. Earlier
uncommitted profile results and retained local cache are preserved. Changes remain
uncommitted; the user did not request publication for this item.
