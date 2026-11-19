# Rank-owned tree ready for bounded four-rank validation

- Timestamp: 2026-11-19T08:48:21.646437+00:00
- Agent: Codex
- Work: SCALING-121
- Status: Prepared

[Report](../experiments/root-merge-tree/report.md), [runner](../experiments/root-merge-tree/run.py),
[readiness](../experiments/root-merge-tree/readiness/checks.json).
Shared bounded routing validates source once/preserves whole orbits and greedy
owner hashes. Old partition wrapper reuses it; packets add bounded memory to
that experimental wrapper, not a changed ordinary client/production default.
Non-coordinator ranks receive only local support; coordinator retains full source/
CPU oracle. Shared reduction gives unique owners, supports1..64/odd/empty owners,
checks canonical extents and propagates failures. MPIbinding2..64 typed bounded
send/receive preserves explicit two-rankPeer; no arithmetic duplicate or client
CUDA/MPI dependency. Per-rank GPU max points follow its owned subtree.

CPU-only tests incl1/2/3/4/5/7/8/16/64 logical owners, empty shards, negative/error
cases and pre-refactor owner hashes pass. Four historical CPU root digests match.
CUDA/MPI-tagged binary compiles (sandbox socket warnings, exit0);3new reviewer/
4prior distributed/5staged/3shared reviewer regressions pass. Python/shell syntax,
local doc links and edited-text whitespace pass. Real MPI/GPU/race execution awaits
scheduled validation. Local CPU hashes are not GPU performance evidence.

Normal warm-up/2timed/artifact pass and small4rank CUDA memcheck keep all shards,
local roots, sends/receives/owned merged states and final/reference. All handoff
bytes/state chain/coverage checked; profiles/progress/case/sample/stride/peer and
per-rank RSS retained. Validation file writes are separate artifact pass, never
compute/network inputs. No existing cache/profile removed.

README checked/updated with current reference/plan/preprocessing/CUDA docs and
experiment reports/index. Same approved hardware:2HPCnodes/2H200,4logical ranks,
2ranks/card,1CPU/rank,2CPU/4GiB requested RAM per node;64MiB owned/rank≤128MiB/card,
1GiB RSS/rank,128MiB cache/2GiB scratch/15min scheduler. Explicit block placement,
single:2 GPU binding/task-sharing verified against installed Slurm25.05 manual.
Four CUDA contexts add unknown contention/setup; not4GPU throughput or a speedup
baseline. Source/reference still centralized, GPU executors remain reusable;
large-product idle-worker coordination/distributedFFT/discovery/generator/publication
and fulls18 admission remain outstanding. No default or performance claim.

Next commit exact source, verify frozen inputs, submit fresh tree-01 and hand off
without monitoring; no larger resource request needed for this correctness gate.
