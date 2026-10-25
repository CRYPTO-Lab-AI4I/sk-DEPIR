# DOC-033: Planner and application encoder boundaries documented

**Timestamp:** 2026-10-25 19:43 UTC
**Agent:** Codex
**Status:** Completed

Added explicit planner and application-encoder boundaries to
[encoder requirements](../docs/encoder-contract.md). The caller chooses a plan;
the application encoder validates and executes it, assigns identities, manages
per-database inputs/keys, and invokes the two PIR encoding stages. Neither PIR
encoder knows the application partition. Failed validation does not silently
trigger replanning.

[Planning](../docs/parameter-planning.md) now returns one or more plans with rate
evidence, using empirical estimates or matching documented encodings. Distinguished
actual code rate from application rate, which also depends on partition/padding
and format. Provisional plans remain exploratory until encoder capacity is verified.
Benchmark evidence is consumed by planning; new measurements are separate work.

Aligned [representation](../docs/data-representation.md) and reviewed/updated the
[README](../README.md). `git diff --check` and documentation path/anchor checks
passed. No runtime code, experiments, or measurements changed. See the
[start message](2026-10-25-1942-codex-doc-033-started.md).
