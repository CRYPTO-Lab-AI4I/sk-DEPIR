# SCALING-128 — completed

UTC: 2026-11-19T16:37:37.794564+00:00

Reviewed 77868 after completion ping: staged-a passed all22 normal passes,
516proof endpoints and4rank CUDA memcheck (zero errors/leaks). Direct-a failed
at case2/sample5/rank1→rank0, both on HPC gnode13. Later direct-b/staged-b unrun.
Captured both262148-byte/65537-word polynomials; pre/post hashes stable locally,
but different across peers. Both snapshots match their logged SHA; 10239 changed
words all received as zero, first byte40960/word10240 (zero-based),77064→0.
No source/runtime cause established; failure does not require inter-node traffic.
No valid full comparison ratio. Source646/evidence29/cache136 hashes verified.
Shared binary/library unchanged; completed staged CPU/support/state/proof checks
independently rerun. Retained cache34.574MiB; raw failed status/logs unchanged.

[Report](../experiments/mpi-buffer-comparison/report.md),
[review and paired captures](../experiments/mpi-buffer-comparison/runs/compare-01/analysis.json),
[review commands/source hashes](../experiments/mpi-buffer-comparison/runs/compare-01/review-checks.json).
Scheduler3min6s, controller185.362s; staged largest partial case median1.217932s,
normal peakRSS410.676MiB/rank, GPUowned15.241MiB/rank (contexts additional).
Direct52 earlier rank measurements pass; no guard triggered. Resources unchanged.

Added paired-capture review and meaningful rejection test for missing endpoint.
6comparison/3tree/4proof/4distributed/5staged/3shared review tests pass (25 total),
Python syntax and whitespace checked. No native source changed; no new compile
or MPI/GPU run. Profile summary succeeded in28.060s near30s deadline; future shared
reporting allowance90s within unchanged720scontroller/900sscheduler, clocks and
resource allocation unchanged. Underlying previous profile-timeout cause unknown.

README checked/updated; main scaling reference/plan, experiment index, previous
stage/tree reports and preprocessing/CUDA docs current. Keep direct marked failed,
staging explicit (512KiB/rank plus CPUcopies; isolated runtime penalty unknown).
Next isolate host-buffer failure independently; use only validated explicit mode
for further bounded distributed integration. Fulls18 still unrun; no new job.
