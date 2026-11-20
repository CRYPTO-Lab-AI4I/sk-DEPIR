# SCALING-154 — prepared

2026-11-20 17:43 UTC; Codex.

78394 independently passed:24normal/24sanitized resident vector rows,24old
host-tile rows,240proof endpoints perarm,715source/40evidence/147cache identities,
zero CUDA errors/leaks,14.217MiB cache. Controller122.265s,normal2.026s/
sanitizer2.772s;RSS372.121/416.020MiB/process,1.617MiBowned/card. Full evidence
retained in [resident vector report](../experiments/distributed-resident-vectors/report.md).
Main reference/table, README, API docs and storage plan updated with completed
results/units and the next prepared gate.

[Complete resident FFT](../experiments/distributed-resident-fft/report.md) now
uses the same Taylor ticket/source/wave/round/level/pass scheduler, sparse window
kernel/launcher and FFT parameter/operation/stage order as the host reference.
Storage interfaces preserve borrowed host sends and use bounded GPU reads/sends.
Resident target commits go straight to GPU, including interleaved inverse views.
Two bounded host buffers +one capped GPU window (C192 uint64:3072hostbytes/rank,
1536GPUbytes/card) replace complete host source/target shards. Local Taylor
sources still take bounded D2H/H2D roundtrips; costs remain unknown. Third slot
preserved; no copied arithmetic/recurrence/default/field-admission change.

Changed shared Taylor storage/plans/order and error tests, resident adapters,
CUDA window API/launcher, shared GPU driver, experiment/reviewer/runner, reports/
reference/README/board. CPU regressions0.076s,race1.298s; native/plugin/HostTiled/
C/CPP/CUDA/MPI builds and all reviewer/selection checks pass. Commands/hashes and
warnings in [readiness](../experiments/distributed-resident-fft/readiness/checks.json).
No local MPI/GPU execution; initial local compile mistakes corrected before final
checks. README checked/updated for the optional API and actual measured status.

Ready to freeze source and submit old complete host-tile FFT reference +resident
normal/memcheck FFT,24rank/pass rows each, s4d4/5/8 ands17d5/9/12, independent
full native/scalar oracle, exact inverse, source routes/proofs/ownership and reuse.
Same2HPCnodes/2H200/4CPU/8GiBhost budget,128MiBowned/card/1GiBRSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minscheduler. No monitoring.
Resident products and larger matched scaling remain next after a pass; stillstep1,
not fulls18 preprocessing support.
