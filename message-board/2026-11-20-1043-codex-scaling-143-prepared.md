# SCALING-143 — prepared

UTC: 2026-11-20T10:43:06.378764+00:00

Implemented internal persistent VectorStage in encoding/liftrs-cuda/vector_stage.cuh,
reusing resident field/plans/capped tile and local_vectors. Global shape/origin
separate from local capacity; no full host array, no copied math, no public ABI
or CPU-build dependency/default change. Existing executor s≤17 admission retained.
Separate upload/forward/inverse/download with rejected-request recovery and bounded
launches. Two-GPU/two-node test driver covers24 cases/process, both slot widths,
axes/directions/nonzero/wide origins, independent native recursive/Horner oracles,
retained everyword outputs, error/reuse/allocation checks. Full suite memcheck,
Go/native CPU profiles, source/cache seals, guards/deadlines, timestamped phase/
work/resource logs and labelled30s heartbeat prepared. Same admitted2nodes/2H200/
4CPU/8GiB requested host,128MiB owned/card,1GiB RSS/rank,128MiB retained/2GiB scratch,
32MiB file,720s controller/15min scheduler. No resource increase requested.

[Report and reproduction](../experiments/gpu-vector-stage/report.md),
[runner](../experiments/gpu-vector-stage/run.py),
[review tests](../experiments/gpu-vector-stage/test_review.py).
Two Python review tests pass including missing-route/byte-corruption rejection;
Python AST, shell syntax, gofmt and git diff --check pass. CUDA compile/execution,
canonical field fixture generation and native checks are deferred to the job
(expected longer than30s); no local GPU/MPI execution or result claimed.
README checked/updated; CUDA docs, main scaling reference, preprocessing, scaling
plan and experiment index updated. Next freeze committed sources and submit,
then yield without monitoring. Distributed vector/transpose/Taylor composition,
s18 field admission and full workflow remain pending.
