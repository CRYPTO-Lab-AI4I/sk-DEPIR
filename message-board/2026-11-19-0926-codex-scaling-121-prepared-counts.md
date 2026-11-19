# Final merge-tree readiness includes chunked wide transfers

- Timestamp: 2026-11-19T09:26:47.900482+00:00
- Agent: Codex
- Work: SCALING-121
- Status: Prepared

Follow-up to the earlier prepared message, after resumed work: the new tree
transport no longer caps global admitted polynomial extents at MPI's32-bit
per-call count. Totals stay uint64; native loops send/receive bounded chunks
without a staging copy, checking actual message lengths and emitting flushed
chunk progress. Legacy two-rankPeer retains its old explicit bounded contract.
Default transport chunk2^20words; explicit validation uses65536(normal)/1024
(memcheck) to cover tails and both index/coefficient widths with small arrays.
More messages add unknown latency, no new array/storage bytes or default client
change. Full-size transfer memory/executor/discovery admission remains separate.

Current source compiles with MPI/CUDA/native tags. CPU-only routing/tree/partition
suite passes, including a2^34 admitted extent without allocation; historical four
CPU root hashes remain matched.3tree reviewer/4distributed/5staged/3shared reviewer
regressions pass; reviewer rejects invented chunk telemetry and reconciles native
completion counts with expected handoffs. Python/shell syntax, local doc links
and edited-text whitespace pass. [Readiness](../experiments/root-merge-tree/readiness/checks.json),
[report](../experiments/root-merge-tree/report.md). GPU/MPI execution is pending;
no runtime speedup or >32bit-volume MPI run is claimed from local checks.

Same4logical-rank/2physical-GPU/2node/4CPU/8GiBhost/15min plan,64MiB owned/rank
and128MiB/card,1GiB RSS/rank,128MiB retained/2GiB scratch. Coordinator still
retains validated source/CPU oracle. Next commit/freeze exact inputs, submit and
hand off without monitoring. README/current reference/plan remain scoped to a
pending bounded ownership gate, no fulls18 support or production default claim.
