# SCALING-169 — completed

2026-11-22T13:48:12.455890+00:00; Codex.

79239 PASS:all32 normal/16 small sanitized outputs,8 zero-error/zero-leak GPU
summaries;790 source/78 evidence/100 cache files reconciled. Source372cf19,
HPCgnode08/gnode09,4workers sharing2H200s,4CPU/8GiB requested RAM.
Controller2min27.292s,cache16.531MiB,normal/sanitized maxRSS378.047/413.871MiB
per process,reported ownedGPU3.095MiB/worker(two/card;contexts/transients extra).
Preserved rawlogs/profiles/cache identities and[report](../experiments/distributed-four-worker/report.md).

README checked/updated;docs/preprocessing.md,current plan and primary
[remaining-work table](../docs/encoding-status.md#remaining-work-before-full-s18-preprocessing-scaling-169)
updated,including actual s18 preprocessing cell. Distinguish componentfour-worker
pass from genuine two-worker s4/s6 pipeline and unrunfulls18.
Audited pair-only discovery guard,two-owner merge,checkpoint control/manifest and
small native-reference bound. Remaining:multi-level root ownership;generalized
fullpipeline discovery/count/routing/Newton/verification;restart/publication at
chosen topology;representative scaling then full-resource admission/fullrun.
Native exactK and arithmetic already exist;not starting those from scratch.

Verification:independent analyze.py workers-01 succeeds;seals/routes/native/
schoolbook bytes/preservation/sanitizers all pass;gitdiff whitespace check passes
excluding unchanged rawlogs. Documentation/results only,no newcode/nativebuild
or GPU tests needed;no newjob launched for user'sremaining-work status request.
Next implementation is scalable rank-owned root reduction/fullpipeline integration.
