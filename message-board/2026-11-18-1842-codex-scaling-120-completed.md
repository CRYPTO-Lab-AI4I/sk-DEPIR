# Buffered persistence77352 passes; no default change

- Timestamp: 2026-11-18T18:42:06.630374+00:00
- Agent: Codex
- Work: SCALING-120
- Status: Completed

Reviewed on user completion ping. [Report](../experiments/preprocessing-persistence/report.md),
[analysis](../experiments/preprocessing-persistence/runs/io-01/analysis.json),
[raw results](../experiments/preprocessing-persistence/runs/io-01/results.json),
[accounting](../experiments/preprocessing-persistence/runs/io-01/accounting.txt).
77352 COMPLETED0:0,1m45s scheduler/103.350s controller, sourcef493cb9,
HPCgnode02/Xeon6710E/2CPU/4GiB requested RAM/0GPU/NFS. All11commands exit0,
no guards.628source/14evidence/106cache seals verify without overrides;12normal
and12small-race cases reconcile. Complete publication/readback, byte/hash/tail,
producer-error/no-publish/retry/corruption, real small CLI recovery/restart/Go
reference and full concurrent race probe pass.

Fixed96MiB/trial,4files: payload application calls1536→24 with4MiB coalescing.
Publication medians direct1/2/4writers0.336030/0.268352/0.198815s versus buffered
0.346113/0.212306/0.184447s. Full readback medians0.895899–1.459786s dominate;
individual trials0.862630–1.8885s approximately. Complete elapsed benefit mixed,
only2samples/configuration, cached small files/profile/counter effects included;
no buffer/writer default or fullsize throughput choice justified.4MiB buffers add
up to16MiB/4writers and copies, no transforms/network/payload/verification pass.

Persistence group RSS41.137MiB, event process high-water40.418MiB;
final retained payload residency96MiB, sampled RSS+retained-cache proxy109.609MiB
(not total memory). Persistence max sampled cgroup current362.695MiB;
whole job sampled706.316MiB/high-water memory.peak718.566MiB includes compilation.
Fast diagnostics' zero sampledRSS means missed observation, not zero actual use.
Cache111.946MiB,106files at `.cache/liftrs/preprocessing-persistence/io-01/`:
all96normal/race manifests, normal96MiB and race512KiB final payloads,2binaries.
CPU profile/raw logs preserved bytewise; only prior cheap verified synthetic probe
payloads pruned as planned, no previous preprocessing/profile/cache lost.

Updated README/preprocessing docs, main scaling reference/table, scaling plan,
experiment index/report and prior root report follow-up. README check: hardware,
measured pending→passed status and limited conclusion updated, existing defaults,
ordinary CPU builds/fulls17 passed and fulls18-unrun tables unchanged. No code,
new job or budget increase. Completion analyzer/seals/arrays verify;3persistence,
5staged and3shared reviewer regressions pass. Python/shell syntax, local doc links
and edited-text whitespace pass; no new GPU run/full code test suite for this
results/documentation review.

Next implementation: scalable support/rank ownership and multi-level GPU merge,
then generator/Newton/transforms and rank-owned atomic/checksummed publication.
Investigate coalesced verification reads while preserving every check; no measured
read-buffer speedup claimed. Do not replace integration with another same-size
I/O sweep. Larger/fullsize/multi-node I/O and GPU allocations need separate
resource admission; fulls18 memory/runtime still unvalidated.
