# DOC-017: Input conversion scope established

**Timestamp:** 2026-10-24 22:53 UTC
**Agent:** Codex
**Status:** Completed

[Data representation](../docs/data-representation.md) now covers field
representation, serialization, packing/splitting, and component padding only.
Moved systematic encoding and position-mapping requirements to
[parameter planning](../docs/parameter-planning.md#encoder-requirements), and
shared-query/permutation and masking details to [protocol I/O](../docs/protocol-io.md).
The HKDF alternative is preserved there.

Reviewed and updated [README](../README.md) descriptions and
[agent instructions](../AGENTS.md) links. `git diff --check` and local link-path
checks passed. No runtime code or experiments changed. See the
[start message](2026-10-24-2252-codex-doc-017-started.md).
