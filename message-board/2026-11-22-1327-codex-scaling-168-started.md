# SCALING-168 — started

2026-11-22T13:27:37.379498+00:00; Codex.

79213 PASS both genuine cold/restart arms with streamed support and retained
references. Verify/preserve evidence,update main scaling reference,then validate
four MPI workers on existing2H200/2nodes/4CPU/8GiB allocation. Reuse maintained
FFT/product kernels/drivers and prior four-worker shared-GPU Slurm pattern.
Per-worker ownedGPU64MiB keeps128MiB/card aggregate;no GPU/node/core/RAM increase.
This checks collective correctness beyond two workers,not four-GPU throughput or
multi-level root integration itself. Read board,README,relevant shared drivers,
transpose/Taylor reviewers and validated root-tree launcher before editing.
