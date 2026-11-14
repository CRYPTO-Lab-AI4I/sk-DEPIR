# SCALING-075 — Local-vector validation passes; repeated calibration prepared

- UTC: 2026-11-14T14:39:47.443215+00:00
- Agent: Codex
- Status: Handoff (awaiting resource agreement)

User ping authorizes job74468 review. Command
`python3 -B experiments/gpu-s17-preprocessing/analyze.py local-01` passes498
historical source/35 evidence/36 retained-cache identities. All small GPU
preparation/high-bit/address/window/peer-lifetime and Go/native/legacy CUDA checks,
vet and18 optimized/18 sanitized shared-recurrence cases pass. CUDA reports zero
errors/leaked bytes. Full hashes/inverses agree for both layouts on1/2GPUs.
[Analysis](../experiments/gpu-s17-preprocessing/runs/local-01/analysis.json),
[report](../experiments/gpu-s17-preprocessing/report.md#completed-local-vector-validation-scaling-075).

32MiB same-allocation warm/profiled/logged whole-call sample per case: peer1GPU
197.515/170.759ms forward/inverse; peer2GPU233.538/197.709ms; local1GPU
194.450/183.013ms; local2GPU189.407/178.511ms. Same65.57/67.14MiB owned GPU payload.
Local reduces2GPU latency18.9%/9.7% versuspeer, but2localGPUs gain only2.6%/2.5%
over1localGPU. One-card local inverse is7.2% slower. Single fixed-order samples
are not a fulls17 forecast or stable speedup. Peer default remains unchanged.

Updated README, preprocessing guide, scaling plan/reference and report. Preserve
profiles, caches and original failed/verified evidence. CPU setup profiling stays
deferred; no fulls17 cache/preparation/encoding or new numerical capacity claim.

Prepared runner/test/analyzer controls for256MiB tiles,3samples,1/2/4GPU prefixes,
both layouts. Largest one-GPU prefix admitted explicitly; sources/params/owned
memory/hashes/inverses/sample identities recorded and validated. Existing defaults
unchanged, prior evidence compatible. [Plan](../experiments/gpu-s17-preprocessing/calibration-plan.json)
requests4H200/4GiB aggregate owned cap (1GiB/card), same4CPU/8GiBhost/6GiBRSS,
2GiBcache/4GiBscratch and600/900/1200s limits. No submission before explicit
agreement; user asked through async tool, citing AGENTS.md budget rule.

Static validation: gofmt, Python AST/positive CLI and negative one-GPU-prefix
admission; updated analyzer still reconcileslocal-01; git diff --check pass.
[Check record](../experiments/gpu-s17-preprocessing/calibration-checks.json).
New larger/repeated Go/CUDA runtime validation remains pending; no extra long
local test or job is started. README check includes bounded results and limits.
