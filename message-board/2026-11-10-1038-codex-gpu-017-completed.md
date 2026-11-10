# GPU-017 — Resident convolution verified

Timestamp: 2026-11-10T10:38:47.873711+00:00
Agent: Codex
Status: completed

Resident convolution integration is complete in the optional CUDA experiment. Changes cover the private CUDA API and bridge, preprocessing dispatch, CPU-only dispatch tests, CUDA product tests and experiment harness. Public default execution remains CPU. Documentation updated in README.md, encoding/liftrs-shared/README.md and the scaling/s16 guides; README check completed.

[Report](../experiments/gpu-products/report.md) links scripts, exact sources and results. Job72839/products-02 passes 100 product checks, full matrix regression, old-library compatibility and CUDA memcheck (zero errors). Complete s10 preparation medians: transform-only GPU 2.735 s; resident at the same cutoff 2.400 s; resident with cutoff1024 1.528 s. Cutoff tuning and residency are distinct contributions. Job72838's test-only fallback failure and its correction remain preserved. Job72841/validation-05 passes repository Go/native suites, focused race/vet, legacy build and fresh-cache checks. CGO-disabled tests and current-source focused CUDA-tag tests pass. Python/shell syntax, documentation links and git diff --check pass.

Caches, profiles, raw logs and source snapshots retained. Measurements are on an H200 NVL, not a commodity-GPU performance claim. No full s16 run or production security claim. Changes are not yet committed or pushed.
