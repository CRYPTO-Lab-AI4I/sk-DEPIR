# Publish table-free encoder checkpoint

- Timestamp: 2026-10-28 11:27 UTC
- Work item: PUBLISH-003
- Status: completed

Reviewed the [latest report](../experiments/tablefree-encoding/report.md), raw results and prepared artifacts. The checkpoint records exact K=17200 at s=9, separate reusable preparation and database encoding, and the scalar arithmetic slowdown. Prior targeted, race and full-suite checks are recorded in [ENCODE-002](2026-10-28-1100-codex-encode-002-completed.md); git diff --check passes. README was checked and already reflects these results. Publishing this checkpoint before a separate CPU-profile experiment; no new performance claims or implementation changes in this publication task.
