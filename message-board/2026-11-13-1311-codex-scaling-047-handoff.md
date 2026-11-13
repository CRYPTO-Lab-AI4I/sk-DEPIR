# Extended CPU retries submitted

- Timestamp: 2026-11-13T13:11:39.494434+00:00
- Agent: Codex
- Work: SCALING-047
- Status: handoff

Cancellation requested for74203/74202/74201; old files retained. Submitted source checkpointfb9bade with explicit extended limits: s16 packed74251 (4h Slurm/3h30m worker), s15 disk74252 (12h/11h30m), s16 disk74253 (48h/47h30m). Controller deadlines are15minutes below Slurm. User selected packed in parallel with a sequential disk queue; both first jobs depend only on termination of74201, and74253 depends on74252 via afterany. Memory/CPU/storage caps unchanged.

[Exact submission commands/policies](../experiments/encoding-table/runs/fill-02/submissions.json), [report](../experiments/encoding-table/report.md#extended-cpu-retries). Results/logs/profiles: experiments/encoding-table/runs/fill-02/<case>/; Slurm logs experiments/encoding-table/slurm-<job>.log; caches .cache/liftrs/encoding-table/fill-02/<case>/. Fresh paths preserve original partial evidence.

Changed run.py policy and dynamic Go timeout, submit.py subset/dependency scheduling and matching Slurm time, report, reference/plan, README and board. README checked/updated. Verification: run.py --check verifies unchanged non-time budgets and deadline ordering; submission --dry-run, Python AST, documentation links and git diff --check passed. Existing encoder implementation unchanged; runtime pending. Scheduling does not depend on analysis of another job. No post-submission monitoring; await user ping. Limits are ceilings, not forecasts or proof of completion.
