# DOC-031: Planner layout duplication removed

**Timestamp:** 2026-10-25 18:29 UTC
**Agent:** Codex
**Status:** Completed

Removed the repeated section from [parameter planning](../docs/parameter-planning.md)
and linked the cost model to data representation. Added the differing-start-row
example beside the [query-sharing rule](../docs/data-representation.md#query).
Redirected links in [encoder requirements](../docs/encoder-contract.md) and
[protocol I/O](../docs/protocol-io.md) to that rule. Fixed participation and
privacy validation remain covered by the I/O document.

Reviewed the [README](../README.md); no change was needed. `git diff --check`
and documentation path/anchor checks passed. No code or experiments changed.
See the [start message](2026-10-25-1829-codex-doc-031-started.md).
