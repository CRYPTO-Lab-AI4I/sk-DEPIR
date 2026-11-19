# SCALING-129 — prepared

UTC: 2026-11-19T16:48:51.893032+00:00

[Host-only replay](../experiments/mpi-host-buffers/report.md) is ready. Replay
verified captured public sender polynomial, fresh Go payloads through existing
World, same-node/cross-node roots and uint64 patterns, normal versus forced Go
reclamation. Staged controls precede direct; failure stops job, captures sync on
both sides before abort. Shared proof handshake extracted once to mpi/proof.go;
GPU-tree helpers delegate without changed semantics. CPU-only runner reuses
shared freezing/guards/seals and skips CUDA setup; zero GPU allocation/dependency.
Useful timestamped progress/barrier-completed counts and 30 s waiting heartbeat.
All CPU profiles and replay/cache retained. No production default changes.

Both host and tagged GPU-tree builds pass within 25 s; host hash/high-bit/proof
unit and original two tree proof tests pass. Python host/comparison/tree/proof/
distributed/staged/common reviewer tests pass (27 total). Scheduler/Python syntax
and diff whitespace checked. No local MPI/GPU execution.
[Commands/hashes](../experiments/mpi-host-buffers/readiness/checks.json).
README checked/updated, main scaling reference/plan, index and reports updated.
Same 2 HPC nodes / 4 CPUs / 8 GiB host, now zero GPUs; 1 GiB RSS/rank,
16 MiB churn/rank plus sub-MiB payload buffers, staged scratch512 KiB/rank.
128 MiB retained / 2 GiB scratch, 720 s controller / 15 min scheduler unchanged.
Estimated2–5 min after scheduling. Forced reclamation is diagnostic stress only;
non-reproduction cannot clear original fault or establish cause. Submit/freeze
and hand off without monitoring. Fulls18 and isolated staging penalty unknown.
