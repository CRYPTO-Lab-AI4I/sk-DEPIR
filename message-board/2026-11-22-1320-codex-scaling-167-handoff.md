# SCALING-167 — handoff

2026-11-22T13:20:32.912423+00:00; Codex.

Packaging fixf6e743e;real snapshot/input preflight recorded53271bf. Retry
**79213**,coordinator-02,submitted within unchanged2nodes/2H200/4CPU/8GiB RAM,
128MiBowned/card,1GiBRSS/rank,128MiBcache/2GiBscratch/32MiBfile,720s/15min.
[Receipt](../experiments/distributed-coordinator-stream/submission-retry.json),
[report](../experiments/distributed-coordinator-stream/report.md),
results `experiments/distributed-coordinator-stream/runs/coordinator-02/`,
log `experiments/distributed-coordinator-stream/slurm-79213.log`,
cache `.cache/liftrs/distributed-coordinator-stream/coordinator-02/` outsideGit.
Expected3–6min excluding queue. No monitoring/status/log/result inspection.

Preserved original79202 receipt/failure evidence;no GPU arms ran there. Real
committed snapshot includes manifest (786files/3,623,064bytes);all12 actual inputs
69,746bytes verified/copied read-only;reviewers pass from frozen source. Narrow
selector admission only;other result/readiness JSON remains excluded. No Go/CUDA
change or repeated native/sanitizer tests. README checked/updated;primary scaling
reference,preprocessing doc,plan and report show failure+actual retry ID.
After completion ping,run analyzer coordinator-02 and verify both arms before
continuing multi-rank/fullsize work. Fulls18 remains unrun.
