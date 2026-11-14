# Completed jobs reviewed; correct CPU fixture and retry validation

- Timestamp: 2026-11-14T08:33:20.854368+00:00
- Agent: Codex
- Work: SCALING-067
- Status: handoff (source checkpoint)

Job74424 passes all targeted preprocessing tests and Go regressions but fails
full native `TestCPUEncodingModules` at s4,t2. Also reconciled older job74419,
which failed at the same module fixture. Resolved479/460 source,16/13 evidence
and26/14 cache hashes respectively; profiles/binaries and failure evidence kept.
The fixture has K97,N256,L256, violating documented2L≤N compact/file/disk
admission. Changed all-mode s4 fixture to t3,K37,L128. Added separate dense-prefix
rejection and resident/wide-vs-Go checks; updated targeted selector. No production
algorithm, admission, default, protocol or memory-time trade-off change.

Disk s16 job74253 succeeds on gnode17 with16 native workers (32 allocated):
encoding36759.871s, load92.555s, verification382.655s, packing60.636s,
controller37504.672s. Sampled RSS19.469GiB; tracked cache59.271GiB;
RSS-plus-tracked-cache78.736GiB is not exact physical RAM.32GiB work files,
8GiB packed output. All918672421 systematic positions,16 curves and packed
readbacks pass; canonical hash matches prior s16.411 measured source and17
sealed evidence hashes reconcile. Independent large retained-file reread remains
pending; queued as bounded analysis, not another encoding. Logical interface
reads3.878TiB/writes0.808TiB, not physical disk counters. Slower mode remains
explicit. [Disk report](../experiments/encoding-table/report.md#completed-disk-s16)
and [scaling reference](../docs/encoding-status.md) now record measured boundaries.

Enhanced existing table analyzer with selected-case review, historical source
resolution and streaming30s byte-progress logs. The native preprocessing retry
can request this analysis through its existing guarded command controller; no
extra large output or budget increase. Documentation covers both initial failed
validations and corrected tests in CPU/preprocessing guides and reports. README
checked and updated; no full s17/CUDA17 or performance-equivalence claim.

Go formatting, diff, Python CLI/AST and shell syntax checks pass. Validation
exceeding30s must run in submitted jobs: retry native-02 (with disk cache analysis)
and modules-02, each unchanged4CPU/8GiB,600s command/900s controller/1200s Slurm
limits. Retain copied sources/profiles/binaries/caches and yield without monitoring.
