# SCALING-072 — Bounded retry and artifact review submitted

- UTC: 2026-11-14T10:58:10.268529+00:00
- Agent: Codex
- Status: Handoff
- Source: `10a362b49c5d178daea6de66d960c77529f6addf`

GPU retry job74466: two H200s, same approved512MiB aggregate/256MiB per-card
payload,4CPU/8GiB host,6GiB RSS,2GiB cache/4GiB scratch and600/900/1200s limits.
[Submission](../experiments/gpu-s17-preprocessing/peer-retry-submission.json),
[report](../experiments/gpu-s17-preprocessing/report.md), results
`experiments/gpu-s17-preprocessing/runs/peer-02/`, log
`experiments/gpu-s17-preprocessing/slurm-74466.log`.

Independent CPU-only artifact review job74467:4CPU/8GiB,6GiB RSS,2GiB new output,
noGPU,600/900/1200s limits. [Submission](../experiments/encoding-table/review-submission.json),
[report](../experiments/encoding-table/report.md), results
`experiments/encoding-table/reviews/review-01/`, log
`experiments/encoding-table/reconcile-slurm-74467.log`. Existing caches read-only;
no full encoding/preprocessing rerun. It verifies all22 selected completed cases.

Both sbatch submissions returned successfully. No scheduler query, log following
or new-result inspection. Await user completion ping before analysis. README
check recorded in the prepared handoff; these submission IDs do not change APIs,
performance claims or supported parameters. No fulls17 admission or GPU-count ceiling.
