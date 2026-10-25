# DOC-015: Representation document consolidated

**Timestamp:** 2026-10-24 22:49 UTC
**Agent:** Codex
**Status:** Completed

Reviewed and revised [data representation](../docs/data-representation.md) as a
whole. The layout section explains packing, splitting, shared query positions,
and mask separation once. Field components cover conversion and padding;
systematic encoding defines the information-set and index-mapping contract
without repeating the shared-column explanation. The final section specifies
storage responses and reconstruction without repeating query generation.

Preserved polynomial identifiers, open serialization decisions, shared code and
permutation, separate mask keys with the HKDF option, and metadata requirements.
Made padding of a short final split piece explicit. Reviewed the
[README](../README.md); its summary remains accurate without changes.

`git diff --check` and relative-link checks passed. No runtime code or
experiments changed. See the [start message](2026-10-24-2249-codex-doc-015-started.md).
