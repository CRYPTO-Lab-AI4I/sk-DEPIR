# SCALING-160 — prepared

2026-11-21T23:09:24.218907+00:00; Codex.

78508 independently passed5arms/12rank outputs, complete native/scalar references
and small CUDA memcheck/leak checks.743source/61original evidence/60cache
identities. Controller148.668s, cache18.609MiB, normal RSS390.387MiB/process,
small sanitizer416.391MiB/process on HPC gnode13/gnode14 (2H200/4CPU/8GiB RAM).
Final distributed product0.087347/0.227747/0.602521s atN4096/16384/65536.
Largest local root0.192915s, input preparation1.763340s includes native/scalar
reference1.544547s. Post-call product/factor owned GPU2.356/3.103MiB/card are
separate observations, not aggregate/physical peaks. [Evidence](../experiments/distributed-root-merge/report.md).

Prepared [operand routing](../experiments/distributed-root-routing/report.md):
shared CUDA-free interval/tile scheduler handles arbitrary owners/rank counts,
per-array memory admission, source preservation and canonical inputs/receives.
MPI adapter reuses existing control/coefficient/proof helpers. Root driver
retains explicitpeer/routed selection; old default stayspeer. No new arithmetic,
FFT or protocol/default change. Saved input shards and exact source-derived
paired route proofs independently establish every destination coefficient.

Projected atN65536: remove131068-byte peer polynomial/rank and halve input
network payload262136→131068bytes; GPU arrays unchanged. More tile handshakes
may cost runtime; not yet measured. Input clock now includes shard assembly,
so compare the new same-allocation peer baseline rather than historical peer-only
exchange. Diagnostic support/native references and host input shards remain.

5arms: peerN65536, small routed sanitizer, routedN4096/16384/65536. Same admitted
2HPCnodes/2H200/4CPU/8GiB RAM;128MiB aggregate owned/card,1GiB RSS/rank,
128MiBcache/2GiBscratch/32MiBfile,720s/15min; estimated3–6min, not measured.
No new allocation budget or fulls18 claim. Reuse validated preflight/guards/logs.

Changed shared router/tests, MPI adapter, root input/reviewer extension, new
experiment and linked docs/reference/README/index/board. CPU0.071s/race1.282s,
final router0.002s/race1.010s, alltag build, actual fixture/router tests and
3routing/3root/4larger/3product/5residentProduct/3vector reviewer tests passed;
Python/shell/gofmt/source whitespace checked. Race command explicitly enables
CGO and uses validated MPI compiler environment before shared suite setup;
focused regression guards inherited CGO=0. Native CUDA unchanged; no local
GPU/MPI execution. [Readiness](../experiments/distributed-root-routing/readiness/checks.json).

Main scaling reference and README checked/updated with78508 result and new scope.
Freeze/submit then hand off without monitoring. After success, integrate rank-owned
reduction and repeated/truncated products for Newton/generator identities, with
s18 arithmetic/support/discovery/checkpoint/publication still separate.
