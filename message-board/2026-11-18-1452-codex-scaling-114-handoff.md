# Approved GPU staged exchange77241 submitted

- Timestamp: 2026-11-18T14:52:15.223767+00:00
- Agent: Codex
- Work: SCALING-114
- Status: Handoff

User approved exact2nodes/2H200/15min resource plan. Submitted **77241** using
slurm/slurm/25.05.1H200+2CPU+4GiB host request/node,128MiB owned/card,64MiB pinned/
2GiBRSS per rank,128MiB retained/2GiB scratch,720s controller/900s scheduler.
Prepared executable sourcesac774a2; runner records actual producer revision and
frozen tracked Git source identities. No source/limit increase beyond approval.
[Receipt](../experiments/gpu-staged-transport/submission.json),
[report](../experiments/gpu-staged-transport/report.md),
[results](../experiments/gpu-staged-transport/runs/staged-01/),
[log](../experiments/gpu-staged-transport/slurm-77241.log).

Updated README, actual main reference, active plan, CUDA/experiment docs and
report to approved/submitted/results-pending. README checked. Runner --check,
sbatch bash -n and git diff --check pass; prior compilation/schema/review tests
remain applicable. No GPU/MPI runtime pass claimed. No fulls18, GPU-direct or
distributed-root/FFT support implied; buffered persistence remains separate.
No scheduler/process polling, log follow, waiting or result analysis after
submission. Await completion ping before review or dependent integration.
