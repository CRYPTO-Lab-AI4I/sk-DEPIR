# DOC-044: Annotated review applied

- Timestamp: 2026-10-25T21:29:07.966064+00:00
- Agent: Codex
- Status: completed

Updated [encoder requirements](../docs/encoder-contract.md): application encoding
owns padding/bit conversion; mathematical encoding consumes exactly K field elements
and a compatible artifact. Clarified the paper's Algorithm 1, Theorem 4.3,
Claim 4.2, and encoding steps versus remaining implementation choices.

Updated [parameter planning](../docs/parameter-planning.md): distinguish candidates
from configurations supported for encoding, emphasize useful and potentially
accurate Monte Carlo evidence, and separate empirical rates, exact calculated
rates, and measured storage. Removed the identical-prior-layout requirement.
The planner never runs code preprocessing; optional resource estimates use
existing evidence and account for reusable setup once per code.

Kept field representation in place in [data representation](../docs/data-representation.md),
as requested. Consolidated query execution rules in [protocol I/O](../docs/protocol-io.md)
with cross-references; retained layout reconstruction and mask separation in data
representation. Moved the query implementation note after the conceptual rules,
restored list spacing, and reflowed prose. Preserved the user's mask wording.
Updated [README](../README.md) for rate evidence, the application encoder's input
conversion, and catalogue links.

Verification: reviewed responsibilities and retained security requirements; checked
relative links and Markdown anchors in docs and README; `git diff --check` passed.
No implementation, paper, experiment, or result artifacts changed; runtime tests
were not needed. Mathematical methods remain to be implemented and validated.
