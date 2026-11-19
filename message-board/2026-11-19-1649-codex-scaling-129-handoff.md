# SCALING-129 — handoff

UTC: 2026-11-19T16:49:51.581708+00:00

Submitted **78000**, source `c48ee74311b857dd1d61ec123b71e1d9919489c9`, host-only MPI replay.
2 HPC nodes / 4 logical ranks / 4 CPUs / 8 GiB host total / zero GPUs; existing
RSS/storage/deadline limits. Staged normal/reclaim controls before direct,
stop on failure, shared proof/capture transport, replay of sealed public data.
Host/GPU-tree builds and local unit/review tests pass; source freeze652 files /
8checked inputs match. No MPI job monitoring or results inspection after submit.
[Report](../experiments/mpi-host-buffers/report.md),
[receipt](../experiments/mpi-host-buffers/submission.json),
[freeze](../experiments/mpi-host-buffers/readiness/snapshot.json).
Results/profiles `experiments/mpi-host-buffers/runs/replay-01/`; progress
`experiments/mpi-host-buffers/slurm-78000.log`; sealed replay/build/captures
`.cache/liftrs/mpi-host-buffers/replay-01/`. Estimated2–5 min after scheduling,
15 min scheduler cutoff. README/main reference/plan/index/report submission links
updated. Forced reclamation is stress only, source cause and runtime penalty
unknown; no production default or fulls18 change. Wait for user completion ping.
