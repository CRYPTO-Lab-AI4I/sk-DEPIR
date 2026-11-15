# SCALING-079 — Persistent workflow retry75342

- UTC: 2026-11-15T22:47:54.183254+00:00
- Agent: Codex
- Status: Handoff
- Source: `63d04426316280d93b748d35bb49f7e74265e532`

Submittedvalidation-02 job75342, pinnedgnode02. Same4H200/4GiB,4CPU/8GiBhost,
6GiBRSS,2GiBcache/4GiBscratch,600/900/1200s guards. Peer probe has60s limit;
TERM/KILL waits bounded5s each, with scheduler/driver cleanup explicitly separate.
[Submission](../experiments/gpu-preprocessing-workflow/retry-submission.json),
[report](../experiments/gpu-preprocessing-workflow/report.md), results
`experiments/gpu-preprocessing-workflow/runs/validation-02/`, log
`experiments/gpu-preprocessing-workflow/slurm-75342.log`, cache
`.cache/liftrs/gpu-preprocessing-workflow/validation-02/`.

sbatch returned75342 successfully. No scheduler queries, new-log reading, polling,
waiting or result analysis after submission. Await user completion ping. README
check/changes and source/verification recorded in the prepared handoff. Original
failure evidence/partial caches/profiles retained; node03 root cause unresolved.
No new fulls17 budget/admission or GPU/admin reset. No protocol/kernel/default
change. CPU setup investigation remains deferred by user agreement.
