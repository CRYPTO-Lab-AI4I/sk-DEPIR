# SCALING-174 — started

UTC: 2026-11-22T18:55:25.484707+00:00. Agent: Codex.

80962 independently passed811 source/74 evidence/473 cache seals,all narrow cold/resume and bounded field18 checks. Controller147.889s;arms cold6.097s,resume4.585s,wide6.090s. Pipeline RSS378.285/377.750MiB per worker;wide379.883MiB. Owned GPU0.064MiB narrow/3.125MiB wide per worker,context/driver extra;cache16.001MiB. [Analysis](../experiments/distributed-pipeline-wide/runs/wide-01/analysis.json). Fulls18 remains unrun.

Next decouple actual root execution from complete diagnostic support/native arrays. Share the maintained generic pipeline GPU executor setup and add rank-owned root merging through existing shard router/product/schedule. Keep existing diagnostic validation/default available;new explicit owned-root mode uses only discovered local roots,counts and intermediate shards. Preserve checks by comparing every leaf/stage/final output against checksum-verified retained native outputs from80962,plus existing canonical/prefix/identity/spectrum checks. Combine genuine cold/restart and bounded field18 tail in one allocation under unchanged budgets. No performance claim from changing validation boundaries;include reference loading/review in controller time. Read message board,README,current plan and executor/routing contracts;no other allocations polled.
