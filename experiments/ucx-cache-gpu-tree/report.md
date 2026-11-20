# GPU merge tree with the UCX cache control

Status: **GPU-tree job 78249 submitted** (SCALING-137), results unreviewed.
[Receipt](submission.json), progress `slurm-78249.log`, results/profiles
`runs/tree-01/`. Implementation `d4621d2`;
[source freeze](readiness/snapshot.json): 662 files / 3,091,958 bytes; all ten
checked input hashes match. No MPI/GPU execution implied by readiness. [Runner](run.py),
[scheduler script](run.sbatch), [shared comparison](../mpi-buffer-comparison/run.py),
[reviewer](../mpi-buffer-comparison/analyze.py),
[shared UCX configuration](../ucx-registration-cache/run.py),
[GPU test](../../encoding/liftrs-shared/root_tree_gpu_test.go).
Implementation revision in submission.json; exact source/evidence/cache seals in results.json.

[Host control 78218](../ucx-registration-cache/report.md#review-of-78218) passed all
four arms after registration caching was disabled. That supports a cache interaction
under the host reproducer, not a permanent fix or GPU/full-size correctness result.
Restore the original GPU-tree workload with the same explicit UCX control before
adopting a configuration or projecting further performance.

Use the shared `configure(context)` preflight: UCX 1.18 version/config recorded,
`UCX_RCACHE_ENABLE=n` required, PML explicitly ucx, previous diagnostic verbosity
retained. Then run the unchanged same-build staged/direct/direct/staged crossover
in one allocation. Build the CUDA library and test executable once. Same genuine
s=8 and synthetic s=17 fixtures; largest support 262,144 exponents, 12 timed passes
per arm plus warm-up/artifact. Same CPU oracle, exact coverage/unique ownership,
canonical bounds, snapshots, paired proofs, four-rank small-case CUDA memcheck and
all CPU profiles. Original host/device guards and progress logging stay active.
No arithmetic, transforms or routing/tree changes.

Both host-only and GPU harnesses now use one MPI diagnostic-settings allowlist.
It records requested values only; native logs identify the selected PML. Review
checks the cache/backend/logging setting on every normal and memcheck rank. A
missing/default cache value fails review, including in partially completed arms.
Keep failures and stop immediately; do not construct ratios from incomplete arms.
All original direct/default failures and useful caches remain preserved.

Per-pass clocks and exclusions as in [original comparison](../mpi-buffer-comparison/report.md):
routing/hash/proof/GPU roots/tree/final CPU equality included; references/setup,
warm-up, artifact I/O and profile summaries excluded from per-pass medians but
retained separately in logs/profiles. Phase maxima are not additive. Matched
arm-median ratios apply only to this cache-disabled diagnostic; they do not isolate
cache-on versus cache-off cost or full-size latency. Logging affects timings.

Budget returns to the **previously admitted** GPU envelope, not a larger one:
**2 HPC nodes / 2 H200s / 4 logical ranks / 4 CPU cores / 8 GiB requested host RAM
total**. Two ranks share each card, 64 MiB owned GPU/rank (128 MiB/card, contexts
additional), 1 GiB process RSS guard/rank. Staging owns 512 KiB/rank normally,
8 KiB/rank under memcheck; direct owns none. Registration-cache disabling changes
unknown UCX internal-memory and registration overhead, not these explicit arrays.
No silent memory/time exchange or production default change. Cache cap 128 MiB,
build scratch 2 GiB, file cap 32 MiB, controller720 s / scheduler15 min; profile
summary90 s. Projected retained cache approximately72 MiB from original comparison.
Estimated **4–7 min after scheduling**, not measured. No four-GPU speedup claim.

```sh
python3 -B experiments/ucx-cache-gpu-tree/run.py --check
module load slurm/slurm/25.05
sbatch --parsable --output=experiments/ucx-cache-gpu-tree/slurm-%j.log \
  experiments/ucx-cache-gpu-tree/run.sbatch tree-01
# After user completion ping only:
python3 -B experiments/mpi-buffer-comparison/analyze.py tree-01 \
  --experiment experiments/ucx-cache-gpu-tree
```

Results/profiles/configuration/arm checkpoints: `runs/tree-01/`; progress:
`slurm-JOBID.log`; replay/state/failure arrays, binary and CUDA library outside Git:
`.cache/liftrs/ucx-cache-gpu-tree/tree-01/`, with seals in results.json. Regenerate
from the recorded revision and fresh name; toolchain Go1.27.3/GCC14.2/CUDA12.9.1/
OpenMPI4.1.8/UCX1.18/Slurm25.05. No cache/profile deletion or agent monitoring.

[Local checks](readiness/checks.json): native MPI/CUDA test binary compiles,
original fingerprint tests and diagnostic-settings snapshot unit pass;32 Python
reviewer tests pass, including rejection of an incorrect GPU-worker cache flag.
Python/scheduler syntax and whitespace pass. No local MPI/GPU execution; actual
GPU comparison and full s=18 capability remain unvalidated. Even a pass must not
be described as production security or an unrestricted upstream-runtime repair.
