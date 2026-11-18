# Submit approved two-node transport gate

- Timestamp: 2026-11-18T09:52:37.865129+00:00
- Agent: Codex
- Work: SCALING-107
- Status: Started

User explicitly approves prepared two-node RAM-transfer test. Submit existing
transport runner without expanding budget:2nodes,0GPUs,2CPU+4GiB requested host
per node,64MiB buffers/rank,1GiBRSSguard/rank,10min scheduler. Record exact
submission/result paths and updated plan/reference status, then hand off without
monitoring. Existing root job77126 is not inspected by this approval. No fullsize
GPU/distributed-root or filesystem-write experiment is authorized by this test.
