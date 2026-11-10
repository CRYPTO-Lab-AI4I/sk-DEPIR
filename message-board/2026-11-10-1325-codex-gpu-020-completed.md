# GPU-020 — larger bounded sweep verified

Timestamp: 2026-11-10T13:25:31.417296+00:00
Agent: Codex
Status: completed

Job72847/scaling-02 passed. [Report](../experiments/gpu-scaling/report.md) and [analysis](../experiments/gpu-scaling/runs/scaling-02/analysis.json) link runner, raw results, profiles and source/binary/cache provenance. Exact CPU/GPU forward/inverse/product arrays match at2^26,2^27,2^28. At2^28: GPU forward2.651 s vs16-core native CPU65.114 s (24.56x); GPU inverse2.533 s; resident product6.669 s. CPU product includes Go pointwise multiplication, so its ratio is separately qualified.100 small regression checks, old-library loading, full-domain metadata/budget rejection and small CUDA memcheck pass. Memcheck does not cover large arrays.279 source hashes and sealed evidence verified by python3 -B experiments/gpu-scaling/analyze.py scaling-02.

Explicit device peak3.001 GiB<4 GiB; sampled process RSS4.141 GiB<12 GiB; private cache8.145 MiB. No guard stop. Profiles, plans, binaries and old caches retained. Updated report, README performance scope, docs/encoding-scaling.md current gate and docs/encoding-s16.md measured limits. README check completed. Documentation links and git diff --check pass; no runtime code change or repeated suites needed for results reconciliation.

The bounded sweep is complete, not full s16 encoding. Four doublings remain to2^32. Review separate device/host/time budget and validation schedule before another increment; contiguous three-array payload at full size is48 GiB, only a model. No new job submitted. All current follow-on files/evidence staged for review; prior a997c4e remains the published checkpoint, with newer work not yet committed/pushed.
