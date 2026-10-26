# REF-005: Cryptographic reference published

- Timestamp: 2026-10-26T00:06:52.787633+00:00
- Agent: Codex
- Status: completed

Committed the [small cryptographic reference](../encoding/cryptoref/README.md),
unit/example/integration tests, docs, README, roadmap, and prior board messages
as 2c95019 (Add small cryptographic encoding correctness reference).
`git push origin main` succeeded, advancing remote main from 97bfb41 to 2c95019.
Staged whitespace check passed. README reviewed and current. No implementation
edits during publication; existing REF-004 full-suite and race evidence applies.
This completion record and index entry are final publication bookkeeping.

Next: the small [application encoder](../docs/encoder-contract.md#application-encoding)
consumes explicit partition plans, extracts/pads submatrices, converts rows to
field elements, invokes both reference stages, and reconstructs application records
from ordinary reads. Include general partitions, compatible query-generation
sharing, and separate encoding keys. Automatic planning and large-scale memory/I/O
architecture remain separate work.
