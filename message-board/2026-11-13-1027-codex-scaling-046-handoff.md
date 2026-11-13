# Table-gap jobs submitted

- Timestamp: 2026-11-13T10:27:33.033345+00:00
- Agent: Codex
- Work: SCALING-046
- Status: handoff

Committed/pushed the reference as61b1857 and experiment/scripts as366f474 before submission. Submitted22 independent cases with sequential CPU and GPU resource chains: CPU74194–74203; GPU74204–74215. Exact case/job mapping, sbatch commands, policies and source checkpoint are in [the submission manifest](../experiments/encoding-table/runs/fill-01/submissions.json); result/phase/profile paths and limits in [the report](../experiments/encoding-table/report.md).

Results will appear in experiments/encoding-table/runs/fill-01/<case>/; Slurm logs at experiments/encoding-table/slurm-<job>.log; durable outputs/builds at .cache/liftrs/encoding-table/fill-01/<case>/. No scheduler/process/log polling, waits for job completion or result analysis performed. Only submission receipts were collected. Wait for the user's completion ping. Static checks passed, but full compilation/runtime checks remain pending inside jobs.

README was checked/updated before submission. Table remains unmeasured N for these cells, with clear R/HPC versus L/laptop evidence. The two-hour deadlines may prevent large backed cases completing; no s≥17 support/measurement claimed. No preprocessing reruns. The s16 disk case uses the user's conditionally authorized64GiB project storage cap; all other cases keep16GiB.
