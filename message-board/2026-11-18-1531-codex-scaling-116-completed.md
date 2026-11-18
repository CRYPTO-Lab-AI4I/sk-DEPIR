# Staged GPU exchange77281 passes; current references updated

- Timestamp: 2026-11-18T15:31:18.807446+00:00
- Agent: Codex
- Work: SCALING-116
- Status: Completed

Reviewed on the user's completion ping; no new job submitted or resource budget
expanded. [Experiment report](../experiments/gpu-staged-transport/report.md),
[raw results](../experiments/gpu-staged-transport/runs/staged-02/results.json),
[verified analysis](../experiments/gpu-staged-transport/runs/staged-02/analysis.json),
[accounting](../experiments/gpu-staged-transport/runs/staged-02/accounting.txt).

77281 COMPLETED0:0,34s scheduler/33.504s controller, source6931f6c. HPCgnode04/05,
1H200NVL/2CPU/4GiB requested host RAM per node. Verified613 Git-source,16evidence,
9cache seals, no overrides; all14commands exit0. Both ranks pass full CPU producer,
GPU ownership/inverse, full CPU result and reverse exchange checks; both small
CUDA memchecks have0errors/0leaks. All24normal and4sanitized measurements reconcile.
At32MiB/rank, median pipeline11.8274ms, matched host2.2691ms, remote MPI2.568GiB/s
per node, effective pipeline1.321GiB/s (16MiB remote/node). Phase maxima are not
additive. Peak normal rankRSS421.016/411.207MiB; ownedGPU33.500MiB/card, pinned64MiB/rank.
These exclude no claimed aggregate job RSS; command-group RSS was not sampled.

Changed shared `experiments/s17-arithmetic/analyze.py` only to permit an explicit
list of collected command-summary fields; strict default retained. Staged reviewer
uses that option, reports missing commandRSS, and leaves producer results intact.
Regression verifies default schema rejection and that explicit omission still
rejects command failures/guards. No protocol/kernel/timing-methodology changes.
Native and Go profiles/matching binaries/metadata retained; cache8.283MiB outside
Git at `.cache/liftrs/gpu-staged-transport/staged-02/`. Raw logs remain bytewise.

Updated actual [scaling reference](../docs/encoding-status.md),
[plan](../docs/encoding-scaling.md), root/CUDA/experiment READMEs and report.
README check: updated bounded hardware/results/status; fulls17 CPU/GPU passed and
fulls18 unrun statuses remain accurate. Previous77241 failure evidence preserved.

Verification: five staged review tests pass; three gpu-code-encoding/shared-review
regressions pass; completion analyzer verifies all seals and checks; Python AST,
local documentation links and tracked diff whitespace checks pass. No GPU rerun
or full-suite run (analysis/schema/documentation change only).

Next: actual normalized-root subsets and complete polynomial transfer/merge
across ranks with independent CPU/canonical checks, followed separately by buffered
persistence. No distributed FFT/root/fulls18 runtime, GPU-direct/four-node bandwidth
or full-scale admission follows from this bounded field-multiply component.
