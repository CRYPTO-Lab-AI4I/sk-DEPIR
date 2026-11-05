# BENCH-001 completed

Timestamp: 2026-11-05T21:25:46+00:00. Agent: Codex.

Implemented [native clock readings](../encoding/liftrs-cpp/clocks.go) for the same
steady clock as stage counters, independent POSIX/wall clocks and aggregate process
CPU time. The [timing harness](../experiments/encoding-timing/main.go) brackets
whole calls and records getrusage CPU/RSS. No encoding arithmetic, preprocessing,
code parameters or existing result files changed.

The [report](../experiments/encoding-timing/report.md) links scripts, exact source/
binary/cache hashes, raw results, profiles, clock controls and power observations.
Short busy/sleep controls pass; six small Go-oracle comparisons pass. All six
s=14 words match the retained hash, stage sums fit same-clock totals, and the four
elapsed clocks agree within 0.376 ms. Historical clock disagreement was not
reproduced; its cause remains unestablished.

A midrun observation found battery at 18% during the first hybrid case, after the
first native case had ended. The user connected AC; periodic observations bracket
the last four cases on AC. Preserve the first unobserved-power/cold-workspace case
(168.947 s) and transition case (132.009 s) but exclude that pair from the AC
comparison. AC pairs: hybrid 93.720/95.375 s, full C++ 84.228/84.715 s. Median time
reduction is 10.66%, with similar aggregate CPU cost and whole-process peak RSS
4.5875 GiB. This is two profiled warm-workspace pairs on this host, not a universal
speedup. No code preprocessing was rerun. Binary, reusable caches and raw native
samples remain local; gzip samples and CPU profiles are in the report directory.

Verification: focused Go race/native encoding tests, native package/experiment
builds, vet, Go fallback encoding suite and clock ABI ASan/UBSan checks pass.
[Verification manifests](../experiments/encoding-timing/verification-analysis.json)
resolve all recorded sources. Evidence/hash analysis, Python compilation, local
links and `git diff --check` pass. The 250 ms calibration controls omit profiling
to avoid profiler setup dominating them; all encoding cases retain profiling.
The missing temporary Go 1.27.3 toolchain was restored from its official archive
with provenance documented in the report.

README check: updated [root README](../README.md),
[native API documentation](../encoding/liftrs-cpp/README.md),
[scaling plan](../docs/encoding-scaling.md) and experiment index with current
measurement boundaries/results. Normalized the board index table spacing.
The next scaling task remains the documented s=15 preparation-memory audit;
no s=15 run or support extension was started. No commit/push was requested here.
