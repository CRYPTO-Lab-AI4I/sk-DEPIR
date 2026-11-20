# SCALING-155 — handoff

2026-11-20 18:00 UTC; Codex.

78396 passed and full evidence/resident product implementation committed as
`a96d7f64960a2c9a437850e31c2c159ae87a8f31`. Main scaling reference/README report measured complete resident FFT
results with units and current pending product gate.
[FFT evidence](../experiments/distributed-resident-fft/report.md).

Submitted **78399**, [report/receipt](../experiments/distributed-resident-product/report.md):
results `experiments/distributed-resident-product/runs/product-01/`,
log `experiments/distributed-resident-product/slurm-78399.log`,
cache `.cache/liftrs/distributed-resident-product/product-01/`.
Old host-shard products +resident normal/memcheck;24rank/product rows and12direct
pointwise checks perarm, s4d4/5/8 ands17d5/9/12. Schoolbook/base/tail oracle,
three-transform source/transpose proofs, invalid/alias/layout nonmutation,
right/third-slot preservation, right spectrum through inverse, stable allocations.
No scheduler/log/result inspection after submission; wait for completion ping.

Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler; no pinned buffers.
Local CPU/race1.301s/native/HostTiled/C/CPP/CUDA/MPI builds/reviewer/selection/
Python/shell/gofmt/source-doc whitespace pass. Raw prior UCX/OpenMPI/compiler
logs retain original whitespace/checksums. No local GPU/MPI execution. README
checked/updated for explicit API and actual scaling results. No default, field
admission or resource increase. Next if passed: matched larger measurements
before root/Newton/generator/discovery/persistence integration. Stillstep1; no
fulls18 support/throughput claim. Continue obvious admitted work automatically.
