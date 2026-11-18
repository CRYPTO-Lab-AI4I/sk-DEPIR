# GPU two-node staged communication gate ready

- Timestamp: 2026-11-18T14:38:53.323806+00:00
- Agent: Codex
- Work: SCALING-113
- Status: Prepared

[Concrete runner/probe/resource plan](../experiments/gpu-staged-transport/report.md)
is ready. Existing CUDA multiplication produces s17public field patterns, pinned
D2H stages direct MPI_Alltoall ownership exchange, then H2D and GPU inverse/owner
check. Independent native polynomial-basis CPU producer oracle checks every word;
full CPU post-consumer and reverse ownership checks. Same-buffer host baseline,
per-phase maxima and complete forward clock, setup/verification outside clock.
No global FFT, normalized roots, GPU-direct capability, fullsize performance or
encoded-format claim. Optional server-only MPI/CUDA, ordinary CPU builds unchanged.

Slurm --mpi=list advertises pmix_v3/v4; OpenMPI4.1.8 ext3x present. Use srun
pmix_v3 with1task/node,1H200/task,explicit CPU/GPU binding and kill-on-bad-exit.
Rank checks MPI/Slurm identity and exactly1visible device before allocation;
avoid forwarding the launch node's GPU ordinal to another node. Rank wrapper/
binary/field metadata are in shared cache, not launch-private /tmp. Normal
1/8/32MiB/rank,warm+3samples; small32KiB two-rank CUDA memcheck,warm+1sample.
GPU producer/consumer/copies/MPI/synchronization are included in pipeline clock;
full CPU/reference/reverse checks outside. Shared reduced rank times deduplicated.
Word-progress logs/UTC/error phases,owned payload caps,RSS bounds and deadlines
are active; no false heartbeat progress or overlap claim. Rank-specific gprof
prefixes prevent PID collision; matching binary/native+Go profiles retained.

Tracked Git blob source freeze avoids ignored/racing worktree inputs and does
not need source overrides.3 review tests pass: paired timing/coverage/host/caps,
remote payload calculation,Git-only source snapshot. Profiled NVCC sm90 compile
and MPI/cudart link pass; local socket warning only on tool query. Go fixture
compile/skip0.003s and actual generation/schema0.009s pass,1572880bytes canonical
field metadata,hash in local-checks.json. Python AST,bash syntax,runner --check,
git diff --check pass. No GPU/MPI computation was executed locally; runtime,
PMIx/GPU binding,memcheck/profile artifacts remain unvalidated.

README checked/updated; actual scaling reference,active step3,CUDA docs and
experiment index state compiled/prepared only. Fulls17 CPU/GPU and fulls18unrun
cells unchanged. Next after this gate: actual independent normalized-root subsets
across ranks with existing helpers/explicit sparse validation and full joined
polynomial transfer/merge,then controlled buffered persistence separately.

Proposed budget:2nodes,1H200+2CPU+4GiB requested host per node(2GPUs/4CPU/8GiBtotal),
128MiB owned/card,64MiB pinned/rank,2GiBRSSguard/rank,128MiB retained artifacts,
2GiB launch scratch,280sMPI command/720scontroller/900sscheduler. Previous two-node
approval was0GPUs/10min; CUDA budget one node. Request explicit agreement for this
combined allocation/cutoff before sbatch. No submission, mode adoption or fullsize
budget increase. Frozen code/runner are committed for review before asking.
