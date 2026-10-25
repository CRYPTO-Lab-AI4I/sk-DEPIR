# DOC-047: Routine decisions resolved; publication ready

- Timestamp: 2026-10-25T23:30:40.105655+00:00
- Agent: Codex
- Status: completed

Resolved routine choices in [data representation](../docs/data-representation.md):
least-significant-bit-first application columns and packed field elements, trailing
zero padding, a UTF-8 version-1 JSON manifest, deterministic coordinate ordering
and client-side group numbering, and explicit CSV conversion.

[Encoder requirements](../docs/encoder-contract.md) now specify deterministic small
reference preprocessing: monomial evaluation matrix, ordered Gaussian elimination,
pivot information set, and message multiplication. Preserve existing first-coordinate-
fastest grid indices, keep dense artifacts in memory, and require checked allocation
limits. Cryptographic randomness uses crypto/rand, per-database keys can be generated
independently, and re-encoding uses fresh masks. Concrete cryptographic suite review
and the scalable backend remain substantive tasks, not arbitrary defaults.

[Protocol I/O](../docs/protocol-io.md) defines ordinary byte-range adapters, uint64
positions, bounds/overflow checks, response order, wrapped Go errors, and caller
cancellation. No server wire protocol is required. [Planning](../docs/parameter-planning.md)
distinguishes minimal logical address widths from adapter serialization. Updated
[README](../README.md), [roadmap](../ROADMAP.md), and the calculation
[report](../experiments/bounded-sample-parameters/report.md) for current status.

Next: implement the bounded dense reference encoder and independently validate
monomial membership, systematic recovery, and exhaustive tiny-instance retrieval.
Small fixtures carry no production-size security claims.

Verification: reproduced [results.csv](../experiments/bounded-sample-parameters/results.csv)
byte-for-byte with its script; verified paper SHA-256 and the 9-bit packing example;
checked current documentation links/anchors and `git diff --check`. Inspected
existing coordinate, masking, permutation, and encoder code; consulted official Go
API docs linked from the contracts. No runtime implementation changed, so full Go
tests were not rerun. Historical board links to replaced PDFs remain historical.

Publication: the user authorized committing and pushing all pending session changes,
including the supplied replacement paper. A separate receipt will record the
commit and push outcome.
