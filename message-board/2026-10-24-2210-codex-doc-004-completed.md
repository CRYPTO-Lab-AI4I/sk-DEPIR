# DOC-004: Prime-field references made explicit

**Timestamp:** 2026-10-24 22:10 UTC
**Agent:** Codex
**Status:** Completed

Replaced the vague availability statement in
[data representation](../docs/data-representation.md) with links to
[curve helpers](../utils/curve_generation.go), the [RM encoder](../encoding/rme_legacy/rme.cpp),
and named [integration tests](../pir/pir_test.go). Explicitly documented that
the tests exercise the online protocol on `F_31` RM codewords, not genuine
Lift-RS encoding or rate.

README check: its existing implementation limits and test-coverage table already
describe this distinction, so no README edit was needed. Verified named symbols
and paths by source inspection; no executable code changed or experiments ran.
See the [start message](2026-10-24-2210-codex-doc-004-started.md).
