# Current s17 outcomes reconciled across reference tables

- Timestamp: 2026-11-17T08:24:02.229092+00:00
- Agent: Codex
- Work: DOC-095
- Status: Completed

Updated [scaling reference](../docs/encoding-status.md) with a prominent current
s17 summary, explicit per-mode statuses, spaced units and HPC job resources.
Preprocessing75832 and resident CPU76645 passed; resident GPU76646 was attempted
and failed its host-memory guard. Ten other allocation/storage modes have no
full s17 measurement and remain explicitly identified as unrun modes.
Rewrote the stale s17 target narrative around completed evidence and the proposed
GPU retry, and marked the chronological milestone sections as historical so old
pending statements cannot be mistaken for current status. Updated the checked
[README](../README.md) with a summary link and corrected older s17-limit/cache
statements. Evidence: [full encoding report](../experiments/s17-full-encoding/report.md)
and [summary](../experiments/s17-full-encoding/runs/first-01/summary.json).

Verification: checked all six detailed s17 rows and ten genuinely unrun modes;
asserted CPU pass, encoding6311.087 seconds and427.845 GiB sampled RSS against
retained summary.json; checked failed GPU status and current-summary job IDs,
removed stale blanket unmeasured claims, and ran `git diff --check`.
Documentation only; no executable changes, new jobs or runtime tests. Resident
CPU packing does not establish direct packed-output mode support. GPU full
encoding time remains unknown; no measured result is inferred for another mode.
