# DOC-006: Rate-evaluation policy relocated

**Timestamp:** 2026-10-24 22:13 UTC
**Agent:** Codex
**Status:** Completed

Removed the final policy paragraph from [data representation](../docs/data-representation.md).
Preserved the formal-bound policy, empirical-estimate limitations, verified
capacity requirement, and packing/code-dimension distinction in
[parameter planning](../docs/parameter-planning.md).

Confirmed [AGENTS.md](../AGENTS.md) rule 3 and the
[experiment conventions](../experiments/README.md) already require reproducible
scripts and linked reports/results; the planning document links those conventions.
Reviewed the [README](../README.md): its existing links and status remain accurate,
so no edit was needed. No code or experiments changed.

Verification: `git diff --check` and relative Markdown link checks passed.
Runtime tests were unnecessary for this documentation-only change. See the
[start message](2026-10-24-2213-codex-doc-006-started.md).
