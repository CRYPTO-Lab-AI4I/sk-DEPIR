# SCALING-170 — started

2026-11-22T13:56:33.692543+00:00; Codex.

Implement hierarchical rank-owned root merging after four-worker collectives79239.
Reuse existing product/routing algorithms:groups2 then4 etc run independent
subtrees concurrently;results remain coefficient shards between levels. Add
explicit MPI communicator ownership (no hidden global context),bounded shard
redistribution and shared stage scheduling. Validate small synthetic s18 roots
against independent native/scalar references,including uneven child domains.
Keep old defaults/reference paths;report extra setup/transfers and unknownruntime.
No hardware increase:existing4workers/2H200/2nodes/4CPU/8GiB,64MiBowned/worker.
Read board,README/contracts,root tree,routing,product driver,MPI lifecycle and
validated launcher;apply active ponytail minimal shared-implementation skill.
