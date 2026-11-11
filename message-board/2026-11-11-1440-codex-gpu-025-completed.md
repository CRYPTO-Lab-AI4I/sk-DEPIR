# GPU-025 — Modular CUDA validation passed

Timestamp: 2026-11-11T14:40:52.157569+00:00
Agent: Codex
Status: completed

User completion ping authorizes results review of job73243. [Analysis](../experiments/cuda-encoding/runs/validation-01/analysis.json) passes312 source identities and sealed evidence/cache hashes. CPU-only, CGO-disabled, CUDA-only and combined suites pass; focused race/vet,80 historical matrix records, independent Go/CUDA small outputs and repeated/error/lifecycle checks pass. Memcheck reports zero errors/leaks. Complete s10 CPU/CUDA words and packed outputs agree across both repeats. CUDA encoding0.1040/0.1036 s versus CPU16 workers0.3101/0.3082 s; GPU setup0.8505 s, CPU0.002745 s. This is not a fulls16 speedup estimate.

Updated [validation report](../experiments/cuda-encoding/report.md), root/CUDA README and [scaling plan](../docs/encoding-scaling.md); checked relevant shared README/API contracts. CPU remains independent of GPU/preprocessing origin. No runtime correction required from this validation; production encoding file hashes gate the dependent fulls16 runner. Full-size GPU performance, cryptographic/application integration and TRL5 remain unverified. Source snapshots and profiles retained; perf instruction profiling remains unavailable on this host. Separate GPU-026 tracks the full-size run and its admission. No commit/push requested on this work item.
