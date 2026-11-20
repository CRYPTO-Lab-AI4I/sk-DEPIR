# SCALING-156 — prepared

2026-11-20 18:53 UTC; Codex.

78399 independently passed24normal/24sanitized resident products,24old reference,
12direct checks/arm,720transpose+5364Taylor endpoints/arm,724source/41evidence/
148cache identities and clean CUDA summaries. Controller153.300s,normalarm3.121s/
sanitizer6.476s;RSS374.566/414.418MiB/process,1.618MiBowned/card,15.912MiBcache.
Largest4096-product samples0.2347–0.2367s resident versus0.1080–0.1111s host,
with asymmetric diagnostic capture. CPUprofile foreign calls78.69% is not GPU
kernel time. [Full evidence](../experiments/distributed-resident-product/report.md).
Main reference/README/plan updated, including a readable current-step table.

Prepared [matched comparison](../experiments/resident-product-scaling/report.md):
s17field syntheticN4096/16384,host/resident,127/4096-word transport,3samples/pair,
alternating backend order;12normal arms plus2N4096/tile4096 sanitizer arms. Same
coefficients/schoolbook/routes/checks, not fulls17 preprocessing. Fresh owners per
arm/dimension, lazy plan first shape and reused second shape; constructor timed
separately. Resident operand I/O uses explicit tile bound. Validation capture is
separately measured; per-rank total-minus-capture and independent maxima retain
all checks. Only small sanitizer subset runs; no claim for unsanitized largecase.

One source/build/preflight/allocation for all named arms; completed results/
profiles persist before later phases. Focused regression checks actual named
profile paths. Tile buffers1016→32768bytes (+about31KiB each);N16384 mainGPUarrays
192KiB/card; everything inside existing128MiBowned/1GiBRSS caps. Projected250–400s
controller inside720s/15min, unknown until measured. No admission/default change.

Changed shared test/controller/reviewer parameterization, new experiment/reviewer/
launch tests, documentation/README/reference and board. CPUrace1.298s,finalalltag
Go/MPIbuild and reviewer tests pass; unchanged CUDA source reuses validated78399
native artifacts locally, rebuilt once by admitted runner. Commands/hashes/
warnings in [readiness](../experiments/resident-product-scaling/readiness/checks.json).
No local GPU/MPI execution. README checked/updated; raw prior logs/profiles kept
verbatim. Ready to commit/freeze and submit, then hand off without monitoring.
Further scaling/integration depends on measured batching/setup/transfer cost;
fulls18 remains unrun, stillstep1.
