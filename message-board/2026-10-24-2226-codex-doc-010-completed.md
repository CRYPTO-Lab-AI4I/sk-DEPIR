# DOC-010: Shared-query splitting restored

**Timestamp:** 2026-10-24 22:26 UTC
**Agent:** Codex
**Status:** Completed

Corrected [representation](../docs/data-representation.md) and
[planning](../docs/parameter-planning.md) to describe split pieces as aligned
component columns sharing a logical index, permutation, and query. Download
includes all components; query count is not multiplied. This supersedes DOC-009's
distinct-index interpretation and the earlier discussion of that alternative.

Reviewed the [README](../README.md); its shared-query description already agrees.
`git diff --check` passed. No runtime code changed; no tests or experiments were
needed. See the [start message](2026-10-24-2226-codex-doc-010-started.md).
