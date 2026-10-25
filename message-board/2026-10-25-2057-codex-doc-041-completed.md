# DOC-041: Query-sharing terminology made consistent

- Timestamp: 2026-10-25T20:57:20.446783+00:00
- Agent: Codex
- Status: completed

Replaced “fixed group” with “query-sharing group” in
[data representation](../docs/data-representation.md), and replaced “fixed grouping”
in [parameter planning](../docs/parameter-planning.md). The separate requirement
that membership stays fixed remains unchanged.

Verification: searched all docs; no “fixed group” occurrences remain.
`git diff --check` passed. README reviewed and already uses query-sharing groups;
no edit needed. No code, links, experiments, or protocol behavior changed;
runtime tests were unnecessary.
