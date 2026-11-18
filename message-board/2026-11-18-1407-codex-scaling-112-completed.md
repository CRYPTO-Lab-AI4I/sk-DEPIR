# Matched subset validation77213 verified

- Timestamp: 2026-11-18T14:07:06.478051+00:00
- Agent: Codex
- Work: SCALING-112
- Status: Completed

77213 COMPLETED0:0,HPCgnode02,4H200/4CPU/8GiB requested host RAM,8min11s
scheduler/488.638s controller. Producer e061a08.601source/51evidence/164original
cache seals reconcile, including only the explicitly named exact ignored-source
snapshot; remaining source Git-bound.192normal rows(128timed/64warm),24sanitized
rows present, matching modes/reference strategy/order and complete CPU/scalar/
file hashes across all repeats. Local exhaustive65536small subsets, malformed,
wide memory/progress, both-mode root/partition and race gates pass. Both strategy
memchecks and inherited CLI memcheck0errors/0leakedbytes; large synthetic cases
not sanitized. Default bitmap native/reference/CLI/restart/race/ASan/vet pass.
No producer execution/default changed during review.

[Report/results/profiles/accounting](../experiments/root-subset-validation/report.md#completed-matched-comparison-scaling-112).
Four warm profiled samples/case/configuration, bitmap/search/search/bitmap process
order, same sorted CPU-oracle validation in both modes. Largest synthetic s17root
262144exponent case four-worker means3.427332s bitmap vs1.039159s sorted, partition
and merge included, observed ratio3.30. Sorted whole-root2.505141s/four-worker
1.039159s ratio2.41; not fulls17/s18 preprocessing or pure GPU speedup. Sorted
largest subtimes:partition0.062647s,joined roots0.667263s,merge0.309244s(~30%total).
Small s8/s12 cases show no general search latency benefit; s12four-worker~1.6%
search overhead with overlapping sample ranges. No density/field-only threshold
or universal mode inferred.

Matched sampled process RSS bitmap4561.707/4797.340MiB(4.455/4.685GiB) versus
sorted519.672/521.914MiB(0.5075/0.5097GiB),about89%less. Process includes CUDA/
runtime/fixture/reference/profile/checks, not fullsize client/preprocessing RAM.
Owned GPU largestfour-worker43.102MiB in both modes, not device physical peak.
Wholecomponent commands78.629/84.134s bitmap,46.465/46.619s sorted. Original four
profiles/binary retained; two representative top summaries from retained matching
subset.test show validator45.53CPU-s(39.41%)→1.32CPU-s(2%), memclr21.02→0.50flatCPU-s.
Wholeprofile samples cover setup/oracles; overlap and opaque cgo waits prevent
GPU/wall attribution. Cache60.852MiB outside Git retained; no prior artifacts deleted.

Recommend explicit sorted validation for the measured sparse worker subsets;
ordinary full-support/cache/discovery defaults remain bitmap. This improves both
time/memory here, not a slower automatic fallback or fullsupport adoption. Root
gate can now proceed to GPU node communication/staging/rank ownership, separately
buffered persistent writes. Fulls18 runtime, final GPU/node count and fullsize
host/transport/checkpoint integration remain unmeasured. No new job/budget change.

Updated actual main reference with latest matched table first, retained77148 as
historical link, active plan/next gate, README, preprocessing internals, route
comparison and experiment index. Fulls17 CPU/GPU and fulls18unrun cells unchanged.
README checked/updated. Completion analyzer,2 review tests,Python AST and git diff
--check pass. Saved pprof summaries are read-only analysis, not new profile runs.
Raw source/evidence/cache files remain bytewise unchanged.
