# SCALING-125 — prepared

UTC: 2026-11-19T10:27:05.135141+00:00

Implemented an explicit `-root-mpi-buffer=staged` experimental mode: reusable
MPI-owned chunk, extra 512 KiB/rank (2 MiB total) normally, 8 KiB/rank under
memcheck, one CPU copy per send/receive, no new GPU transforms/disk payload.
Runtime penalty unknown; direct mode and production defaults unchanged.
Count/proof control records now use C-local buffers in both modes.

Extended the shared proof/capture handshake to uint64 support packets before
range checks. Both payload types have phase/sequence/width-bound paired capture
and a post-fsync acknowledgement. Index hashing retains even invalid high bits.
Reviewer checks paired proof identities and received-shard coverage; raw failed
77782 evidence preserved. No claim that a hypothesis is the proven cause.

Changed MPI transport, tagged test/proof helpers, shared tree runner/reviewer,
new [experiment](../experiments/mpi-owned-staging/report.md), relevant README,
CUDA/preprocessing docs, [scaling reference](../docs/encoding-status.md), plan and
experiment index. README checked and updated.

[Readiness](../experiments/mpi-owned-staging/readiness/checks.json): tagged native/
MPI/CUDA build passed under 25 s deadline; two Go fingerprint tests passed;
3 transfer / 3 tree / 4 distributed / 5 staged / 3 common runner review tests pass;
scheduler syntax and diff whitespace checks pass. No local MPI/GPU execution.
Same admitted 2 HPC node / 2 H200 / 4 logical rank / 4 CPU / 8 GiB requested host,
64 MiB owned/rank, 1 GiB RSS/rank, 128 MiB retained / 2 GiB scratch, 15 min scheduler.
A new bounded run will be handed off without monitoring; full s18 remains unrun.
