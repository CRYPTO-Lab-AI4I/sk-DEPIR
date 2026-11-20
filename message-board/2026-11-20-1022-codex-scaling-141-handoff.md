# SCALING-141 — handoff

UTC: 2026-11-20T10:22:32.569796+00:00

Submitted **78279**, bounded32/64-bit sharded transpose validation, source
`e96ed6ab5e4286e43cf5106fb93ba10d10eda3c6`. Same2 HPC nodes/4 ranks/CPUs/8 GiB host/zeroGPU envelope,
1GiB RSS/rank,128MiB retained/2GiB scratch/32MiB file,720s controller/15min scheduler.
Largest actual explicit arrays1.125MiB/rank plus8KiB optional MPI staging; full
host-reference footprint1TiB/16ranks is projection only, not chosen production
layout/admission. Two modes, eight global-coordinate operations per mode, paired
metadata/proofs, everyword/output shard verification, old root/support lane
regressions, profiles/guards/progress/strict logs retained. Independent CPU/race/
descriptor checks pass; existing GPU/native harness compiles;2new plus36 prior
Python reviewer tests pass. No local MPI/GPU execution.
[Report](../experiments/distributed-transpose/report.md),
[receipt](../experiments/distributed-transpose/submission.json),
[freeze](../experiments/distributed-transpose/readiness/snapshot.json).
Results/profiles/rank logs `experiments/distributed-transpose/runs/transpose-01/`;
progress `experiments/distributed-transpose/slurm-78279.log`; reusable binary/
output shards/failure captures `.cache/liftrs/distributed-transpose/transpose-01/`.
Snapshot672files/10checked inputs match; README/main reference/plan/preprocessing/
shared docs/index updated. Estimated2–4min after scheduling; no monitoring or
results inspection after submission. This is the first collective FFT data-movement
reference/control implementation, not a distributed GPU FFT/preprocessing fallback.
GPU vector/Taylor composition and fulls18 remain unimplemented; no default change.
Wait for user completion ping.
