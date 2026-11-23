# SCALING-180 handoff: proof-recording job81738

UTC: 2026-11-23T00:28:53.526934+00:00; agent: Codex.

Committed implementation 08f1100 and readiness cd681de. Job **81738** submitted: [receipt](../experiments/distributed-proof-recording/submission.json), [report](../experiments/distributed-proof-recording/report.md), results `experiments/distributed-proof-recording/runs/recording-01/`, log `experiments/distributed-proof-recording/slurm-81738.log`. Same two HPC nodes/two H200s/four cores/8 GiB requested RAM; 64 MiB owned GPU and 1 GiB RSS per worker; 128 MiB cache/2 GiB scratch/32 MiB file; 720-second controller/15-minute Slurm cutoff. Estimated 4–7 minutes excluding queue. No polling or GPU result claim. Local completion registration remains outside Git.

851 frozen build inputs, 383 code hashes identical to compiled readiness tree, 14 frozen reviewer suites, tagged native tests and 108014 historical native bytes plus actual s8 inputs passed preflight. Raw evidence whitespace preserved; authored files checked. README and primary scaling reference updated with81729 passed and81738 pending; source/storage/plan documentation records scope and limitations.

After terminal notice, independently reconcile receipt/source/evidence/cache, inspect all six arms and native totals, compare artifact/digest/count/log-volume/runtime/memory results, then update the main reference. Continue production parameter/rank/lifecycle/non-oracle execution and bounded persistence/admission work. Full s18 remains unrun and unadmitted.
