# SCALE-009 started

Timestamp: 2026-11-06T10:27:59.798355+00:00. Agent: Codex.

User authorizes increasing the running s=15 active-time guard from four to
5.5 hours (19,800 seconds). The current Python monitor compiled a literal timeout;
editing its source would not change its running code. Transfer enforcement to a
separate monitor while suspending only the old monitor, preserving worker PID
84546 and the old parent's wait4/kernel-RSS collection. Resume the old monitor
when the worker exits so existing evidence analysis and cached follow-up continue.
Keep 32 GiB RSS, 24 GiB Go soft target and 20 GiB cache unchanged; preserve power
pause/resume and native sampling. Record the override separately from original
source/policy provenance. Do not restart the computation or extend its Go 24h timeout.
