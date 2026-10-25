# DOC-037: Paper and implementation references added

- Timestamp: 2026-10-25T20:25:50.960249+00:00
- Agent: Codex
- Status: completed

Added contextual references across all four specifications:

- [Data representation](../docs/data-representation.md): paper input alphabet,
  existing grid indices, field arithmetic, and CSV serialization.
- [Encoder contract](../docs/encoder-contract.md): Definition 4.4, monomial
  algorithm, Section 4.2.4 preprocessing/systematic encoding, Figure 5 masking,
  Definition 5.2 permutation, current RM placeholder and unused C++ helper,
  cryptographic helpers, and existing tests.
- [Parameter planning](../docs/parameter-planning.md): current field and parameter
  types, Figure 5 cost-model basis, and benchmark implementation/limitations.
- [Protocol I/O](../docs/protocol-io.md): Figure 5 and Section 5.2, query generation,
  curve helpers, response types, local reads, and decoding.

References explicitly distinguish paper constructions, prototype behavior, and
planned library requirements. In particular, documented that current `Params.K`
means curve degree, not code dimension, and that the slice/index mask interface
does not implement the planned database/version domains.

Verification: read the referenced paper passages and relevant code paths; checked
all docs file links, heading anchors, and PDF page bounds; `git diff --check`
passed. Reviewed README architecture, implementation limits, and benchmarks;
no README change needed. No code or experiments changed; runtime tests skipped.
