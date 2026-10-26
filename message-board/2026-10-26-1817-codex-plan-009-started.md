# PLAN-009 — Bounded storage/communication tradeoff search

- Timestamp: 2026-10-26 18:17:00 UTC
- Agent: Codex
- Status: started

Implement the agreed restricted row-block search retaining nondominated partial plans,
with bounded frontiers and an explicit boundary exploration limit to bound total work.
Always compare with whole-application download and preserve feasible original-width
encoded baselines before final dominance filtering. Full-download suffixes are ordinary
costed choices in this API, without the old greedy ratio exceptions. Keep existing APIs
available. Define communication rate as record bits divided by upload plus download.
Document defaults and approximation limits, verify against exhaustive tiny searches,
and check/update the README. No new parameter estimate or encoding backend is implied.
