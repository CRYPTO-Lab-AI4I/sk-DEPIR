# Bounded buffered preprocessing persistence

- Timestamp: 2026-11-18T17:07:30.504547+00:00
- Agent: Codex
- Work: SCALING-119
- Status: Started

User authorizes next separate persistence gate. Reuse shared CLI payload/checksum/
fsync/atomic publication and restart readback; explicit experiment-only buffering,
no default/format change. Compare64KiB/4MiB buffers with1/2/4 writers, four separate
24MiB synthetic public canonical-coefficient files (96MiB fixed total/case), matched
orders, complete checksum/readback and failure checks. Per-file single ownership;
no filesystem intermediate compute exchange. Retain profiles/final synthetic
payloads; prune only this new probe's verified cheap synthetic payloads between
cases with manifests/checksums retained, never expensive preprocessing/prior caches.
Use1HPCnode/2CPU/4GiB RAM,0GPU,2GiB RSS guard,128MiB retained/2GiB scratch,32MiB
per-file/15minute scheduler ceiling, within existing envelope. This bounded size
validates correctness/call aggregation, not sustained full-size storage bandwidth.
Read board, README, current plan, publication/recovery/codecs and prior I/O evidence.
Do not explicitly link the earlier file-access report in the plan. Long job handoff
without monitoring remains mandatory; no capacity estimate authorizes more budget.
