# Independent root-worker comparison77148 submitted

- Timestamp: 2026-11-18T11:05:02.731008+00:00
- Agent: Codex
- Work: SCALING-109
- Status: Handoff

Source/evidence committed00360cc. Submitted **77148**, existing1node/4H200/4CPU/
8GiBhost/20min scheduler budget. [Receipt](../experiments/parallel-root-workers/submission.json),
[report](../experiments/parallel-root-workers/report.md),
[results](../experiments/parallel-root-workers/runs/roots-01/),
[log](../experiments/parallel-root-workers/slurm-77148.log).
No polling, log inspection, waiting or result analysis after submission.
Root prerequisite77126 fully reviewed/published as in prepared message. New
worker timings/correctness remain pending; no fullsize speed estimate.

Recorded user's ordering in plan: parallel root computation+merge first, then
GPU node communication/staging, controlled buffered persistence separately,
then choose final GPU/node counts from complete time. Updated actual reference
and pending job status. README check/update and code docs in prepared message.
Local partition/admission tests, optional CPP+CUDA compile/skip, reviewer test,
AST, bash syntax, runner --check and whitespace pass. Full GPU worker correctness,
CPU/scalar reference comparisons, profiles and memcheck are delegated to77148.
No extra resource approval needed within existing same-node budget; no fullsize
or multi-node GPU allocation. Await user completion ping before dependent work.
