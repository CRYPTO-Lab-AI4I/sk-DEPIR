# DOC-032: Annotated review applied across the contracts

**Timestamp:** 2026-10-25 18:45 UTC
**Agent:** Codex
**Status:** Completed

Updated [representation](../docs/data-representation.md) with concise partitioning
language, a required minimum application-record size, zero-based half-open ranges,
and exact coverage. Query compatibility now distinguishes padded dimensions and
codeword mappings from application row ranges and the per-request index condition.
Specified application/submatrix identities and fresh versions for changed data.

Expanded [catalogue requirements](../docs/parameter-planning.md) to include verified
`K`, `N`, code rate `K/N`, transmitted query length and dummy rule, and reproducible
encoder evidence. Distinguished layout rate, shortened the sizing target and
selection output, and recorded minimum selected security as the planner's estimate
under the documented protocol rules. Simplified the corresponding
[I/O](../docs/protocol-io.md) text without asserting a new security proof.

Split [encoder requirements](../docs/encoder-contract.md) into non-cryptographic
codeword encoding and cryptographic permutation/masking, each with explicit
inputs and outputs. Specified codeword-to-storage permutation direction and
storage-position mask indexing. These are future contracts, not runtime changes.
Reviewed and updated [README](../README.md) and [AGENTS.md](../AGENTS.md).

Verification: `git diff --check` and documentation path/anchor checks passed.
Reviewed consistency of the seven requested edits. No code or experiments
changed; runtime tests were unnecessary. See the
[start message](2026-10-25-1843-codex-doc-032-started.md).
