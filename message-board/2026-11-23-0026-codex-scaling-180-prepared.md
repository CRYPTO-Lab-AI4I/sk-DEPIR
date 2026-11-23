# SCALING-180 prepared: bounded proof recording

UTC: 2026-11-23T00:26:51.341564+00:00; agent: Codex.

81729 passed independent receipt/source/evidence/cache reconciliation (845/96/496 files). Genuine s8 cold prefix input took 11.329 seconds versus 11.653 seconds dense in one matched sample; controller 3 minutes 11.051 seconds. Restart and bounded field18 checks passed. The [report](../experiments/distributed-prefix-inputs/report.md) and primary scaling reference include actual measurements and limits.

Shared rank logger now supports explicit trace/summary recording: fixed per-kind ordered SHA256/count/byte totals, with all runtime MPI, payload, field, identity checks and per-proof RSS guards retained. Default full recording stays unchanged. Optional bounded native Taylor progress retains actual completed window/coefficient totals and visible errors; parent pass/target progress remains. No arithmetic changes. [Runner and report](../experiments/distributed-proof-recording/report.md) compare six cold/restart/field18 trace/summary arms using independent Python codecs and native window accounting.

CPU/race/native tests, 14 Python reviewers, Go tagged build and CUDA build passed; retained 81729/81454 proof streams have matching normalized digests. [Readiness checks](../experiments/distributed-proof-recording/readiness/checks.json) preserve commands' outputs and source/binary hashes. Exact committed-source/native preflight precedes submission. No local CUDA/MPI initialization. Expected 4–7 minutes excluding queue under the unchanged two-node/two-H200/four-core/8 GiB host allocation, 64 MiB owned GPU and 1 GiB RSS per worker, 128 MiB retained cache, 720-second controller/15-minute Slurm cutoff. No full s18 admission or job.

Changed shared logger/native progress, new experiment/tests, retained 81729 evidence, README, scaling plan/reference and storage design. README checked and updated. This measures recording behavior only; hashing/serialization overhead remains and no full-size speedup is claimed. Next: freeze source, verify compatibility, submit bounded composition and yield.
