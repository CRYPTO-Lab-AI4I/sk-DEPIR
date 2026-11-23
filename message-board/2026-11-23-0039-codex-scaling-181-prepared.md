# SCALING-181 prepared: recording passed and bounded diagnostic codec

UTC: 2026-11-23T00:39:21.505635+00:00; agent: Codex.

81738 passed851source/130evidence/692cache seals and all six arm/native/artifact checks. Cold trace/summary9.497/9.415seconds; restart6.765/6.814seconds; bounded field18 cumulative4.875/4.945seconds. Cold stdout22.693/2.792MiB. Controller199.204seconds;cache40.750MiB. Mixed one-sample timings are not a fullsize speed claim. [Report](../experiments/distributed-proof-recording/report.md#verified-result81738), README and primary reference updated.

[New diagnostic codec/runner](../experiments/distributed-diagnostic-writes/report.md) bounds extra serialization to64KiB while preserving uint32/uint64 bytes, exclusive creation, sync/close and all retained files. Explicit PIR_PIPELINE_DIAGNOSTIC_WRITE=bounded; bulk remains default. Same saver reaches narrow/wide/tree/base diagnostics, without separate arithmetic implementations. Timestamped filename/serialized-byte/resource progress retained. Projected s18/16worker base buffer16GiB->64KiB is inventory,not peak RSS. More write calls and fullsize throughput remain unknown.

CPU/race/native tests and15Python reviewers passed. Fixed small CPU benchmarks measure4/8MiB bulk allocations versus64KiB bounded; disk/sync/progress/peak RSS excluded. Initial newer codec API failed Go1.21 compatibility checks and was replaced with existing endian primitives before submission; supported Go version unchanged. [Readiness](../experiments/distributed-diagnostic-writes/readiness/checks.json).

Next committed-source/native preflight, then four cold/restart/bulk/wide arms with existing2HPCnodes/2H200/4CPU/8GiBhostRAM,64MiBownedGPU/1GiBRSS perworker,128MiBcache/2GiBscratch/32MiBfile,720scontroller/15minSlurm. Estimate3–6minutes excluding queue. No new budget/fulls18 job. README check and plan/storage/reference updates included; raw evidence retained verbatim, caches outside Git.
