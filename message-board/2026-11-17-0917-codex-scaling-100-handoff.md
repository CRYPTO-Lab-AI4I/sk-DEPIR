# Stream publication validation76710 submitted

- Timestamp: 2026-11-17T09:17:03.366038+00:00
- Agent: Codex
- Work: SCALING-100
- Status: Handoff

Implementation/recipe committed as1deb2ef. Submitted **76710**, freshexport-01,
existing4H200/1GiB percard,4CPU/8GiBhost/6GiBRSS,2GiBretained/4GiBscratch,
20min Slurm. No budget increase, fulls18 job or default change.
[Receipt](../experiments/preprocessing-export/submission.json),
[report](../experiments/preprocessing-export/report.md).
Results `experiments/preprocessing-export/runs/export-01/`;
log `experiments/preprocessing-export/slurm-76710.log`;
cache `.cache/liftrs/preprocessing-export/export-01/`.

Focused native/Go-only tests and CPU/CUDA compile checks pass; GPU publication,
regression and sanitizer checks remain pending this job. Active reference/plan,
README, API/CLI docs and prior primitive report updated to name the new step/job.
No polling/log analysis after submission, and no inspection of76706. On completion
ping run `python3 -B experiments/preprocessing-export/analyze.py export-01`,
reconcile source/evidence/cache, verify all stream/resident/independent bytes,
restart/readback/failure cases and memcheck, and report whole-workflow boundaries.
Then measure host-backed tiled transform and split-product costs per plan; do
not infer full-size saving/speed or admit s18 from these small checks.
