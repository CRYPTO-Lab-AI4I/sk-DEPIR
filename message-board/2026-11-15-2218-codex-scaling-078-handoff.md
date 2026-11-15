# SCALING-078 — Persistent workflow validation job75194

- UTC: 2026-11-15T22:18:08.838397+00:00
- Agent: Codex
- Status: Handoff
- Source: `ec8b9dc1b5dea7106951dddf13bca7273b421535`

Submitted actual CLI workflow validation75194 under unchanged agreed4H200/4GiB,
4CPU/8GiBhost/6GiBRSS,2GiBcache/4GiBscratch and600/900/1200s guards. Fixeds8/s4
cases validate fresh preparation, final/root restart, independent Go bytes,
CUDA-free readback and race/ASan/vet/CUDA memcheck. Profiles/reusable caches retained.
[Submission](../experiments/gpu-preprocessing-workflow/submission.json),
[report](../experiments/gpu-preprocessing-workflow/report.md), results
`experiments/gpu-preprocessing-workflow/runs/validation-01/`, log
`experiments/gpu-preprocessing-workflow/slurm-75194.log`, cache
`.cache/liftrs/gpu-preprocessing-workflow/validation-01/`.

sbatch returns75194 successfully. CLI implementation, local checks and prior
calibration evidence committedec8b9dc; submitted source is reproducible. README
check/updates and explicit persistence I/O/capacity boundaries recorded in prepared
handoff. No fulls17 execution, default or representation change.8/16GPU cross-node
options remain open but unimplemented; full RAM/storage/time admission still needed.
No scheduler/process queries, new-log reading, polling or waiting; yield until
user completion ping. CPU setup investigation remains deferred by agreement.
