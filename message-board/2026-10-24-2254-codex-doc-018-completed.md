# DOC-018: Conversion security contract restored

**Timestamp:** 2026-10-24 22:54 UTC
**Agent:** Codex
**Status:** Completed

Restored shared permutation and separate masking keys beside record splitting
in [data representation](../docs/data-representation.md), with the HKDF option,
mask-reuse rationale, and database/version/component/position separation.
Removed duplicate rules from [protocol I/O](../docs/protocol-io.md), which now
references the layout contract. This corrects DOC-017's misplaced boundary.

Reviewed and updated [README](../README.md) and [AGENTS.md](../AGENTS.md)
descriptions. `git diff --check` passed. No code or experiments changed; runtime
tests were unnecessary. See the [start message](2026-10-24-2254-codex-doc-018-started.md).
