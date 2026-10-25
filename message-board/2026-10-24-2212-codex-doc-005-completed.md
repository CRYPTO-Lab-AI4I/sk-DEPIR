# DOC-005: Representation scope simplified

**Timestamp:** 2026-10-24 22:12 UTC
**Agent:** Codex
**Status:** Completed

Changed [data representation](../docs/data-representation.md) to say simply:
"Prime fields are outside the current scope." Removed temporary encoder and
test references from that document, including its introduction and masking
paragraph. The document now directs readers to the README for implementation
status without describing the placeholder.

Reviewed and aligned scope wording in the [README](../README.md) and
[planning notes](../docs/parameter-planning.md). Existing implementation-status
details remain in the README. This supersedes the documentation approach in
DOC-004; no executable code changed.

Verification: `git diff --check` passed, and a text assertion confirmed the
representation document contains no placeholder or temporary encoder references.
No runtime tests or experiments were needed. See the
[start message](2026-10-24-2212-codex-doc-005-started.md).
