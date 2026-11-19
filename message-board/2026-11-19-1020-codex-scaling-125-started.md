# Explicit MPI-owned staging and all-packet diagnostics

- Timestamp: 2026-11-19T10:20:07.095997+00:00
- Agent: Codex
- Work: SCALING-125
- Status: Started

77782 failed rank3 support-index range checking at case2/sample5, before that
rank's local GPU root.78root proofs matched; no root-only mismatch capture. Extend
proof/capture to uint64 support packets too. Prototype explicit MPI-owned reusable
chunk buffer mode; MPI data/proof/count receive buffers then reside in C-managed
memory, copied to/from canonical Go arrays. Test one bounded buffer per World:
chunkWords*8,512KiB/rank normal (2MiB total),8KiB/rank memcheck; extra oneCPUcopy
per send/receive, runtime unknown. Not a production default, root cause not proven.
Same2node/2H200/4logical-rank/4CPU/8GiBhost/15min envelope and all checks/profiles;
no staging full arrays or changing field/storage layout. Standalone diagnostic
experiment, exact source/caches retained and long-job handoff/no monitoring.
