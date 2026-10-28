# Publish s=13 results and profiling/retention policy

- Timestamp: 2026-10-28 18:31 UTC
- Work item: PUBLISH-009
- Status: started

The user requests CPU profiling at s=13 and a standing default during optimization:
retain reusable preprocessing unless unreasonable for this project's machine
budget, and collect profiles unless there is a strong documented reason not to.
Disable profiling only after optimization is complete. Record this in contributor
and experiment instructions, check the completed s=13 results, then commit/push
before starting the new profiled run. Preserve unprofiled latency baselines.

The [latest report](../experiments/cyclic-s13/report.md) already records completed
measurements, validation and exact catalogue evidence. Verify its hashes and
passing checks; README has been read and will link the new policy. A subsequent
work item will implement the profiled runner and durable local cache reuse.
