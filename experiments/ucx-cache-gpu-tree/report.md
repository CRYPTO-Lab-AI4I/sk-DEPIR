# GPU merge tree with the UCX cache control

Status: **78261 passed the full four-arm GPU crossover** (SCALING-140).
[Retry receipt](retry-submission.json), [raw results](runs/tree-02/results.json),
[review](runs/tree-02/analysis.json), [accounting](runs/tree-02/scheduler-accounting.txt),
[progress](slurm-78261.log). Implementation `e9f31f5`; verified
**664 source / 123 original evidence / 339 cache hashes**. All commands exited zero;
binary/library hashes, worker configuration and cross-arm final arrays match.
[Runner](run.py), [strict shared reviewer](../mpi-buffer-comparison/analyze.py).
Earlier 78249 raw failure and its explicit derived recovery remain below.

## Review of 78261

HPC gnode14/gnode15, 2 H200s, 4 logical ranks / CPU cores, 4 GiB requested host
RAM per node. Scheduler **5 min 5 s**, controller **303.568 s**. Every arm passed
22 normal passes / 516 endpoint checks, plus four-rank small-case CUDA memcheck
with zero errors and zero leaked bytes. Totals: **88 normal passes / 2,064 normal
endpoint checks**, another **96 memcheck endpoint checks**, and **16 clean CUDA
memcheck processes**. All 32 normal/memcheck worker configurations and native
PML selections confirm UCX with `UCX_RCACHE_ENABLE=n`. All rank streams parse
strictly, with no recovery or discarded malformed event. CPU profiles retained.

| Round | Mode | Largest synthetic-case median pipeline | Normal largest logged process peak RSS/rank |
|---|---|---:|---:|
| staged-a | MPI-owned staging | 1.247468 s | 412.055 MiB |
| direct-a | Direct Go payload | 1.153021 s | 411.613 MiB |
| direct-b | Direct Go payload | 1.246852 s | 410.836 MiB |
| staged-b | MPI-owned staging | 1.155454 s | 412.703 MiB |

Each row has 12 timed samples of the 262,144-exponent synthetic s=17 fixture,
excluding warm-up and artifact passes. Smaller s=8/17 fixtures also pass.
Peak owned GPU arrays **15.241 MiB/rank**, contexts additional; retained cache
**71.843 MiB**. Explicit staging owns an extra512 KiB/rank (8 KiB under memcheck);
UCX internal memory is additional and not isolated by this comparison.

The staged/direct ratio reverses: **1.0819** in pair A (staged8.19% slower),
**0.9267** in pair B (staged7.33% faster). This single-allocation crossover shows
**no consistent mode speed advantage**. It is not independent allocation replication,
a cache-on/off overhead estimate, four-GPU scaling, or full s=17/s=18 preprocessing.
The earlier default-cache direct failures stay failed; this establishes a bounded
validated configuration, not a universal runtime repair or production security.

**Next implementation:** continue the scalable preprocessing pipeline rather than
repeat this small transport sweep. Preserve explicit configuration and both
validated modes. `reduceRootProduct` still gives a receiver full host polynomials
and a single-executor multiply; senders currently leave the reduction. The next
module must keep worker ranks available for a collective large-product executor,
reusing existing 32/64-bit field/FFT kernels and verified RAM transport. Define deterministic
operation IDs, coefficient ownership and bounded transfer/workspace accounting;
compare small collective products and transforms against independent CPU results
before generator/Newton integration. Then remove coordinator-wide discovery/
support replication, and compose rank-owned checkpoints/final publication with
existing checksum/atomic helpers. Full-size memory/communication/runtime remain
unknown and require a separate admitted plan. No new job submitted in this review.

## Review of 78249

HPC gnode14/gnode15, 2 H200s, 4 logical ranks / CPU cores, 4 GiB requested host
RAM per node. Scheduler **4 min 16 s**, controller **255.343 s**. No native process,
CPU reference or resource guard failed in the three launched normal rounds.

