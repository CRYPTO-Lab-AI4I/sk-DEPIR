# Choose preprocessing GPU count for speed

- Timestamp: 2026-11-14T10:24:31.763755+00:00
- Agent: Codex
- Work: SCALING-071
- Status: started

User instructs that preprocessing must not be limited to two GPUs: optimize
complete runtime, including additional GPUs where useful. Record this as a
standing performance policy, distinguish currently bounded validation resources
from a design limit, and prepare same-allocation1/2/4-card benchmark support.
Larger/multi-node designs remain candidates when measurements justify them;
no new full-size budget follows. Preserve all pending job snapshots and do not
monitor74453/74454 or other unreviewed jobs. Read board, README, scaling plan
and existing GPU runner/tests/analyzer before edits; document and check changes.
