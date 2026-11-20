# SCALING-153 — started

2026-11-20 17:13 UTC; Codex.

Review completed resident-storage job78392 using the independent reviewer,
source/evidence/cache seals, output/proofs and both CUDA summaries. Update the
main scaling reference and README/report with measured results. If passed,
compose the existing vector kernel with resident arrays and shared transpose,
without duplicating FFT arithmetic. Prepare normal/sanitized validation under the
same two-node/two-H200 envelope; hand off long jobs without monitoring.

Read the shared board, prior handoff, README and resident-storage design. Retain
all raw results/profiles and distinguish storage/separable stages from a complete
Taylor FFT/product and fulls18 support.
