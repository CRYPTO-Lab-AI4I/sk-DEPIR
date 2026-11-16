# SCALING-083 — First full s17 preprocessing job75819 submitted

- UTC: 2026-11-16T10:16:59.906403+00:00
- Agent: Codex
- Status: Handoff
- Source: `0ac7da7b4a33bad32b6ff0bed4046b2aa02e2360`

Submittedfulls17,m2,t5 with explicitlocal layout under exactly the presented
approvedbudget:4H200112GiBowned/card,448GiB aggregate cap, exclusive128allocated
CPUcores/32activeworkers,864GiBrequest/800GiBRSS/720GiBGo soft,512GiBretained/
16GiBscratch,45hworker/46hcontroller/48hSlurm cutoffs. No node pin, excludegnode03.
[Submission](../experiments/gpu-s17-full-preprocessing/submission.json),
[report](../experiments/gpu-s17-full-preprocessing/report.md), results
`experiments/gpu-s17-full-preprocessing/runs/first-01/`, Slurm log
`experiments/gpu-s17-full-preprocessing/slurm-75819.log`, persistent checkpoints/
code/profiles `.cache/liftrs/gpu-s17-full-preprocessing/s17-m2-t5/`.

User chose local after numeric proposal and explicit local/peer trade-off review;
approval saved in budget.json before submission. CLI/profile/atomic checkpoints,
startup peer probe, count/source/host/device/storage/time/headroom guards and
bounded shutdown retained. Source committed0ac7da7. Approved check-mode and
controller tests/bash/diff pass; prior unapproved check evidence preserved.
README checked/status updated, scaling plan/reference updated for submitted state.
No protocol/math/default change or budget increase beyond the proposal.

sbatch returned75819 successfully. No scheduler queries, log/progress reading,
polling, waiting or result analysis after submission. Yield until user's completion
ping. Full peak/latency and verified cache still unmeasured.48h is a cutoff, not
runtime estimate. No database encoding.8/16GPU cross-node alternatives remain
future unimplemented options; CPU setup investigation stays deferred.
