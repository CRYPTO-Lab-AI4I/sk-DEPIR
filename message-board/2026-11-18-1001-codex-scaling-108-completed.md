# Two-node RAM transport77127 verified

- Timestamp: 2026-11-18T10:01:09.610116+00:00
- Agent: Codex
- Work: SCALING-108
- Status: Completed

77127 COMPLETED0:0 in5s, HPCgnode13/gnode14,0GPU,2CPU+4GiB requested host/node.
Producer e06a33d. [Report](../experiments/preprocessing-transport/report.md#completed-host-memory-exchange-scaling-108)
links raw data, accounting and new completion analyzer.3source/4evidence logs and
retained binary seals reconcile, no overrides; raw logs remain byte-identical.
Both distinct ranks complete all1/8/32MiB cases,1warm+3measured each; full34bit
pattern/order and inverse exchange checks pass. No intermediate array files.

Median slowest-rank forward latency:99.122us/0.662793ms/2.447160ms at1/8/32MiB
input per rank. Remote payload is half each input, yielding4.926/5.894/6.385GiB/s
per node.32MiB case aggregate bidirectional sent payload12.770GiB/s; not per-node
bandwidth. Shared reduced rank times are deduplicated to3samples/size, not6.
Peak OS process RSS297.734/287.867MiB per rank, includes MPI, within1GiB guards.
Scheduler launcher-step RSS is not rank memory and is not used as such. Retained
binary/cache and exact MPI configuration preserved; selected adapter/protocol
and GPU-direct support are not instrumented. No four-node/saturation/transform/
root/storage throughput or fulls18 runtime claimed from this host-memory test.

Added reusable analyzer with source Git/evidence/binary seals and exact log-record
comparison;3 review tests pass, including missing/same-node/RSS and factor-of-two
remote-byte/sample-count cases. Python AST and git diff --check pass. No producer
execution or arithmetic code changed. Raw output formatting is retained.
README checked/updated with bounded measured evidence. Updated actual main
scaling-reference route table/transport paragraph, active plan step8/gate, route
comparison and experiment index; fulls17 CPU/GPU and s18unrun cells remain accurate.
No capacity model changes. Next GPU staging/rank ownership and independent root
workers depend on shared-root backend validation; controlled buffered persistence
remains separate. Root77126 was not inspected: completion ping was interpreted
as latest submitted transport77127. No new job or budget increase.
