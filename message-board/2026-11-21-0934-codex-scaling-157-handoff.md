# SCALING-157 — handoff

2026-11-21 09:34 UTC; Codex.

78432 passed; full raw/review/profile evidence and optional local Taylor accessor
committed as `d327890fcca5da069ad0abf038e47a7cf16ae843`. Main scaling reference/README/report now show measured
matched table, resident slowdown and the explicit next optional gate. Both
comparison paths use GPU arithmetic; host means main arrays stay in host RAM.

Submitted **78460**, [report/receipt](../experiments/resident-taylor-local/report.md):
results `experiments/resident-taylor-local/runs/local-01/`,
log `experiments/resident-taylor-local/slurm-78460.log`,
cache `.cache/liftrs/resident-taylor-local/local-01/`.
14arms:legacyproducts,local full-extension FFTnormal/memcheck,localproductsnormal/
memcheck/cache guards,both widths/layouts,then9matched host/oldresident/localresident
samples atN4096/16384,tile4096. Guard upload/scatter/local-copy invalidation stays
out of benchmark setup. Shared source routes/proofs/recurrence and independent
oracles retained. New mode explicit; no default/field-admission change.

Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler;no pinned buffers.
No extra principal array; conditional canonical scans/local address costs explicit,
net benefit unknown. Completed phases/profiles persist before later arms. No
scheduler/log/result inspection after submission; wait for completion ping.

Local CPU0.076s/race1.299s/native/HostTiled/C/CPP/CUDA/MPI builds/reviewers/Python/
shell/gofmt/source-doc whitespace pass. Raw prior OpenMPI/UCX logs retained verbatim.
No local GPU/MPI execution. README checked/updated. Next depends on validated
local-access benefit before larger products/pipeline integration. Stillstep1;
no fullsize/s18 performance/support claim. Continue obvious admitted work.
