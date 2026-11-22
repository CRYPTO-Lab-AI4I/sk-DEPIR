# SCALING-168 — handoff

2026-11-22T13:36:25.799582+00:00; Codex.

Implementation139d75e,snapshot preflight372cf19;submitted **79239**,workers-01.
[Receipt](../experiments/distributed-four-worker/submission.json),
[report](../experiments/distributed-four-worker/report.md),
results `experiments/distributed-four-worker/runs/workers-01/`,
log `experiments/distributed-four-worker/slurm-79239.log`,
cache `.cache/liftrs/distributed-four-worker/workers-01/` outsideGit.
4MPIworkers sharing2H200 on2nodes,4CPU/8GiB requested RAM unchanged;
64MiBowned/worker=128MiB/card aggregate,1GiBRSS/worker,128MiBcache/
2GiBscratch/32MiBfile,720s/15min. Expected3–6min excluding queue.
No monitoring/status/log/result inspection. Snapshot790files/3,633,152bytes,
required paths and reviewer suite verified from actual committed source.

README checked/updated;primary encoding-status,plan,preprocessing/report show ID.
Prior79213 PASS and failed79202 records preserved. This validates new multi-peer
collectives,not four-GPU throughput or multi-level root/discovery integration.
After completion ping:analyze workers-01,reconcile seals/all4worker routes/outputs/
small sanitizer summaries,update mainreference;then continue rank-owned roots/
discovery integration. Fulls18 remains unrun.
