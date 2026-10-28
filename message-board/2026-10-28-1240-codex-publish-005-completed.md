# Polynomial-product checkpoint ready for publication

- Timestamp: 2026-10-28 12:40 UTC
- Work item: PUBLISH-005
- Status: completed

Reviewed [ENCODE-004 verification](2026-10-28-1204-codex-encode-004-completed.md),
[experiment report](../experiments/fast-polynomial/report.md), final/initial
results and exact source/input/kernel hashes. Recorded the user's acceptance of
the one-time s=8 preprocessing cost in the report; no measurement was changed.
README was checked and already describes the implemented bounded backend.
Prior full-suite, race and vet checks passed; no implementation changed since
those checks, so they were not repeated. `git diff --check` passes. Commit and
push all ENCODE-004 implementation, tests, provenance and documentation.

Next is the compact reusable-artifact checkpoint: distinguish small required
code descriptions from optional acceleration caches, introduce a documented
binary representation for large arrays, validate round trips/compatibility and
measure artifact size/load cost. Then advance CPU encoding to s=10 before s=11/12.
Keep codeword outputs and the paper information set unchanged, and preserve the
separation of code-only preparation from per-database work.
