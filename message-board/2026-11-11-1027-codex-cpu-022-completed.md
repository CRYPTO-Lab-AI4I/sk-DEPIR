# CPU-022 — CPU stage profile verified

Timestamp: 2026-11-11T10:27:48.103367+00:00
Agent: Codex
Status: completed

User completion ping authorized inspection. Job72955/profile-01 passed:286 source hashes, all sealed evidence/cache hashes, six full canonical-word hashes and same-clock native accounting verified by experiments/cpu-encoding-profile/analyze.py. Median encoding248.675/140.746/91.164 s at8/16/32 workers. Taylor49.921/40.937/38.933 s; at32 workers42.7% of total. Rows/columns scale about3.9x/3.8x from8 to32; Taylor1.28x. Native buffer/pointwise/assembly2.3% combined; memory accesses also exist inside transforms. Peak sampled RSS5.594 GiB<12 GiB. Original/staged preprocessing, binary and profiles retained.

Updated [report](../experiments/cpu-encoding-profile/report.md), README measured scope and docs/encoding-scaling.md next CPU target. README check completed. Evidence self-check, shell syntax, documentation links and whitespace pass; no runtime change or new hardware job. Native Go profiles remain unsymbolized/Cgo and perf is restricted; stage scaling is not instruction attribution. Fixed-order two-sample medians are a profiling screen, not a robust s16 forecast.

Recommended next work: CPU Taylor within-block parallelism using the existing validated gather algebra, retaining current recurrence as reference and requiring full transform/canonical output checks. No optimization implemented or submitted automatically. Current results/experiment/docs are staged, not yet committed/pushed. Continue obeying>30-second job handoff rule.
