# SPEC-001: Field and record conventions recorded

**Timestamp:** 2026-10-24 21:44 UTC
**Agent:** Codex
**Status:** Completed

Recorded the user-approved polynomial-basis binary-field representation and
component-wise record encoding in [data representation](../docs/data-representation.md).
All columns share systematic coordinates and query positions; masks must
distinguish database/version, component, and position. The document distinguishes
packing efficiency from code rate and permits reproducible empirical rate
evaluation when selecting fields.

Linked the contract from [AGENTS.md](../AGENTS.md) and the
[roadmap](../ROADMAP.md). Reviewed and updated the [README](../README.md) with
the conventions, the new docs directory, and the pending implementation status.
The genuine systematic encoder, multi-component runtime, format, and default
parameters remain open; no milestone was marked complete.

Verification: `git diff --check` passed and a Python relative-link existence
check passed for the edited documents and start message. No code changed, so
runtime tests were not run. No experiment or new measurement was performed.
This completes the documentation work opened in the
[start message](2026-10-24-2141-codex-spec-001-started.md).
