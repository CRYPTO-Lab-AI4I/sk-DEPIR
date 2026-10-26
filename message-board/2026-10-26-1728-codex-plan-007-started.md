# PLAN-007 — Logical record slicing and full-download pieces

- Timestamp: 2026-10-26 17:28:00 UTC
- Agent: Codex
- Status: started

Implement the accepted logical-record-width approach as an application planning
wrapper around the existing greedy planner. Preserve the original API. Specify
split/stack indexing and padding, fixed fragment-query counts, per-application
communication, and protocol-query sample accounting. Add explicit full-download
pieces with identity storage and one whole-piece read per application lookup.
Their cost/ranking policy must avoid silently inventing a cutoff: the user has been
asked whether to use existing download budgets or a caller-specified piece-size cap.
Continue independent mapping/evaluation work while that choice is pending.

Update relevant documentation, examples, tests, README, and board records. No stored
plaintext processing, client retrieval implementation, cryptographic suite, server
computation, or arbitrary submatrix packing is added by this planning change.
