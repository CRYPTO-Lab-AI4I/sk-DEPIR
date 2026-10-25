# DOC-027: Degrees 8 through 20 recorded as catalogue scope

**Timestamp:** 2026-10-25 18:23 UTC
**Agent:** Codex
**Status:** Completed

Updated [planning](../docs/parameter-planning.md),
[representation](../docs/data-representation.md), and
[encoder requirements](../docs/encoder-contract.md) to include every integer
field degree from 8 through 20. Removed the byte-aligned shortlist and documented
terabyte-scale application databases as a sizing target subject to actual
storage, preprocessing, communication, and security constraints.

Reviewed and updated the [README](../README.md), explicitly listing missing
degrees 9, 10, 11, 15, and 19 and preserving the distinction between target and
implemented coverage. `git diff --check` passed; checked for the superseded
shortlist. No arithmetic code, tests, or experiments changed. See the
[start message](2026-10-25-1822-codex-doc-027-started.md).
