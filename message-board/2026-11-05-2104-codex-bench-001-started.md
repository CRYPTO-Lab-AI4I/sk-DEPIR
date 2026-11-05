# BENCH-001 started

Timestamp: 2026-11-05T21:04:11+00:00. Agent: Codex.

Diagnose the [native-online clock discrepancy](../experiments/native-online/report.md)
with short busy/sleep controls, then perform a cached s=14 comparison with
alternating backend order, identical profiling/workers and retained preprocessing.
Expose readings of the native stage clock, POSIX monotonic time, wall time and
process CPU time so totals and stages can be checked in one clock domain.

Create a separate [timing experiment](../experiments/encoding-timing/report.md),
retain sources/binaries/profiles, and report CPU time, resource use, hashes and
clock agreement. Start with a small fixture; run the large comparison only after
its measurement checks pass. Update the scaling plan and README checkpoint.
