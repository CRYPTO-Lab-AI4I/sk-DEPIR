# SCALING-152 — prepared

2026-11-20 17:05 UTC; Codex.

Implemented the first [GPU-resident storage gate](../experiments/distributed-gpu-storage/report.md):
three reusable capped GPU shards, bounded device transfer tile, row/interleaved
local copy and peer packing/scattering. Host/CUDA adapters share the extracted
rank communication schedule, descriptors/proofs and failure propagation. Optional
CUDA API keeps ordinary CPU clients independent. No FFT/field recurrence copied.

Changed shared transpose and tests, optional CUDA API/bridge/executor and new
rank storage code, reproducible experiment/controller/reviewer, contributor rule,
README/API docs, storage design, plan and **main scaling reference**. The user’s
standing instruction to proceed on obvious next steps is recorded without waiving
budget admission or the long-job handoff rule.

Validation: CPU regressions0.076s, race1.311s; CUDA optional API/C bridge/native
plugin and full CPP/CUDA/MPI test binary compile.19 Python reviewer/launcher/
summary tests pass; Python/shell/gofmt/whitespace checks pass.
[Exact commands, hashes, warnings and retained artifacts](../experiments/distributed-gpu-storage/readiness/checks.json).
Compiler wrapper socket restrictions emitted warnings; final builds exit0.
Initial local compile errors were corrected before the final checks. No MPI/GPU
execution locally. README checked/updated for the new explicit experimental API
and pending gate; no production default or fulls18 claim changed.

Ready to commit/freeze source and submit normal+sanitized storage validation inside
the existing2nodes/2H200/4CPU/8GiBhost budget,128MiBowned/card,1GiBRSS/rank,
128MiBcache/2GiBscratch,720scontroller/15minscheduler.13geometries,104passes/rank
per execution, both widths/all layouts/empty owners/tails, bit-preserving global
oracle, source/third-slot preservation and no allocation growth. Then hand off
unmonitored. Resident vector/Taylor/product integration remains next after a pass.
