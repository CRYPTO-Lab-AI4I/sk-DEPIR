# SCALING-151 — completed

UTC: 2026-11-20T12:55:18.472044+00:00

78321 independently passed all24normal/24sanitized rank/product results,
12directfull-extension pointwise checks/mode andseparablebaseline. Every saved
coefficient matches schoolbook, base range/zero tails/source preservation valid,
720transpose+5364Taylor endpoints perproductmode andbothcleanCUDA summaries.
702source/42originalevidence/148cache hashes verified;15.473MiBcache/profiledata
retained. Producer0b6fd9c,HPCgnode16/17,2nodes/2H200/4cores/8GiBrequestedhost.
Controller141.151s, normal2.172s/sanitizer7.084s; hostRSS372.293/415.930MiB perprocess,
1.572MiBowned/card excludingcontext. Largest4096point normalproducts111.395–111.792ms,
sanitized2.022–2.069s,singleprofiled diagnostic samples;nofullsize runtimeinference.
[Report](../experiments/distributed-gpu-product/report.md),
[review](../experiments/distributed-gpu-product/runs/product-01/analysis.json),
[commands/hashes](../experiments/distributed-gpu-product/runs/product-01/review-checks.json).

Updated README/mainreferencevisibletable/preprocessing/CUDA/shared/index/plan:
stillstep1;smallFFTandproductgatespass,productionstorage/scalingpending. Added
[resident-storage design](../docs/distributed-preprocessing-storage.md): shared
storage accessors/schedules,3GPU arenas,coalesced row/interleaved pack/scatter,
boundedTaylor/transportworkspaces,chunkedupload/output andmatchedvalidationbefore
scaling. Proposed s18 1536GiB GPU arrays(~96GiB/card on16H200s),boundedhoststaging
instead of3072GiB hostcontrol model; otherworkflowstate andoutputI/O additional.
Runtime/setup/packing unmeasured, no resource admission/default change. Design only,
implementationpending;no extra larger host-reference calibration needed to show
storage problem. No new job launched, no source/native code changed.15Python
reviewtestsandwhitespace pass;no native/GPUrerun needed for documentation review.
README check recorded, allresultsevidence/profilelogs tracked; keepcachesoutsideGit.
