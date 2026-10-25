# DOC-019: Documentation reconciled around one data model

**Timestamp:** 2026-10-24 22:58 UTC
**Agent:** Codex
**Status:** Completed

Rewrote [representation](../docs/data-representation.md) to define PIR database
columns and field-component columns consistently, cover packed/split/component
padding, and list required layout metadata. Shared permutation, separate mask
keys, the HKDF alternative, and mask domains remain beside record splitting.
Serialization choices are explicitly unresolved.

Rewrote [planning](../docs/parameter-planning.md) around inputs, evidence, a
symbol table, layout costs, and selection output. The general width
`W = sum_c s*ceil(B_c/s)` counts per-column padding; packing and splitting are
separate special cases, and the minimum-packing shortcut is scoped accordingly.
Removed conversation history and duplicate policy statements.

Expanded [protocol I/O](../docs/protocol-io.md) into intended logical requirements:
ordered query entries, response indexing, duplicate-position correspondence,
structural validation, and reconstruction. Wire format and error encoding remain
open; this does not claim implementation or response authentication.

Preserved the pending systematic encoder contract in
[encoder requirements](../docs/encoder-contract.md), moved development order to
the [roadmap](../ROADMAP.md#implementation-sequence), and updated the
[README](../README.md) and [agent instructions](../AGENTS.md) links/descriptions.
No milestone was marked complete.

Verification: `git diff --check` passed; relative paths and heading anchors in
all seven affected documents resolved. Reviewed formulas for both layout cases
and checked that security and representation decisions were retained. No code,
new measurements, or experiments were added; runtime tests were unnecessary.
See the [start message](2026-10-24-2256-codex-doc-019-started.md).
