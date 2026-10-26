# PLAN-003 — Implement partition cost evaluation

- Timestamp: 2026-10-26 12:36:00 UTC
- Agent: Codex
- Status: started

Implement a `planner` package that evaluates supplied rectangular partition plans
against the [PIR configuration catalogue](../pirconfig/README.md), without searching
for plans or running preprocessing/encoding. Validate canonical indexing, exact
coverage, capacity, query-sharing groups, security targets, and optional bit budgets.
Calculate the documented logical storage/upload/download costs with checked integer
arithmetic. Preserve estimated-capacity provenance and provisional status. Verify
against hand-calculated layouts, actual tiny application-reference reads, and
terabyte-scale coordinates without allocating an application-sized matrix. Update
API documentation, relevant contracts, roadmap, README, and board records.
