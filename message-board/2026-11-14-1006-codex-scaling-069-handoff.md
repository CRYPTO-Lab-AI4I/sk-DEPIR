# One-H200 GPU preprocessing validation submitted

- Timestamp: 2026-11-14T10:06:56.146274+00:00
- Agent: Codex
- Work: SCALING-069
- Status: handoff

GPU-first implementation/table checkpoint a1801fc, typed-GRES correction8828856.
Slurm rejected an untyped `gpu:1` request before allocation; corrected to H200
without changing count/caps. Submitted job74453, source `88288564db7977781c043be4e6d82f426a0242bc`, one H200,
512MiB aggregate/per-card owned payload,4CPU/8GiB host request,6GiB RSS guard,
4GiB Go soft limit,2GiB new retained/4GiB scratch,600/900/1200s deadlines.
[Submission](../experiments/gpu-s17-preprocessing/submission.json) and
[report](../experiments/gpu-s17-preprocessing/report.md) record commands/results.
Logs `experiments/gpu-s17-preprocessing/slurm-74453.log`, results/profiles
`experiments/gpu-s17-preprocessing/runs/gpu-01/`, caches/binaries
`.cache/liftrs/gpu-s17-preprocessing/gpu-01/`.

No monitoring or dependent analysis. Jobs74447/74448 remain uninspected. Two-GPU
request remains pending; do not treat elapsed time or default selection as
approval. If approved later, peer validation must keep the agreed bounded
512MiB aggregate (256MiB/card) unless another explicit budget is reviewed.
No full s17 preprocessing or multi-node implementation/admission follows.

Source GPU implementation includes all heavy preparation stages and same-node
peer shards, but only compilation/linking/host-simulated arithmetic is verified
so far. Tests include complete small Go comparisons, high-index GPU windows/
logical address probes, dense2^22-value tiles, old client regressions and memory
checks. GPU runtime, real peer behavior and performance remain unverified.
README checked/updated; scaling tables now label runtime units and distinct
memory measures instead of positional numbers. No new performance claim from
source availability. Await user completion ping to reconcile and fix failures.
