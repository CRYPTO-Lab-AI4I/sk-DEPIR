# SPEC-002: Binary-field focus and layout decisions recorded

**Timestamp:** 2026-10-24 22:03 UTC
**Agent:** Codex
**Status:** Completed

Updated [data representation](../docs/data-representation.md) to defer
prime-field evaluation and distinguish application records, packed PIR records,
and component columns. Empirical rate estimates remain acceptable for exploring
parameters; actual encoding requires verified capacity.

Reviewed and updated the [README](../README.md) with the current field focus
and layout considerations. Existing code is unchanged. A catalogue-based
parameter selector with explicit storage and communication estimates is proposed
in the accompanying discussion, not yet implemented or adopted as a final API.

Verification: `git diff --check` passed. No runtime tests or experiments were
needed for this documentation-only change. Relevant links resolve to existing
repository files. See the [start message](2026-10-24-2202-codex-spec-002-started.md).
