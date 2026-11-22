# SCALING-170 — prepared

2026-11-22T14:28:48.301323+00:00; Codex.

Implemented [hierarchical sharded roots](../experiments/distributed-root-hierarchy/report.md).
Shared power-of-two group stage plan2..64 and domain admission,reused old owner
routing through generalized child-shard intervals,no intermediate gather. Explicit
MPI Split/Counts handles own communicators/staging;legacy codecs/tags shared,
children close before parents,errors abort original world. Native chunks identify
global/group rank andsize. GPUcomposition reuses maintained product driver and
verified-result hook,not new arithmetic. Old paths/defaults remain.

Normal16/4092 synthetic s18 roots,groups2+2 then4,unequal4096/2048 child domains;
small16-root CUDA memcheck in one allocation. Independent native/scalar leaves/
subtrees/final outputs and actual source-byte/routing proofs retained. Physical
budget unchanged2nodes/2H200/4CPU/8GiBhost,4workers,64MiBowned each=128MiB/card,
1GiBRSS/worker,128MiBcache/2GiBscratch/32MiBfile,720s/15min. Extra per-stage
setup/host transfers and unknownruntime documented;no selected slower default.
Expected3–6min excluding queue,not fulls18time.

[Readiness](../experiments/distributed-root-hierarchy/readiness/checks.json):
CPU/race oracle2/4/8worker routing pluslegacyowner/tree tests pass;MPI closed-handle
admission and alltagsbuild pass;actualfield18 native counts/domains/rootsetup pass
0.08s aftercatching/fixingnil fixture phasehook locally. FocusedPythonhierarchy,
oldproduct/fourworker reviewers andSlurm syntax/check pass. No localMPI/GPUinit,
no unchanged CUDA recompilation/full historicalroot replay. Integrated job pending.
README checked/updated;primary encoding-status table/remaining checklist,
preprocessing doc,scaling plan and report updated. Fulls18 remains unrun.
Freeze source,verify actual Git-blob snapshot/reviewer paths,submit hierarchy-01,
recordID andyield without monitoring.
