# SCALING-145 — handoff

UTC: 2026-11-20T11:06:56.967064+00:00

Submitted **78293**, source2a86336, vectors-03 split sanitizer retry.
[Report](../experiments/gpu-vector-stage/report.md),
[receipt](../experiments/gpu-vector-stage/submission-03.json).
Results/profiles/logs experiments/gpu-vector-stage/runs/vectors-03/ and
slurm-78293.log; retained .cache/liftrs/gpu-vector-stage/vectors-03/.
Same2HPC nodes/2H200/4CPU/8GiBhost,128MiB-owned/card,1GiB-RSS/process,
128MiBcache/2GiBscratch/32MiBfile,720s-controller/15min-scheduler.
Normal24cases/rank retained; sanitizer20small+onewide-tail-axis/rank with2wide
transforms.540s stepallowance; measured previous136s/transform predicts~5min
sanitizer plus~1min setup/normal/profile, not guaranteed ETA.
No monitoring/results inspection. README/reference/plan/CUDA/preprocessing/index
submission status current.4Python tests, AST/shell/whitespace pass; native CUDA
compile/new sanitizer outcome pending.78286 rawfailure and48normalpasses retained.
Wait for user completion ping; full distributed FFT and s18 remain pending.
