# DOC-040: Grouping eligibility and operation separated

- Timestamp: 2026-10-25T20:54:46.834032+00:00
- Agent: Codex
- Status: completed

[data representation](../docs/data-representation.md) now lists only common
encoding/query configuration and compatible row ranges as grouping eligibility.
The shared permutation is assigned after choosing groups. Fixed participation
and message sizes are specified in Query; separate masking keys are explained
alongside mask identities and versions. All requirements are preserved.

Reviewed planner, encoder, and protocol references: they remain consistent with
the distinction and require no edits. README checked; no update needed.
Verified all docs relative file links and heading anchors; `git diff --check`
passed. No runtime tests or experiments for this documentation-only clarification.
