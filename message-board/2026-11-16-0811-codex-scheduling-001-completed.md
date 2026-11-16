# SCHEDULING-001 — Job75342 placement restriction released

- UTC: 2026-11-16T08:11:37.766609+00:00
- Agent: Codex
- Status: Completed

User authorizes any accessible node. Updated existing pendingjob75342 with
ReqNodeList empty and ExcNodeListgnode03, avoiding its unresolved GPU fault.
One-time config verification confirms ReqNodeList=(null), excludedgnode03,
4H200/4CPU/8GiB,1node,gpuq,20minute limit unchanged. Original4GiB owned GPU/
6GiB RSS/2GiB cache/4GiB scratch/worker/controller limits and all job arguments,
IDs/results/log/cache paths remain. No cancel/relaunch or resource increase.

[Scheduler update](../experiments/gpu-preprocessing-workflow/scheduling-update.json)
and [experiment report](../experiments/gpu-preprocessing-workflow/report.md)
record original pin, diagnostic cause(3of4GPUs busy, private job visibility),
current override and reproduction without a pin. Future jobs may use accessible
eligible nodes by default, excluding documented unresolved faulty ones. Prior
submission records remain historical provenance, not overwritten.

README checked; no edit needed because this is operational placement only, with
no API/setup/supported parameter/performance/project-layout change. git diff
--check passes. Only placement/budget configuration verified once after mutation;
no progress/log reading, follow-up scheduler polling or GPU reset. Continue to
await user's completion ping for75342. Fulls17 still unvalidated/unadmitted.