| Round | Normal correctness / proofs | CUDA memcheck | Largest partial-case median | Normal largest logged process peak RSS/rank |
|---|---|---|---:|---:|
| staged-a | 22 passes, 516 endpoint checks passed | Four ranks: zero errors/leaks | 1.187015 s | 410.336 MiB |
| direct-a | 22 passes, 516 endpoint checks passed | Four ranks: zero errors/leaks | 1.191109 s | 410.844 MiB |
| direct-b | 22 passes, 516 checks passed after explicit log recovery | Not run | 1.180788 s | 410.605 MiB |
| staged-b | Not run | Not run | Unknown | Unknown |

All retained final arrays across the three normal rounds match the CPU oracle
and one another; UCX cache control and exact ownership/state checks pass. Each
largest case has 12 timed samples. Peak owned GPU arrays remain **15.241 MiB/rank**
(contexts additional). These are partial, profiled synthetic support fixtures,
not full s=17/s=18 preprocessing or a registration-cache overhead measurement.
The incomplete four-arm sequence cannot establish a stable mode speed difference.

### Explicit recovery of one interleaved event

In raw `direct-b-normal.log`, lines1499–1500 split a native progress event at
`"t` / `otal"`. A delayed OpenMPI help-aggregation warning was inserted between
the two fragments by merged output. [review_interleaved.py](review_interleaved.py)
accepts exactly this known warning and adjacent same-rank continuation, validates
the complete expected event, and writes a **separate**
[derived view](runs/tree-01/direct-b-normal.recovered.log). The warning is kept on
its own line; raw bytes/file remain untouched. Original/derived SHA256, complete
fragments, warning and recovered event are recorded in analysis.json. Source/
evidence/cache seals are verified before recovery; no generic skip or best-effort
parser is used. Re-review all direct-b CPU/shape/state/proof/byte-hash checks.

```sh
python3 -B experiments/ucx-cache-gpu-tree/review_interleaved.py
```

**Completed retry (SCALING-140): job 78261 passed**, reviewed above.
[Retry receipt](retry-submission.json), progress `slurm-78261.log`, results/profiles
`runs/tree-02/`. Source `e9f31f5`; [frozen inputs](readiness/retry-snapshot.json):
664 files / 3,099,895 bytes, all eight checked hashes match.
Repeat the full crossover with separate per-rank stdout
and stderr files. The explicit `PIR_RANK_LOG_FILES=1` wrapper option preserves the
launcher log, all worker streams and profiles; legacy host-runner mode remains.
A strict collector requires all eight files and parses every worker JSON event;
malformed worker JSON still fails. CUDA summaries are counted across all worker
streams. This prevents launcher diagnostics from sharing a worker event stream.
Native logging within a process remains unchanged; no arithmetic or clock change.
Direct file I/O may affect diagnostic timing; no cross-job speedup is inferred.
The same 128 MiB cache / 2 GiB scratch / 32 MiB per-file limits and GPU/host/deadline
envelope remain. Expected output volume stays small; each rank stream is bounded.
A fresh run preserves the incomplete original and produces a matched comparison
under one consistent logging method. No production default or full-size claim.

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

SCALING-139 local checks: four logging tests verify exact recovery/rejection,
strict complete stream collection, malformed-event rejection and actual wrapper
stdout/stderr isolation plus legacy behavior. Existing 32 reviewer tests pass
(36 total); shell/Python syntax and whitespace checked. No local MPI/GPU execution.
No native/Go source change in this logging fix; repeat uses existing tagged builds
and correctness checks inside the admitted job.

Retry review after user completion ping:

```sh
python3 -B experiments/mpi-buffer-comparison/analyze.py tree-02 \
  --experiment experiments/ucx-cache-gpu-tree
```

Retry arrays/library/binary remain at `.cache/liftrs/ucx-cache-gpu-tree/tree-02/`;
original tree-01 evidence remains intact. Scheduler15 min; estimated4–7 min after
scheduling. Logging file I/O overhead unknown; no monitoring after submission.
