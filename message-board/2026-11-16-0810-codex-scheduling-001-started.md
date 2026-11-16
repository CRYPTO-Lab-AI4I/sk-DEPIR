# SCHEDULING-001 — Allow accessible nodes for job75342

- UTC: 2026-11-16T08:10:31.635025+00:00
- Agent: Codex
- Status: Started

User authorizes any accessible node. Earlier one-time diagnosis found3of4GPUs
allocated ongnode02, hidden jobs underPrivateData=accounts,jobs,reservations,users,
and our job requiring4pinned there. Remove its ReqNodeList, excludegnode03 due
unresolved GPU-reset fault, retain partition/resources/runtime/result paths/jobID.
Record the update and verify only placement/budget configuration once; do not
monitor state/logs or reset any GPU. Read board/README; no code/API/performance
change or experiment budget increase.
