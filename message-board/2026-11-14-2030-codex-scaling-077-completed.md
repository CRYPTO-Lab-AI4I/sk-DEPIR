# SCALING-077 — Repeated four-GPU calibration reconciled

- UTC: 2026-11-14T20:30:00.587242+00:00
- Agent: Codex
- Status: Completed

Job74635 passes after user completion ping;498 source/35 evidence/36 cache
identities reconcile. All18 tile hashes/inverses, small complete four-card
preparation/high-bit/window/address tests, peer lifetime, Go/native/legacy CUDA
regressions, vet and18 optimized/18 sanitized recurrence checks pass. CUDA reports
zero errors/leaked bytes. [Analysis](../experiments/gpu-s17-preprocessing/runs/calibration-01/analysis.json),
[report](../experiments/gpu-s17-preprocessing/report.md#completed-four-gpu-calibration-scaling-077).
Command: python3 -B experiments/gpu-s17-preprocessing/analyze.py calibration-01.

256MiB tiles,3samples/layout/count, same gnode02NV6 allocation; forward/inverse
medians in seconds: peer1=0.865674/0.754588, peer2=0.995857/0.884177,
peer4=1.088009/0.978584; local1=0.855847/0.830580,
local2=0.915740/0.890322, local4=0.778385/0.753873. Equal owned allocations per
layout:513.57/515.14/518.28MiB aggregate at1/2/4cards; context overhead additional.
Sampled entireGPUworkflow hostRSS1.615GiB; whole controller~491s including builds.
Four-local versus four-peer improves28.5%/23.0%, but versus one-peer forward10.1%
and inverse<0.1%; sum of medians5.4%. No end-to-end preprocessing speedup claimed.

Analyzer now reproducibly reports medians/ranges by layout/count while retaining
raw samples and old-result compatibility. Independent assertions verify all18
samples and six summary rows, Python AST and git diff --check pass. No math,
API/default, layout, memory or experiment budget changes. No new job submitted.
README checked/updated; report, preprocessing guide and scaling reference/plan
updated with explicit units, boundaries, exact source/evidence links and limits.
Profiles/caches retained. Fulls17 remains unrun/unadmitted; CPU setup review deferred.

Repeated count-calibration is complete. Next persistent full-workflow runner and
its bounded validation/admission: discovery/support publication, support-bound
root save/reload, complete identities, streaming final cache/readback and restart/
resource/source accounting. Do not infer full-size scaling from small grids;
source geometry, including logging-wave bounds, is explained as inference.
