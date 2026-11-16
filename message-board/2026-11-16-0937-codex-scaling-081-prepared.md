# SCALING-081 — Count-only admission and sizing prepared

- UTC: 2026-11-16T09:37:49.742988+00:00
- Agent: Codex
- Status: Handoff (bounded count submission follows)

Workflow result checkpoint40fdaaa committed; requested HTTPS push failed again
for absent session credentials. Work resumed after user interruption. No remote
publication claim; continue independent authorized admission work.

Added exact GoodMonomialCount to native/GPU bindings and shared CountCodeMonomials,
reusing the existing C APIs and locked count helper used by GoodMonomials. No
mathematical kernels or ABI/default changes. Private count constructor bypasses
only unused full host-grid admission; session cannot escape to data operations.
Normal NewCodePreprocessor admission preserved and tested. GPU counting requires
explicit resident devices, no CPU fallback. Native count/reference remains.

CLI --dimension-only publishes dimension provenance/profile with no support/root/
code arrays. Counts are not certified support or reusable code. Temporary device
bitmap is not copied/persisted; later full discovery re-enumerates, not skipped.
[Admission experiment](../experiments/s17-preprocessing-admission/report.md) wraps
the shared complete-workflow runner/analyzer with separate output/cache roots,
repeating original small/race/sanitizer checks and count-only memcheck before
s8/s16/s17 counts. s8 must match Go support, s16 must match918672421. Fulls17
count needs~2.5GiB modeled device state, no~29GiB host index list/128GiB grid.
Same4H200/4GiB,4CPU/8GiBhost/6GiBRSS,2GiBcache/4GiBscratch,600/900/1200s caps.

[Size calculator](../experiments/s17-preprocessing-admission/sizes.py) derives
retained12K+12N+61 bytes, ≤288GiB+61 forK≤N/2, without claiming peak RSS/runtime.
[Budget draft](../experiments/s17-preprocessing-admission/full-budget-draft.json)
is not approved/submitted: exclusive128core node/32workers,864GiB request/800GiB
RSS/720GiB Go soft,4GPU112GiB percard,512GiB retained/16GiB scratch,48h cutoff.
Exact K, heap audit, guards and numeric full agreement still needed. Cross-node
8/16GPU options retained but unimplemented. Metadata snapshot records configured
68GPUs, partition100h max, filesystem free space; none establishes health/quota/
reservation. Scaling reference corrects old67-card configuration snapshot.

Changed paths: native/CUDA/shared Go preparation bindings, CLI/tests/docs,
shared workflow controller/analyzer, admission wrappers/sizing/draft/local checks,
README and preprocessing/scaling guides/reference. README checked and updated.
[Local checks](../experiments/s17-preprocessing-admission/local-checks.json):
independent-Go count, known s10 dimension and ordinary host-cap rejection, full
native persistence/reference tests, CUDA-free tests, CUDA-tag compile/link only,
Python/CLI/calculator/controller tests, prior completed75342 analyzer compatibility,
gofmt/bash/diff checks pass. No GPU count execution or fulls17 latency claimed.
Profiles/cache retained; no long local job. Submit bounded admission count, record
ID then yield without monitoring until completion ping. No full-resource increase.
