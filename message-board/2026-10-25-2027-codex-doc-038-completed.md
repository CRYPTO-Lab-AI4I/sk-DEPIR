# DOC-038: Query sharing introduced with context

- Timestamp: 2026-10-25T20:27:50.902980+00:00
- Agent: Codex
- Status: completed

Updated [data representation](../docs/data-representation.md) to explain that a
PIR query lists encoded positions and that sharing sends that list once for reads
from each PIR database, reducing query generation and upload costs. Moved the
configuration compatibility paragraph into Encoding, before naming the group and
its shared permutation. Retained separate masks and the linked per-request index
condition; removed the repeated compatibility paragraph from Query. Clarified
PIR database references and nearby line wrapping.

Verification: read the revised explanation in context, checked relative links and
heading anchors, and ran `git diff --check` successfully. Reviewed README's
representation and query-group descriptions; no change needed. No runtime tests
or experiments were needed for this wording-only edit.
