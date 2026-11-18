# Bounded distributed-root validation77285 submitted

- Timestamp: 2026-11-18T15:51:01.393089+00:00
- Agent: Codex
- Work: SCALING-117
- Status: Handoff

Submitted **77285** using [runner](../experiments/distributed-root/run.py) /
[scheduler script](../experiments/distributed-root/run.sbatch).
[Receipt](../experiments/distributed-root/submission.json),
[report](../experiments/distributed-root/report.md).
Implementation committed asf80a685. Actual immutable Git build snapshot readiness
verifies621files/2941820bytes and matches all10 checked implementation input hashes;
[check](../experiments/distributed-root/readiness/snapshot.json). No rebuild or
GPU/MPI execution claimed by this snapshot check.

Results/profiles `experiments/distributed-root/runs/roots-01/`; progress/log
`experiments/distributed-root/slurm-77285.log`; cache
`.cache/liftrs/distributed-root/roots-01/`. Same approved allocation:2HPCnodes,
1H200/2CPU/4GiB requested host RAM each,128MiB ownedGPU/card,2GiB RSS guard/rank,
64MiB staging ceiling (no new pinned buffer),128MiB cache/2GiB scratch,
15minute scheduler. No resource increase or new default/fullsize claim.

Updated report/current scaling reference/plan and README/CUDA/experiment index
with the submitted ID. README checked; ordinary CPU builds unchanged. Local
historical CPU-root digests, CPU-only partition tests, tagged compile and12Python
review regressions passed, as recorded in the prepared message/report.

User handoff: no scheduler/process polling, log following or result inspection.
Wait for completion ping before analyzing77285. Deferred validation is complete
cross-node GPU root/merge correctness/timing, both CUDA memchecks and retained
rank arrays/profiles. Fulls18 support, generator/spectrum, scalable sharding/
merge-tree and buffered persistence gates remain outstanding.
