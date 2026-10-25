# DOC-046: Query reuse is client computation only

- Timestamp: 2026-10-25T23:27:40.938222+00:00
- Agent: Codex
- Status: completed

Updated [representation](../docs/data-representation.md) and
[protocol I/O](../docs/protocol-io.md) so the client computes one query per
query-sharing group but sends each PIR database its own copy and requests every
position occurrence. Storage handles ordinary reads without group knowledge.
Grouped responses are client-side bookkeeping.

Corrected [upload accounting](../docs/parameter-planning.md) to
`sum_g J_g*L_g*ceil(log2(N_g))`, where J_g is group size. The single-configuration
example now includes J in upload. Download/storage unchanged. Explicitly distinguish
logical counts from remote-storage transport overhead. Updated [README](../README.md).
This supersedes earlier once-per-group upload assumptions; sharing reduces client
query generation, not transmitted address counts.

Verification: reviewed single/group cases (J=1 restores the single-database formula),
searched for stale once-per-group/upload-saving claims, checked docs/README links
and heading anchors, and ran `git diff --check`; all passed. No implementation or
experiment changes; runtime tests unnecessary for this specification change.
