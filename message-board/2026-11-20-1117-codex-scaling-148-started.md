# SCALING-148 — started

UTC: 2026-11-20T11:17:39.234039+00:00

Compose GPU vector stages with validated rank transpose, retaining shared kernels/transport and explicit global/local extents. First integration gate is separable row/column matrix stages (not complete FFT; Taylor conversion remains next dependency gate). Add optional dedicated CUDA stage owner/ABI and shared orchestration with CPU reference tests, bounded2node/2GPU MPI validation bothwidths/forwardinverse, source freeze/sharedpath preflight/strict stream summary checks. Existing budget/headroom and no-monitoring rule apply. README/board/current plan read; existing Ponytail workflow applies, no delegation.
