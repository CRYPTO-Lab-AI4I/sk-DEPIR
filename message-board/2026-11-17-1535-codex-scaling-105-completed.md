# Complete s18 route comparison recorded

- Timestamp: 2026-11-17T15:35:32.685038+00:00
- Agent: Codex
- Work: SCALING-105
- Status: Completed

[New reproducible analysis](../experiments/s18-preprocessing-routes/report.md)
includes model.py/test_model.py/model.json and a retained-profile top summary.
Inputs/source SHA256 identities bind existing exact count, bounded-component,
full-s17 measurements and shared arithmetic. No GPU/execution/default change,
new experiment, resource allocation or giant payload rehash. Source worktree
started at0aac5a9. Existing caches/profiles remain intact. The resumed CPU profile
matches its original producer cache seal; pprof summary is saved, not a new run.

Audit finds three oversized Newton products (N/2,N,N) in addition to final
identity, requiring3/7/13 prefix-aware block pairs versus14 full identity pairs.
Four-GPU hybrid also lacks tiled base-grid discovery and final-spectrum export
integration. Root products fit N/4/384GiB aggregate,96GiB/card. Last Newton square
and output stages are separated:633.446GiB principal product host arrays plus
64GiB-minus4bytes split temporary gives697.446GiB. Old128GiB prefix can die before
output allocation;825.446GiB is a GC-retention envelope, not required live memory.
Candidate support release after roots lowers stage-aware split Newton to576GiB;
current release only at export is too late. Actual RSS/GC/cache remains unmeasured.
Discovery770.893GiB is a separate conservative host model; tiled spectrum may
retain512GiB host arena plus P/G, so256.001GiB export payload is not its peak.

Sixteen-H200/four-node distributed model has96GiB product arrays/card and balanced
host discovery/Newton192.723/158.362GiB per node, before staging/runtime/cache.
Host/polynomial state and root work must be distributed; gathering all at one
coordinator invalidates that model. Transport, distributed transforms, checkpoints
and atomic canonical publication remain unimplemented. Hypothetical network
payload scenarios explicitly exclude unmeasured exchanges/latency; not HPC
bandwidth claims. Canonical950.170GiB plus conservative two-generations-per-phase
recovery528.001GiB gives1478.170GiB payload storage before headroom. No storage
increase approved. Root throughput scenarios36.0/40.9h use K or Klog²K work scaling
from measured s17root8h33min57s, not complete-route runtime forecasts or bounds.

Recommend distributed host/root/GPU route for speed; current small tiled/split
penalties must not become fullsize multipliers. Async question asks user to select
distributed16GPU vs four-GPU hybrid before material trade-off adoption; pending.
Next implementation gates: common support lifetimes, independent exact root
subtrees, concrete budgeted cross-node transport/redistribution and integration.
No jobs submitted; new multi-node/fullsize budget review remains separate.

Updated README, actual main encoding-status route table and s18 preprocessing-cell
link, active scaling plan step8/step5, and experiment index. Fulls17 CPU/GPU passed
cells and s19/20 rows unchanged. README check: public scaling guidance updated.
Verification:2 model tests pass; model output reproduces byte-for-byte; input/source
identities, Python AST, new local documentation links and git diff --check pass.
No CUDA/CPU execution regression rerun needed for analysis/docs-only changes.
