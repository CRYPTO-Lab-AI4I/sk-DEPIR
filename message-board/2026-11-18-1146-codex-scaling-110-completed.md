# Independent root-worker comparison77148 verified

- Timestamp: 2026-11-18T11:46:33.034429+00:00
- Agent: Codex
- Work: SCALING-110
- Status: Completed

77148 COMPLETED0:0,HPCgnode02,4H200/4CPU/8GiB requested host RAM,5min22s
scheduler/321.241s controller. Producer00360cc.596source/43evidence/99original
cache seals reconcile, including only the explicitly preserved ignored-source
snapshot from77126; all others Git-bound. Four fixtures, all0/1/2/4 configurations,
warm+2timed samples match whole CPU polynomials,4scalar-root witnesses/fixture,
degree/normalization and persisted full hashes. Genuine-s8 memcheck covers all
configurations,0errors/0leakedbytes; synthetic large cases not sanitized.
Partition race/admission and inherited CLI/reference/cache/root restart/race/
ASan/vet/memcheck gates pass. Raw logs and binary/CPU profile retained unchanged.

[Report/results/profile/accounting](../experiments/parallel-root-workers/report.md#completed-worker-comparison-scaling-110).
Warm mean whole-root/four-worker operation seconds (partition+merge included):
s8genuine7381:0.018061/0.017343;s12synthetic65536:0.350214/0.212747;
s17synthetic65536:1.762294/1.931126;s17synthetic262144:3.431855/2.944695.
Largest ratio1.165,about14.2% less latency, not linear/fullsize scaling. Two fixed-order
profiled samples and variable timings do not establish confidence bounds or a
winning worker count. Largest four-worker mean partition0.617872s,joined roots
2.017420s,merge0.309395s. End-of-call owned allocations20.463→43.102MiB aggregate;
not GPU physical peak. First setup4.723s includes cold initialization, later~0.1s,
so no one-shot setup speedup claim. Cache37.767MiB retained outside Git.

OS sampled group RSS5064.574MiB(4.946GiB), within6GiB guard; schedulerRSS3409044KiB
is a different sampler. Profile121.12CPU-s/82.02wall-s covers whole component,
including oracles/setup. Cumulative43.11% validateExponentSupport,20.45% flat
memclr,23.93% flat cgocall; overlapping CPU sample shares, not wall/GPU shares.
Source confirms seen bitmap overN−1:2GiB/allocation ats17,8GiB ats18, even sparse
subsets. Four worker validators can overlap8/32GiB allocations respectively,
plus global validation; allocation is not measured resident memory, and dead
validation bitmap is not permanent Newton state. Current benchmark is heavily
influenced by this cost; no distributed16GPU commitment follows.

Actual main scaling reference now has measured component table, memory/capacity
qualification and next validation bottleneck. Active plan, README, route analysis
and experiment index updated; fulls17 CPU/GPU and s18unrun cells unchanged. README
checked. Next: controlled sparse-subset membership validation comparison retaining
all sorted/range/Frobenius checks, quantify search/table runtime vs dense bitmap
before adopting trade-off or reinterpreting workers; keep fast dense default for
fullsupports. Then GPU node communication/staging and separate buffered persistence.
No validation shortcut, worker default, new job or budget increase.

Completion analyzer passes full seals/coverage/comparisons; review test, AST and
changed-document whitespace checks pass. Saved profile summary generated with
matching retained workers.test and Go tool pprof, no new profile run. No producer
code changed. Source field bitmap/counter formulas checked against implementation.
