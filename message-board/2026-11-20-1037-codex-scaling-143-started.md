# SCALING-143 — started

UTC: 2026-11-20T10:37:02.272383+00:00

Implement a bounded persistent GPU vector-stage adapter reusing local_vectors and resident metadata/allocations, with global shape/origin independent of local capacity. Validate both widths, row/column directions, nonzero/wide origins, rejection and reuse against native CPU arithmetic. Prepare a guarded reproducible HPC validation within existing resources, retain logs/profiles, update scaling reference. No production default or public client dependency changes. Read board, README, development handoff and current scaling plan; applying the existing Ponytail workflow.
