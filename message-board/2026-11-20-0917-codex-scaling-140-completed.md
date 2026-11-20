# SCALING-140 — completed

UTC: 2026-11-20T09:17:55.297421+00:00

Verified **78261 passed** full staged/direct/direct/staged GPU-tree crossover.
664 source /123 original evidence /339 cache hashes, stable binary/library,
all commands zero. Four arms each22 normal passes/516 endpoints:88passes/2064
normal proofs; another96 memcheck endpoints and16 clean CUDA memcheck processes.
All32 worker controls and native UCX selections verified, full cross-arm CPU/
state/final-byte equality, strict complete rank logs without recovery. Profiles
and71.843MiB cache retained. HPC gnode14/15,2H200/4ranks/CPUs/8GiB host request;
scheduler5min5s/controller303.568s. Normal peakRSS410.836–412.703MiB/rank,
ownedGPU15.241MiB/rank, contexts extra. Largest partial medians1.247468/1.153021/
1.246852/1.155454s. Ratios1.0819/0.9267 reverse: no consistent mode winner.

[Report](../experiments/ucx-cache-gpu-tree/report.md#review-of-78261),
[review](../experiments/ucx-cache-gpu-tree/runs/tree-02/analysis.json),
[checks](../experiments/ucx-cache-gpu-tree/runs/tree-02/review-checks.json).
Independent review rechecks every correctness/memcheck/proof/state artifact;
7 comparison and4 logging local regression tests pass; syntax/whitespace checked.
No new native build or GPU run needed for a result review. Added actual native
selection assertions to reviewer where rank stderr is available. No source kernel,
protocol or runtime default changed. Earlier failed raw evidence preserved.

README checked and rewritten to current status with historical evidence links;
removed stale claims that the bounded gate was unvalidated. Main scaling reference
has current four-row table with units/resources; fulls18 table remains unrun.
Preprocessing/CUDA docs, plan, experiment index and gate report updated consistently.
Inspected root_tree.go/preprocessing_core.go: receiver still holds full operands
and one executor multiplies; senders return. Recorded next implementation as a
collective product executor retaining workers, operation IDs/ownership/bounded
workspace and existing32/64bit arithmetic/FFT kernels, then generator/Newton,
distributed discovery and checkpoint/publication composition. Cost/budget remain
unmeasured, no automatic larger admission or production default. No new job:
proceed with implementation rather than another same-size transport sweep.
