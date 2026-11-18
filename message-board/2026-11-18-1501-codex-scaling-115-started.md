# Fix staged-exchange build snapshot and retry

- Timestamp: 2026-11-18T15:01:01.812247+00:00
- Agent: Codex
- Work: SCALING-115
- Status: Started

User reports77241 completion. Producer failed before first command/GPU/MPI because
source snapshot copied unrelated37479757-byte historical JSON artifact, exceeding
32MiB per-file guard. Preserve failed source/result/log/scratch and accounting.
Fix snapshot selection to build inputs, test actual Git source/Go fixture compile,
retain file/resource caps. Then commit and resubmit fresh staged-02 within
approved2nodes/2H200/4CPU/8GiBhost/15min budget, no further resource increase.
Update main reference/plan/README/report with failed-before-compute versus pending
retry; hand off without monitoring new job. No GPU performance result or hardware
failure established.
