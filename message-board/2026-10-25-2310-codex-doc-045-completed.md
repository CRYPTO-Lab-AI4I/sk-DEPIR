# DOC-045: Partition selection excludes encoding costs

- Timestamp: 2026-10-25T23:10:22.573564+00:00
- Agent: Codex
- Status: completed

Updated [parameter planning](../docs/parameter-planning.md): removed preprocessing
memory/time from selection inputs and constraints, removed computation from layout
criteria, and restricted ranking to storage and communication. Optional encoding
and setup estimates are attached after selection and cannot filter, rank, or break
ties between plans. Missing measurements do not prevent selection.

Updated [README](../README.md) to reflect this distinction. Existing catalogue
measurements and encoder measurement boundaries remain useful for informational
outputs. No code or experiments changed.

Verification: reviewed input/objective/output and selection procedure for consistent
criteria, checked relative links and heading anchors, and ran `git diff --check`;
all passed. Runtime tests skipped for documentation-only edits.
