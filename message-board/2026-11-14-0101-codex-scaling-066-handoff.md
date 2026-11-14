# Preprocessing validation submitted

- Timestamp: 2026-11-14T01:01:57.357126+00:00
- Agent: Codex
- Work: SCALING-066
- Status: handoff

Source checkpoint `ec12f1b89ee1db4c7eebdbcadf33a7b437948b80` implements the shared native-width preprocessing
extension, documented in [the guide](../docs/preprocessing.md). Submitted
CPU-only job74424,4cores/8GiB,6GiB sampled RSS guard,2GiB retained cache/4GiB
scratch,600s command/900s controller/1200s Slurm limits. No larger budget or full
s17 domain run. [Submission](../experiments/s17-preprocessing/submission.json)
and [report](../experiments/s17-preprocessing/report.md) identify exact source,
runner and result locations. Logs `experiments/s17-preprocessing/slurm-74424.log`,
results `experiments/s17-preprocessing/runs/native-01/`, reusable small artifacts
and matching binaries `.cache/liftrs/s17-preprocessing/native-01/`.

No monitoring, result inspection or dependent analysis: wait for completion ping.
Job74419 remains uninspected. Local native/static checks passed as recorded in
[static evidence](../experiments/s17-preprocessing/static-checks.json); Go/native,
race/sanitizer and CUDA build checks are pending. Full s17 preprocessing and
CUDA17/distributed execution remain unvalidated/unimplemented respectively.
README checked and updated in source checkpoint; no further README change needed
for this submission metadata. Next reconcile evidence after user ping, address
failures, then validate CUDA32 execution and plan the wider fastest-server gate.
