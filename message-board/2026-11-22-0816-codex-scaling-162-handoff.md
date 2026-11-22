# SCALING-162 — handoff

2026-11-22T08:16:35.956053+00:00; Codex.

Committed integrated pilot as `32d0d12`, submitted **78898**, `pilot-01`.
[Receipt](../experiments/distributed-preprocessing-pilot/submission.json),
[report](../experiments/distributed-preprocessing-pilot/report.md),
[scheduler log](../experiments/distributed-preprocessing-pilot/slurm-78898.log);
results: `experiments/distributed-preprocessing-pilot/runs/pilot-01/`.

One allocation,2integrated arms: normal genuine s4/s6 and smallest s4 CUDA
memcheck/leak. Existing GPU discovery/local roots, sharded root/Newton/identity/
spectrum executor, coordinated prefix commit/reload and complete canonical
artifact comparison. No separate kernel calibration series. Shared host Newton
schedule preserved, new sparse square uses existing base multiplication and
checked operand transport; no extra FFT or new CUDA arithmetic.

Same admitted2HPCnodes/2H200/4CPU/8GiB host,128MiB owned/card,1GiB RSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720s/15min; estimate3–6min controller excluding
queue. Existing preflight/guards/profiling/log retention reused. Native fixture,
CPU/race and focused boundary checks passed; only docs/receipt changed after
freeze. README/main reference/index/plan updated with submission.

No polling, live-log inspection or monitoring after submission. Await completion
ping. Coordinator discovery/publication,2rank single root level and in-process
reload remain explicit limits; fulls18 admission,partitioning,recovery/scaling
are not demonstrated. After pass prioritize those assembled-pipeline boundaries,
adding isolated tests only for concrete new correctness/performance blockers.
