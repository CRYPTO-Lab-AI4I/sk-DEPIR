# SCALING-158 — handoff

2026-11-21T11:58:21.589616+00:00; Codex.

Committed source/evidence as `5525143`; submitted **78469**, `larger-01`.
[Receipt](../experiments/resident-product-larger/submission.json),
[report](../experiments/resident-product-larger/report.md),
[scheduler log](../experiments/resident-product-larger/slurm-78469.log);
results: `experiments/resident-product-larger/runs/larger-01/`.

22 arms compare host-stored GPU and local-access resident products, with exact
structured dense oracles at 16,384 / 65,536 / 262,144 values, small sanitizer
cases and arbitrary-input references. Same admitted two HPC nodes / two H200s /
four CPU cores / 8 GiB requested host RAM; 128 MiB owned GPU/card, 1 GiB RSS/rank,
128 MiB cache, 2 GiB scratch, 32 MiB/file; 720-second controller / 15-minute cutoff.
Estimated controller time 250–500 seconds, not measured. No resource increase.

README, main scaling reference, experiment index and plan updated with submission.
Prepared checks remain valid: handoff edits only documentation/receipt.
No scheduler polling, log inspection or monitoring after submission. Await user
completion ping. Larger timings remain structured-input results, not full s18
preprocessing evidence. After success, advance root/Newton integration and s18
field admission; no production default change.
