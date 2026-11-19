# Owned root-tree validation77704 submitted; no monitoring

- Timestamp: 2026-11-19T09:29:11.047983+00:00
- Agent: Codex
- Work: SCALING-121
- Status: Handoff

Submitted **77704**, [receipt](../experiments/root-merge-tree/submission.json),
[report](../experiments/root-merge-tree/report.md), [runner](../experiments/root-merge-tree/run.py).
Implementation6c7d28f. [Frozen build-input check](../experiments/root-merge-tree/readiness/snapshot.json)
verifies635files/3022873bytes and all11 readiness input hashes. Actual job records
its exact Git source at start; subsequent receipt/documentation commits only.

Results/profiles `experiments/root-merge-tree/runs/tree-01/`, progress
`experiments/root-merge-tree/slurm-77704.log`, cache
`.cache/liftrs/root-merge-tree/tree-01/`. Existing budget:2HPCnodes/2H200 total,
4logical ranks sharing2cards,1CPU/rank,2CPU/4GiB requested RAM per node,
64MiB owned/rank≤128MiB/card,1GiB RSS/rank,128MiB retained/2GiB scratch,
15min scheduler. Block placement/single:2 binding/task sharing. This tests unique
ownership and2 reduction levels, not4GPU throughput or fulls18 support.

Normal4samples incl warm-up/2timed/separate snapshot pass; small4rank memcheck
warm-up/snapshot. Chunk totalsuint64/per-operation bounded counts; explicit65536
normal/1024memcheck word chunks exercise width/tail counters. All source/shards/
transfers/owned states/final/CPU references and profiles retained. Coordinator
source/oracle remain resident; discovery/large-product worker service/FFT and
full generator/publication integration/admission remain future work.

Local CPU-only ownership/routing/partition and64bit extent tests pass, historical
four CPU root hashes match; tagged MPI/CUDA binary compiles.3tree/4distributed/
5staged/3shared reviewer regressions pass. Real MPI/GPU/memcheck/race results await
completion ping. README checked/updated with actual primary scaling reference,
plan, preprocessing/CUDA docs, experiment index/report and board.

No scheduler/process polling, log following or result inspection after submission.
Wait for user completion ping. No production default/format change, cache deletion,
extra resource allocation or fullsize performance claim.
