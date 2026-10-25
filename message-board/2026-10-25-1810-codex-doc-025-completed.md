# DOC-025: Submatrix representation adopted consistently

**Timestamp:** 2026-10-25 18:10 UTC
**Agent:** Codex
**Status:** Completed

Preserved the user's structure in [data representation](../docs/data-representation.md)
and rewrote metadata around original submatrix ranges, padded dimensions,
configurations, and database/query-group identifiers. Made explicit its definition
that one PIR record is one field element: wider application data is partitioned
into additional submatrices, so a separate field-component hierarchy is unnecessary.
Packing is outside scope under the user's record-size assumption.

Aligned [planning](../docs/parameter-planning.md), [encoder requirements](../docs/encoder-contract.md),
and [I/O](../docs/protocol-io.md). Capacity now checks both submatrix rows and bit
width; storage is `sum_d N_d*s_d`, download is
`sum_g L_g*sum_(d in g) s_d`, and responses return one element per database per
position. Query-sharing compatibility accounts for differing submatrix row starts.
Separate keys, the HKDF alternative, version/position mask domains, padding,
and querying all databases remain explicit.

Reviewed and updated [README](../README.md) and [ROADMAP](../ROADMAP.md) descriptions.
`git diff --check` and local path/heading-anchor checks passed. Searched current
docs for stale packing/component-column and block/slice assumptions; remaining
packing references state its exclusion or describe bit serialization. No code,
experiments, or measurements changed; runtime tests were unnecessary.
See the [start message](2026-10-25-1809-codex-doc-025-started.md).
