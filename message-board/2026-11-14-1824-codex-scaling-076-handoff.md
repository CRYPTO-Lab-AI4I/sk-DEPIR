# SCALING-076 — Four-H200 calibration job74635 submitted

- UTC: 2026-11-14T18:24:18.699096+00:00
- Agent: Codex
- Status: Handoff
- Source: `fc159970eb1734888cbd1e83ebce6b150daf61ce`

Submitted exact approved calibration-01 on4H200 with4GiB aggregate owned cap,
1GiB/card,256MiB tiles,3repetitions,1/2/4GPU prefixes andpeer/local layouts.
4CPU/8GiBhost/6GiBRSS/4GiBGosoft,2GiBcache/4GiBscratch and600/900/1200s limits
unchanged. [Submission](../experiments/gpu-s17-preprocessing/calibration-submission.json),
[report](../experiments/gpu-s17-preprocessing/report.md). Results
`experiments/gpu-s17-preprocessing/runs/calibration-01/`, log
`experiments/gpu-s17-preprocessing/slurm-74635.log`, retained cache
`.cache/liftrs/gpu-s17-preprocessing/calibration-01/`.

Verification: run.py --check with exact approved flags and git diff --check pass;
sbatch returned jobID74635. README checked and status updated for approved
bounded calibration; maintained reference updated without new measurements.
No implementation/default or fulls17 admission changes. No polling, waiting,
following logs or result inspection; await completion ping. Existing profiles/
caches retained. CPU setup investigation remains deferred as agreed.
