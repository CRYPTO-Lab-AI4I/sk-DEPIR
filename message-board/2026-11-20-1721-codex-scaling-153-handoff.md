# SCALING-153 — handoff

2026-11-20 17:21 UTC; Codex.

78392 storage validation independently passed; full raw/review/profile evidence
and resident-vector integration committed as `4fe10dcd0be5e9155ded77f1bd951999c3c78739`. Main reference/README
updated with208normal/208sanitized storage results, explicit runtime/memory units,
and the new pending vector gate. See [storage report](../experiments/distributed-gpu-storage/report.md).

Submitted **78394**, [receipt/report](../experiments/distributed-resident-vectors/report.md):
results `experiments/distributed-resident-vectors/runs/resident-01/`,
log `experiments/distributed-resident-vectors/slurm-78394.log`,
cache `.cache/liftrs/distributed-resident-vectors/resident-01/`.
Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler; no pinned buffers.
Old host-tile regression plus resident normal/memcheck,24rank/pass rows per arm.
No scheduler/log/result monitoring after submission; wait for completion ping.

Local CPU/race regressions, native/C/CPP/CUDA/MPI builds, reviewer/selection tests,
Python/shell/gofmt and source/document whitespace checks pass. Raw78392 OpenMPI/
UCX/compiler logs contain trailing whitespace/blank EOF lines: preserved verbatim
for checksum evidence rather than normalized. No GPU/MPI run locally. README
checked/updated; no default, field-admission or resource increase.

Next after passed review: resident Taylor-window storage, complete resident FFT/
products, then matched larger scaling. Stillstep1; no fulls18 feasibility/runtime
claim. Continue obvious admitted next steps without another go-ahead.
