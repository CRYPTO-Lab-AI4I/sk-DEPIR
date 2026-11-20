# SCALING-144 — prepared

UTC: 2026-11-20T10:46:19.373877+00:00

78285 raw failed preserved. Reconciled680 source/17 original evidence/3 cache
hashes; fixture/native CUDA build and Python review checks passed. Failed launcher
exit127 on remote node missing local /tmp source wrapper; no GPU kernel/test,
CUDA memcheck or native CPU profile produced. Controller40.262s. Copy rank entry
from frozen source to shared cache and pass explicit --chdir=shared cache. Added
launcher regression proving remote command excludes node-local source paths;
3 Python tests pass, shell syntax/whitespace checked. Reviewer labels corrected
from inherited unrelated test names to actual identity/stage review scope.
README checked/updated with main reference/plan/preprocessing/CUDA/index/report.
[Report](../experiments/gpu-vector-stage/report.md),
[review](../experiments/gpu-vector-stage/runs/vectors-01/analysis.json).
No arithmetic/default/resource changes. Retry vectors-02 in existing2node/2H200/
4CPU/8GiBhost/128MiB-owned-card/1GiB-RSS-process/128MiBcache/2GiBscratch/720s-controller/
15min-scheduler budget. Commit exact sources before submission, then yield without
monitoring. Full s18/distributed FFT remains pending.
