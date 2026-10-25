# DOC-021: Block-and-slice framework integrated

**Timestamp:** 2026-10-25 16:03 UTC
**Agent:** Codex
**Status:** Completed

Integrated consecutive row blocks of catalogue-supported capacities, unequal
column slices, optional packing, and zero-row padding into
[data representation](../docs/data-representation.md). Mask keys now explicitly
distinguish every resulting database, including its block and slice; shared
permutations are scoped to compatible query groups.

Updated [planning](../docs/parameter-planning.md) with per-database parameters
and the collection costs `S = sum N_d*w_d`,
`D = sum_g L_g*sum_(d in g) w_d`, and
`U = sum_g L_g*ceil(log2(N_g))`. The packing shortcut is now explicitly local:
grouping choices can change total upload. All costs include irrelevant-block
responses and separate padding, metadata, and transport overheads appropriately.

Aligned [encoder requirements](../docs/encoder-contract.md) and
[protocol I/O](../docs/protocol-io.md) with configurations per database, compatible
shared queries, querying every group, and reconstruction from the target block.
Cover-query generation and joint/repeated-query privacy remain explicit
validation requirements, not security claims established by these edits.

Reviewed and updated the [README](../README.md). `git diff --check` and local
path/heading-anchor checks passed. Checked for stale single-configuration
assumptions and reviewed the general formulas against their single-group special
case. No executable code, experiments, or measurements were added; runtime tests
were unnecessary. See the [start message](2026-10-25-1601-codex-doc-021-started.md).
