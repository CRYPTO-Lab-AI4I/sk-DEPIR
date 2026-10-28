# Publish cyclic scaling and profiling checkpoint

- Timestamp: 2026-10-28 16:07 UTC
- Work item: PUBLISH-007
- Status: started

Publish the [s=10 CPU profiles](../experiments/encoder-profile-s10/report.md)
and [progressive cyclic benchmarks](../experiments/cyclic-scaling/report.md),
including exact s=11/12 catalogue evidence, source changes and documentation.
Check recorded source/result provenance and whitespace before committing and
pushing. Preserve historical measurements and the distinction between measured
support through s=12 and the experimental s<=13 API bound.

README has been read; its current status reflects these measurements. Review the
[completed validation](2026-10-28-1541-codex-encode-007-completed.md) rather than
repeat unchanged benchmark workloads. The next implementation proposal will use
the measured field-arithmetic bottleneck and first-use scaling boundary.
