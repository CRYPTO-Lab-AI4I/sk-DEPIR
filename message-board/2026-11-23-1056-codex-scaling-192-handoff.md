# SCALING-192 handoff

UTC: 2026-11-23T10:56:53.736908+00:00. Agent: Codex. Status: handoff.

82101verified passed; source/evidence/cache seals, all17manifests/68shards,
76baseline files/native/proof/final bytes and fresh CPU loader match. Main reference
and reports updated. Implemented shared cold/restart arithmetic with coordinated
support verification and immutable archive hard-link reuse; all code documented.

Submitted **82102** fromed99872, implementation9006b3b. [Receipt](../experiments/distributed-workflow-resume/submission.json),
[report](../experiments/distributed-workflow-resume/report.md). Results
`experiments/distributed-workflow-resume/runs/resume-01/`, Slurm log
`experiments/distributed-workflow-resume/slurm-82102.log`.
Unchanged2HPCnodes/2H200/4cores/8GiBrequested hostRAM/4workers;64MiBownedGPU/
worker,1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm,
excludegnode03. Estimated3–6min excluding queue. Completion notifier local/ignored;
no model polling. README checked/current claim updated and primary scaling table
shows pending82102, not arithmetic-resume success.

On terminal notice verify receipt/seals, all four arms, remaining13Newton steps/
161367uploaded words, proof/native-work equivalence with pilot restart, source/
coordinator-hash agreement, all remaining checkpoints/identity/final bytes and
no-copy archive. Then parameterized command/complete phase and working-set admission,
representative larger topology and full s18 resources. No fullsize or node-failure
recovery validation claim; ordinary CPU client remains CUDA/MPI-free.
