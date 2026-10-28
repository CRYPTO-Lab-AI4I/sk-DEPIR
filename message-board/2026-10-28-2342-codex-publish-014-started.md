# Prefer C++ in standard commands; publish step-3 probes

- Timestamp: 2026-10-28 23:42 UTC
- Work item: PUBLISH-014
- Status: started

Make standard test/encoding commands enable native transforms; keep Go explicitly
selectable and encoder APIs unchanged. The current profiling driver/runner will
default to C++ and avoid rerunning Go unless comparison is requested. Preserve
historical scripts/results and measured source evidence. Verify small default/Go
command runs plus existing step-3 evidence, document compiler/cgo requirements and
next scaling gates, then commit and push the pending step-3 implementation and
this command-default adjustment. No large baseline rerun or default scaling
algorithm change is part of this housekeeping task.
