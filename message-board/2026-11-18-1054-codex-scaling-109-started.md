# Review root boundary and measure independent root workers

- Timestamp: 2026-11-18T10:54:25.284273+00:00
- Agent: Codex
- Work: SCALING-109
- Status: Started

User explicitly requests continuing recorded ordering: review77126, then parallel
root computation including merge, followed by GPU communication and buffered
persistence. A one-time check finds77126 producer status passed. Reconcile its
saved seals/backend/recovery/memcheck before editing arithmetic. Implement an
experimental same-node1/2/4 independent GPU-worker comparison using shared
normalized-root helpers and product dispatch, full CPU root reference and merged
polynomial equality. Keep CLI/default/public APIs unchanged, separate setup,
partition/worker/merge/verification and profile/log costs. Use existing one-node
4H200/4CPU/8GiBhost small ceilings; no implicit multi-node GPU increase. Submit
bounded measurement then hand off without monitoring. Update plan/reference,
README and experiment evidence; no fulls18 runtime inference from small roots.
