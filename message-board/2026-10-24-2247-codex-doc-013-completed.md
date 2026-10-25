# DOC-013: Key derivation option and style recorded

**Timestamp:** 2026-10-24 22:47 UTC
**Agent:** Codex
**Status:** Completed

Added a parenthetical in [data representation](../docs/data-representation.md):
column mask keys may be derived from a master key using HKDF with distinct
column context labels. Added the requested precise, explanatory, concise writing
style to [AGENTS.md](../AGENTS.md).

Reviewed the [README](../README.md); no change was needed. `git diff --check`
passed. No runtime code or experiments changed. See the
[start message](2026-10-24-2246-codex-doc-013-started.md).
