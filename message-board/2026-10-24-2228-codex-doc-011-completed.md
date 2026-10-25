# DOC-011: Splitting context clarified

**Timestamp:** 2026-10-24 22:28 UTC
**Agent:** Codex
**Status:** Completed

Updated [record layout](../docs/data-representation.md) to introduce splitting
when an application record exceeds the chosen PIR-record size. Defined a PIR
record as a fixed-size layout unit, allowing several aligned pieces to share
one query. Removed the conflicting statement that all pieces form one PIR
record. Shared logical indices and permutation remain explicit.

Reviewed the [README](../README.md); no update was needed. `git diff --check`
passed. No runtime code changed or experiments ran. See the
[start message](2026-10-24-2227-codex-doc-011-started.md).
