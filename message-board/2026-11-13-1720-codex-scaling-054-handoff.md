# Corrected matrix validation retry submitted

- Timestamp: 2026-11-13T17:20:22.321015+00:00
- Agent: Codex
- Work: SCALING-054
- Status: handoff

Reviewed74312 after completion ping: wide Go/C plan/matrix/parallel Taylor/example/lifetime race pass; shared regression suite fails only package-relative artifact opens/then example panic from wrong working directory. Later direct C matrix/sanitizer/legacy stages did not run. Reconciled420 sources to71db6a7,25 evidence and7 cache identities with failed status retained. [Analyzer/report](../experiments/s17-matrix/report.md#first-run-and-corrected-retry).

Fixed controller to run the retained shared binary in encoding/liftrs-shared within its copied repository, with explicit artifact-path preflights. No algorithm/test selection/resource/timing-boundary changes. Common hash reconciliation supports explicit forensic failed-run checks while still enforcing passed commands by default; earlier scalar/native analyzers rerun successfully. Failed logs/profiles/caches preserved; shared profile may be incomplete after panic. Scaling reference/plan record partial status. README checked: existing pending-validation text remains accurate, no edit required.

Static controller --check, Python/shell syntax, artifact existence, local links and authored whitespace checks pass. Exact fix/evidence checkpointdc57f45. Submitted job74330 /matrix-02,4CPU/noGPU/8GiB request/6GiB RSS,2GiB cache/4GiB scratch,10m worker/15m controller/20m Slurm. [Submission](../experiments/s17-matrix/retry-submission.json); outputs experiments/s17-matrix/runs/matrix-02/, log experiments/s17-matrix/slurm-74330.log, cache .cache/liftrs/s17-matrix/matrix-02/. Full validation remains open, no s17 full encoding claim. Other jobs untouched; no post-submit monitoring; await user completion ping.
