# Spectrum export prototype ready for bounded GPU validation

- Timestamp: 2026-11-17T08:34:59.292932+00:00
- Agent: Codex
- Work: SCALING-096
- Status: Prepared

Expanded the [active plan](../docs/encoding-scaling.md#s18-preprocessing-memory-design-and-comparison-scaling-096)
with host lifetimes/export, full publication integration, host-backed tiled FFT,
split-product measurements, and a four-versus-sixteen-H200 decision. Updated
current s17 plan statuses from reviewed evidence and the scaling reference.

Implemented explicit `Preparation.SpectrumTo` with shared upload/forward GPU
arithmetic, bounded aligned host tile, optional begin/read plugin ABI, canonical
little-endian slots, synchronous writer errors and serialized ownership. Resident
output still uses the same shared transform; defaults and full limits unchanged.
Added independent scalar Horner/failure/reuse tests at s8/s17,1/2/4GPU peer/local,
a balanced32 MiB component comparison and CUDA memcheck through a reused guarded
controller. Host-lifetime audit distinguishes caller support, root/P aliases and
already-released product operands; no caller-owned arrays are silently discarded.

[Experiment report/scripts/model](../experiments/preprocessing-memory/report.md)
records exact bounds, methodology and limitations. Final-stage s18 host-array
projection889.446→377.447 GiB with1 MiB tile is not RSS or full feasibility;
discovery and products still need further work. Penalty remains unknown.

Verification: CUDA Go package and shared CPU/CUDA test package compile-only
checks passed (no GPU execution); memory-model self-check/reproduction, Python
AST, launcher shell syntax, existing controller policy check and
`git diff --check` passed. README, CUDA package README and experiment index updated.
CUDA C++ compilation/runtime/sanitizer checks remain for the submitted validation.
Budget stays within existing4H200/4GiB aggregate,4CPU/8GiBhost,2GiBretained/
4GiBscratch,20min scheduler bounds. Producer commit precedes submission so exact
source hashes can reconcile. No fulls18 job or s17 GPU retry is launched here.
