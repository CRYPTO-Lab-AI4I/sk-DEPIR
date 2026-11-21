# SCALING-160 — handoff

2026-11-21T23:10:27.242774+00:00; Codex.

Committed78508 evidence/router implementation as `868a580`; submitted **78662**,
`routing-01`. [Receipt](../experiments/distributed-root-routing/submission.json),
[report](../experiments/distributed-root-routing/report.md),
[scheduler log](../experiments/distributed-root-routing/slurm-78662.log);
results: `experiments/distributed-root-routing/runs/routing-01/`.

5arms: whole-peer referenceN65536, small routed CUDA sanitizer, routed normal
N4096/16384/65536. Shared CUDA-free router, existing local GPU factors/product
engine and MPI control/codec/proofs; native/scalar references and every coefficient
remain checked. Explicit peer/routed selection; no default change. Source-derived
proof and saved-input review verifies ownership without full peer materialization.

Same admitted2HPCnodes/2H200/4CPU/8GiB requested RAM,128MiB aggregate owned/card,
1GiB RSS/rank,128MiB cache/2GiB scratch/32MiB file,720s/15min; estimate3–6min
controller excluding queue. Projected host/network savings, extra handshakes and
unknown runtime are explicit; same-allocation peer reference uses matching input
assembly timing. Diagnostic support/native references remain replicated; no
production root/Newton, s18 arithmetic admission or fulls18 result implied.

Local CPU/race, alltag build, actual fixtures/router and reviewer tests passed.
README/main reference/index/plan updated with result/submission. Only documentation
and receipt changed after freeze. No polling, live-log inspection or monitoring
after submission. Await completion ping. Then integrate rank-owned reduction and
repeated/truncated products for generator/Newton work under the recorded plan.
