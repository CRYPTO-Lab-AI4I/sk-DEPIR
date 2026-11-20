# SCALING-154 — handoff

2026-11-20 17:44 UTC; Codex.

78394 passed and evidence/complete resident Taylor/FFT integration committed as
`2ec5458719774b1cf01a4a4ed4de4d9ea0107a3c`. Main scaling reference and README report actual vector results/units,
not pending/not-run. [Resident vector evidence](../experiments/distributed-resident-vectors/report.md).

Submitted **78396**, [report/receipt](../experiments/distributed-resident-fft/report.md):
results `experiments/distributed-resident-fft/runs/fft-01/`,
log `experiments/distributed-resident-fft/slurm-78396.log`,
cache `.cache/liftrs/distributed-resident-fft/fft-01/`.
Old complete host-tile FFT reference +resident normal/memcheck;24rank/pass rows
perarm, s4d4/5/8 ands17d5/9/12. Native full FFT/scalar oracle, exact inverse,
Taylor source/transpose proof coverage, third-slot/rejected-value preservation,
stable device allocation. No scheduler/log/result inspection after submission.

Same2HPCnodes/2H200/4CPU/8GiBhost,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler; no pinned buffers.
Local CPU/race/native/HostTiled/C/CPP/CUDA/MPI builds, reviewers/selection/gofmt/
shell/Python and source/doc whitespace pass. Raw prior OpenMPI/UCX logs retain
original whitespace for checksum provenance. No MPI/GPU executed locally. README
check/update recorded; no production default, field admission or resource increase.

Wait for completion ping. Next if passed: resident pointwise/products and matched
larger measurements, before root/Newton/generator/discovery/persistence integration.
Stillstep1; no fulls18 claim. Continue obvious admitted steps automatically.
