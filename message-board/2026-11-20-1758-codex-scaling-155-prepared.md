# SCALING-155 — prepared

2026-11-20 17:58 UTC; Codex.

78396 independently passed:24normal/24sanitized resident FFT rows,24old complete
host-tile reference;240transpose+1788Taylor endpoints/arm;719source/41evidence/
148cache identities verified;clean CUDA summaries. Controller141.091s,
normal2.371s/sanitizer3.691s;RSS373.934/411.531MiB/process,1.618MiBowned/card,
16.077MiBretained. [Full evidence](../experiments/distributed-resident-fft/report.md).
Main scaling table, README, plan/design and API documentation updated.

Implemented [resident products](../experiments/distributed-resident-product/report.md):
shared operation spans/three-FFT product order, native existing pointwise kernel,
three slots left/right/scratch, right spectrum preserved through pointwise/inverse.
Field scans are shared with vectors; pointwise adds no device array. Inputs/reset/
upload/result readback and diagnostic right-spectrum capture are timed explicitly.
No copied FFT or arithmetic/default/field-admission change. Existing test driver
shares coefficients/schoolbook oracle/routes between host and resident modes.
Direct pointwise checks admitmin(ownedWords,capacity,17), verify whole zero tails,
alias/layout/range rejection, right/third-slot preservation and reuse.

Changed shared product schedule/adapter/failure tests, optional CUDA pointwise
API/validation, shared test driver/new helper, reproducible experiment/reviewer/
reference selector, docs/reports/README/board. CPU0.075s,finalrace1.301s;
native/HostTiled/C/CPP/CUDA/MPI builds and all reviewer/selection/Python/shell/
gofmt/source-doc whitespace pass. No local GPU/MPI execution. Exact commands,
source/artifact hashes and warnings in [readiness](../experiments/distributed-resident-product/readiness/checks.json).
README checked/updated for actual78396 measurements and new explicit API.

Ready to freeze source and submit old host-shard products +resident normal/
memcheck,24rank/product rows each,12direct checks/arm, s4d4/5/8 ands17d5/9/12,
independent schoolbook/base/tail checks, three-transform source/proof coverage,
right-spectrum preservation and stable allocation. Same2HPCnodes/2H200/4CPU/
8GiBhost,128MiBowned/card/1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,
720scontroller/15minscheduler. Hand off without monitoring. Larger matched
measurements and preprocessing integration remain next; stillstep1, no fulls18.
