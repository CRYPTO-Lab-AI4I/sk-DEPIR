# SCALING-074 — Local-vector comparison prepared

- UTC: 2026-11-14T11:29:13.388386+00:00
- Agent: Codex
- Status: Handoff (validation submission follows)

Priority decision and reviewed22-case/peer results committed as728ddc6. Requested
HTTPS push failed because no credentials are available. Existing SSH port22
connection timed out; strict SSH443 failed missing trusted host key. No bypass
or remote change. User notified via asynchronous authentication request; work
continues independently. This is authentication/connectivity, not auto-review rejection.

Implemented optional local-vector scheduling in CUDA preprocessing, reusing the
same vector recurrence, peer transpose kernels and value/scratch buffers. Preserve
global row vector offsets, local coefficient interleaving and Taylor barriers.
Default remainspeer. Experimental runner `--vector-layout local` snapshots the
explicit environment selector at engine construction; tiny shapes retain peer
execution when too few vectors exist to partition. Document process-wide selector
and concurrency restrictions in the preprocessing guide. No API/ABI, canonical
representation or CPU/client build/default changes.

Changed paths: CUDA preparation.cu/preparation_kernels.cuh/transform.cuh,
shared preprocessing GPU tests; GPU experiment run.py/analyze.py/check.cpp,
static/host check evidence and report; preprocessing guide, scaling plan/reference,
README. New tile tests compare both schedules and every available GPU prefix
with independent sparse evaluations, full dense forward hash/inverse and equal
allocation-payload assertions. Scheduled local small preparation/high-width and
peer lifetime checks run under CUDA memcheck, plus Go/native regressions/vet and
18 shared-recurrence optimized/sanitizer host cases. Previous9-case evidence stays
compatible with analyzer. No old results are rewritten.

[Report](../experiments/gpu-s17-preprocessing/report.md#local-vector-comparison-scaling-074),
[checks](../experiments/gpu-s17-preprocessing/local-vector-checks.json).
Host18-case full independent Horner/inverse checks, CUDA shared O0 link, Python
AST/CLI, gofmt and git diff --check pass. GPU execution is pending; no speedup
claim. README checked and updated for the explicit comparison and its limits.

Trade-off: same2N/3N device payload and host/cache/storage,8 additional transposes
(4local/4gather-restore),16N word reads/writes (~2TiB/fulls17 FFT). Earlier1TiB
figure counted only local passes; actual peer-link traffic is a subset. Setup/
latency unknown; no slower mode adopted or default changed. The implementation
retains existing guards/profiles/cache persistence and durable progress logs.

Next submitlocal-01 on the already approved two-H200 caps, record jobID and yield
without monitoring. No fulls17 preparation admitted. Remaining gates and deferred
CPU setup regression investigation are explicitly recorded in the plan.
