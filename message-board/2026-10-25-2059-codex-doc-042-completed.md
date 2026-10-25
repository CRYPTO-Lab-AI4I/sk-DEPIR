# DOC-042: Coordinate identifiers and mask material separated

- Timestamp: 2026-10-25T20:59:58.857694+00:00
- Agent: Codex
- Status: completed

[data representation](../docs/data-representation.md) now identifies PIR databases
by their submatrix row/column ranges within the application database. Removed
random 128-bit application/version identifiers and extra identifier metadata.
Retained separate masking material and prevention of mask reuse across encodings,
without choosing a new key/nonce format.

Updated [encoder requirements](../docs/encoder-contract.md) and
[protocol I/O](../docs/protocol-io.md) to remove database/version-identifier
requirements. Checked parameter planning and README; their existing statements
remain accurate and needed no edit. Historical messages are unchanged.

Verification: searched current docs for obsolete 128-bit and database/version
requirements, checked file links and heading anchors, and ran `git diff --check`;
all passed. No runtime tests or experiments needed for documentation edits.
