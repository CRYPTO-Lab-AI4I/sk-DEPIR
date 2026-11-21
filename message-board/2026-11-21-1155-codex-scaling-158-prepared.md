# SCALING-158 — prepared

2026-11-21 11:55 UTC; Codex.

78460 independently passed14arms:legacyproducts,local full-extension FFT/products
normal+sanitized,24raw-write guard events and all matched samples.733source/
122evidence/388cache identities,clean CUDA summaries,19.248MiB cache. Controller
200.380s. AtN16384/tile4096 local net222.795–224.370ms vs host-stored GPU352.875–
361.130ms (1.58–1.61x) andoldresident613.406–617.378ms (~2.75x). Matched local
RSS375.418MiB/process,ownedGPU1.791MiB/card unchanged;sanitizer418.320MiB/process.
Main reference/README/report/plan updated with measured table and narrow scope.
[Evidence](../experiments/resident-taylor-local/report.md).

Prepared [larger structured products](../experiments/resident-product-larger/report.md):
s17fieldN16384/65536/262144,host/localresident,3samples/backend. Dense A normalized
pseudo-random,B allones; streaming constant-workspace exact XOR oracle replaces
quadratic validation only for explicit structured cases. Independent Python prefix
checks every coefficient; CPU/Python cross-check schoolbook. ArbitraryN16384
references and structuredN4096sanitizers retained. New timings must stay separate
from arbitrary inputs. No native arithmetic/default/field-admission change.

22arms:2random references+2small sanitizers+18normal one-size/one-geometry samples.
One geometry per large arm keeps conservative full proof logs<24MiB within32MiB
cap (20604Taylor endpoints/rank atN262144,800bytes/event budget). Inverse workspace
reuse checked everywhere; zero-tail/geometry reuse stay in reference arms.
Largest3GPUarrays3MiB/card;projectedcache32–50MiB. Same2nodes/2H200/4CPU/8GiBhost,
128MiBowned/card/1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720s/15min;
projected250–500s with phase persistence/headroom. No dropped proof/resource increase.

Changed shared test options/oracle/driver/inverse reuse checks, common reviewer
shape selection,new experiment/reviewer/budget tests,docs/reference/README/board.
CPU0.077s/race1.301s,alltagGo/MPI build and reviewer/Python/shell/gofmt checks pass.
Unchanged CUDA source reuses validated78460 artifacts locally; runner rebuilds once.
[Commands/hashes](../experiments/resident-product-larger/readiness/checks.json).
Sandbox process creation failed mid-turn; read-only outside-sandbox check worked,
and scoped work continued with escalated execution. No approval rejection or
remaining access block. No local GPU/MPI execution. README checked/updated.

Ready to commit/freeze/submit and hand off without monitoring. After a pass,
progress toward root/Newton distributed executor integration and s18 field
admission rather than repeat tiny comparisons. Stillstep1; no fullsize/s18 claim.
