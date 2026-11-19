# SCALING-125 — handoff

UTC: 2026-11-19T10:28:19.077926+00:00

Submitted **77799**, explicit MPI-owned staging and both-width proof/capture,
implementation `6a32c38e76f55d887405a4b5b1f97b6059713862`. Same admitted 2 HPC nodes / 2 H200 / 4 logical
ranks / 4 CPUs / 8 GiB host envelope; 512 KiB extra host scratch/rank and one
copy per send/receive. Runtime penalty unknown; direct/default paths unchanged.
No job monitoring or result inspection after submission.

[Receipt](../experiments/mpi-owned-staging/submission.json),
[report](../experiments/mpi-owned-staging/report.md),
[source readiness](../experiments/mpi-owned-staging/readiness/snapshot.json).
Results `experiments/mpi-owned-staging/runs/staged-01/`; progress
`experiments/mpi-owned-staging/slurm-77799.log`; retained payload/cache at
`.cache/liftrs/mpi-owned-staging/staged-01/`. Hashes/source provenance recorded by
runner. Reviewed README and kept main scaling reference/plan/experiment index
current. Local checks passed as recorded in prepared message. Original 77704/
77782 faults remain unresolved; no full s18 readiness claim. Wait for user ping.
