# SCALING-135 — handoff

UTC: 2026-11-20T08:20:09.633945+00:00

Submitted **78218**, UCX registration-cache disabled host replay, source
`19e6143b55b86ea6e1633d1a83353afb00093b16`. Same 2 HPC nodes /4 ranks/CPUs /8 GiB requested host /zeroGPUs,
unchanged RSS/storage/deadline envelope. Observed UCX backend retained, explicit
UCX_RCACHE_ENABLE=n, actual-node UCX1.18 config and worker controls verified.
[Report](../experiments/ucx-registration-cache/report.md),
[receipt](../experiments/ucx-registration-cache/submission.json),
[freeze](../experiments/ucx-registration-cache/readiness/snapshot.json).
Results/profiles `experiments/ucx-registration-cache/runs/no-cache-01/`; progress
`experiments/ucx-registration-cache/slurm-78218.log`; retained replay/build/captures
`.cache/liftrs/ucx-registration-cache/no-cache-01/`. Snapshot658 files /5 checked
inputs match; host build/unit and31 Python tests pass. README/main reference/plan/
index/reports updated. Estimated1–3min after scheduling;15min cutoff. Registration
cost/internal pinned memory unknown, diagnostic only. Original failure unresolved;
no default/fullsize capability change. No job monitoring or results inspection
after submission; wait for user completion ping.
