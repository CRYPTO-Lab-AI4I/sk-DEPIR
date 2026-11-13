# Published completed evidence and s17 aims

- Timestamp: 2026-11-13T13:23:45.241052+00:00
- Agent: Codex
- Work: SCALING-048
- Status: completed

Added [s17 aims](../docs/encoding-status.md#aim-for-s17), linked from [README](../README.md): both CPU and GPU encoding have plausible capacity after the port. Recommend first CPU resident reuse (~427GiB host), then assess two-card streamed GPU (~256GiB aggregate device/~221GiB host) versus three-card resident reuse (~398GiB device/~221GiB host). GPU counts are payload floors; sharding, communication and headroom require implementation/measurement. One-card host-backed ~349GiB host is an optional slower alternative. Preprocessing peak/time unknown, three-array GPU model384GiB, so whole-workflow feasibility is not established. Wider34-bit arithmetic/coordinates/cache compatibility and independent checks are prerequisites. No new performance forecast or allocation approval.

Archive only19 original fill-01 cases already confirmed successful in the user's prior status check:7 CPU cases, all12 GPU cases, their Slurm logs and result/profile evidence. [Experiment report](../experiments/encoding-table/report.md#evidence-publication-checkpoint) marks publication versus pending full hash/cross-mode reconciliation. Active retry and partial cancelled-attempt files remain local for later reconciliation; no scheduler, process, active logs or retry results inspected. Large caches/outputs remain outside Git at reported paths.

README checked/updated. Verification: documentation links and git diff --check pass; documentation/evidence publication only, no runtime tests or jobs launched. Source reports/model assumptions remain linked from the reference. Table numeric updates await reconciliation rather than converting summary pass flags into new measurements.
