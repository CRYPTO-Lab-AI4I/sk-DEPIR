# GPU staged two-node exchange gate

- Timestamp: 2026-11-18T14:17:21.232086+00:00
- Agent: Codex
- Work: SCALING-113
- Status: Started

User requests continuing plan after matched subset validation. Prepare two-node,
one-GPU/rank pipeline: GPU field computation using maintained arithmetic, pinned
host staging, direct MPI exchange, device reconstruction/consumption and full
independent CPU checks. No shared-filesystem exchange arrays, distributed FFT,
fullsize run or CPU-client dependency. Reuse installed OpenMPI/CUDA and canonical
Go field metadata. Compare host-only, staged GPU exchange and computation timing,
retain matching native/fixture profiles, trace phase/count/RSS and bounded caps.
Finish concrete runners/review/tests/resource plan before second-node GPU budget
question; earlier two-node approval has zero GPUs, existing CUDA budget one node.
Update main reference/plan/README, commit and await approval if required. No
monitoring of unrelated jobs, slower default or inference of fulls18 runtime.
