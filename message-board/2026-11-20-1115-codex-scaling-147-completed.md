# SCALING-147 — completed

UTC: 2026-11-20T11:15:26.264555+00:00

78293 component validation independently passed:48normal/42split-sanitized cases,
all saved bytes/shapes/origins,680source/22originalevidence/368cache hashes. Both
sanitizer commands exit0 with one zero-error/zero-leak summary each, actually in
stdout. Rawcontrollerfailed preserved: stderr-only summaryassert. Strict paired
stream parser fixed, rejects missing/duplicate/nonzero summaries;5Python tests
pass. Normal profile reporting moved before later sanitizer work for future runs.
No CUDA/arithmetic/default changes or GPU rerun. Generated two native CPU reports
from retained normal profiles in <1s total; extra review commands/hashes retained.
HPCgnode06/07,2nodes/2H200/4CPU/8GiBrequestedhost, same admittedlimits. Normal31.210s,
sanitizer287.750s, controller357.989s before reporting failure;131.109MiBnormal /
167.059MiBsanitized hostRSS/process,4.569MiBowned/card;51.831MiBcache. Largest normal
wide vector forward1.812–1.836s/inverse1.742–1.765s singleprofiled/setup-excluded samples.
Split sanitizer42cases omits widezeroorigin/extrareuse calls (normalpass only).
[Report](../experiments/gpu-vector-stage/report.md),
[review](../experiments/gpu-vector-stage/runs/vectors-03/analysis.json),
[review commands](../experiments/gpu-vector-stage/runs/vectors-03/review-checks.json).
README checked/updated, mainreference/plan/preprocessing/CUDA/index current.
Python5tests/AST/shell/whitespace pass; no GPU/nativecode changed requiring new
run. No newjob launched. Next compose GPUvector/transposes/Taylor, independent
fullforward/inverse thenproducts/s18admission. Fulls18 stillunrun; no extrapolated
fullsize performance or largerbudget inferred from componentpass.
