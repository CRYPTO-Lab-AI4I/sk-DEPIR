# SCALING-126 — completed

UTC: 2026-11-19T13:37:23.511295+00:00

Reviewed job 77799 after user completion ping. Bounded normal validation and CUDA
memcheck passed; raw workflow failed afterward on 30 s profile-summary timeout.
No original status/log/result was changed. Verified 642 source / 23 evidence /
87 cache seals, 22 normal passes / 88 rank measurements, 516 paired normal proof
endpoints, exact shard coverage and CPU equality/state snapshots. Four memcheck
processes: zero errors/leaks, 24 additional proofs. Actual staging sizes checked.
HPC gnode13/gnode14: 2 H200, 4 logical ranks/CPUs, 4 GiB requested host RAM/node.
Largest partial case median 1.201005 s; normal largest logged peak RSS 411.309 MiB
per rank, owned GPU arrays 15.241 MiB per rank; retained cache 27.567 MiB.
Scheduler 2 min 43 s; normal command 38.258 s; memcheck command 4.283 s.

[Report](../experiments/mpi-owned-staging/report.md),
[review](../experiments/mpi-owned-staging/runs/staged-01/analysis.json),
[commands/hashes](../experiments/mpi-owned-staging/runs/staged-01/review-checks.json).
All four original profiles read offline under 20 s per command; recovered summaries
are additional derived files. Timeout cause is unconfirmed. Future shared tree
runner uses `-symbolize=none` for offline summaries after measured execution;
full profiles retained, no kernel/benchmark-clock change. Native symbols are not
resolved by this summary. Shared reviewer now validates completed stages even if
later reporting fails, preserving the original workflow status and reason.
Regression test covers later reporting failures versus failed/guarded stages.
4 transfer / 3 tree / 4 distributed / 5 staged / 3 common review tests passed;
all profile reads, scheduler accounting, diff and Python syntax checks passed.
No MPI/GPU rerun or new job submitted. No native/Go compilation needed: only
Python orchestration/review and documentation changed.

Updated README (checked), main encoding scaling reference/plan, preprocessing/CUDA
docs and experiment index. Direct-mode failures remain unexplained; full s18
preprocessing remains unrun. Staging is explicit, 512 KiB scratch/rank plus CPU
copies; isolated runtime penalty unknown. Next controlled direct/staged comparison
uses identical instrumentation before any default decision. Transport warnings
retained/documented; no verified RDMA or GPU throughput claim.
