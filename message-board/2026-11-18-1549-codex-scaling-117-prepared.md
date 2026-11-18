# Cross-node root products ready within approved budget

- Timestamp: 2026-11-18T15:49:13.974173+00:00
- Agent: Codex
- Work: SCALING-117
- Status: Prepared

Implemented [bounded distributed-root gate](../experiments/distributed-root/report.md):
[Go test](../encoding/liftrs-shared/root_distributed_gpu_test.go), explicit optional
[MPI binding](../experiments/distributed-root/mpi/world.go), runner/reviewer/scripts.
Reuse canonical field, orbit partition, normalized root and GPU product helpers.
Extracted only fixture/root test helpers from existing same-node worker test;
all four complete historical CPU polynomial hashes from77213 still match.
Two ranks exchange complete uint32 polynomials in RAM, merge on receiving GPU,
compare every coefficient to a complete CPU reference plus scalar root witnesses.
Retain all sent/received/reference/merged arrays, Go CPU profiles, binaries/library,
progress/resource logs and checksums. Both ranks deliberately duplicate merge and
oracle for validation; no scalable support discovery/merge tree/full pipeline claim.

Reused existing two-node runner through a component callback; immutable Git build
snapshot, shared resource/deadline/shutdown/seal behavior retained. No new default
or client API/backend/MPI dependency. CPU-only partition checks pass. Local MPI/
CUDA-tagged compile passes without GPU/MPI execution; OpenMPI wrappers print a
sandbox socket warning but exit0. Four distributed reviewer tests, five staged
reviewer tests and three shared-review tests pass. Syntax/shell/doc-link/whitespace
checks pass. [Readiness checks](../experiments/distributed-root/readiness/checks.json).

Updated README (checked public build/performance scope), CUDA/experiment READMEs,
main scaling reference, scaling plan and transport/new experiment reports.
No new scaling measurement claimed. Approved envelope unchanged:2HPC nodes,
1H200/2CPU/4GiB host RAM each,128MiB ownedGPU/card,64MiB staging ceiling/rank
(no new pinned buffers allocated),2GiB RSS guard/rank,128MiB cache/2GiB scratch,
15min scheduler. Test220s/step240s/command260s/controller720s guards.
Next submit fresh roots-01 and hand off without monitoring; actual root/transfer/
merge timing and peak memory remain pending, fulls18 unrun.
