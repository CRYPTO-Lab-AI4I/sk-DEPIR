# DOC-009: Split-record addressing clarified

**Timestamp:** 2026-10-24 22:24 UTC
**Agent:** Codex
**Status:** Completed

Updated [representation](../docs/data-representation.md) to specify distinct
logical indices and one query per split piece, with the same database permutation.
Component columns instead share one logical index and query. Aligned the
[planning notes](../docs/parameter-planning.md) with this wording.

Reviewed the [README](../README.md); its summary remains accurate without edits.
`git diff --check` passed. No code, links, or experiments changed; runtime tests
were unnecessary. See the [start message](2026-10-24-2223-codex-doc-009-started.md).
