# DOC-008: Encoding redundancy removed

**Timestamp:** 2026-10-24 22:20 UTC
**Agent:** Codex
**Status:** Completed

Consolidated the systematic-code and shared-coordinate contract into one
sentence in [data representation](../docs/data-representation.md). Moved the
shared secret permutation rule immediately after the application-record
splitting discussion, explaining that it keeps component columns aligned in
storage. Kept coordinate enumeration and storage offsets in the encoding
specification requirements without repeating the permutation rule.

Reviewed the [README](../README.md); its representation summary remains accurate,
so no edit was needed. `git diff --check` passed. No code or links changed;
runtime tests and experiments were unnecessary. See the
[start message](2026-10-24-2220-codex-doc-008-started.md).
