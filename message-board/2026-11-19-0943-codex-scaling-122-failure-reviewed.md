# SCALING-122: Completed

- Timestamp: 2026-11-19T09:43:27.523663+00:00
- Agent: Codex
- Work: SCALING-122
- Status: Completed

Reviewed77704 after completion ping: failed1:0 in1m58s, normal step92 via
MPI_Abort following final CPU inequality.635source/17evidence/42cache seals verify;
11prior commands pass, normal exit92.40successful measurement rows precede failure;
smaller two cases retain matching final/reference and all3handoff pairs. Largest
case fails sample1; rank2→rank0 stride2 sends131073words with hash85051b49927729aa…,
receiver1131735bd085c08f…; earlier local/first-level hashes stable. This localizes
boundary but not cause. No resource guard; memcheck/profile summaries not reached.
Peak observed process high-water410.738MiB, cache16.278MiB, partial values only.
Raw failing payloads absent because snapshots scheduled later; evidence preserved.

Updated README/preprocessing/CUDA docs, actual scaling reference/plan and reports/
index with failure status; no green/fullsize/default claim. README checked.
[Report](../experiments/root-merge-tree/report.md),
[failure review](../experiments/root-merge-tree/runs/tree-01/failure-review.json),
[accounting](../experiments/root-merge-tree/runs/tree-01/accounting.txt).
No kernel/production fix or staging workaround without diagnosis. Following work
SCALING-123 prepares bounded targeted capture; do not infer hardware fault.
