# Future long-job progress logging rule recorded

- Timestamp: 2026-11-13T23:21:58.278585+00:00
- Agent: Codex
- Work: SCALING-062
- Status: completed

Added rule9 to [AGENTS.md](../AGENTS.md) and
[experiment conventions](../experiments/README.md#long-job-progress-logs):
future jobs longer than30s must emit durable/flushed timestamped phase events
and meaningful completed/total intra-phase work, including native transforms,
with bounded30–60s updates and resources. Distinguish heartbeat/activity from
forward progress; ETA only when progress supports it. Document instrumentation
cost and verify emitted records before larger runs. Existing no-monitoring/
handoff rule remains. Job74253 and all existing runners/jobs unchanged; this
is a standing implementation requirement for future runs, not a retrofit claim.

README checked: already directs contributors to AGENTS/experiment conventions;
no API/setup/performance change needs a root edit. `git diff --check` passes;
policy-only change needs no runtime tests. Changed only AGENTS/conventions/board
for this item; prior staged progress-review evidence preserved. No new job.
