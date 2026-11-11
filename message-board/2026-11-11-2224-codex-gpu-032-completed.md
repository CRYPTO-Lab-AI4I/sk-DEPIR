# GPU-032 — Validation completed

Timestamp: 2026-11-11T22:24:24.907361+00:00
Agent: Codex
Status: completed

User confirmed all jobs finished. [Analysis](../experiments/gpu-transform-memory/runs/rows-01/analysis.json) reconciles333 source identities and all evidence/cache digests. All square/rectangular native/Go forward matches and inverse roundtrips, zero transpose time in direct layout, exact4N savings, closure/old-ABI/unsupported-operation checks and existing default/reuse online regressions pass. Memcheck0 errors/0 leaked bytes; no source correction needed.

At2^25 values working arrays256→128 MiB; owned269,233,796→135,016,068 bytes (same798,340-byte tables/plan). Forward medians0.583993519→0.596274239 s (+2.10%); inverse0.572088774→0.588209546 s (+2.82%). Meaningful larger tested cases show roughly2–8% penalties; tiny cases faster. Three fixed-order repeats, not performance bounds/fullonline evidence. Keep fast default when memory fits; direct remains explicit --layout direct/compare or capacity-driven. No policy exception or default switch.

Allocation2m14s, gnode14/H200 NVL driver580.126.20,16 CPU/16 GiB host/512 MiB device cap, exit0:0/no guards. Controller group RSS0.691 GiB versus Slurm0.31G reflects different metrics. [Report](../experiments/gpu-transform-memory/report.md), root/CUDA README and scaling plan updated; raw records/accounting/profiles/source snapshots and256 MiB fixture pair retained. GPU experiment did not inspect CPU job; both were separately analyzed only after this all-complete ping. Self-checks/doc paths/whitespace pass.

Transform-only evidence through2^25 values, not full2^32 encoding or s17 support. Next design gate is safe bounded canonical assembly/output that does not recreate full scratch; then code/spectrum chunking. Narrowed base-value compaction with bounded staging/disjoint regions is one unimplemented option, needing dependency/alias proof and conversion/transfer cost accounting. Extra time must be explicit before any slower end-to-end choice. No new job/budget/implementation/publication triggered by result review.
