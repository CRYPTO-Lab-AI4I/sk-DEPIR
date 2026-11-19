# SCALING-127 — handoff

UTC: 2026-11-19T13:48:10.882884+00:00

Submitted **77868**, same-build staged/direct/direct/staged comparison, source
`334cdfa140b22f218cab750bac8b1b8b3b7eb669`. Same admitted 2 HPC nodes / 2 H200 / 4 logical ranks / 4 CPUs /
8 GiB host total, 15 min scheduler. Native build once, fresh rank processes per
arm, strict shared CPU/proof/state/memcheck checks, stop-on-failure. 24 local
review/orchestration tests passed; exact Git snapshot verifies 646 files and all
five checked input hashes. No local MPI/GPU execution, no job monitoring.

[Report](../experiments/mpi-buffer-comparison/report.md),
[receipt](../experiments/mpi-buffer-comparison/submission.json),
[source freeze](../experiments/mpi-buffer-comparison/readiness/snapshot.json).
Results/profiles: `experiments/mpi-buffer-comparison/runs/compare-01/`;
progress: `experiments/mpi-buffer-comparison/slurm-77868.log`; retained cache:
`.cache/liftrs/mpi-buffer-comparison/compare-01/`. Runtime estimate 4–7 min after
scheduling; comparison results remain unreviewed. Main reference/plan, README
(checked), index and previous experiment report updated with submission.
Staging adds copies and 512 KiB/rank; isolated runtime penalty unknown. No default
change, original failures unresolved and full s18 remains unrun. Wait for user ping.
