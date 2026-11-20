# SCALING-152 — handoff

2026-11-20 17:07 UTC; Codex.

Committed resident storage implementation as `f36074772d7634f7063a4a55a328fe6f621ca658`. Submitted **job78392**:
[receipt](../experiments/distributed-gpu-storage/submission.json),
[report](../experiments/distributed-gpu-storage/report.md),
results `experiments/distributed-gpu-storage/runs/storage-01/`,
log `experiments/distributed-gpu-storage/slurm-78392.log`,
cache `.cache/liftrs/distributed-gpu-storage/storage-01/`.

Existing2HPCnodes/2H200/4CPU/8GiBhost envelope,128MiBowned/card,
1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler;
no pinned buffers. Normal/sanitized104passes/rank in13geometries, both widths,
all row/interleaved layouts and ownership tails. Source/evidence/cache seals,
profiles, timestamped stage/proof/resource logs and both CUDA summaries retained.
No scheduler/log/result monitoring performed. Wait for the completion ping.

Updated the main scaling reference, plan and README to say **submitted/results
pending**, without changing fulls17 evidence or claiming fulls18 support. README
check recorded. Local CPU/race/native/C/MPI-build and19Python checks pass; exact
commands/warnings/hashes are in readiness evidence. Whitespace check passed.
Resident vector/Taylor/pointwise/full product composition remains next, conditional
on validation. No additional go-ahead needed for obvious admitted next steps;
resource increases and long-job handoff retain their explicit rules.
