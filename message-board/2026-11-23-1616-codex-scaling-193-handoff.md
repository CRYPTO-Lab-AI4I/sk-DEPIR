# SCALING-193 handoff

UTC: 2026-11-23T16:16:26.120486+00:00. Agent: Codex. Status: handoff.

82102 verified passed, reports/primary reference updated. Implemented documented
parameterized MPI server command with strict policy, actual shared cold/restart
runtime, explicit memory/log modes, secure default sampling, mandatory profiles,
path/config agreement, periodic RSS/storage/deadline guards and exact-K storage/
shape admission before GPU discovery. Complete fullsize host/GC/native working-set
and topology admission still remain; current guards are not a complete model.

Submitted **82181** froma9fb7c2, implementation84df606. [Receipt](../experiments/distributed-command/submission.json),
[report](../experiments/distributed-command/report.md). Results
`experiments/distributed-command/runs/command-01/`; Slurm log
`experiments/distributed-command/slurm-82181.log`. Unchanged2HPCnodes/2H200/4cores/
8GiBrequested hostRAM/4workers;64MiBownedGPU/worker,1GiBRSS/worker,128MiBcache/
2GiBscratch/32MiBfile,720scontroller/15minSlurm,excludegnode03. Expected3–7min
excluding queue. Completion notification local/ignored; no agent polling.
README/preprocessing/current reference/plan/storage docs checked and updated to
submitted, not passed. CPU/race/config/tagged/native tests and26frozen reviewers
pass;931source inputs/412compiled code hashes match.

On terminal notice verify receipt/seals and pilot/actual-command cold/restart
proof/native work, checkpoint/source/publication/shard/final bytes, coordinator
support hashing and archive aliases. No command diagnostic support/root-stage/
identity copies are retained; measurement boundary change documented. Then finish
complete host-phase admission with real local counts and representative larger
rank/topology before fulls18 resource agreement. Fulls18 unrun/unadmitted.
