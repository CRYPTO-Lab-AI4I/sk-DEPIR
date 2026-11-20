# SCALING-135 — prepared

UTC: 2026-11-20T08:19:01.675789+00:00

Keep observed UCX PML and same host replay, explicitly UCX_RCACHE_ENABLE=n; retain identical logging/proofs/payloads/allocation controls. Linked UCX1.18 supports option; systemucx_info1.20 is not used. Actual-node config and allworker flags checked. Host build/unit and31 Python tests pass, no local MPI run. Same2nodes/4cores/8GiB host/zeroGPUs and existing storage/deadline limits. Repeated registration may cost time/internal pinned memory; unknown, diagnostic-only. No default/cause/fullsize claim. [Report](../experiments/ucx-registration-cache/report.md), [checks](../experiments/ucx-registration-cache/readiness/checks.json). Freeze/submit then no monitoring.
