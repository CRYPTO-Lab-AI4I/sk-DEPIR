# SCALING-156 — handoff

2026-11-20 18:55 UTC; Codex.

78399 passed; raw/review/profile evidence and matched scaling experiment committed
as `e76ef8aac43f94ab8fe16a29c06c7f333f2bef5b`. Main scaling reference/README show completed resident products and
actual memory/runtime units; resident small-product slowdown remains explicit.
[78399 evidence](../experiments/distributed-resident-product/report.md).

Submitted **78432**, [report/receipt](../experiments/resident-product-scaling/report.md):
results `experiments/resident-product-scaling/runs/matched-01/`,
log `experiments/resident-product-scaling/slurm-78432.log`,
cache `.cache/liftrs/resident-product-scaling/matched-01/`.
12normal arms: s17fieldN4096/16384,host/resident,tile127/4096,3samples/pair with
backend order alternated;2sanitizer arms:N4096/tile4096 only. Same coefficients/
schoolbook/proof checks, total/net/capture timing, separate constructor setup,
right-spectrum preservation and stable allocation. No fulls17 preprocessing claim.

Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler; no pinned buffers.
Buffer growth1016→32768bytes is within existing caps; no admission/default change.
Completed phases/profiles saved before later arms. No scheduler/log/result
monitoring after submission; wait for completion ping.

Local alltag build, CPUrace1.298s, reviewer/named-profile-path regression/Python/
shell/gofmt/source-doc whitespace pass. Unchanged CUDA source reuses validated
78399 artifacts locally; admitted job rebuilds once. Commands/hashes/warnings in
readiness. Raw prior logs/profile whitespace/checksums retained. README checked/
updated; no local GPU/MPI execution. Next depends on matched measured batching/
setup/transfer costs; larger scaling and complete preprocessing integration remain.
Stillstep1; no fulls18 throughput/support claim. Continue obvious admitted work.
