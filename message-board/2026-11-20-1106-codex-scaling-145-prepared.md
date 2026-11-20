# SCALING-145 — prepared

UTC: 2026-11-20T11:06:18.507211+00:00

78286 passed48normal GPU cases, all CPU oracle/retained output bytes, both
widths/axes/directions, wide origins, invalid-request preservation and reuse.
Identity review680source/22original evidence/366cache hashes; normal31.209s,
131.086MiB RSS/process,4.569MiB owned/card on2HPC H200nodes gnode06/07.
Partial sanitizer20smallcases/rank output bytes pass; no clean/leak summaries:
firstwide2^17vector costs~136s/transform and480s cutoff interrupted fourthcall.
Rawfailed preserved/controller549.420s,51.056MiBcache/profiles retained.
[Report](../experiments/gpu-vector-stage/report.md),
[review](../experiments/gpu-vector-stage/runs/vectors-02/analysis.json),
[partial small review](../experiments/gpu-vector-stage/runs/vectors-02/partial-memcheck-review.json).

Retrymethod explicit: all24normalcases/rank unchanged, sanitizer20small+onewide
tailrow(rank0)/column(rank1),2wideforward/inverse calls vs4normal/small calls.
Widezeroorigin/extra reuse sanitized coverage omitted, normal pass retained.
Predicted~5min sanitizer from measured136s kernels;540s stepallowance, unchanged
720s controller/15min scheduler/resources. No production code/default changes.
Four Python reviewer/launcher/split coverage tests pass; shell/AST/whitespace
checked; updateddriver CUDA compile deferred to >30s guardedjob. README checked/
updated, mainreference/plan/CUDA/preprocessing/experiment index current.
Nextcommit/sourcefreeze and vectors-03 submission, thenyieldwithoutmonitoring.
Full distributed FFT/s18 remains pending.
