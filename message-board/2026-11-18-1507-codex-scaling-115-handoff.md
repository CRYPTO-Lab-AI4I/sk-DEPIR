# GPU staged-exchange retry77281 submitted

- Timestamp: 2026-11-18T15:07:34.779371+00:00
- Agent: Codex
- Work: SCALING-115
- Status: Handoff

Fix/failed evidence committed6931f6c. Submitted **77281**,fresh staged-02,
replacing77241 failed source preparation. [Receipt](../experiments/gpu-staged-transport/submission-retry-02.json),
[report](../experiments/gpu-staged-transport/report.md),
[results](../experiments/gpu-staged-transport/runs/staged-02/),
[log](../experiments/gpu-staged-transport/slurm-77281.log).
Resource/kernel/timing budget unchanged:2nodes/2H200/4CPU/8GiBhost,128MiBowned/
card,64MiBpinned/2GiBRSS/rank,128MiBretained/2GiBscratch,15min scheduler.
No source/per-file/storage guard was widened; source preflight now precedes reads.

README, actual main reference, active plan and experiment/CUDA docs distinguish
failed77241(no compute) from pending77281. Prior raw data/scratch/caches retained.
README checked; verification and actual-snapshot builds in prepared message,
whitespace pass. No GPU/MPI runtime evidence yet. No polling/log follow/wait/results
inspection after submission. Await completion ping for retry before dependent work.
