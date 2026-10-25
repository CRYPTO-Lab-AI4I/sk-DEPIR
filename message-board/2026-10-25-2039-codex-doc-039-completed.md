# DOC-039: Fixed query-sharing conditions documented

- Timestamp: 2026-10-25T20:39:40.575547+00:00
- Agent: Codex
- Status: completed

[data representation](../docs/data-representation.md#query-sharing-groups) now
lists four conditions: common encoding/query configuration, compatible original
row ranges, fixed group membership and participation, and separate masks.
Explained the local PIR index and the pairwise row-range criterion: ranges are
disjoint or share their starting row, with ending rows allowed to differ. Added
examples and explicitly retained general partitions across groups. Query describes
normal and irrelevant-data handling without suggesting request-dependent grouping.

Updated [planner checks](../docs/parameter-planning.md),
[protocol participation](../docs/protocol-io.md), and
[encoder cross-reference](../docs/encoder-contract.md). Reviewed README's existing
query-group and representation descriptions; no change required.

Verification: checked the range criterion against `j-a` (equal indices on an
overlap exactly when starts agree), reviewed examples and all-request wording,
checked docs relative links and Markdown anchors, and ran `git diff --check`.
No runtime tests or experiments: documentation only. This specifies compatibility
and participation requirements, not a new security proof.
