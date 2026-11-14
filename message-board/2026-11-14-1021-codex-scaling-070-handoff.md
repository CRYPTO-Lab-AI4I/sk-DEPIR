# Two-H200 peer validation submitted

- Timestamp: 2026-11-14T10:21:42.425761+00:00
- Agent: Codex
- Work: SCALING-070
- Status: handoff

Submitted job74454, peer-01, source `d5d2a3b1b42f0fe1a5001e443d46f1b0a2a8d017`, with the user-approved two H200s,
256MiB/card/512MiB aggregate owned allocation cap,4CPU/8GiB host request,
6GiB RSS/4GiB Go soft limit,2GiB retained/4GiB scratch,600/900/1200s deadlines.
Existing runner overrides typed GRES to `gpu:h200:2`; no code/method/budget change.
[Submission](../experiments/gpu-s17-preprocessing/peer-submission.json) and
[report](../experiments/gpu-s17-preprocessing/report.md) record full commands.
Results/profiles `experiments/gpu-s17-preprocessing/runs/peer-01/`, Slurm log
`experiments/gpu-s17-preprocessing/slurm-74454.log`, retained cache/binaries
`.cache/liftrs/gpu-s17-preprocessing/peer-01/`. Compare one/two-card tile hashes
and timings inside the same allocation; complete small/wide correctness and
memory checks remain part of the existing runner.

No monitoring, job state/log reads or dependent analysis. Job74453 and other
unreviewed jobs remain untouched. README checked; no public capability or result
change before validation, so no README edit. Full s17 is not admitted. Next wait
for user completion ping and reconcile exact source/evidence/cache identities.
