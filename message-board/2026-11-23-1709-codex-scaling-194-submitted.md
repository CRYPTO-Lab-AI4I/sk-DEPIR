# SCALING-194 handoff: validation82194 submitted

UTC: 2026-11-23T17:09:56.023415+00:00. Agent: Codex. Status: handoff.

User approved the exact small eight-worker validation; job **82194** accepted from
0b63d72, implementationaff14a1. [Receipt](../experiments/distributed-host-admission/submission.json),
[report](../experiments/distributed-host-admission/report.md). Results
`experiments/distributed-host-admission/runs/host-01/`; Slurm log
`experiments/distributed-host-admission/slurm-82194.log`.
Resources:2HPCnodes/2H200,8cores/16GiBrequestedRAM total,8workers,
32MiBownedGPU/worker=128MiB/card,1GiBRSS/worker,768MiBdeclaredhostreserve,
zeroCUDA-pinned,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm,
excludegnode03. Expected3–7min excluding queue. Not a full s18 run.

README checked and current status updated; scaling reference/current table,
preprocessing/plan/storage docs reflect approved/submitted, not passed. Source/
readiness unchanged from939inputs/414compiled hashes/27frozen suites and CPU/native
checks. No new runtime check or GPU experiment before submission was necessary.
Local ignored completion notifier registered; no agent job polling.

After terminal notice verify receipt/seals,8worker output/checkpoint/shard/source/
publication/native-work/host-phase checks and coordinator hash/archive aliases,
then continue representative fullsize reserve/node packing/storage/time admission.
Named inventory+declared reserve is not measured fullsize peak RSS. Fulls18 remains
unrun/unadmitted.
