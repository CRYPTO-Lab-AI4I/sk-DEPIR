# DOC-036: Context and terminology pass completed

- Timestamp: 2026-10-25T20:23:04.604461+00:00
- Agent: Codex
- Status: completed

In [data representation](../docs/data-representation.md), explain partitioning
before introducing a linked “a planner” and application encoder. Give indexing
its own section; retain exact coverage in partitioning. Introduce query-sharing
groups, systematic encoding, HKDF, and PRF where first used.

Reviewed all four docs. In [encoder requirements](../docs/encoder-contract.md),
introduce the overall workflow, link preprocessing at its first use, and explain
the information set and code dimension notation. In
[parameter planning](../docs/parameter-planning.md), explain query groups, rate,
dummy positions, samples, and field notation; use `S_max` for the sample bound
to distinguish it from stored bits `S`. In [protocol I/O](../docs/protocol-io.md),
introduce group indices, dummy positions, and the role of private query state.
These are explanatory edits; no contracts or security parameters changed.

Verification: read every `docs/*.md` file; checked all relative file links and
Markdown heading anchors; `git diff --check` passed. Reviewed README's specification
and architecture descriptions; no update needed. No runtime tests or experiments
were run for this documentation-only change.
