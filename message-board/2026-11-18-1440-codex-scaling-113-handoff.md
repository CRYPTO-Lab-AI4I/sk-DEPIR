# GPU staged gate awaiting resource agreement

- Timestamp: 2026-11-18T14:40:16.875110+00:00
- Agent: Codex
- Work: SCALING-113
- Status: Handoff

Concrete implementation/report/local checks committed **ac774a2**. Async approval
question requests2nodes,1H200+2CPU+4GiB host/node,128MiB owned/card,64MiB pinned/
2GiBRSS per rank,15min scheduler. This changes prior zero-GPU/10min two-node
budget; no sbatch submitted before agreement. [Exact runner/resource plan](../experiments/gpu-staged-transport/report.md).

NVCC compile+MPI/cudart profiled link,3 review tests, Go fixture generation/schema,
AST,bash syntax,runner --check,whitespace pass. Runtime GPU/MPI/PMIx binding,
full field-oracle checks and small two-rank memcheck remain unrun. Slurm pmix_v3
plugin and OpenMPI ext3x were inspected without allocating or monitoring jobs.
Rank-specific Slurm GPU binding, shared-cache wrapper and Git-blob build snapshot
are explicit. No ignored-source override is needed. Local metadata/source hashes
are in local-checks.json; no useful cache/profile removed.

README/active plan/actual reference/experiment+CUDA docs updated as in prepared
message. No fullsize/default/capacity change or write-concurrency job. After
approval submit staged-01, recordID/results and hand off without monitoring.
No new substantive work or source edits after the approval question; this message
records its pending status only. Await reply rather than infer agreement.
