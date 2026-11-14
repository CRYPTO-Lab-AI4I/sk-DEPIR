# SCALING-072 — Completed-job review and peer retry prepared

- UTC: 2026-11-14T10:57:03.813985+00:00
- Agent: Codex
- Status: Handoff (prepared, submission follows)

Reviewed completed jobs74447/74448/74453/74454 after the user's ping. Native
preprocessing and CPU refactor correctness/race/sanitizers pass; GPU1 passes,
GPU2 arithmetic agrees but memcheck error704 fails the job. Exact source/evidence/
cache counts and command provenance are in the linked reports and run analyses.
Disk s16 now passes complete independent retained-cache rehash. All22 selected
table-case saved records report passed; bulk source/evidence/cache/cross-mode
reconciliation is prepared without another encoding run. Cancelled original
partial evidence and replacement logs/profiles remain archived, not relabeled.

Changes: shared process-scoped CUDA peer leases in
[preparation.cu](../encoding/liftrs-cuda/preparation.cu), lifetime regression tests,
GPU runner/analyzer selectors and pthread linking; module docs and reports.
[GPU report](../experiments/gpu-s17-preprocessing/report.md) preserves failed
results and records the32MiB same-allocation comparison: two GPUs17.8%/19.8%
slower than one. This is component timing, not a fulls17 prediction. Peer-local
bulk redistribution is recorded as a candidate with explicit extra transpose
traffic, unchanged proposed3N device payload and unknown runtime benefit.

[CPU report](../experiments/cpu-encoding-modules/report.md) flags setup regressions
27–44% (~2–3ms); encoding medians1–3% slower at s10. Functional validation is not
a no-regression claim. [Native report](../experiments/s17-preprocessing/report.md)
records corrected bounded preparation and the completed disk checksum review.
[CPU-only artifact controller](../experiments/encoding-table/reconcile.py) reuses
existing analysis, selects replacements before reading and logs guarded progress.
Its report documents4CPU/8GiB,6GiB RSS,2GiB new output and600/900/1200s limits.

Verification: shared CUDA O0 link exit0 under25s, Go formatting, Python AST/CLI,
bash syntax,22-case replacement metadata self-check and git diff --check pass;
[check record](../experiments/gpu-s17-preprocessing/peer-lease-checks.json).
GPU execution of the correction remains pending the two-card retry under the
existing approved caps; no larger experiment or fulls17 launch is admitted.
README checked and updated, along with CPU/preprocessing guides, scaling plan
and maintained reference. Profiles/caches and failed evidence retained.

Next: submit peer-02 on two H200s and independent CPU-only review-01. Record IDs,
then yield without polling, following logs or reading new results until ping.
