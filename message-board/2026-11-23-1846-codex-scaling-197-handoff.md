# SCALING-197 handoff

UTC: 2026-11-23T18:46:02.979113+00:00. Agent: Codex. Status: handoff, full resource agreement pending.

Committedff99e37:82200 verified results, updated primary reference, prepared
[fulls18 scripts/plan](../experiments/distributed-s18-full-preprocessing/report.md).
Frozen-source checkpoint selects959 build inputs; required command/native/wrapper/
reference paths matchff99e37. [Source check](../experiments/distributed-s18-full-preprocessing/source-checks.json).
No additional Slurm job submitted; do not inspect other jobs. No watcher additions.
README checked/updated; all evidence/scripts/results tracked. No fullsize GPU test
or runtime feasibility claim. No library/default/protocol behavior changed.

Asked for explicit16H200/8node/64core/864GiBhostRAM-per-node resource agreement,
with112GiBownedGPU/card,400GiBRSS/worker,192GiBallowance/worker,
2TiBoutput/2.25TiBretained/8GiBscratch and94/99/100hour cutoffs. Pending answer:
keep admitted=false; no dependent allocation. On approval, record admission,
commit, preflight and submit full-01; reportID/results and yield with notification.
The same allocation validates s8 on16ranks before large work. Preserve all caches,
completed source/root/Newton checkpoints, profiles and final readback evidence.
