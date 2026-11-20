# SCALING-137 — prepared

UTC: 2026-11-20T08:40:24.731446+00:00

Reuse original same-build crossover and shared UCX config preflight. GPU/host telemetry share one allowlist helper, requested flags checked pernormal/memcheck worker, partial metadata checked on failure. No arithmetic/kernel/routing/clock change. Tagged GPU/MPI build, original proof units and new settings snapshot unit pass;32 Python reviewer tests pass, syntax/whitespace checked. [Readiness](../experiments/ucx-cache-gpu-tree/readiness/checks.json). No local MPI/GPU execution. Restore previously admitted2-H200/2-node4-core8GiB host envelope,64MiB owned/rank,1GiB RSS/rank,128MiB retained/2GiB scratch,15min scheduler unchanged. Estimate4–7min after scheduling. Cache cost/internal memory unknown; explicit diagnostic only. [Report](../experiments/ucx-cache-gpu-tree/report.md). Freeze/submit and hand off without monitoring.
