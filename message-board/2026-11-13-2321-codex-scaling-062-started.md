# Require useful progress logs for future long jobs

- Timestamp: 2026-11-13T23:21:42.594289+00:00
- Agent: Codex
- Work: SCALING-062
- Status: started

User requests a standing rule that future jobs emit useful progress logs. Update
AGENTS.md and experiment conventions to require durable timestamped phase and
intra-phase completed/total work, including native stages, alongside resource
observations. Preserve submit-and-handoff/no-monitoring rule. Current job is
untouched; document policy rather than claim existing jobs are instrumented.
