# DOC-007: Representation document reorganized

**Timestamp:** 2026-10-24 22:15 UTC
**Agent:** Codex
**Status:** Completed

Reorganized [data representation](../docs/data-representation.md) into binary
fields, record layout, systematic encoding and positions, retrieval, and mask
domains. Field scope now opens the binary-fields section; representation and
serialization have separate subsections. Defined application records, PIR
records, and component columns before describing encoding.

Moved field candidates and the full-occupancy effective-rate formula into
[parameter planning](../docs/parameter-planning.md). Existing contracts are
preserved, including shared systematic coordinates, padding, ordering, mask
separation, and open serialization decisions. No runtime behavior changed.

Reviewed the [README](../README.md): its descriptions and links remain accurate,
so no edit was necessary. `git diff --check` and relative Markdown link checks
passed. No runtime tests or experiments were needed. See the
[start message](2026-10-24-2215-codex-doc-007-started.md).
