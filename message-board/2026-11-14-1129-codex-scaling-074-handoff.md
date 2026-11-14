# SCALING-074 — Local-vector GPU validation submitted

- UTC: 2026-11-14T11:29:49.715640+00:00
- Agent: Codex
- Status: Handoff
- Revision: `62e626b3d89438909223dfea52e2eb7ccd96facb`

Submittedjob74468, `local-01 --devices2 --vector-layout local`, under unchanged
approved two-H200512MiB aggregate/256MiB per-device,4CPU/8GiB host,6GiB RSS,
2GiB cache/4GiB scratch and600/900/1200s limits. Compare peer/local1/2GPU tile
schedules; verify local small/wide/peer-lifetime behavior and sanitizers/regressions.
[Submission](../experiments/gpu-s17-preprocessing/local-submission.json),
[report](../experiments/gpu-s17-preprocessing/report.md). Results
`experiments/gpu-s17-preprocessing/runs/local-01/`, log
`experiments/gpu-s17-preprocessing/slurm-74468.log`, retainedcache
`.cache/liftrs/gpu-s17-preprocessing/local-01/`.

sbatch returned successfully. No state queries, new-log reads, waiting or result
inspection; await completion ping. README checked/updated in62e626b; submission
metadata adds no behavior change. No fulls17 run admitted. Publishing remains
blocked by GitHub authentication; user has been notified and asked to restore
credentials. Commits728ddc6/62e626b retain the work locally. No auto-review rejection.
