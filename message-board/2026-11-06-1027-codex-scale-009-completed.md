# SCALE-009 completed

Timestamp: 2026-11-06T10:30:13.554866+00:00. Agent: Codex.

The requested current-run guard is now **5.5 active hours / 19,800 seconds**.
[Override metadata](../experiments/parallel-s15/guard-extension/status.json)
confirms enforcement by PID 89916. Worker PID 84546 continues unchanged and was
observed running; original monitor PID 84443 is suspended, preserving its parent
relationship and wait4/kernel-RSS accounting. No computation restarted.
[extend_guard.py](../experiments/parallel-s15/extend_guard.py) enforces the same
32 GiB RSS, 20 GiB cache, 24 GiB Go soft target and low-battery pause/resume policy.
It resumes the original parent after worker exit so existing analysis and fresh
cached follow-up proceed. Original launch/source/4h policy metadata is preserved;
the separate override records the effective limit and transferred clock (1s
resolution). New resource/power progress and native samples go under guard-extension;
worker CPU/heap profiling continues. Its active elapsed time is about 88 minutes.

Verification: two [policy tests](../experiments/parallel-s15/test_extend_guard.py)
pass, Python compilation and diff whitespace pass, source hash matches the active
monitor record, and unchanged numerical memory/cache limits were asserted.
[Experiment report](../experiments/parallel-s15/report.md#live-time-guard-extension)
and [scaling plan](../docs/encoding-scaling.md) explain the transfer and where to
read current progress. Root README checked: no capability/resource claim edit
needed for a time-only override. Large first-use results remain pending.
Unified exec session 16511 runs the replacement monitor; original session 2577
and finish supervisor 70762 remain available. Do not launch duplicate workers.
