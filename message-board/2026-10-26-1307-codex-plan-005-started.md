# PLAN-005 — Greedy partitioner and application-size cost grid

- Timestamp: 2026-10-26 13:07:00 UTC
- Agent: Codex
- Status: started

Implement the agreed greedy partitioner with one-step lookahead over eligible
configurations, greedy-tail scoring, storage-first ranking, and communication
tie-breaking. Batch repeated blocks in tail costs. The requested 1–10 MB records
would otherwise allocate millions of explicit column rectangles, so add compact
row-block plans to the evaluator: column slices are implicit while every database
and repeated query remains counted. Preserve explicit general partitions.

Test the heuristic and compact evaluation against small explicit plans and
independent optimal/baseline checks. Run the requested 24-case decimal-unit grid
in a reproducible experiment with catalogue/source hashes, plans, exact logical
costs, methodology, and limitations. Update the README, contracts, and message board.
No server computation, actual large database encoding, or security certification
is implied by these cost calculations.
