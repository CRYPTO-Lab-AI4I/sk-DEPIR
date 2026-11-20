# SCALING-139 — handoff

UTC: 2026-11-20T08:54:59.486060+00:00

Submitted **78261**, fresh tree-02 crossover with strict per-rank stdout/stderr,
source `e9f31f57194268e67e34c3340c4386d2eab7dbee`. Same admitted2H200/2nodes/4ranks/CPUs/8GiB host,
64MiB owned/rank,1GiB RSS/rank,128MiB cache/2GiB scratch,15min scheduler.
One build and original four-arm correctness/memcheck checks, explicit UCX cache
control unchanged; no arithmetic/Go/CUDA/native source or clock changes.
[Report](../experiments/ucx-cache-gpu-tree/report.md),
[retry receipt](../experiments/ucx-cache-gpu-tree/retry-submission.json),
[freeze](../experiments/ucx-cache-gpu-tree/readiness/retry-snapshot.json).
Results/profiles `experiments/ucx-cache-gpu-tree/runs/tree-02/`; launcher progress
`experiments/ucx-cache-gpu-tree/slurm-78261.log`; each phase also retains8 rank
stream files. Arrays/binary/library `.cache/liftrs/ucx-cache-gpu-tree/tree-02/`.
Snapshot664 files/8 checked inputs match;36 local Python tests pass. Original
78249 raw failure retained; three normal checks validated with explicit derived
recovery, two full rounds/memchecks passed. README/main reference/plan/index
updated with measured partial result and pending retry. Estimated4–7min after
scheduling; direct file I/O overhead unknown, no cross-run/default/fullsize claim.
No monitoring or result inspection after submission; wait for user ping.
