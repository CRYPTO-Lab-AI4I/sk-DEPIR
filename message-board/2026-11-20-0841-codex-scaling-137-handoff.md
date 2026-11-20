# SCALING-137 — handoff

UTC: 2026-11-20T08:41:36.847628+00:00

Submitted **78249**, same-build GPU tree crossover with explicit UCX cache control,
source `d4621d2f3059817d4446606f01bd14e8672fafec`. Previously admitted2 HPC nodes/2 H200/4 ranks/CPUs/8 GiB host,
64 MiB owned/rank,1 GiB RSS/rank,128 MiB retained/2 GiB scratch,15 min scheduler.
No resource increase. Shared UCX preflight verifies linked1.18 config; per-worker
requested controls and native selection logs retained. One build, four ordered
staged/direct/direct/staged arms, original CPU/state/proof/memcheck/profile checks.
[Report](../experiments/ucx-cache-gpu-tree/report.md),
[receipt](../experiments/ucx-cache-gpu-tree/submission.json),
[freeze](../experiments/ucx-cache-gpu-tree/readiness/snapshot.json).
Results/profiles `experiments/ucx-cache-gpu-tree/runs/tree-01/`; progress
`experiments/ucx-cache-gpu-tree/slurm-78249.log`; retained arrays/binary/library
`.cache/liftrs/ucx-cache-gpu-tree/tree-01/`. Snapshot662 files/10checked inputs
match; tagged build, proof/settings units and32 Python tests pass. Host78218 pass
verified/documented, README/main reference/plan/index/report updated. Estimated
4–7min after scheduling; no default/fullsize/production claim. Registration cost
and internal pinned memory unknown. No job monitoring/results inspection after
submission; wait for user completion ping.
