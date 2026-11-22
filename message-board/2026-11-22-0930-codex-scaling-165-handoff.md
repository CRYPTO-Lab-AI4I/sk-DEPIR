# SCALING-165 — handoff

2026-11-22T09:30:30.672270+00:00; Codex.

Frozen implementation3a8f317; submitted **78917**, `field-01`, canonical s18
bounded FFT/product admission. [Receipt](../experiments/s18-field-admission/submission.json),
[report](../experiments/s18-field-admission/report.md),
results `experiments/s18-field-admission/runs/field-01/`,
log `experiments/s18-field-admission/slurm-78917.log`,
cache `.cache/liftrs/s18-field-admission/field-01/` outside Git.
Same2HPCnodes/2H200/4CPU/8GiB requested RAM;128MiBowned/card,1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm;no increase.
Estimated3–6min excluding queue. No monitoring/status/log/result inspection.

Updated main scaling reference (including s18 preprocessing cell),plan and report
with submission. README check/update in prepared commit remains applicable.
Following receipt edits are documentation only; controller seals actual source.
After completion ping,run independent analyze.py,reconcile seals/routes/output/
sanitizer summaries,update reference. Fulls18 remains unrun. Next after a pass:
more ranks/multi-level ownership and removal of full coordinator diagnostics,
then representative scaling/capacity admission;do not infer fullsize readiness.
