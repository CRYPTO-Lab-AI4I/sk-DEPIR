# SCALING-073 — Peer validation and all22-case reconciliation pass

- UTC: 2026-11-14T11:17:21.101428+00:00
- Agent: Codex
- Status: Completed

Reviewed completed jobs74466/74467 after user completion ping, without scheduler
monitoring. [GPU analysis](../experiments/gpu-s17-preprocessing/runs/peer-02/analysis.json)
passes498 source/35 evidence/36 retained-cache identities; source10a362b.
CUDA memcheck reports zero errors/leaked bytes; overlap/close/reopen peer tests,
small full preparation, s17 high-bit/windows, full tile hashes/inverses, legacy
CUDA and Go/native regressions/vet/shared-recurrence sanitizers pass.
Command: `python3 -B experiments/gpu-s17-preprocessing/analyze.py peer-02`.

[All-case review](../experiments/encoding-table/reviews/review-01/results.json)
passes22 cases at s10/14/15/16 with canonical cross-mode hash agreement, in380.660s.
Independent local verification of the saved review checks7 historical source
identities,23 evidence logs, both submission-manifest identities, all22 per-case
analysis hashes and four shared canonical hashes; see its analysis.json. Cache
rehashing occurred inside74467 and is not repeated locally. No encodings rerun.
Existing scripts/reports retain exact source revisions, commands, units, budgets,
profiling and caches. Original failed peer evidence remains unchanged.

Updated [scaling reference](../docs/encoding-status.md): publish all22 measured
cells with explicit time/host RSS/device memory/file-residency units and correct
measurement boundaries, preserve laptop/HPC labels and s17–20 projections.
New32MiB peer component sample is17.4%/19.7% slower on2GPU than1GPU, not a full-size
forecast. CPU setup regressions remain open. Reports and scaling plan record
current validation and next local-vector/bulk-peer performance work. README
checked and updated for successful retry and completed batch. No public API,
mode default, memory strategy or numeric experiment budget changed.

Verification: result analyzers/hash assertions and git diff --check pass; all22
published cells derive directly from reconciled saved phase/command JSON.
No new job submitted. Fulls17 preprocessing/cache/encoding remains unrun and
unadmitted; no full-size speedup claim. User receives the requested update table.
