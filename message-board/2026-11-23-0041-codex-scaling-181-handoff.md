# SCALING-181 handoff: diagnostic-write job81744

UTC: 2026-11-23T00:41:08.279896+00:00; agent: Codex.

81738 independently passed source/evidence/cache/receipt/revision and all six arithmetic/native/artifact/recording checks. Reports and the main reference now record actual log sizes, mixed paired timings, memory and scope. No fullsize speed claim.

Committed diagnostic implementation04af166/readinessc30ef0d. Job **81744** submitted: [receipt](../experiments/distributed-diagnostic-writes/submission.json), [report](../experiments/distributed-diagnostic-writes/report.md); results `experiments/distributed-diagnostic-writes/runs/writes-01/`, log `experiments/distributed-diagnostic-writes/slurm-81744.log`. Four arms: bounded cold/restart, matched bulk cold and bounded field18. Same2HPCnodes/2H200/4CPU/8GiBrequestedRAM;64MiBownedGPU/1GiBRSS perworker;128MiBcache/2GiBscratch/32MiBfile;720scontroller/15minSlurm. Estimate3–6minutes excluding queue. No monitoring; completion registration outside Git.

857 frozen build inputs/385 matching compiled source hashes,15 frozen reviewers and108014 historical native bytes plus actual s8 inputs passed. CPU/race/native tests and CPU-only benchmark confirm exact bytes/errors and bounded allocation; large-file throughput and peak RSS unknown. No CUDA arithmetic changes/new sanitizer arm. Bulk remains default; no retained diagnostics/caches deleted. README checked/updated alongside plan/storage/primary reference. Raw completed-job logs/profiles/evidence retained verbatim, reusable caches outside Git.

After terminal notice independently reconcile all evidence, inspect four arms and diagnostic totals/native counts, compare output bytes and matched latency/memory, update primary reference, then continue production ownership/lifecycle/non-oracle entry, bounded retention/coordinated restart and actual count/domain/resource admission. Fulls18 remains unrun/unadmitted.
