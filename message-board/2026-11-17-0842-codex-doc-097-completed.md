# Active scaling tables corrected and history separated

- Timestamp: 2026-11-17T08:42:20.789693+00:00
- Agent: Codex
- Work: DOC-097
- Status: Completed

Updated the actual [scaling reference](../docs/encoding-status.md), rather than
adding another milestone to its end. Resident CPU/GPU reuse is now the first
result column; s17 CPU76645 is passed and GPU76646 attempted/failed. The ten
unmeasured s17 alternative-mode cells say projection only / no full measurement;
no resident result is substituted for another mode. Moved superseded milestone
and forecast material, including a stale component-section statement that full
s17 preparation was not run, to [encoding history](../docs/encoding-history.md).
Retained the batch timing/memory boundary alongside active tables. Corrected
stale next-work prose, hypothetical-width language and s17/s18 catalogue counts;
updated affected rounded s17 GPU payload/host projections with exactK. Measured
values remain unchanged. The active summary names the full encoding source and
review commit. [README](../README.md) and [plan link](../docs/encoding-scaling.md)
now distinguish the active reference from historical calibration.

Verification: asserted CPU/GPU pass/fail status and CPU6311.087 seconds/
427.845 GiB against retained full-run summary; checked all six detailed s17 rows,
all ten alternative projections, and primary resident columns. Compared every
s10/14/15/16 resident measurement cell before/after column reordering to ensure
its mode and data moved together. Checked removed stale claims/counts, retained
archive milestone headings, relative documentation file links, and
`git diff --check`. README checked and updated. Documentation only; no tests or
jobs launched, and no inspection of running job76704 logs/results.
