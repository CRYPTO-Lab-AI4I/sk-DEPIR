# SCALING-143 — handoff

UTC: 2026-11-20T10:43:59.977254+00:00

Submitted **78285**, source0711805, persistent GPU vector-stage validation.
[Report](../experiments/gpu-vector-stage/report.md),
[receipt](../experiments/gpu-vector-stage/submission.json).
Results/profiles: experiments/gpu-vector-stage/runs/vectors-01/;
job log: experiments/gpu-vector-stage/slurm-78285.log;
retained field/binary/output/profile cache: .cache/liftrs/gpu-vector-stage/vectors-01/.
Same2HPC nodes/2H200/4CPU/8GiBhost envelope,128MiB owned/card,1GiB RSS/process,
128MiB retained/2GiB scratch/32MiB file,720s controller/15min scheduler.
No monitoring/results inspection.24cases per node, repeated under full CUDA
memcheck, independent CPU oracle and retained everyword comparisons. Single
profiled diagnostic timings only; not a complete FFT/preprocessing result.
README/main reference/plan/preprocessing/CUDA docs/index submission state current.
Python reviewer2tests, AST, shell syntax, gofmt and whitespace checks passed.
Native compilation, fixture generation, GPU correctness/error/reuse and sanitizers
remain pending this job. Wait for user completion ping before review/composition.
