# SCALING-164 — handoff

2026-11-22T09:13:48.111654+00:00; Codex.

Committed passed78909 evidence and restart/publication integration as `98695bc`.
Submitted **78915**, `restart-01`;
[receipt](../experiments/distributed-restart-pilot/submission.json),
[report](../experiments/distributed-restart-pilot/report.md),
[scheduler log](../experiments/distributed-restart-pilot/slurm-78915.log).
Results: `experiments/distributed-restart-pilot/runs/restart-01/`.

One allocation,2arms: cold genuine s4/s6 then fresh MPI-process prefix8 resume;
both stream canonical/generator artifacts and recheck complete identity/spectrum/
native bytes. New loading validates params/support/hash/field/padding/memory
before GPU, skips discovery/roots explicitly. Stream mode avoids full publication
arrays, with capped buffers and extra unifiedG output; defaults retained.

Same admitted2HPCnodes/2H200/4CPU/8GiB requested RAM,128MiBowned/card,1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720s/15min; estimate3–6min excluding queue.
Native CUDA unchanged from passed78909; no repeated kernel sanitizer. Focused
CPU/race/host-boundary tests, alltag build and native fixtures passed. Only
docs/receipt changed after source freeze. README/main reference/index/plan
updated; current bounded focus isstep4 recovery/publication.

No polling, live-log inspection or monitoring after submission. Await user
completion ping. Fixed rank count/fresh process is the tested target; arbitrary
node failure,elasticity,s18 parameters,larger rank/root levels,diagnostic global
array removal and representative scaling remain separate.
