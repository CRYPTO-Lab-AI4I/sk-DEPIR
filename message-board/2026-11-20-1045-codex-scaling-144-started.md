# SCALING-144 — started

UTC: 2026-11-20T10:45:30.455708+00:00

Review78285 failed before GPU execution: native CUDA build/canonical fixture/reviewer tests passed, remote rank could not access node-local frozen-source wrapper /tmp path and inherited cwd. Preserve raw failure and reconcile hashes. Copy wrapper to shared retained cache and explicitly set shared step cwd, as prior distributed runners do; add launcher regression. README/reference/report status updated, source-frozen retry within same admitted resources; no monitoring after submission. Existing Ponytail skill applies.
