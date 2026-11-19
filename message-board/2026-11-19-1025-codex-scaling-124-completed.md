# SCALING-124 — completed review

UTC: 2026-11-19T10:25:25.583209+00:00

Reviewed job 77782 after the user's completion ping. It failed after 2 min 8 s
on a support-index range check, largest case/sample 5/rank 3, before GPU work.
All 78 polynomial proofs matched (independently recomputed agreement); support
packets lacked proofs/capture. 52 completed measurement rows, no paired failure
snapshot; memcheck/profile summaries not reached. No guard failure.

Preserved raw evidence/cache; verified 640 source / 17 evidence / 50 cache seals.
[Report](../experiments/root-transfer-diagnostics/report.md),
[review](../experiments/root-transfer-diagnostics/runs/capture-01/analysis.json).
Updated README (checked), experiment index, preprocessing/CUDA docs and the main
[scaling reference](../docs/encoding-status.md) plus plan. Original failure and
full s=18 unrun status remain. Next explicit staging/all-packet diagnostic has
its own started message SCALING-125. No new job submitted during this review.
