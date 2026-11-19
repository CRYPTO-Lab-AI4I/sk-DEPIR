# Rank-owned support and multi-level GPU root reduction

- Timestamp: 2026-11-19T08:13:53.599134+00:00
- Agent: Codex
- Work: SCALING-121
- Status: Started

User authorizes next scalable ownership/tree work. Reuse verified support/root/
GPU product and MPI helpers, remove non-coordinator full-support/reference copies
and duplicate final merges. Add bounded streamed routing and deterministic owner
reduction supporting arbitrary2..64 ranks, shared pure-Go tests and opt-in MPI/GPU
validation. Coordinator validation fixture is still full and discovery/large
transforms remain separate gaps; do not claim fulls18 support.

Four logical ranks across2HPC nodes/2H200 can validate2 tree levels within existing
budget:2ranks share each card,1CPU/rank,2CPU/4GiB requested host RAM per node,
64MiB owned/executor≤128MiB/card,1GiB RSS/rank≤2GiB/node,128MiB retained/2GiB
scratch/15min scheduler. Explicitly not4GPU speedup or a production throughput
allocation. Retain independent CPU/scalar/canonical references on coordinator,
all shard/root/final identities, profiles, progress and small memchecks. No
filesystem arrays for inter-rank exchange, no current backend/default change.
Read message board, current scaling plan, root fixture/partition and MPI code,
README and existing reports. Hand off any long job without monitoring.
