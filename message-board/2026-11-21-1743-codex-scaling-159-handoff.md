# SCALING-159 — handoff

2026-11-21T17:43:11.061495+00:00; Codex.

Committed implementation and completed78469 evidence as `0c16446`. Submitted
**78508**, `roots-01`; [receipt](../experiments/distributed-root-merge/submission.json),
[report](../experiments/distributed-root-merge/report.md),
[scheduler log](../experiments/distributed-root-merge/slurm-78508.log).
Results: `experiments/distributed-root-merge/runs/roots-01/`.

5arms: arbitrary regression, small root CUDA sanitizer, normal root merges
N4096/16384/65536, each with N−4 roots so the final result spans both ranks.
Reuse local GPU factor construction and the shared resident product driver;
complete native/scalar references, saved support/operand/reference artifacts and
all existing communication/reuse checks. No new arithmetic or fulls18 claim.
Bounded replicated support/native reference and whole local-operand exchange
remain explicit; production partitioned ownership and Newton integration follow.

Same admitted2HPCnodes/2H200/4CPU/8GiB host,128MiB aggregate owned GPU/card,
1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file,720s/15min. Estimate3–6min
controller, excluding queue. Factor workspace released before final product.
README/main scaling reference/index/plan updated with submission. Existing local
CPU/race, alltag build, real fixture partition and reviewer checks passed; only
documentation/receipt changed after freeze. No polling, log inspection or
monitoring after submission. Await user completion ping before dependent work.
